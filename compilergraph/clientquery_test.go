// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Portions copied and modified from: https://github.com/golang/go/blob/master/src/text/template/parse/lex_test.go

package compilergraph

import (
	"testing"

	"github.com/google/cayley"
	"github.com/serulian/compiler/compilerutil"
	"github.com/stretchr/testify/assert"
)

func createClientQueryTestLayer(t *testing.T) *GraphLayer {
	store, err := cayley.NewMemoryGraph()
	assert.Nil(t, err, "Could not construct Cayley graph")

	gl := &GraphLayer{
		id:                compilerutil.NewUniqueId(),
		prefix:            "testprefix",
		cayleyStore:       store,
		nodeKindPredicate: "node-kind",
		nodeKindEnum:      TestNodeTypeTagged,
	}

	// Add some nodes.
	firstNode := gl.CreateNode(TestNodeTypeFirst)
	secondNode := gl.CreateNode(TestNodeTypeFirst)
	thirdNode := gl.CreateNode(TestNodeTypeFirst)
	fourthNode := gl.CreateNode(TestNodeTypeFirst)

	firstNode.Decorate("test-id", "first")
	secondNode.Decorate("test-id", "second")
	thirdNode.Decorate("test-id", "third")
	fourthNode.Decorate("test-id", "fourth")

	firstNode.Decorate("loves", "cake")
	secondNode.Decorate("loves", "cake")
	thirdNode.Decorate("loves", "pie")
	fourthNode.Decorate("loves", "pie")

	firstNode.Decorate("calories", "10")
	secondNode.Decorate("calories", "100")
	thirdNode.Decorate("calories", "120")
	fourthNode.Decorate("calories", "200")

	return gl
}

func getTestNode(gl *GraphLayer, name string) GraphNode {
	return gl.StartQuery(name).In("test-id").GetNode()
}

func TestBasicClientQuery(t *testing.T) {
	gl := createClientQueryTestLayer(t)

	// Find the node that likes pie and has calories between 100 and 150.
	node, found := gl.StartQuery().
		Has("loves", "pie").
		HasWhere("calories", WhereGTE, "100").
		HasWhere("calories", WhereLTE, "150").
		TryGetNode()

	assert.True(t, found, "Missing expected node")
	assert.Equal(t, getTestNode(gl, "third"), node, "Expected third node")
}

func TestBasicClientQuery2(t *testing.T) {
	gl := createClientQueryTestLayer(t)

	// Find the node that likes pie and has calories greater than 150.
	node, found := gl.StartQuery().
		Has("loves", "pie").
		HasWhere("calories", WhereGTE, "150").
		TryGetNode()

	assert.True(t, found, "Missing expected node")
	assert.Equal(t, getTestNode(gl, "fourth"), node, "Expected fourth node")
}

func TestFilteredClientQuery(t *testing.T) {
	gl := createClientQueryTestLayer(t)

	veganNode := gl.CreateNode(TestNodeTypeSecond)
	veganNode.Decorate("diet-type", "vegan")

	kosherNode := gl.CreateNode(TestNodeTypeSecond)
	kosherNode.Decorate("diet-type", "kosher")

	// Find the node that likes pie with calories greater than 150 and is vegan.
	fifthNode := gl.CreateNode(TestNodeTypeFirst)
	fifthNode.Decorate("loves", "pie")
	fifthNode.Connect("has-diet", veganNode)
	fifthNode.Decorate("calories", "200")

	firstNode := getTestNode(gl, "first")
	firstNode.Connect("has-diet", veganNode)

	secondNode := getTestNode(gl, "second")
	secondNode.Connect("has-diet", kosherNode)

	thirdNode := getTestNode(gl, "third")
	thirdNode.Connect("has-diet", veganNode)

	fourthNode := getTestNode(gl, "fourth")
	fourthNode.Connect("has-diet", kosherNode)

	filter := func(q *GraphQuery) Query {
		return q.Out("has-diet").Has("diet-type", "vegan")
	}

	node, found := gl.StartQuery().
		Has("loves", "pie").
		FilterBy(filter).
		HasWhere("calories", WhereGTE, "150").
		TryGetNode()

	assert.True(t, found, "Missing expected node")
	assert.Equal(t, fifthNode, node, "Expected fifth node")
}