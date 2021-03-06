// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grok

import (
	"regexp"
	"strings"

	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/graphs/typegraph"
)

type sourceRangeCapable interface {
	SourceRange() (compilercommon.SourceRange, bool)
}

type multipleSourceRangesCapable interface {
	SourceRanges() []compilercommon.SourceRange
}

// sourceRangesOf returns the source ranges found for the given capable instance, if any.
func sourceRangesOf(src sourceRangeCapable) []compilercommon.SourceRange {
	instance, supportsMultiple := src.(multipleSourceRangesCapable)
	if supportsMultiple {
		return instance.SourceRanges()
	}

	sourceRange, hasSourceRange := src.SourceRange()
	if hasSourceRange {
		return []compilercommon.SourceRange{sourceRange}
	}

	return []compilercommon.SourceRange{}
}

// sourceRangesForTypeRef returns source ranges for the type referenced, if any.
func sourceRangesForTypeRef(typeref typegraph.TypeReference) []compilercommon.SourceRange {
	if !typeref.IsNormal() {
		return []compilercommon.SourceRange{}
	}

	return sourceRangesOf(typeref.ReferredType())
}

// trimDocumentation trims the given documentation string, removing excess whitespace and any documentation following
// an empty line.
func trimDocumentation(documentation string) string {
	parts := strings.Split(documentation, "\n\n")
	return strings.TrimSpace(parts[0])
}

// highlightParameter highlights the given parameter name found in the given documentation, by replacing all ticked
// forms that *aren't* the parameter with italics forms instead.
func highlightParameter(documentation string, paramName string) string {
	if paramName == "" {
		return documentation
	}

	var parameterRegex = regexp.MustCompile("([^`]|^)`(" + paramName + ")`([^`]|$)")
	var allParametersRegex = regexp.MustCompile("([^`]|^)`([a-zA-Z0-9_]+)`([^`]|$)")
	var SENTINAL = "#SENTINAL#"

	documentation = parameterRegex.ReplaceAllString(documentation, "$1"+SENTINAL+"$3")
	documentation = allParametersRegex.ReplaceAllString(documentation, `$1**$2**$3`)
	documentation = strings.Replace(documentation, SENTINAL, "`"+paramName+"`", -1)
	return documentation
}
