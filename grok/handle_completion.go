// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grok

import (
	"strings"

	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/graphs/scopegraph/proto"
	"github.com/serulian/compiler/graphs/srg"
	"github.com/serulian/compiler/parser"
)

// GetCompletions returns the autocompletion information for the given activationString at the given location.
// If the activation string is empty, returns all context-sensitive defined names.
func (gh Handle) GetCompletions(activationString string, sal compilercommon.SourceAndLocation) (CompletionInformation, error) {
	builder := &completionBuilder{
		handle:           gh,
		activationString: activationString,
		sal:              sal,
		completions:      make([]Completion, 0),
	}

	// Find the node at the location.
	sourceGraph := gh.scopeResult.Graph.SourceGraph()
	node, found := sourceGraph.FindNodeForLocation(sal)
	if !found {
		return builder.build(), nil
	}

	switch {
	case strings.HasPrefix(activationString, "<"):
		// SML open or close expression.
		gh.addSmlCompletions(node, activationString, builder)

	case strings.HasSuffix(activationString, "."):
		// Autocomplete under an expression.
		gh.addAccessCompletions(node, activationString, builder)

	case strings.HasSuffix(activationString, "<"):
		// Autocomplete of types.
		gh.addContextCompletions(node, builder, func(scope srg.SRGContextScopeName) bool {
			return scope.NamedScope().ScopeKind() == srg.NamedScopeType
		})

	case strings.HasPrefix(activationString, "from ") || strings.HasPrefix(activationString, "import "):
		// Imports.
		importSnippet, ok := buildImportSnippet(activationString)
		if ok {
			importSnippet.populateCompletions(builder, sal.Source())
		}

	default:
		// Context autocomplete.
		gh.addContextCompletions(node, builder, func(scope srg.SRGContextScopeName) bool {
			return strings.HasPrefix(strings.ToLower(scope.LocalName()), strings.ToLower(activationString))
		})
	}

	return builder.build(), nil
}

// srgScopeFilter defines a filter function for filtering names found in scope.
type srgScopeFilter func(scope srg.SRGContextScopeName) bool

// addAccessCompletions adds completions based on an access expression underneath a node's context.
func (gh Handle) addAccessCompletions(node compilergraph.GraphNode, activationString string, builder *completionBuilder) {
	expressionString := strings.TrimSuffix(strings.TrimSuffix(activationString, "?."), ".")

	// Parse the activation string into an expression.
	source := compilercommon.InputSource(node.Get(parser.NodePredicateSource))
	startRune := node.GetValue(parser.NodePredicateStartRune).Int()

	parsed, ok := gh.scopeResult.Graph.SourceGraph().ParseExpression(expressionString, source, startRune)
	if !ok {
		return
	}

	parentImplementable, hasParentImplementable := gh.structureFinder.TryGetContainingImplemented(node)
	if !hasParentImplementable {
		return
	}

	// Scope the expression underneath the context of the node found.
	scopeInfo, isValid := gh.scopeResult.Graph.BuildTransientScope(parsed, parentImplementable)
	if !isValid {
		return
	}

	// Grab the static/non-static members of the resolved type and add them as completions.
	var lookupType = gh.scopeResult.Graph.TypeGraph().VoidTypeReference()
	var isStatic = false

	switch scopeInfo.Kind {
	case proto.ScopeKind_VALUE:
		lookupType = scopeInfo.ResolvedTypeRef(gh.scopeResult.Graph.TypeGraph())
		isStatic = false

	case proto.ScopeKind_STATIC:
		lookupType = scopeInfo.StaticTypeRef(gh.scopeResult.Graph.TypeGraph())
		isStatic = true

	case proto.ScopeKind_GENERIC:
		return

	default:
		panic("Unknown scope kind")
	}

	if !lookupType.IsNormal() {
		return
	}

	for _, member := range lookupType.ReferredType().Members() {
		if member.IsStatic() == isStatic && member.IsAccessibleTo(source) {
			builder.addMember(member, lookupType)
		}
	}
}

// addContextCompletions adds completions based on the node's context.
func (gh Handle) addContextCompletions(node compilergraph.GraphNode, builder *completionBuilder, filter srgScopeFilter) {
	for _, scopeName := range gh.structureFinder.ScopeInContext(node) {
		if filter(scopeName) {
			builder.addScopeOrImport(scopeName)
		}
	}
}

// addSmlCompletions adds completions for an SML expressions.
func (gh Handle) addSmlCompletions(node compilergraph.GraphNode, activationString string, builder *completionBuilder) {
	// First lookup the parent SML expression (if any). If one is found, offer it as the closing
	// completion if applicable.
	if activationString == "<" || strings.HasPrefix(activationString, "</") {
		smlExpression, isUnderExpression := gh.structureFinder.TryGetContainingNode(node, parser.NodeTypeSmlExpression)
		if isUnderExpression {
			// Grab the name of the expression and add it as a completion.
			smlPathExpression := smlExpression.GetNode(parser.NodeSmlExpressionTypeOrFunction)
			pathString, hasPathString := srg.IdentifierPathString(smlPathExpression)
			if hasPathString {
				if activationString == "<" {
					builder.addSnippet("Close SML Tag", "/"+pathString+">")
				} else {
					if strings.HasPrefix(pathString, activationString[2:]) {
						builder.addSnippet("Close SML Tag", pathString[len(activationString)-2:]+">")
					}
				}
			}
		}
	}

	// If only the close option was requested, nothing more to do.
	if strings.HasPrefix(activationString, "</") {
		return
	}

	// Otherwise, if requested, lookup all valid SML functions and types within the context.
	gh.addContextCompletions(node, builder, func(scope srg.SRGContextScopeName) bool {
		switch scope.NamedScope().Kind() {
		case parser.NodeTypeClass:
			// TODO: Check for a Declare constructor?
			return strings.HasPrefix(scope.LocalName(), activationString[1:])

		case parser.NodeTypeFunction:
			return strings.HasPrefix(scope.LocalName(), activationString[1:])
		}

		return false
	})
}