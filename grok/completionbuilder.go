// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grok

import (
	"fmt"

	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/graphs/srg"
	"github.com/serulian/compiler/graphs/typegraph"
)

// completionBuilder defines a helper for easier construction of completions,
// whether they be snippets, members, or scopes.
type completionBuilder struct {
	handle           Handle
	activationString string
	sourcePosition   compilercommon.SourcePosition
	completions      []Completion
}

func (cb *completionBuilder) addSnippet(title string, code string) *completionBuilder {
	return cb.addCompletion(Completion{
		Kind:          SnippetCompletion,
		Title:         title,
		Code:          code,
		SourceRanges:  []compilercommon.SourceRange{},
		TypeReference: cb.handle.scopeResult.Graph.TypeGraph().VoidTypeReference(),
	})
}

func (cb *completionBuilder) addTypeOrMember(typeOrMember typegraph.TGTypeOrMember) *completionBuilder {
	if typeOrMember.IsType() {
		return cb.addType(typeOrMember.(typegraph.TGTypeDecl))
	}

	return cb.addMember(typeOrMember.(typegraph.TGMember), cb.handle.scopeResult.Graph.TypeGraph().VoidTypeReference())
}

func (cb *completionBuilder) addMember(member typegraph.TGMember, lookupType typegraph.TypeReference) *completionBuilder {
	docString, _ := member.Documentation()

	return cb.addCompletion(Completion{
		Kind:          MemberCompletion,
		Title:         member.Name(),
		Code:          member.Name(),
		Documentation: trimDocumentation(docString),
		TypeReference: member.MemberType().TransformUnder(lookupType),
		SourceRanges:  sourceRangesOf(member),
		Member:        &member,
	})
}

func (cb *completionBuilder) addType(typedef typegraph.TGTypeDecl) *completionBuilder {
	docString, _ := typedef.Documentation()

	return cb.addCompletion(Completion{
		Kind:          TypeCompletion,
		Title:         typedef.Name(),
		Code:          typedef.Name(),
		Documentation: trimDocumentation(docString),
		TypeReference: cb.handle.scopeResult.Graph.TypeGraph().VoidTypeReference(),
		SourceRanges:  sourceRangesOf(typedef),
		Type:          &typedef,
	})
}

func (cb *completionBuilder) completionKindForNamedScope(namedScope srg.SRGNamedScope) (CompletionKind, *typegraph.TGMember, *typegraph.TGTypeDecl) {
	switch namedScope.ScopeKind() {
	case srg.NamedScopeType:
		srgType, ok := namedScope.GetType()
		if !ok {
			panic("Could not retrieve SRG type")
		}

		foundType, ok := cb.handle.scopeResult.Graph.TypeGraph().GetTypeOrModuleForSourceNode(srgType.GraphNode)
		if !ok {
			return ValueCompletion, nil, nil
		}

		casted := foundType.(typegraph.TGTypeDecl)
		return TypeCompletion, nil, &casted

	case srg.NamedScopeMember:
		srgMember, ok := namedScope.GetMember()
		if !ok {
			panic("Could not retrieve SRG member")
		}

		foundMember, ok := cb.handle.scopeResult.Graph.TypeGraph().GetTypeMemberForSourceNode(srgMember.GraphNode)
		if !ok {
			return ValueCompletion, nil, nil
		}

		return MemberCompletion, &foundMember, nil

	case srg.NamedScopeImport:
		return ImportCompletion, nil, nil

	case srg.NamedScopeValue:
		return ValueCompletion, nil, nil

	case srg.NamedScopeParameter:
		return ParameterCompletion, nil, nil

	case srg.NamedScopeVariable:
		return VariableCompletion, nil, nil

	default:
		return ValueCompletion, nil, nil
	}
}

func (cb *completionBuilder) addImport(packageName string, sourceKind string) *completionBuilder {
	if sourceKind == "" {
		return cb.addCompletion(Completion{
			Kind:          ImportCompletion,
			Title:         packageName,
			Code:          packageName,
			Documentation: "",
			SourceRanges:  []compilercommon.SourceRange{},
			TypeReference: cb.handle.scopeResult.Graph.TypeGraph().VoidTypeReference(),
		})
	}

	return cb.addCompletion(Completion{
		Kind:          ImportCompletion,
		Title:         fmt.Sprintf("%s (%s)", packageName, sourceKind),
		Code:          fmt.Sprintf("%s`%s`", sourceKind, packageName),
		Documentation: "",
		SourceRanges:  []compilercommon.SourceRange{},
		TypeReference: cb.handle.scopeResult.Graph.TypeGraph().VoidTypeReference(),
	})
}

func (cb *completionBuilder) addScopeOrImport(scopeOrImport srg.SRGContextScopeName) *completionBuilder {
	namedScope := scopeOrImport.NamedScope()

	// Lookup the documentation for the scope.
	var docString = ""
	documentation, hasDocumentation := namedScope.Documentation()
	if hasDocumentation {
		docString = documentation.String()
	}

	// Lookup the declared type for the scope.
	var typeref = cb.handle.scopeResult.Graph.TypeGraph().VoidTypeReference()

	returnType, hasReturnType := namedScope.DefinedReturnType()
	declaredType, hasDeclaredType := namedScope.DeclaredType()

	if hasReturnType {
		returnTypeRef, _ := cb.handle.scopeResult.Graph.ResolveSRGTypeRef(returnType)
		typeref = cb.handle.scopeResult.Graph.TypeGraph().FunctionTypeReference(returnTypeRef)
	} else if hasDeclaredType {
		typeref, _ = cb.handle.scopeResult.Graph.ResolveSRGTypeRef(declaredType)
	} else if namedScope.IsFunction() {
		typeref = cb.handle.scopeResult.Graph.TypeGraph().FunctionTypeReference(cb.handle.scopeResult.Graph.TypeGraph().VoidTypeReference())
	} else {
		// Check if there is scope for the node and, if so, grab the type from there. This handles
		// the dynamic case, such as inferred variable types.
		nodeScope, hasScope := cb.handle.scopeResult.Graph.GetScope(namedScope.Node())
		if hasScope {
			switch namedScope.ScopeKind() {
			case srg.NamedScopeVariable:
				fallthrough

			case srg.NamedScopeValue:
				typeref = nodeScope.AssignableTypeRef(cb.handle.scopeResult.Graph.TypeGraph())

			default:
				typeref = nodeScope.ResolvedTypeRef(cb.handle.scopeResult.Graph.TypeGraph())
			}
		}
	}

	name, hasName := namedScope.Name()
	if !hasName {
		return cb
	}

	localName, hasLocalName := scopeOrImport.LocalName()
	if !hasLocalName {
		return cb
	}

	completionKind, member, typedef := cb.completionKindForNamedScope(namedScope)
	return cb.addCompletion(Completion{
		Kind:          completionKind,
		Title:         name,
		Code:          localName,
		Documentation: highlightParameter(trimDocumentation(docString), name),
		TypeReference: typeref,
		SourceRanges:  sourceRangesOf(namedScope),
		Member:        member,
		Type:          typedef,
	})
}

func (cb *completionBuilder) addCompletion(completion Completion) *completionBuilder {
	cb.completions = append(cb.completions, completion)
	return cb
}

func (cb *completionBuilder) build() CompletionInformation {
	return CompletionInformation{cb.activationString, cb.completions}
}
