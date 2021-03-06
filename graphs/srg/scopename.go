// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package srg

//go:generate stringer -type=NamedScopeKind

import (
	"fmt"
	"sort"

	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/packageloader"
	"github.com/serulian/compiler/sourceshape"
)

// SRGScopeOrImport represents a named scope or an external package import.
type SRGScopeOrImport interface {
	Name() (string, bool)
	IsNamedScope() bool // Whether this is a named scope.
	AsNamedScope() SRGNamedScope
	AsPackageImport() SRGExternalPackageImport
}

// SRGExternalPackageImport represents a reference to an imported name from another package
// within the SRG.
type SRGExternalPackageImport struct {
	packageInfo packageloader.PackageInfo // The external package.
	name        string                    // The name of the imported member.
	srg         *SRG                      // The parent SRG.
}

// Package returns the package under which the name is being imported.
func (ns SRGExternalPackageImport) Package() packageloader.PackageInfo {
	return ns.packageInfo
}

// Name returns the name of the imported member.
func (ns SRGExternalPackageImport) Name() (string, bool) {
	return ns.name, true
}

// ImportedName returns the name being accessed under the package.
func (ns SRGExternalPackageImport) ImportedName() string {
	return ns.name
}

func (ns SRGExternalPackageImport) IsNamedScope() bool {
	return false
}

func (ns SRGExternalPackageImport) AsNamedScope() SRGNamedScope {
	panic("Not a named scope!")
}

func (ns SRGExternalPackageImport) AsPackageImport() SRGExternalPackageImport {
	return ns
}

// SRGNamedScope represents a reference to a named scope in the SRG (import, variable, etc).
type SRGNamedScope struct {
	compilergraph.GraphNode
	srg *SRG // The parent SRG.
}

// GetNamedScope returns SRGNamedScope for the given SRG node. Panics on failure to lookup.
func (g *SRG) GetNamedScope(nodeId compilergraph.GraphNodeId) SRGNamedScope {
	return SRGNamedScope{g.layer.GetNode(nodeId), g}
}

// NamedScopeKind defines the various kinds of named scope in the SRG.
type NamedScopeKind int

const (
	NamedScopeType      NamedScopeKind = iota // The named scope refers to a type.
	NamedScopeMember                          // The named scope refers to a module member.
	NamedScopeImport                          // The named scope refers to an import.
	NamedScopeParameter                       // The named scope refers to a parameter.
	NamedScopeValue                           // The named scope refers to a read-only value exported by a statement.
	NamedScopeVariable                        // The named scope refers to a variable statement.
)

// IsNamedScope returns true if this is a named scope (always returns true).
func (ns SRGNamedScope) IsNamedScope() bool {
	return true
}

// AsNamedScope returns the named scope as a named scope object.
func (ns SRGNamedScope) AsNamedScope() SRGNamedScope {
	return ns
}

// Node returns the underlying node.
func (ns SRGNamedScope) Node() compilergraph.GraphNode {
	return ns.GraphNode
}

// AsPackageImport returns the named scope as a package import (always panics).
func (ns SRGNamedScope) AsPackageImport() SRGExternalPackageImport {
	panic("Not an imported package!")
}

// Title returns a nice title for the given named scope.
func (ns SRGNamedScope) Title() string {
	switch ns.ScopeKind() {
	case NamedScopeType:
		return "type"

	case NamedScopeMember:
		return "member"

	case NamedScopeImport:
		return "import"

	case NamedScopeValue:
		return "value"

	case NamedScopeParameter:
		return "parameter"

	case NamedScopeVariable:
		return "variable"

	default:
		panic("Unknown kind of named scope")
	}
}

// IsAssignable returns whether the scoped node is assignable.
func (ns SRGNamedScope) IsAssignable() bool {
	switch ns.ScopeKind() {
	case NamedScopeType:
		fallthrough

	case NamedScopeMember:
		// Note: Only the type graph knows whether a member is assignable, so this always returns false.
		fallthrough

	case NamedScopeImport:
		fallthrough

	case NamedScopeValue:
		fallthrough

	case NamedScopeParameter:
		return false

	case NamedScopeVariable:
		return true

	default:
		panic("Unknown kind of named scope")
	}
}

// IsStatic returns whether the scoped node is static.
func (ns SRGNamedScope) IsStatic() bool {
	switch ns.ScopeKind() {
	case NamedScopeType:
		return true

	case NamedScopeMember:
		return ns.Kind() == sourceshape.NodeTypeConstructor

	case NamedScopeImport:
		return true

	case NamedScopeValue:
		fallthrough

	case NamedScopeParameter:
		fallthrough

	case NamedScopeVariable:
		return false

	default:
		panic("Unknown kind of named scope")
	}
}

// AccessIsUsage returns true if the named scope refers to a member or variable that
// is used immediately via the access. For example, a variable or property access will
// "use" that member, while a function or constructor is not used until invoked.
func (ns SRGNamedScope) AccessIsUsage() bool {
	switch ns.ScopeKind() {
	case NamedScopeType:
		return false

	case NamedScopeMember:
		return ns.Kind() == sourceshape.NodeTypeProperty

	case NamedScopeImport:
		return false

	case NamedScopeValue:
		fallthrough

	case NamedScopeParameter:
		fallthrough

	case NamedScopeVariable:
		return true

	default:
		panic("Unknown kind of named scope")
	}
}

// ScopeKind returns the kind of the scoped node.
func (ns SRGNamedScope) ScopeKind() NamedScopeKind {
	switch ns.Kind() {

	/* Types */
	case sourceshape.NodeTypeClass:
		return NamedScopeType

	case sourceshape.NodeTypeInterface:
		return NamedScopeType

	case sourceshape.NodeTypeNominal:
		return NamedScopeType

	case sourceshape.NodeTypeStruct:
		return NamedScopeType

	case sourceshape.NodeTypeAgent:
		return NamedScopeType

	/* Generic */
	case sourceshape.NodeTypeGeneric:
		return NamedScopeType

	/* Import */
	case sourceshape.NodeTypeImportPackage:
		return NamedScopeImport

	/* Members */
	case sourceshape.NodeTypeVariable:
		return NamedScopeMember

	case sourceshape.NodeTypeField:
		return NamedScopeMember

	case sourceshape.NodeTypeFunction:
		return NamedScopeMember

	case sourceshape.NodeTypeConstructor:
		return NamedScopeMember

	case sourceshape.NodeTypeProperty:
		return NamedScopeMember

	case sourceshape.NodeTypeOperator:
		return NamedScopeMember

	/* Parameter */
	case sourceshape.NodeTypeParameter:
		return NamedScopeParameter

	case sourceshape.NodeTypeLambdaParameter:
		return NamedScopeParameter

	/* Named Value */
	case sourceshape.NodeTypeNamedValue:
		return NamedScopeValue

	case sourceshape.NodeTypeAssignedValue:
		return NamedScopeValue

	/* Variable */
	case sourceshape.NodeTypeVariableStatement:
		return NamedScopeVariable

	default:
		panic(fmt.Sprintf("Unknown scoped name %v", ns.Kind()))
	}
}

// SourceRange returns the range of the named scope in source, if any.
func (ns SRGNamedScope) SourceRange() (compilercommon.SourceRange, bool) {
	return ns.srg.SourceRangeOf(ns.GraphNode)
}

// Documentation returns the documentation comment found on the scoped node, if any.
func (ns SRGNamedScope) Documentation() (SRGDocumentation, bool) {
	comment, found := ns.srg.getDocumentationCommentForNode(ns.GraphNode)
	if !found {
		return SRGDocumentation{}, false
	}

	switch ns.Kind() {
	case sourceshape.NodeTypeParameter:
		fallthrough

	case sourceshape.NodeTypeLambdaParameter:
		docInfo, hasDocInfo := comment.Documentation()
		if !hasDocInfo {
			return SRGDocumentation{}, false
		}

		name, hasName := ns.Name()
		if !hasName {
			return SRGDocumentation{}, false
		}

		return docInfo.ForParameter(name)

	default:
		return comment.Documentation()
	}
}

// Code returns a code-like summarization of the referenced scope, for human consumption.
func (ns SRGNamedScope) Code() (compilercommon.CodeSummary, bool) {
	name, hasName := ns.Name()
	if !hasName {
		return compilercommon.CodeSummary{}, false
	}

	switch ns.ScopeKind() {
	case NamedScopeType:
		srgType := SRGType{ns.GraphNode, ns.srg}
		return srgType.Code()

	case NamedScopeMember:
		srgMember := SRGMember{ns.GraphNode, ns.srg}
		return srgMember.Code()

	case NamedScopeImport:
		srgImport := SRGImport{ns.GraphNode, ns.srg}
		return srgImport.Code()

	case NamedScopeParameter:
		srgParameter := SRGParameter{ns.GraphNode, ns.srg}
		return srgParameter.Code()

	case NamedScopeValue:
		return compilercommon.CodeSummary{"", name, false}, true

	case NamedScopeVariable:
		declaredType, hasDeclaredType := ns.DeclaredType()
		if hasDeclaredType {
			return compilercommon.CodeSummary{"", fmt.Sprintf("var<%s> %s", declaredType.String(), name), true}, true
		}

		return compilercommon.CodeSummary{"", "var " + name, false}, true

	default:
		panic("Unknown kind of named scope")
	}
}

// Name returns the name of the scoped node.
func (ns SRGNamedScope) Name() (string, bool) {
	switch ns.Kind() {

	case sourceshape.NodeTypeClass:
		return ns.TryGet(sourceshape.NodeTypeDefinitionName)

	case sourceshape.NodeTypeInterface:
		return ns.TryGet(sourceshape.NodeTypeDefinitionName)

	case sourceshape.NodeTypeNominal:
		return ns.TryGet(sourceshape.NodeTypeDefinitionName)

	case sourceshape.NodeTypeStruct:
		return ns.TryGet(sourceshape.NodeTypeDefinitionName)

	case sourceshape.NodeTypeAgent:
		return ns.TryGet(sourceshape.NodeTypeDefinitionName)

	case sourceshape.NodeTypeImportPackage:
		return ns.TryGet(sourceshape.NodeImportPredicatePackageName)

	case sourceshape.NodeTypeGeneric:
		return ns.TryGet(sourceshape.NodeGenericPredicateName)

	case sourceshape.NodeTypeProperty:
		fallthrough

	case sourceshape.NodeTypeConstructor:
		fallthrough

	case sourceshape.NodeTypeVariable:
		fallthrough

	case sourceshape.NodeTypeField:
		fallthrough

	case sourceshape.NodeTypeFunction:
		return ns.TryGet(sourceshape.NodePredicateTypeMemberName)

	case sourceshape.NodeTypeOperator:
		return ns.TryGet(sourceshape.NodeOperatorName)

	case sourceshape.NodeTypeParameter:
		return ns.TryGet(sourceshape.NodeParameterName)

	case sourceshape.NodeTypeLambdaParameter:
		return ns.TryGet(sourceshape.NodeLambdaExpressionParameterName)

	case sourceshape.NodeTypeVariableStatement:
		return ns.TryGet(sourceshape.NodeVariableStatementName)

	case sourceshape.NodeTypeNamedValue:
		return ns.TryGet(sourceshape.NodeNamedValueName)

	case sourceshape.NodeTypeAssignedValue:
		return ns.TryGet(sourceshape.NodeNamedValueName)

	default:
		panic(fmt.Sprintf("Unknown scoped name %v", ns.Kind()))
	}
}

// GetType returns the type pointed to by this scope, if any.
func (ns SRGNamedScope) GetType() (SRGType, bool) {
	if ns.ScopeKind() == NamedScopeType {
		return SRGType{ns.GraphNode, ns.srg}, true
	}

	return SRGType{}, false
}

// GetMember returns the member pointed to by this scope, if any.
func (ns SRGNamedScope) GetMember() (SRGMember, bool) {
	switch ns.Kind() {
	case sourceshape.NodeTypeProperty:
		fallthrough

	case sourceshape.NodeTypeConstructor:
		fallthrough

	case sourceshape.NodeTypeVariable:
		fallthrough

	case sourceshape.NodeTypeField:
		fallthrough

	case sourceshape.NodeTypeFunction:
		return SRGMember{ns.GraphNode, ns.srg}, true
	}

	return SRGMember{}, false
}

// GetParameter returns the parameter pointed to by this scope, if any.
func (ns SRGNamedScope) GetParameter() (SRGParameter, bool) {
	switch ns.Kind() {
	case sourceshape.NodeTypeParameter:
		fallthrough

	case sourceshape.NodeTypeLambdaParameter:
		return SRGParameter{ns.GraphNode, ns.srg}, true
	}

	return SRGParameter{}, false
}

// IsFunction returns whether the named scope refers to a function.
func (ns SRGNamedScope) IsFunction() bool {
	return ns.Kind() == sourceshape.NodeTypeFunction
}

// DefinedReturnType returns the defined return type of the scoped node, if any.
func (ns SRGNamedScope) DefinedReturnType() (SRGTypeRef, bool) {
	if member, isMember := ns.GetMember(); isMember {
		return member.DefinedReturnType()
	}

	return SRGTypeRef{}, false
}

// DeclaredType returns the declared type of the scoped node, if any.
func (ns SRGNamedScope) DeclaredType() (SRGTypeRef, bool) {
	if member, isMember := ns.GetMember(); isMember {
		return member.DeclaredType()
	}

	if parameter, isParameter := ns.GetParameter(); isParameter {
		return parameter.DeclaredType()
	}

	// Note: variables and named values can have their types inferred at scope time, so the scope
	// graph will need to be queried for those types.
	return SRGTypeRef{}, false
}

// ResolveNameUnderScope attempts to resolve the given name under this scope. Only applies to imports.
func (ns SRGNamedScope) ResolveNameUnderScope(name string) (SRGScopeOrImport, bool) {
	if ns.Kind() != sourceshape.NodeTypeImportPackage {
		return SRGNamedScope{}, false
	}

	packageInfo, err := ns.srg.getPackageForImport(ns.GraphNode)
	if err != nil {
		return SRGNamedScope{}, false
	}

	if !packageInfo.IsSRGPackage() {
		return SRGExternalPackageImport{packageInfo.packageInfo, name, ns.srg}, true
	}

	moduleOrType, found := packageInfo.FindTypeOrMemberByName(name)
	if !found {
		return SRGNamedScope{}, false
	}

	return SRGNamedScope{moduleOrType.GraphNode, ns.srg}, true
}

// ScopeNameForNode returns an SRGNamedScope for the given SRG node. Note that the node
// must be a named node in the SRG or this can cause a panic.
func (g *SRG) ScopeNameForNode(srgNode compilergraph.GraphNode) SRGNamedScope {
	return SRGNamedScope{srgNode, g}
}

// FindReferencesInScope finds all identifier expressions that refer to the given name, under the given
// scope.
func (g *SRG) FindReferencesInScope(name string, node compilergraph.GraphNode) compilergraph.NodeIterator {
	// Note: This filter ensures that the name is accessible in the scope of the given node by checking that
	// the node referencing the name is contained by the given node.
	containingFilter := func(q compilergraph.GraphQuery) compilergraph.Query {
		startRune := node.GetValue(sourceshape.NodePredicateStartRune).Int()
		endRune := node.GetValue(sourceshape.NodePredicateEndRune).Int()

		return q.
			HasWhere(sourceshape.NodePredicateStartRune, compilergraph.WhereGT, startRune).
			HasWhere(sourceshape.NodePredicateEndRune, compilergraph.WhereLT, endRune)
	}

	return g.layer.StartQuery(name).
		In(sourceshape.NodeIdentifierExpressionName).
		IsKind(sourceshape.NodeTypeIdentifierExpression).
		Has(sourceshape.NodePredicateSource, node.Get(sourceshape.NodePredicateSource)).
		FilterBy(containingFilter).
		BuildNodeIterator()
}

// FindNameInScope finds the given name accessible from the scope under which the given node exists, if any.
func (g *SRG) FindNameInScope(name string, node compilergraph.GraphNode) (SRGScopeOrImport, bool) {
	// Attempt to resolve the name as pointing to a parameter, var statement, loop var or with var.
	srgNode, srgNodeFound := g.findAddedNameInScope(name, node)
	if srgNodeFound {
		return SRGNamedScope{srgNode, g}, true
	}

	// If still not found, try to resolve as a type or import.
	nodeSource := node.Get(sourceshape.NodePredicateSource)
	parentModule, parentModuleFound := g.FindModuleBySource(compilercommon.InputSource(nodeSource))
	if !parentModuleFound {
		panic(fmt.Sprintf("Missing module for source %v", nodeSource))
	}

	// Try to resolve as a local member.
	srgTypeOrMember, typeOrMemberFound := parentModule.FindTypeOrMemberByName(name, ModuleResolveAll)
	if typeOrMemberFound {
		return SRGNamedScope{srgTypeOrMember.GraphNode, g}, true
	}

	// Try to resolve as an imported member.
	localImportNode, localImportFound := parentModule.findImportWithLocalName(name)
	if localImportFound {
		// Retrieve the package for the imported member.
		packageInfo, err := g.getPackageForImport(localImportNode)
		if err != nil {
			return SRGNamedScope{}, false
		}

		resolutionName := localImportNode.Get(sourceshape.NodeImportPredicateSubsource)

		// If an SRG package, then continue with the resolution. Otherwise,
		// we return a named scope that says that the name needs to be furthered
		// resolved in the package by the type graph.
		if packageInfo.IsSRGPackage() {
			packageTypeOrMember, packagetypeOrMemberFound := packageInfo.FindTypeOrMemberByName(resolutionName)
			if packagetypeOrMemberFound {
				return SRGNamedScope{packageTypeOrMember.GraphNode, g}, true
			}
		}

		return SRGExternalPackageImport{packageInfo.packageInfo, resolutionName, g}, true
	}

	// Try to resolve as an imported package.
	importNode, importFound := parentModule.findImportByPackageName(name)
	if importFound {
		return SRGNamedScope{importNode, g}, true
	}

	return SRGNamedScope{}, false
}

// scopeResultNode is a sorted result of a named scope lookup.
type scopeResultNode struct {
	node       compilergraph.GraphNode
	startIndex int
}

type scopeResultNodes []scopeResultNode

func (slice scopeResultNodes) Len() int {
	return len(slice)
}

func (slice scopeResultNodes) Less(i, j int) bool {
	return slice[i].startIndex > slice[j].startIndex
}

func (slice scopeResultNodes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// findAddedNameInScope finds the {parameter, with, loop, var} node exposing the given name, if any.
func (g *SRG) findAddedNameInScope(name string, node compilergraph.GraphNode) (compilergraph.GraphNode, bool) {
	nodeSource := node.Get(sourceshape.NodePredicateSource)
	nodeStartIndex := node.GetValue(sourceshape.NodePredicateStartRune).Int()

	// Note: This filter ensures that the name is accessible in the scope of the given node by checking that
	// the node adding the name contains the given node.
	containingFilter := func(q compilergraph.GraphQuery) compilergraph.Query {
		startRune := node.GetValue(sourceshape.NodePredicateStartRune).Int()
		endRune := node.GetValue(sourceshape.NodePredicateEndRune).Int()

		return q.
			In(sourceshape.NodePredicateTypeMemberParameter,
				sourceshape.NodeLambdaExpressionInferredParameter,
				sourceshape.NodeLambdaExpressionParameter,
				sourceshape.NodePredicateTypeMemberGeneric,
				sourceshape.NodeStatementNamedValue,
				sourceshape.NodeAssignedDestination,
				sourceshape.NodeAssignedRejection,
				sourceshape.NodePredicateChild,
				sourceshape.NodeStatementBlockStatement).
			InIfKind(sourceshape.NodeStatementBlockStatement, sourceshape.NodeTypeResolveStatement).
			HasWhere(sourceshape.NodePredicateStartRune, compilergraph.WhereLTE, startRune).
			HasWhere(sourceshape.NodePredicateEndRune, compilergraph.WhereGTE, endRune)
	}

	nit := g.layer.StartQuery(name).
		In("named").
		Has(sourceshape.NodePredicateSource, nodeSource).
		IsKind(sourceshape.NodeTypeParameter, sourceshape.NodeTypeNamedValue, sourceshape.NodeTypeAssignedValue,
			sourceshape.NodeTypeVariableStatement, sourceshape.NodeTypeLambdaParameter, sourceshape.NodeTypeGeneric).
		FilterBy(containingFilter).
		BuildNodeIterator(sourceshape.NodePredicateStartRune, sourceshape.NodePredicateEndRune)

	// Sort the nodes found by location and choose the closest node.
	var results = make(scopeResultNodes, 0)
	for nit.Next() {
		node := nit.Node()
		startIndex := nit.GetPredicate(sourceshape.NodePredicateStartRune).Int()

		// If the node is a variable statement or assigned value, we have do to additional checks
		// (since they are not block scoped but rather statement scoped).
		if node.Kind() == sourceshape.NodeTypeVariableStatement || node.Kind() == sourceshape.NodeTypeAssignedValue {
			endIndex := nit.GetPredicate(sourceshape.NodePredicateEndRune).Int()
			if node.Kind() == sourceshape.NodeTypeAssignedValue {
				if parentNode, ok := node.TryGetIncomingNode(sourceshape.NodeAssignedDestination); ok {
					endIndex = parentNode.GetValue(sourceshape.NodePredicateEndRune).Int()
				} else if parentNode, ok := node.TryGetIncomingNode(sourceshape.NodeAssignedRejection); ok {
					endIndex = parentNode.GetValue(sourceshape.NodePredicateEndRune).Int()
				} else {
					panic("Missing assigned parent")
				}
			}

			// Check that the startIndex of the variable statement is <= the startIndex of the parent node
			if startIndex > nodeStartIndex {
				continue
			}

			// Ensure that the scope starts after the end index of the variable. Otherwise, the variable
			// name could be used in its initializer expression (which is expressly disallowed).
			if nodeStartIndex <= endIndex {
				continue
			}
		}

		results = append(results, scopeResultNode{node, startIndex})
	}

	if len(results) == 1 {
		// If there is a single result, return it.
		return results[0].node, true
	} else if len(results) > 1 {
		// Otherwise, sort the list by startIndex and choose the one closest to the scope node.
		sort.Sort(results)
		return results[0].node, true
	}

	return compilergraph.GraphNode{}, false
}
