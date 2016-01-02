// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webidl

import (
	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/webidl/parser"
)

// irgASTNode represents a parser-compatible AST node, backed by an IRG node.
type irgASTNode struct {
	graphNode compilergraph.GraphNode // The backing graph node.
}

// Connect connects an IRG AST node to another IRG AST node.
func (ast *irgASTNode) Connect(predicate string, other parser.AstNode) parser.AstNode {
	ast.graphNode.Connect(predicate, other.(*irgASTNode).graphNode)
	return ast
}

// Decorate decorates an IRG AST node with the given value.
func (ast *irgASTNode) Decorate(predicate string, value string) parser.AstNode {
	ast.graphNode.Decorate(predicate, value)
	return ast
}

// buildASTNode constructs a new node in the IRG.
func (g *WebIRG) buildASTNode(source compilercommon.InputSource, kind parser.NodeType) parser.AstNode {
	graphNode := g.layer.CreateNode(kind)
	return &irgASTNode{
		graphNode: graphNode,
	}
}