// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package srg

import (
	"testing"

	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/parser"

	"github.com/stretchr/testify/assert"
)

func TestFindNodeForPosition(t *testing.T) {
	testSRG := getSRG(t, "tests/position/position.seru")
	cit := testSRG.AllComments()
	for cit.Next() {
		comment := SRGComment{cit.Node(), testSRG}
		parent := comment.ParentNode()

		source := parent.Get(parser.NodePredicateSource)
		startRune := parent.GetValue(parser.NodePredicateStartRune).Int()

		sal := compilercommon.NewSourceAndLocation(compilercommon.InputSource(source), startRune)
		node, found := testSRG.FindNodeForLocation(sal)
		if !assert.True(t, found, "Missing node with comment %s", comment.Contents()) {
			continue
		}

		if !assert.Equal(t, node.GetNodeId(), parent.GetNodeId(), "Mismatch of node with comment %s", comment.Contents()) {
			continue
		}
	}
}