// generated by stringer -type=NodeType; DO NOT EDIT

package typegraph

import "fmt"

const _NodeType_name = "NodeTypeErrorNodeTypeClassNodeTypeInterfaceNodeTypeExternalInterfaceNodeTypeNominalTypeNodeTypeModuleNodeTypeMemberNodeTypeOperatorNodeTypeReturnableNodeTypeGenericNodeTypeAttributeNodeTypeReportedIssueNodeTypeTagged"

var _NodeType_index = [...]uint8{0, 13, 26, 43, 68, 87, 101, 115, 131, 149, 164, 181, 202, 216}

func (i NodeType) String() string {
	if i < 0 || i >= NodeType(len(_NodeType_index)-1) {
		return fmt.Sprintf("NodeType(%d)", i)
	}
	return _NodeType_name[_NodeType_index[i]:_NodeType_index[i+1]]
}
