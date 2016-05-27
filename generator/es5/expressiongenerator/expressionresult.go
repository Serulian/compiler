// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package expressiongenerator

import "github.com/serulian/compiler/generator/escommon/esbuilder"

// ExpressionResult represents the result of generating an expression.
type ExpressionResult struct {
	inlineExpr esbuilder.ExpressionBuilder // The built inline expression.
	wrappers   []*expressionWrapper        // If this expression is async, the wrappers around the inline expr.
	isPromise  bool                        // Whether the result is a promise.
}

// IsPromise returns true if the generated expression is a promise.
func (er ExpressionResult) IsPromise() bool {
	return er.isPromise
}

// IsAsync returns true if the generated expression is asynchronous.
func (er ExpressionResult) IsAsync() bool {
	return len(er.wrappers) > 0
}

// Build returns the builder for this expression.
func (er ExpressionResult) Build() esbuilder.SourceBuilder {
	return er.BuildWrapped("", nil)
}

// BuildWrapped returns the builder for this expression. If specified, the expression will be wrapped
// via the given template string. The original expression builder reference will be placed into
// a template field with name "ResultExpr", while any data passed into this method will be placed
// into "Data".
func (er ExpressionResult) BuildWrapped(wrappingTemplateStr string, data interface{}) esbuilder.SourceBuilder {
	var result esbuilder.SourceBuilder = er.inlineExpr

	// If specified, wrap the expression via the template string.
	if wrappingTemplateStr != "" {
		fullData := struct {
			ResultExpr esbuilder.SourceBuilder
			Data       interface{}
		}{result, data}

		result = esbuilder.Template("getbuilder", wrappingTemplateStr, fullData)
	}

	// For each expression wrapper (in *reverse order*), wrap the result expression via the wrapper.
	// This is used to generate promise and other async wrappings.
	for rindex, _ := range er.wrappers {
		wrapper := er.wrappers[len(er.wrappers)-rindex-1]

		templateStr := `({{ emit .PromisingExpression }}).then(function({{ .ResultName }}) {
	{{ range $idx, $expr := .IntermediateExpressions }}
		{{ emit $expr }};
	{{ end }}
	{{ if .IsTopLevel }}
		{{ emit .WrappedExpression }}
	{{ else }}
		return ({{ emit .WrappedExpression }});
	{{ end }}
})`

		data := struct {
			// PromisingExpression is the expression of type promise.
			PromisingExpression esbuilder.ExpressionBuilder

			// ResultName is the name of the result of the promising expression.
			ResultName string

			// WrappedExpression is the expression being wrapped.
			WrappedExpression esbuilder.SourceBuilder

			// IsTopLevel returns whether this is the top-level wrapping of the expression.
			IsTopLevel bool

			// IntermediateExpressions returns the intermediate expression that should be emitted
			// before the wrapped expression is invoked.
			IntermediateExpressions []esbuilder.ExpressionBuilder
		}{wrapper.promisingExpr, wrapper.resultName, result, rindex == 0, wrapper.intermediateExpressions}

		result = esbuilder.Template("wrapper", templateStr, data)
	}

	return result
}
