// generated by stringer -type=NodeType; DO NOT EDIT

package parser

import "fmt"

const _NodeType_name = "NodeTypeErrorNodeTypeFileNodeTypeCommentNodeTypeDecoratorNodeTypeImportNodeTypeClassNodeTypeInterfaceNodeTypeGenericNodeTypeFunctionNodeTypeConstructorNodeTypePropertyNodeTypeOperatorNodeTypeFieldNodeTypePropertyBlockNodeTypeParameterNodeTypeStatementBlockNodeTypeLoopStatementNodeTypeConditionalStatementNodeTypeReturnStatementNodeTypeBreakStatementNodeTypeContinueStatementNodeTypeVariableStatementNodeTypeWithStatementNodeTypeMatchStatementNodeTypeAssignStatementNodeTypeMatchStatementCaseNodeTypeAwaitExpressionNodeTypeArrowExpressionNodeTypeLambdaExpressionNodeBitwiseXorExpressionNodeBitwiseOrExpressionNodeBitwiseAndExpressionNodeBitwiseShiftLeftExpressionNodeBitwiseShiftRightExpressionNodeBitwiseNotExpressionNodeBooleanOrExpressionNodeBooleanAndExpressionNodeBooleanNotExpressionNodeComparisonEqualsExpressionNodeComparisonNotEqualsExpressionNodeComparisonLTEExpressionNodeComparisonGTEExpressionNodeComparisonLTExpressionNodeComparisonGTExpressionNodeNullComparisonExpressionNodeDefineRangeExpressionNodeBinaryAddExpressionNodeBinarySubtractExpressionNodeBinaryMultiplyExpressionNodeBinaryDivideExpressionNodeBinaryModuloExpressionNodeMemberAccessExpressionNodeNullableMemberAccessExpressionNodeDynamicMemberAccessExpressionNodeStreamMemberAccessExpressionNodeCastExpressionNodeFunctionCallExpressionNodeSliceExpressionNodeNumericLiteralExpressionNodeStringLiteralExpressionNodeBooleanLiteralExpressionNodeTemplateStringLiteralExpressionNodeListExpressionNodeMapExpressionNodeMapExpressionEntryNodeTypeIdentifierExpressionNodeTypeTypeReferenceNodeTypeStreamNodeTypeNullableNodeTypeVoidNodeTypeIdentifierPathNodeTypeIdentifierAccessNodeTypeTagged"

var _NodeType_index = [...]uint16{0, 13, 25, 40, 57, 71, 84, 101, 116, 132, 151, 167, 183, 196, 217, 234, 256, 277, 305, 328, 350, 375, 400, 421, 443, 466, 492, 515, 538, 562, 586, 609, 633, 663, 694, 718, 741, 765, 789, 819, 852, 879, 906, 932, 958, 986, 1011, 1034, 1062, 1090, 1116, 1142, 1168, 1202, 1235, 1267, 1285, 1311, 1330, 1358, 1385, 1413, 1448, 1466, 1483, 1505, 1533, 1554, 1568, 1584, 1596, 1618, 1642, 1656}

func (i NodeType) String() string {
	if i < 0 || i >= NodeType(len(_NodeType_index)-1) {
		return fmt.Sprintf("NodeType(%d)", i)
	}
	return _NodeType_name[_NodeType_index[i]:_NodeType_index[i+1]]
}
