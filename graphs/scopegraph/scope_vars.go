// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scopegraph

import (
	"fmt"

	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/graphs/scopegraph/proto"
	"github.com/serulian/compiler/graphs/typegraph"
	"github.com/serulian/compiler/sourceshape"
)

var _ = fmt.Printf

type requiresInitializerOption int

const (
	requiresInitializer requiresInitializerOption = iota
	noRequiredInitializer
)

// scopeField scopes a field member in the SRG.
func (sb *scopeBuilder) scopeField(node compilergraph.GraphNode, context scopeContext) proto.ScopeInfo {
	return sb.scopeDeclaredValue(node, "Field", noRequiredInitializer, compilergraph.Predicate(sourceshape.NodePredicateTypeFieldDefaultValue), context)
}

// scopeVariable scopes a variable module member in the SRG.
func (sb *scopeBuilder) scopeVariable(node compilergraph.GraphNode, context scopeContext) proto.ScopeInfo {
	return sb.scopeDeclaredValue(node, "Variable", requiresInitializer, compilergraph.Predicate(sourceshape.NodePredicateTypeFieldDefaultValue), context)
}

// scopeVariableStatement scopes a variable statement in the SRG.
func (sb *scopeBuilder) scopeVariableStatement(node compilergraph.GraphNode, context scopeContext) proto.ScopeInfo {
	return sb.scopeDeclaredValue(node, "Variable", requiresInitializer, compilergraph.Predicate(sourceshape.NodeVariableStatementExpression), context)
}

// getDeclaredVariableType returns the declared type of a variable statement, member or type field (if any).
func (sb *scopeBuilder) getDeclaredVariableType(node compilergraph.GraphNode) (typegraph.TypeReference, bool) {
	declaredTypeNode, hasDeclaredType := node.StartQuery().
		Out(sourceshape.NodeVariableStatementDeclaredType, sourceshape.NodePredicateTypeMemberDeclaredType).
		TryGetNode()

	if !hasDeclaredType {
		return sb.sg.tdg.AnyTypeReference(), false
	}

	// Load the declared type.
	declaredType, rerr := sb.sg.ResolveSRGTypeRef(sb.sg.srg.GetTypeRef(declaredTypeNode))
	if rerr != nil {
		return sb.sg.tdg.AnyTypeReference(), false
	}

	return declaredType, true
}

// scopeDeclaredValue scopes a declared value (variable statement, variable member, type field).
func (sb *scopeBuilder) scopeDeclaredValue(node compilergraph.GraphNode, title string, option requiresInitializerOption, exprPredicate compilergraph.Predicate, context scopeContext) proto.ScopeInfo {
	var exprScope *proto.ScopeInfo
	exprNode, hasExpression := node.TryGetNode(exprPredicate)
	if hasExpression {
		// Scope the expression.
		exprScope = sb.getScope(exprNode, context)
		if !exprScope.GetIsValid() {
			return newScope().Invalid().GetScope()
		}
	}

	// Load the declared type, if any.
	declaredType, hasDeclaredType := sb.getDeclaredVariableType(node)
	if !hasDeclaredType {
		if exprScope == nil {
			return newScope().Invalid().GetScope()
		}

		return newScope().Valid().AssignableResolvedTypeOf(exprScope).GetScope()
	}

	varName, hasVarName := node.TryGet(sourceshape.NodeVariableStatementName)
	if !hasVarName {
		return newScope().Invalid().GetScope()
	}

	// Compare against the type of the expression.
	if hasExpression {
		exprType := exprScope.ResolvedTypeRef(sb.sg.tdg)
		if serr := exprType.CheckSubTypeOf(declaredType); serr != nil {
			sb.decorateWithError(node, "%s '%s' has declared type '%v': %v", title, varName, declaredType, serr)
			return newScope().Invalid().GetScope()
		}
	} else if option == requiresInitializer {
		// Make sure if the type is non-nullable that there is an expression.
		if !declaredType.IsNullable() {
			sb.decorateWithError(node, "%s '%s' must have explicit initializer as its type '%v' is non-nullable", title, varName, declaredType)
			return newScope().Invalid().Assignable(declaredType).GetScope()
		}
	}

	return newScope().Valid().Assignable(declaredType).GetScope()
}
