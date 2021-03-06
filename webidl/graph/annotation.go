// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/webidl/parser"
)

// IRGAnnotation wraps a WebIDL annotation.
type IRGAnnotation struct {
	compilergraph.GraphNode
	irg *WebIRG // The parent IRG.
}

// Name returns the name of the annotation.
func (i *IRGAnnotation) Name() string {
	return i.GraphNode.Get(parser.NodePredicateAnnotationName)
}

// Value returns the value of the annotation, if any.
func (i *IRGAnnotation) Value() (string, bool) {
	return i.GraphNode.TryGet(parser.NodePredicateAnnotationDefinedValue)
}

// Parameters returns all the parameters declared on the annotation.
func (i *IRGAnnotation) Parameters() []IRGParameter {
	pit := i.GraphNode.StartQuery().
		Out(parser.NodePredicateAnnotationParameter).
		BuildNodeIterator()

	var parameters = make([]IRGParameter, 0)
	for pit.Next() {
		parameter := IRGParameter{pit.Node(), i.irg}
		parameters = append(parameters, parameter)
	}

	return parameters
}

// SourceRange returns the source range of the parameter in source.
func (i *IRGAnnotation) SourceRange() (compilercommon.SourceRange, bool) {
	return i.irg.SourceRangeOf(i.GraphNode)
}
