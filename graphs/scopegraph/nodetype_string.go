// Code generated by "stringer -type=NodeType"; DO NOT EDIT

package scopegraph

import "fmt"

const _NodeType_name = "NodeTypeErrorNodeTypeWarningNodeTypeResolvedScopeNodeTypeSecondaryLabelNodeTypeTagged"

var _NodeType_index = [...]uint8{0, 13, 28, 49, 71, 85}

func (i NodeType) String() string {
	if i < 0 || i >= NodeType(len(_NodeType_index)-1) {
		return fmt.Sprintf("NodeType(%d)", i)
	}
	return _NodeType_name[_NodeType_index[i]:_NodeType_index[i+1]]
}
