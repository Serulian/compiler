// Code generated by "stringer -type=NodeType"; DO NOT EDIT.

package parser

import "fmt"

const _NodeType_name = "NodeTypeErrorNodeTypeFileNodeTypeCommentNodeTypeDecoratorNodeTypeImportNodeTypeImportPackageNodeTypeClassNodeTypeInterfaceNodeTypeNominalNodeTypeStructNodeTypeAgentNodeTypeGenericNodeTypeAgentReferenceNodeTypeFunctionNodeTypeVariableNodeTypeConstructorNodeTypePropertyNodeTypeOperatorNodeTypeFieldNodeTypePropertyBlockNodeTypeParameterNodeTypeMemberTagNodeTypeArrowStatementNodeTypeStatementBlockNodeTypeLoopStatementNodeTypeConditionalStatementNodeTypeReturnStatementNodeTypeYieldStatementNodeTypeRejectStatementNodeTypeBreakStatementNodeTypeContinueStatementNodeTypeVariableStatementNodeTypeWithStatementNodeTypeSwitchStatementNodeTypeMatchStatementNodeTypeAssignStatementNodeTypeResolveStatementNodeTypeExpressionStatementNodeTypeSwitchStatementCaseNodeTypeMatchStatementCaseNodeTypeNamedValueNodeTypeAssignedValueNodeTypeAwaitExpressionNodeTypeLambdaExpressionNodeTypeSmlExpressionNodeTypeSmlAttributeNodeTypeSmlDecoratorNodeTypeSmlTextNodeTypeConditionalExpressionNodeTypeLoopExpressionNodeBitwiseXorExpressionNodeBitwiseOrExpressionNodeBitwiseAndExpressionNodeBitwiseShiftLeftExpressionNodeBitwiseShiftRightExpressionNodeBitwiseNotExpressionNodeBooleanOrExpressionNodeBooleanAndExpressionNodeBooleanNotExpressionNodeKeywordNotExpressionNodeRootTypeExpressionNodeComparisonEqualsExpressionNodeComparisonNotEqualsExpressionNodeComparisonLTEExpressionNodeComparisonGTEExpressionNodeComparisonLTExpressionNodeComparisonGTExpressionNodeNullComparisonExpressionNodeIsComparisonExpressionNodeAssertNotNullExpressionNodeInCollectionExpressionNodeDefineRangeExpressionNodeBinaryAddExpressionNodeBinarySubtractExpressionNodeBinaryMultiplyExpressionNodeBinaryDivideExpressionNodeBinaryModuloExpressionNodeMemberAccessExpressionNodeNullableMemberAccessExpressionNodeDynamicMemberAccessExpressionNodeStreamMemberAccessExpressionNodeCastExpressionNodeFunctionCallExpressionNodeSliceExpressionNodeGenericSpecifierExpressionNodeTaggedTemplateLiteralStringNodeTypeTemplateStringNodeNumericLiteralExpressionNodeStringLiteralExpressionNodeBooleanLiteralExpressionNodeThisLiteralExpressionNodePrincipalLiteralExpressionNodeNullLiteralExpressionNodeValLiteralExpressionNodeListExpressionNodeSliceLiteralExpressionNodeMappingLiteralExpressionNodeMappingLiteralExpressionEntryNodeStructuralNewExpressionNodeStructuralNewExpressionEntryNodeMapExpressionNodeMapExpressionEntryNodeTypeIdentifierExpressionNodeTypeLambdaParameterNodeTypeTypeReferenceNodeTypeStreamNodeTypeSliceNodeTypeMappingNodeTypeNullableNodeTypeVoidNodeTypeAnyNodeTypeStructReferenceNodeTypeIdentifierPathNodeTypeIdentifierAccessNodeTypeTagged"

var _NodeType_index = [...]uint16{0, 13, 25, 40, 57, 71, 92, 105, 122, 137, 151, 164, 179, 201, 217, 233, 252, 268, 284, 297, 318, 335, 352, 374, 396, 417, 445, 468, 490, 513, 535, 560, 585, 606, 629, 651, 674, 698, 725, 752, 778, 796, 817, 840, 864, 885, 905, 925, 940, 969, 991, 1015, 1038, 1062, 1092, 1123, 1147, 1170, 1194, 1218, 1242, 1264, 1294, 1327, 1354, 1381, 1407, 1433, 1461, 1487, 1514, 1540, 1565, 1588, 1616, 1644, 1670, 1696, 1722, 1756, 1789, 1821, 1839, 1865, 1884, 1914, 1945, 1967, 1995, 2022, 2050, 2075, 2105, 2130, 2154, 2172, 2198, 2226, 2259, 2286, 2318, 2335, 2357, 2385, 2408, 2429, 2443, 2456, 2471, 2487, 2499, 2510, 2533, 2555, 2579, 2593}

func (i NodeType) String() string {
	if i < 0 || i >= NodeType(len(_NodeType_index)-1) {
		return fmt.Sprintf("NodeType(%d)", i)
	}
	return _NodeType_name[_NodeType_index[i]:_NodeType_index[i+1]]
}
