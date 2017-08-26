// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dombuilder

import (
	"fmt"

	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/generator/es5/codedom"
	"github.com/serulian/compiler/graphs/scopegraph"
	"github.com/serulian/compiler/graphs/scopegraph/proto"
	"github.com/serulian/compiler/graphs/typegraph"
	"github.com/serulian/compiler/parser"
)

var _ = fmt.Printf

type exprModifier func(codedom.Expression) codedom.Expression

var operatorMap = map[compilergraph.TaggedValue]string{
	parser.NodeBinaryAddExpression:      "+",
	parser.NodeBinarySubtractExpression: "-",
	parser.NodeBinaryDivideExpression:   "/",
	parser.NodeBinaryMultiplyExpression: "*",
	parser.NodeBinaryModuloExpression:   "%",

	parser.NodeBitwiseAndExpression:        "&",
	parser.NodeBitwiseNotExpression:        "~",
	parser.NodeBitwiseOrExpression:         "|",
	parser.NodeBitwiseXorExpression:        "^",
	parser.NodeBitwiseShiftLeftExpression:  "<<",
	parser.NodeBitwiseShiftRightExpression: ">>",

	parser.NodeComparisonEqualsExpression:    "==",
	parser.NodeComparisonNotEqualsExpression: "!=",
	parser.NodeComparisonGTEExpression:       ">=",
	parser.NodeComparisonLTEExpression:       "<=",
	parser.NodeComparisonGTExpression:        ">",
	parser.NodeComparisonLTExpression:        "<",
}

// buildRootTypeExpression builds the CodeDOM for a root type expression.
func (db *domBuilder) buildRootTypeExpression(node compilergraph.GraphNode) codedom.Expression {
	childExprNode := node.GetNode(parser.NodeUnaryExpressionChildExpr)
	childScope, _ := db.scopegraph.GetScope(childExprNode)
	childType := childScope.ResolvedTypeRef(db.scopegraph.TypeGraph())

	childExpr := db.buildExpression(childExprNode)
	return codedom.NominalUnwrapping(childExpr, childType, node)
}

// buildFunctionCall builds the CodeDOM for a function call.
func (db *domBuilder) buildFunctionCall(node compilergraph.GraphNode) codedom.Expression {
	childExprNode := node.GetNode(parser.NodeFunctionCallExpressionChildExpr)
	childScope, _ := db.scopegraph.GetScope(childExprNode)

	// Check if the child expression has a static scope. If so, this is a type conversion between
	// a nominal type and a base type.
	if childScope.GetKind() == proto.ScopeKind_STATIC {
		wrappedExprNode := node.GetNode(parser.NodeFunctionCallArgument)
		wrappedExprScope, _ := db.scopegraph.GetScope(wrappedExprNode)
		wrappedExprType := wrappedExprScope.ResolvedTypeRef(db.scopegraph.TypeGraph())

		wrappedExpr := db.buildExpression(wrappedExprNode)

		targetTypeRef := childScope.StaticTypeRef(db.scopegraph.TypeGraph())

		// If the targetTypeRef is not nominal or structural, then we know we are unwrapping.
		if !targetTypeRef.IsNominalOrStruct() {
			return codedom.NominalUnwrapping(wrappedExpr, wrappedExprType, node)
		} else {
			return codedom.NominalRefWrapping(wrappedExpr, wrappedExprType, targetTypeRef, node)
		}
	}

	// Collect the expressions for the arguments.
	ait := node.StartQuery().
		Out(parser.NodeFunctionCallArgument).
		BuildNodeIterator()

	arguments := db.buildExpressions(ait, buildExprCheckNominalShortcutting)
	childExpr := db.buildExpression(childExprNode)

	// If the function call is to a member, then we return a MemberCall.
	namedRef, isNamed := db.scopegraph.GetReferencedName(childScope)
	if isNamed && !namedRef.IsLocal() {
		member, _ := namedRef.Member()

		if childExprNode.Kind() == parser.NodeNullableMemberAccessExpression {
			return codedom.NullableMemberCall(childExpr, member, arguments, node)
		}

		return codedom.MemberCall(childExpr, member, arguments, node)
	}

	// Otherwise, this is a normal function call.
	return codedom.InvokeFunction(childExpr, arguments, scopegraph.PromisingAccessFunctionCall, db.scopegraph, node)
}

// buildSliceExpression builds the CodeDOM for a slicer or indexer expression.
func (db *domBuilder) buildSliceExpression(node compilergraph.GraphNode) codedom.Expression {
	// Check if this is a slice vs an index.
	_, isIndexer := node.TryGetNode(parser.NodeSliceExpressionIndex)
	if isIndexer {
		return db.buildIndexerExpression(node)
	}

	return db.buildSlicerExpression(node)
}

// buildIndexerExpression builds the CodeDOM for an indexer call.
func (db *domBuilder) buildIndexerExpression(node compilergraph.GraphNode) codedom.Expression {
	indexExprNode := node.GetNode(parser.NodeSliceExpressionIndex)
	indexExpr := db.buildExpressionWithOption(indexExprNode, buildExprCheckNominalShortcutting)

	childExpr := db.getExpression(node, parser.NodeSliceExpressionChildExpr)

	scope, _ := db.scopegraph.GetScope(node)
	operator, _ := scope.CalledOperator(db.scopegraph.TypeGraph())

	opExpr := codedom.MemberReference(childExpr, operator, node)
	return codedom.MemberCall(opExpr, operator, []codedom.Expression{indexExpr}, node)
}

// buildSlicerExpression builds the CodeDOM for a slice call.
func (db *domBuilder) buildSlicerExpression(node compilergraph.GraphNode) codedom.Expression {
	childExpr := db.getExpression(node, parser.NodeSliceExpressionChildExpr)
	leftExpr := db.getExpressionOrDefault(node, parser.NodeSliceExpressionLeftIndex, codedom.LiteralValue("null", node))
	rightExpr := db.getExpressionOrDefault(node, parser.NodeSliceExpressionRightIndex, codedom.LiteralValue("null", node))

	scope, _ := db.scopegraph.GetScope(node)
	operator, _ := scope.CalledOperator(db.scopegraph.TypeGraph())

	opExpr := codedom.MemberReference(childExpr, operator, node)
	return codedom.MemberCall(opExpr, operator, []codedom.Expression{leftExpr, rightExpr}, node)
}

// buildAssertNotNullExpression builds the CodeDOM for an assert not null (expr!) operator.
func (db *domBuilder) buildAssertNotNullExpression(node compilergraph.GraphNode) codedom.Expression {
	childExpr := db.getExpression(node, parser.NodeUnaryExpressionChildExpr)
	return codedom.RuntimeFunctionCall(codedom.AssertNotNullFunction, []codedom.Expression{childExpr}, node)
}

// buildNullComparisonExpression builds the CodeDOM for a null comparison (??) operator.
func (db *domBuilder) buildNullComparisonExpression(node compilergraph.GraphNode) codedom.Expression {
	leftExpr := db.getExpression(node, parser.NodeBinaryExpressionLeftExpr)
	rightExpr := db.getExpression(node, parser.NodeBinaryExpressionRightExpr)
	return codedom.BinaryOperation(leftExpr, "??", rightExpr, node)
}

// buildInCollectionExpression builds the CodeDOM for an in collection operator.
func (db *domBuilder) buildInCollectionExpression(node compilergraph.GraphNode) codedom.Expression {
	valueExpr := db.getExpression(node, parser.NodeBinaryExpressionLeftExpr)
	childExpr := db.getExpression(node, parser.NodeBinaryExpressionRightExpr)

	scope, _ := db.scopegraph.GetScope(node)
	operator, _ := scope.CalledOperator(db.scopegraph.TypeGraph())

	return codedom.MemberCall(codedom.MemberReference(childExpr, operator, node), operator, []codedom.Expression{valueExpr}, node)
}

// buildIsComparisonExpression builds the CodeDOM for an is comparison operator.
func (db *domBuilder) buildIsComparisonExpression(node compilergraph.GraphNode) codedom.Expression {
	generatedLeftExpr := db.getExpression(node, parser.NodeBinaryExpressionLeftExpr)

	// Check for a `not` subexpression. If found, we invert the check.
	op := "=="
	rightExpr := node.GetNode(parser.NodeBinaryExpressionRightExpr)
	if rightExpr.Kind() == parser.NodeKeywordNotExpression {
		op = "!="
		rightExpr = rightExpr.GetNode(parser.NodeUnaryExpressionChildExpr)
	}

	generatedRightExpr := db.buildExpression(rightExpr)
	return codedom.NominalWrapping(
		codedom.BinaryOperation(generatedLeftExpr, op, generatedRightExpr, node),
		db.scopegraph.TypeGraph().BoolType(),
		node)
}

// buildBooleanBinaryExpression builds the CodeDOM for a boolean unary operator.
func (db *domBuilder) buildBooleanBinaryExpression(node compilergraph.GraphNode, op string) codedom.Expression {
	boolType := db.scopegraph.TypeGraph().BoolTypeReference()
	leftExpr := codedom.NominalUnwrapping(db.getExpression(node, parser.NodeBinaryExpressionLeftExpr), boolType, node)
	rightExpr := codedom.NominalUnwrapping(db.getExpression(node, parser.NodeBinaryExpressionRightExpr), boolType, node)
	return codedom.NominalWrapping(
		codedom.BinaryOperation(leftExpr, op, rightExpr, node),
		db.scopegraph.TypeGraph().BoolType(),
		node)
}

// buildBooleanUnaryExpression builds the CodeDOM for a native unary operator.
func (db *domBuilder) buildBooleanUnaryExpression(node compilergraph.GraphNode, op string) codedom.Expression {
	boolType := db.scopegraph.TypeGraph().BoolTypeReference()
	childExpr := codedom.NominalUnwrapping(db.getExpression(node, parser.NodeUnaryExpressionChildExpr), boolType, node)
	return codedom.NominalWrapping(
		codedom.UnaryOperation(op, childExpr, node),
		db.scopegraph.TypeGraph().BoolType(),
		node)
}

// buildNativeBinaryExpression builds the CodeDOM for a native unary operator.
func (db *domBuilder) buildNativeBinaryExpression(node compilergraph.GraphNode, op string) codedom.Expression {
	leftExpr := db.getExpression(node, parser.NodeBinaryExpressionLeftExpr)
	rightExpr := db.getExpression(node, parser.NodeBinaryExpressionRightExpr)
	return codedom.BinaryOperation(leftExpr, op, rightExpr, node)
}

// buildNativeUnaryExpression builds the CodeDOM for a native unary operator.
func (db *domBuilder) buildNativeUnaryExpression(node compilergraph.GraphNode, op string) codedom.Expression {
	childExpr := db.getExpression(node, parser.NodeUnaryExpressionChildExpr)
	return codedom.UnaryOperation(op, childExpr, node)
}

// buildUnaryOperatorExpression builds the CodeDOM for a unary operator.
func (db *domBuilder) buildUnaryOperatorExpression(node compilergraph.GraphNode, modifier exprModifier) codedom.Expression {
	scope, _ := db.scopegraph.GetScope(node)
	operator, _ := scope.CalledOperator(db.scopegraph.TypeGraph())

	if operator.IsNative() {
		return db.buildNativeUnaryExpression(node, operatorMap[node.Kind()])
	}

	childScope, _ := db.scopegraph.GetScope(node.GetNode(parser.NodeUnaryExpressionChildExpr))
	parentType := childScope.ResolvedTypeRef(db.scopegraph.TypeGraph())

	childExpr := db.getExpression(node, parser.NodeUnaryExpressionChildExpr)
	callExpr := codedom.MemberCall(codedom.StaticMemberReference(operator, parentType, node), operator, []codedom.Expression{childExpr}, node)
	if modifier != nil {
		return modifier(callExpr)
	}

	return callExpr
}

func (db *domBuilder) buildOptimizedBinaryOperatorExpression(node compilergraph.GraphNode, operator typegraph.TGMember, parentType typegraph.TypeReference, leftExpr codedom.Expression, rightExpr codedom.Expression) (codedom.Expression, bool) {
	// Verify this is a supported native operator.
	opString, hasOp := operatorMap[node.Kind()]
	if !hasOp {
		return nil, false
	}

	// Verify we have a native binary operator we can optimize.
	if !parentType.IsNominal() {
		return nil, false
	}

	isNumeric := false

	switch {
	case parentType.IsDirectReferenceTo(db.scopegraph.TypeGraph().IntType()):
		// Since division of integers requires flooring, turn this off for int div.
		// TODO: Have the optimizer add the proper Math.floor call.
		if opString == "/" {
			return nil, false
		}

		isNumeric = true

	case parentType.IsDirectReferenceTo(db.scopegraph.TypeGraph().BoolType()):
		fallthrough

	case parentType.IsDirectReferenceTo(db.scopegraph.TypeGraph().StringType()):
		fallthrough

	default:
		return nil, false
	}

	// Handle the various kinds of operators.
	resultType := db.scopegraph.TypeGraph().BoolTypeReference()

	switch node.Kind() {
	case parser.NodeComparisonEqualsExpression:
		fallthrough

	case parser.NodeComparisonNotEqualsExpression:
		// Always allowed.
		break

	case parser.NodeComparisonLTEExpression:
		fallthrough
	case parser.NodeComparisonLTExpression:
		fallthrough
	case parser.NodeComparisonGTEExpression:
		fallthrough
	case parser.NodeComparisonGTExpression:
		// Only allowed for number.
		if !isNumeric {
			return nil, false
		}

	default:
		returnType, _ := operator.ReturnType()
		resultType = returnType.TransformUnder(parentType)
	}

	unwrappedLeftExpr := codedom.NominalUnwrapping(leftExpr, parentType, node)
	unwrappedRightExpr := codedom.NominalUnwrapping(rightExpr, parentType, node)

	compareExpr := codedom.BinaryOperation(unwrappedLeftExpr, opString, unwrappedRightExpr, node)
	return codedom.NominalRefWrapping(compareExpr,
		resultType.NominalDataType(),
		resultType,
		node), true
}

// buildBinaryOperatorExpression builds the CodeDOM for a binary operator.
func (db *domBuilder) buildBinaryOperatorExpression(node compilergraph.GraphNode, modifier exprModifier) codedom.Expression {
	scope, _ := db.scopegraph.GetScope(node)
	operator, _ := scope.CalledOperator(db.scopegraph.TypeGraph())

	if operator.IsNative() {
		return db.buildNativeBinaryExpression(node, operatorMap[node.Kind()])
	}

	leftExpr := db.getExpression(node, parser.NodeBinaryExpressionLeftExpr)
	rightExpr := db.getExpression(node, parser.NodeBinaryExpressionRightExpr)

	leftScope, _ := db.scopegraph.GetScope(node.GetNode(parser.NodeBinaryExpressionLeftExpr))
	parentType := leftScope.ResolvedTypeRef(db.scopegraph.TypeGraph())

	optimized, wasOptimized := db.buildOptimizedBinaryOperatorExpression(node, operator, parentType, leftExpr, rightExpr)
	if wasOptimized {
		return optimized
	}

	callExpr := codedom.MemberCall(codedom.StaticMemberReference(operator, parentType, node), operator, []codedom.Expression{leftExpr, rightExpr}, node)
	if modifier != nil {
		return modifier(callExpr)
	}

	return callExpr
}
