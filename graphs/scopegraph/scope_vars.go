// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scopegraph

import (
	"fmt"

	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/graphs/scopegraph/proto"
	"github.com/serulian/compiler/graphs/typegraph"
	"github.com/serulian/compiler/parser"
)

var _ = fmt.Printf

// scopeField scopes a field member in the SRG.
func (sb *scopeBuilder) scopeField(node compilergraph.GraphNode) proto.ScopeInfo {
	return sb.scopeDeclaredValue(node, "Field")
}

// scopeVariable scopes a variable module member in the SRG.
func (sb *scopeBuilder) scopeVariable(node compilergraph.GraphNode) proto.ScopeInfo {
	return sb.scopeDeclaredValue(node, "Variable")
}

// scopeVariableStatement scopes a variable statement in the SRG.
func (sb *scopeBuilder) scopeVariableStatement(node compilergraph.GraphNode) proto.ScopeInfo {
	return sb.scopeDeclaredValue(node, "Variable")
}

// getDeclaredVariableType returns the declared type of a variable statement, member or type field (if any).
func (sb *scopeBuilder) getDeclaredVariableType(node compilergraph.GraphNode) (typegraph.TypeReference, bool) {
	declaredTypeNode, hasDeclaredType := node.StartQuery().
		Out(parser.NodeVariableStatementDeclaredType, parser.NodePredicateTypeMemberDeclaredType).
		TryGetNode()

	if !hasDeclaredType {
		return sb.sg.tdg.AnyTypeReference(), false
	}

	// Load the declared type.
	typeref := sb.sg.srg.GetTypeRef(declaredTypeNode)
	declaredType, rerr := sb.sg.tdg.BuildTypeRef(typeref)
	if rerr != nil {
		panic(rerr)
		return sb.sg.tdg.AnyTypeReference(), false
	}

	return declaredType, true
}

// scopeDeclaredValue scopes a declared value (variable statement, variable member, type field).
func (sb *scopeBuilder) scopeDeclaredValue(node compilergraph.GraphNode, title string) proto.ScopeInfo {
	var exprScope *proto.ScopeInfo = nil

	exprNode, hasExpression := node.TryGetNode(parser.NodeVariableStatementExpression)
	if hasExpression {
		// Scope the expression.
		exprScope = sb.getScope(exprNode)
		if !exprScope.GetIsValid() {
			return newScope().Invalid().GetScope()
		}
	}

	// Load the declared type, if any.
	declaredType, hasDeclaredType := sb.getDeclaredVariableType(node)
	if !hasDeclaredType {
		if exprScope == nil {
			panic("Somehow ended up with no declared type and no expr scope")
		}

		return newScope().Valid().AssignableResolvedTypeOf(exprScope).GetScope()
	}

	// Compare against the type of the expression.
	if hasExpression {
		exprType := exprScope.ResolvedTypeRef(sb.sg.tdg)
		if serr := exprType.CheckSubTypeOf(declaredType); serr != nil {
			sb.decorateWithError(node, "%s '%s' has declared type '%v': %v", title, node.Get(parser.NodeVariableStatementName), declaredType, serr)
			return newScope().Invalid().GetScope()
		}
	} else {
		// Make sure if the type is non-nullable that there is an expression.
		if !declaredType.IsNullable() {
			sb.decorateWithError(node, "%s '%s' must have explicit initializer as its type '%v' is non-nullable", title, node.Get(parser.NodeVariableStatementName), declaredType)
			return newScope().Invalid().Assignable(declaredType).GetScope()
		}
	}

	return newScope().Valid().Assignable(declaredType).GetScope()
}