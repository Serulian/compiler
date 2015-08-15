// generated by stringer -type=tokenType; DO NOT EDIT

package parser

import "fmt"

const _tokenType_name = "tokenTypeErrortokenTypeEOFtokenTypeWhitespacetokenTypeNewlinetokenTypeSinglelineCommenttokenTypeMultilineCommenttokenTypeDotAccessOperatortokenTypeArrowAccessOperatortokenTypeNullDotAccessOperatortokenTypeNullOrValueOperatortokenTypeLeftBracetokenTypeRightBracetokenTypeLeftParentokenTypeRightParentokenTypeLeftBrackettokenTypeRightBrackettokenTypeEqualstokenTypePlustokenTypeMinustokenTypeDivtokenTypeTimestokenTypeLessThantokenTypeGreaterThantokenTypeLTEtokenTypeGTEtokenTypeEqualsEqualstokenTypeNotEqualstokenTypeNottokenTypeTildetokenTypePipetokenTypeAndtokenTypeBooleanOrtokenTypeBooleanAndtokenTypeNumericLiteraltokenTypeStringLiteraltokenTypeTemplateStringLiteraltokenTypeBooleanLiteraltokenTypeIdentifertokenTypeKeyword"

var _tokenType_index = [...]uint16{0, 14, 26, 45, 61, 87, 112, 138, 166, 196, 224, 242, 261, 279, 298, 318, 339, 354, 367, 381, 393, 407, 424, 444, 456, 468, 489, 507, 519, 533, 546, 558, 576, 595, 618, 640, 670, 693, 711, 727}

func (i tokenType) String() string {
	if i < 0 || i >= tokenType(len(_tokenType_index)-1) {
		return fmt.Sprintf("tokenType(%d)", i)
	}
	return _tokenType_name[_tokenType_index[i]:_tokenType_index[i+1]]
}
