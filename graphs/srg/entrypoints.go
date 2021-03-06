// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package srg

import (
	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/sourceshape"
)

// EntrypointImplementations returns an iterator of all SRG members/impls in the SRG with bodies.
func (g *SRG) EntrypointImplementations() SRGImplementableIterator {
	iterator := g.layer.StartQuery().
		Out(sourceshape.NodePredicateBody).
		In(sourceshape.NodePredicateBody).
		BuildNodeIterator()

	return SRGImplementableIterator{iterator, g}
}

// EntrypointVariables returns an iterator of all vars in the SRG that are entrypoints for
// scoping (currently variables and fields).
func (g *SRG) EntrypointVariables() SRGMemberIterator {
	iterator := g.layer.StartQuery().IsKind(sourceshape.NodeTypeVariable, sourceshape.NodeTypeField).BuildNodeIterator()
	return SRGMemberIterator{iterator, g}
}

// ImplicitLambdaExpressions returns an iterator of all implicit lambda expressions defined in the SRG.
func (g *SRG) ImplicitLambdaExpressions() compilergraph.NodeIterator {
	return g.layer.StartQuery().
		IsKind(sourceshape.NodeTypeLambdaExpression).
		With(sourceshape.NodeLambdaExpressionChildExpr).
		BuildNodeIterator()
}
