// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package statemachine

import (
	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/parser"
)

// generateFunctionCall generates the state machine for a function call.
func (sm *stateMachine) generateFunctionCall(node compilergraph.GraphNode, parentState *state) {
	// Generate the expression for the function itself that will be called.
	childExpr := node.GetNode(parser.NodeFunctionCallExpressionChildExpr)
	callExprInfo := sm.generate(childExpr, parentState)

	// For each of the arguments (if any) generate the expressions.
	ait := node.StartQuery().
		Out(parser.NodeFunctionCallArgument).
		BuildNodeIterator()

	argumentInfo := sm.generateIterator(ait, callExprInfo.endState)

	// Create a new state to which we'll jump after the function returns.
	returnValueVariable := sm.addVariable("$returnValue")
	returnReceiveState := sm.newState()
	returnReceiveState.pushExpression(returnValueVariable)

	// In the call state, add source to call the function and jump to the return state once complete.
	callState := sm.getEndState(callExprInfo.endState, argumentInfo)

	data := asyncFunctionCallData{
		CallExpr:            callExprInfo.expression,
		Arguments:           argumentInfo,
		ReturnValueVariable: returnValueVariable,
		ReturnState:         returnReceiveState,
	}
	sm.addAsyncFunctionCall(callState, data)

	sm.markStates(node, parentState, returnReceiveState)
}

// generateNullComparisonExpression generates the state machine for a null comparison operation expression.
func (sm *stateMachine) generateNullComparisonExpression(node compilergraph.GraphNode, parentState *state) {
	// Generate the state for the child expressions.
	leftHandInfo := sm.generate(node.GetNode(parser.NodeBinaryExpressionLeftExpr), parentState)
	rightHandInfo := sm.generate(node.GetNode(parser.NodeBinaryExpressionRightExpr), leftHandInfo.endState)

	// Add a call to compare the expressions as an expression itself.
	data := struct {
		LeftExpr  string
		RightExpr string
	}{leftHandInfo.expression, rightHandInfo.expression}

	rightHandInfo.endState.pushExpression(sm.templater.Execute("nullcompare", `
		$op.nullcompare({{ .LeftExpr }}, {{ .RightExpr }})
	`, data))

	sm.markStates(node, parentState, rightHandInfo.endState)
}

// generateBinaryOperatorExpression generates the state machine for a code-defined binary operator.
func (sm *stateMachine) generateBinaryOperatorExpression(node compilergraph.GraphNode, parentState *state, exprTemplateStr string) {
	scope, _ := sm.scopegraph.GetScope(node)
	operator, _ := scope.CalledOperator(sm.scopegraph.TypeGraph())

	// Generate the state for the child expressions.
	leftNode := node.GetNode(parser.NodeBinaryExpressionLeftExpr)
	leftScope, _ := sm.scopegraph.GetScope(leftNode)

	leftHandInfo := sm.generate(leftNode, parentState)
	rightHandInfo := sm.generate(node.GetNode(parser.NodeBinaryExpressionRightExpr), leftHandInfo.endState)

	// Create a new state to which we'll jump after the operator function returns.
	returnValueVariable := sm.addVariable("$returnValue")
	returnReceiveState := sm.newState()

	if exprTemplateStr == "" {
		returnReceiveState.pushExpression(returnValueVariable)
	} else {
		returnReceiveState.pushExpression(sm.templater.Execute("binaryopexpr", exprTemplateStr, returnValueVariable))
	}

	// Add a call to the operator node.
	data := asyncFunctionCallData{
		CallExpr:            sm.pather.GetStaticMemberPath(operator, leftScope.ResolvedTypeRef(sm.scopegraph.TypeGraph())),
		Arguments:           []generatedStateInfo{leftHandInfo, rightHandInfo},
		ReturnValueVariable: returnValueVariable,
		ReturnState:         returnReceiveState,
	}
	sm.addAsyncFunctionCall(rightHandInfo.endState, data)

	sm.markStates(node, parentState, returnReceiveState)
}

// generateUnaryOperatorExpression generates the state machine for a code-defined binary operator.
func (sm *stateMachine) generateUnaryOperatorExpression(node compilergraph.GraphNode, parentState *state) {
	scope, _ := sm.scopegraph.GetScope(node)
	operator, _ := scope.CalledOperator(sm.scopegraph.TypeGraph())

	// Generate the state for the child expression.
	childNode := node.GetNode(parser.NodeUnaryExpressionChildExpr)
	childScope, _ := sm.scopegraph.GetScope(childNode)
	childInfo := sm.generate(childNode, parentState)

	// Create a new state to which we'll jump after the operator function returns.
	returnValueVariable := sm.addVariable("$returnValue")
	returnReceiveState := sm.newState()
	returnReceiveState.pushExpression(returnValueVariable)

	// Add a call to the operator node.
	data := asyncFunctionCallData{
		CallExpr:            sm.pather.GetStaticMemberPath(operator, childScope.ResolvedTypeRef(sm.scopegraph.TypeGraph())),
		Arguments:           []generatedStateInfo{childInfo},
		ReturnValueVariable: returnValueVariable,
		ReturnState:         returnReceiveState,
	}
	sm.addAsyncFunctionCall(childInfo.endState, data)

	sm.markStates(node, parentState, returnReceiveState)
}

// generateNativeBinaryExpression generates the state machine for a binary operator that is natively invoked.
func (sm *stateMachine) generateNativeBinaryExpression(node compilergraph.GraphNode, parentState *state, op string) {
	leftHandInfo := sm.generate(node.GetNode(parser.NodeBinaryExpressionLeftExpr), parentState)
	rightHandInfo := sm.generate(node.GetNode(parser.NodeBinaryExpressionRightExpr), leftHandInfo.endState)

	data := struct {
		LeftExpr  string
		RightExpr string
		Operator  string
	}{leftHandInfo.expression, rightHandInfo.expression, op}

	rightHandInfo.endState.pushExpression(sm.templater.Execute("binaryop", `
		({{ .LeftExpr }}) {{ .Operator }} ({{ .RightExpr }})
	`, data))

	sm.markStates(node, parentState, rightHandInfo.endState)
}

// generateNativeUnaryExpression generates the state machine for a unary operator that is natively invoked.
func (sm *stateMachine) generateNativeUnaryExpression(node compilergraph.GraphNode, parentState *state, op string) {
	childInfo := sm.generate(node.GetNode(parser.NodeUnaryExpressionChildExpr), parentState)

	data := struct {
		ChildExpr string
		Operator  string
	}{childInfo.expression, op}

	childInfo.endState.pushExpression(sm.templater.Execute("unaryop", `
		{{ .Operator }}({{ .ChildExpr }})
	`, data))

	sm.markStates(node, parentState, childInfo.endState)
}
