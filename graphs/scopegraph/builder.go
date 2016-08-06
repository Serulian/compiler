// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scopegraph

import (
	"fmt"
	"strconv"

	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/graphs/scopegraph/proto"
	"github.com/serulian/compiler/graphs/typegraph"
	"github.com/serulian/compiler/parser"

	"github.com/streamrail/concurrent-map"
)

// scopeAccessOption defines the kind of access under which the scope
// exists.
type scopeAccessOption int

const (
	scopeGetAccess scopeAccessOption = iota
	scopeSetAccess
)

// scopeContext represents the currently operating context for scoping, allowing for
// scope-specific overrides of such items as types of expressions.
type scopeContext struct {
	// The access option for the current scope.
	accessOption scopeAccessOption

	// overrideTypes is (if not nil) the map of the overridden type for an expression
	// under this context.
	overrideTypes *map[compilergraph.GraphNodeId]typegraph.TypeReference
}

// getTypeOverride returns the type override for the given expression node, if any.
func (sc scopeContext) getTypeOverride(exprNode compilergraph.GraphNode) (typegraph.TypeReference, bool) {
	if sc.overrideTypes == nil {
		return typegraph.TypeReference{}, false
	}

	ot := *sc.overrideTypes
	value, found := ot[exprNode.NodeId]
	return value, found
}

// withAccess returns the scope context with the access option set to that given.
func (sc scopeContext) withAccess(access scopeAccessOption) scopeContext {
	return scopeContext{
		accessOption:  access,
		overrideTypes: sc.overrideTypes,
	}
}

// withTypeOverride returns the scope context with the type of the given expression node
// overridden.
func (sc scopeContext) withTypeOverride(exprNode compilergraph.GraphNode, typeref typegraph.TypeReference) scopeContext {
	overrideTypes := map[compilergraph.GraphNodeId]typegraph.TypeReference{}

	if sc.overrideTypes != nil {
		existing := *sc.overrideTypes
		for key, value := range existing {
			overrideTypes[key] = value
		}
	}

	overrideTypes[exprNode.NodeId] = typeref

	return scopeContext{
		accessOption:  sc.accessOption,
		overrideTypes: &overrideTypes,
	}
}

// scopeHandler is a handler function for scoping an SRG node of a particular kind.
type scopeHandler func(node compilergraph.GraphNode, context scopeContext) proto.ScopeInfo

// scopeBuilder defines a type for easy scoping of the SRG.
type scopeBuilder struct {
	sg                     *ScopeGraph
	nodeMap                cmap.ConcurrentMap
	modifier               compilergraph.GraphLayerModifier
	inferredParameterTypes cmap.ConcurrentMap

	Status bool
}

// newScopeBuilder returns a new scope builder for the given scope graph.
func newScopeBuilder(sg *ScopeGraph) *scopeBuilder {
	return &scopeBuilder{
		sg:                     sg,
		nodeMap:                cmap.New(),
		modifier:               sg.layer.NewModifier(),
		inferredParameterTypes: cmap.New(),
		Status:                 true,
	}
}

// saveScopes saves all the created scopes to the graph.
func (sb *scopeBuilder) saveScopes() {
	sb.modifier.Apply()
	sb.modifier = sb.sg.layer.NewModifier()
}

// getScopeHandler returns the scope building handler for nodes of the given type.
func (sb *scopeBuilder) getScopeHandler(node compilergraph.GraphNode) scopeHandler {
	switch node.Kind() {
	// Members.
	case parser.NodeTypeVariable:
		return sb.scopeVariable

	case parser.NodeTypeField:
		return sb.scopeField

	// Statements.
	case parser.NodeTypeStatementBlock:
		return sb.scopeStatementBlock

	case parser.NodeTypeBreakStatement:
		return sb.scopeBreakStatement

	case parser.NodeTypeContinueStatement:
		return sb.scopeContinueStatement

	case parser.NodeTypeYieldStatement:
		return sb.scopeYieldStatement

	case parser.NodeTypeReturnStatement:
		return sb.scopeReturnStatement

	case parser.NodeTypeRejectStatement:
		return sb.scopeRejectStatement

	case parser.NodeTypeConditionalStatement:
		return sb.scopeConditionalStatement

	case parser.NodeTypeLoopStatement:
		return sb.scopeLoopStatement

	case parser.NodeTypeWithStatement:
		return sb.scopeWithStatement

	case parser.NodeTypeVariableStatement:
		return sb.scopeVariableStatement

	case parser.NodeTypeSwitchStatement:
		return sb.scopeSwitchStatement

	case parser.NodeTypeMatchStatement:
		return sb.scopeMatchStatement

	case parser.NodeTypeAssignStatement:
		return sb.scopeAssignStatement

	case parser.NodeTypeExpressionStatement:
		return sb.scopeExpressionStatement

	case parser.NodeTypeNamedValue:
		return sb.scopeNamedValue

	case parser.NodeTypeAssignedValue:
		return sb.scopeAssignedValue

	case parser.NodeTypeArrowStatement:
		return sb.scopeArrowStatement

	case parser.NodeTypeResolveStatement:
		return sb.scopeResolveStatement

	// Await expression.
	case parser.NodeTypeAwaitExpression:
		return sb.scopeAwaitExpression

	// SML expression.
	case parser.NodeTypeSmlExpression:
		return sb.scopeSmlExpression

	case parser.NodeTypeSmlText:
		return sb.scopeSmlText

	// Flow expressions.
	case parser.NodeTypeConditionalExpression:
		return sb.scopeConditionalExpression

	case parser.NodeTypeLoopExpression:
		return sb.scopeLoopExpression

	// Access expressions.
	case parser.NodeCastExpression:
		return sb.scopeCastExpression

	case parser.NodeMemberAccessExpression:
		return sb.scopeMemberAccessExpression

	case parser.NodeNullableMemberAccessExpression:
		return sb.scopeNullableMemberAccessExpression

	case parser.NodeDynamicMemberAccessExpression:
		return sb.scopeDynamicMemberAccessExpression

	case parser.NodeStreamMemberAccessExpression:
		return sb.scopeStreamMemberAccessExpression

	// Operator expressions.
	case parser.NodeDefineRangeExpression:
		return sb.scopeDefineRangeExpression

	case parser.NodeSliceExpression:
		return sb.scopeSliceExpression

	case parser.NodeBitwiseXorExpression:
		return sb.scopeBitwiseXorExpression

	case parser.NodeBitwiseOrExpression:
		return sb.scopeBitwiseOrExpression

	case parser.NodeBitwiseAndExpression:
		return sb.scopeBitwiseAndExpression

	case parser.NodeBitwiseShiftLeftExpression:
		return sb.scopeBitwiseShiftLeftExpression

	case parser.NodeBitwiseShiftRightExpression:
		return sb.scopeBitwiseShiftRightExpression

	case parser.NodeBitwiseNotExpression:
		return sb.scopeBitwiseNotExpression

	case parser.NodeBinaryAddExpression:
		return sb.scopeBinaryAddExpression

	case parser.NodeBinarySubtractExpression:
		return sb.scopeBinarySubtractExpression

	case parser.NodeBinaryMultiplyExpression:
		return sb.scopeBinaryMultiplyExpression

	case parser.NodeBinaryDivideExpression:
		return sb.scopeBinaryDivideExpression

	case parser.NodeBinaryModuloExpression:
		return sb.scopeBinaryModuloExpression

	case parser.NodeBooleanAndExpression:
		return sb.scopeBooleanBinaryExpression

	case parser.NodeBooleanOrExpression:
		return sb.scopeBooleanBinaryExpression

	case parser.NodeBooleanNotExpression:
		return sb.scopeBooleanUnaryExpression

	case parser.NodeComparisonEqualsExpression:
		return sb.scopeEqualsExpression

	case parser.NodeComparisonNotEqualsExpression:
		return sb.scopeEqualsExpression

	case parser.NodeComparisonLTEExpression:
		return sb.scopeComparisonExpression

	case parser.NodeComparisonLTExpression:
		return sb.scopeComparisonExpression

	case parser.NodeComparisonGTEExpression:
		return sb.scopeComparisonExpression

	case parser.NodeComparisonGTExpression:
		return sb.scopeComparisonExpression

	case parser.NodeNullComparisonExpression:
		return sb.scopeNullComparisonExpression

	case parser.NodeAssertNotNullExpression:
		return sb.scopeAssertNotNullExpression

	case parser.NodeIsComparisonExpression:
		return sb.scopeIsComparisonExpression

	case parser.NodeInCollectionExpression:
		return sb.scopeInCollectionExpression

	case parser.NodeFunctionCallExpression:
		return sb.scopeFunctionCallExpression

	case parser.NodeGenericSpecifierExpression:
		return sb.scopeGenericSpecifierExpression

	case parser.NodeRootTypeExpression:
		return sb.scopeRootTypeExpression

	// Literal expressions.
	case parser.NodeBooleanLiteralExpression:
		return sb.scopeBooleanLiteralExpression

	case parser.NodeNumericLiteralExpression:
		return sb.scopeNumericLiteralExpression

	case parser.NodeStringLiteralExpression:
		return sb.scopeStringLiteralExpression

	case parser.NodeListExpression:
		return sb.scopeListLiteralExpression

	case parser.NodeSliceLiteralExpression:
		return sb.scopeSliceLiteralExpression

	case parser.NodeMappingLiteralExpression:
		return sb.scopeMappingLiteralExpression

	case parser.NodeMapExpression:
		return sb.scopeMapLiteralExpression

	case parser.NodeNullLiteralExpression:
		return sb.scopeNullLiteralExpression

	case parser.NodeThisLiteralExpression:
		return sb.scopeThisLiteralExpression

	case parser.NodeTypeLambdaExpression:
		return sb.scopeLambdaExpression

	case parser.NodeValLiteralExpression:
		return sb.scopeValLiteralExpression

	case parser.NodeStructuralNewExpression:
		return sb.scopeStructuralNewExpression

	case parser.NodeStructuralNewExpressionEntry:
		return sb.scopeStructuralNewExpressionEntry

	// Template string.
	case parser.NodeTaggedTemplateLiteralString:
		return sb.scopeTaggedTemplateString

	case parser.NodeTypeTemplateString:
		return sb.scopeTemplateStringExpression

	// Named expressions.
	case parser.NodeTypeIdentifierExpression:
		return sb.scopeIdentifierExpression

	default:
		panic(fmt.Sprintf("Unknown SRG node in scoping: %v", node.Kind()))
	}
}

// getScope returns the scope for the given node, building (and waiting) if necessary.
func (sb *scopeBuilder) getScope(node compilergraph.GraphNode, context scopeContext) *proto.ScopeInfo {
	// Check the map cache for the scope.
	found, ok := sb.nodeMap.Get(string(node.NodeId))
	if ok {
		result := found.(proto.ScopeInfo)
		return &result
	}

	built := <-sb.buildScope(node, context)
	return &built
}

// buildScope builds the scope for the given node, returning a channel
// that can be watched for the result.
func (sb *scopeBuilder) buildScope(node compilergraph.GraphNode, context scopeContext) chan proto.ScopeInfo {
	// Execute the handler in a gorountine and return the result channel.
	handler := sb.getScopeHandler(node)
	resultChan := make(chan proto.ScopeInfo)
	go (func() {
		result := handler(node, context)
		if !result.GetIsValid() {
			sb.Status = false
		}

		// Add the scope to the map and on the node.
		sb.nodeMap.Set(string(node.NodeId), result)

		scopeNode := sb.modifier.CreateNode(NodeTypeResolvedScope)
		scopeNode.DecorateWithTagged(NodePredicateScopeInfo, &result)
		scopeNode.Connect(NodePredicateSource, node)

		resultChan <- result
	})()

	return resultChan
}

// decorateWithError decorates an *SRG* node with the specified scope warning.
func (sb *scopeBuilder) decorateWithWarning(node compilergraph.GraphNode, message string, args ...interface{}) {
	warningNode := sb.modifier.CreateNode(NodeTypeWarning)
	warningNode.Decorate(NodePredicateNoticeMessage, fmt.Sprintf(message, args...))
	warningNode.Connect(NodePredicateNoticeSource, node)
}

// decorateWithError decorates an *SRG* node with the specified scope error.
func (sb *scopeBuilder) decorateWithError(node compilergraph.GraphNode, message string, args ...interface{}) {
	errorNode := sb.modifier.CreateNode(NodeTypeError)
	errorNode.Decorate(NodePredicateNoticeMessage, fmt.Sprintf(message, args...))
	errorNode.Connect(NodePredicateNoticeSource, node)
}

// decorateWithSecondaryLabel decorates an *SRG* node with a secondary scope label.
func (sb *scopeBuilder) decorateWithSecondaryLabel(node compilergraph.GraphNode, label proto.ScopeLabel) {
	labelNode := sb.modifier.CreateNode(NodeTypeSecondaryLabel)
	labelNode.Decorate(NodePredicateSecondaryLabelValue, strconv.Itoa(int(label)))
	labelNode.Connect(NodePredicateLabelSource, node)
}

// GetWarnings returns the warnings created during the build pass.
func (sb *scopeBuilder) GetWarnings() []compilercommon.SourceWarning {
	var warnings = make([]compilercommon.SourceWarning, 0)

	it := sb.sg.layer.StartQuery().
		IsKind(NodeTypeWarning).
		BuildNodeIterator()

	for it.Next() {
		warningNode := it.Node()

		// Lookup the location of the SRG source node.
		warningSource := sb.sg.srg.GetNode(warningNode.GetValue(NodePredicateNoticeSource).NodeId())
		location := sb.sg.srg.NodeLocation(warningSource)

		// Add the error.
		msg := warningNode.Get(NodePredicateNoticeMessage)
		warnings = append(warnings, compilercommon.NewSourceWarning(location, msg))
	}

	return warnings
}

// GetErrors returns the errors created during the build pass.
func (sb *scopeBuilder) GetErrors() []compilercommon.SourceError {
	var errors = make([]compilercommon.SourceError, 0)

	it := sb.sg.layer.StartQuery().
		IsKind(NodeTypeError).
		BuildNodeIterator()

	for it.Next() {
		errNode := it.Node()

		// Lookup the location of the SRG source node.
		errorSource := sb.sg.srg.GetNode(errNode.GetValue(NodePredicateNoticeSource).NodeId())
		location := sb.sg.srg.NodeLocation(errorSource)

		// Add the error.
		msg := errNode.Get(NodePredicateNoticeMessage)
		errors = append(errors, compilercommon.NewSourceError(location, msg))
	}

	return errors
}
