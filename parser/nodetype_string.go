// generated by stringer -type=NodeType; DO NOT EDIT

package parser

import "fmt"

const _NodeType_name = "NodeTypeErrorNodeTypeFileNodeTypeCommentNodeTypeDecoratorNodeTypeImportNodeTypeClassNodeTypeInterfaceNodeTypeNominalNodeTypeStructNodeTypeGenericNodeTypeFunctionNodeTypeVariableNodeTypeConstructorNodeTypePropertyNodeTypeOperatorNodeTypeFieldNodeTypePropertyBlockNodeTypeParameterNodeTypeMemberTagNodeTypeArrowStatementNodeTypeStatementBlockNodeTypeLoopStatementNodeTypeConditionalStatementNodeTypeReturnStatementNodeTypeRejectStatementNodeTypeBreakStatementNodeTypeContinueStatementNodeTypeVariableStatementNodeTypeWithStatementNodeTypeMatchStatementNodeTypeAssignStatementNodeTypeExpressionStatementNodeTypeMatchStatementCaseNodeTypeNamedValueNodeTypeAwaitExpressionNodeTypeLambdaExpressionNodeBitwiseXorExpressionNodeBitwiseOrExpressionNodeBitwiseAndExpressionNodeBitwiseShiftLeftExpressionNodeBitwiseShiftRightExpressionNodeBitwiseNotExpressionNodeBooleanOrExpressionNodeBooleanAndExpressionNodeBooleanNotExpressionNodeRootTypeExpressionNodeComparisonEqualsExpressionNodeComparisonNotEqualsExpressionNodeComparisonLTEExpressionNodeComparisonGTEExpressionNodeComparisonLTExpressionNodeComparisonGTExpressionNodeNullComparisonExpressionNodeIsComparisonExpressionNodeAssertNotNullExpressionNodeInCollectionExpressionNodeDefineRangeExpressionNodeBinaryAddExpressionNodeBinarySubtractExpressionNodeBinaryMultiplyExpressionNodeBinaryDivideExpressionNodeBinaryModuloExpressionNodeMemberAccessExpressionNodeNullableMemberAccessExpressionNodeDynamicMemberAccessExpressionNodeStreamMemberAccessExpressionNodeCastExpressionNodeFunctionCallExpressionNodeSliceExpressionNodeGenericSpecifierExpressionNodeTaggedTemplateLiteralStringNodeTypeTemplateStringNodeNumericLiteralExpressionNodeStringLiteralExpressionNodeBooleanLiteralExpressionNodeThisLiteralExpressionNodeNullLiteralExpressionNodeValLiteralExpressionNodeListExpressionNodeSliceLiteralExpressionNodeStructuralNewExpressionNodeStructuralNewExpressionEntryNodeMapExpressionNodeMapExpressionEntryNodeTypeIdentifierExpressionNodeTypeLambdaParameterNodeTypeTypeReferenceNodeTypeStreamNodeTypeSliceNodeTypeMappingNodeTypeNullableNodeTypeVoidNodeTypeAnyNodeTypeIdentifierPathNodeTypeIdentifierAccessNodeTypeTagged"

var _NodeType_index = [...]uint16{0, 13, 25, 40, 57, 71, 84, 101, 116, 130, 145, 161, 177, 196, 212, 228, 241, 262, 279, 296, 318, 340, 361, 389, 412, 435, 457, 482, 507, 528, 550, 573, 600, 626, 644, 667, 691, 715, 738, 762, 792, 823, 847, 870, 894, 918, 940, 970, 1003, 1030, 1057, 1083, 1109, 1137, 1163, 1190, 1216, 1241, 1264, 1292, 1320, 1346, 1372, 1398, 1432, 1465, 1497, 1515, 1541, 1560, 1590, 1621, 1643, 1671, 1698, 1726, 1751, 1776, 1800, 1818, 1844, 1871, 1903, 1920, 1942, 1970, 1993, 2014, 2028, 2041, 2056, 2072, 2084, 2095, 2117, 2141, 2155}

func (i NodeType) String() string {
	if i < 0 || i >= NodeType(len(_NodeType_index)-1) {
		return fmt.Sprintf("NodeType(%d)", i)
	}
	return _NodeType_name[_NodeType_index[i]:_NodeType_index[i+1]]
}
