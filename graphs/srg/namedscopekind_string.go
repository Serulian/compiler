// generated by stringer -type=NamedScopeKind; DO NOT EDIT

package srg

import "fmt"

const _NamedScopeKind_name = "NamedScopeTypeNamedScopeMemberNamedScopeImportNamedScopeParameterNamedScopeValueNamedScopeVariable"

var _NamedScopeKind_index = [...]uint8{0, 14, 30, 46, 65, 80, 98}

func (i NamedScopeKind) String() string {
	if i < 0 || i >= NamedScopeKind(len(_NamedScopeKind_index)-1) {
		return fmt.Sprintf("NamedScopeKind(%d)", i)
	}
	return _NamedScopeKind_name[_NamedScopeKind_index[i]:_NamedScopeKind_index[i+1]]
}
