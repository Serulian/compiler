// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compilergraph

import (
	"fmt"
	"strconv"

	"github.com/google/cayley"
)

// GraphNodeId represents an ID for a node in the graph.
type GraphNodeId string

// GraphNode represents a single node in a graph layer.
type GraphNode struct {
	NodeId GraphNodeId // Unique ID for the node.
	layer  *GraphLayer // The layer that owns the node.
}

// taggedValue defines an interface for
type TaggedValue interface {
	Name() string                   // The unique name for this kind of value.
	Value() string                  // The string value.
	Build(value string) interface{} // Builds a new tagged value from the given value string.
}

// Connect decorates the given graph node with a predicate pointing at the given target node.
func (gn *GraphNode) Connect(predicate string, target GraphNode) {
	gn.Decorate(predicate, string(target.NodeId))
}

// Decorate decorates the given graph node with a predicate pointing at the given target.
func (gn *GraphNode) Decorate(predicate string, target string) {
	fullPredicate := gn.layer.prefix + "-" + predicate
	gn.layer.cayleyStore.AddQuad(cayley.Quad(string(gn.NodeId), fullPredicate, target, gn.layer.prefix))
}

// DecorateWithTagged decorates the given graph node with a predicate pointing to a tagged value.
// Tagged values are typically used for values that would otherwise not be unique (such as enums).
func (gn *GraphNode) DecorateWithTagged(predicate string, value TaggedValue) {
	gn.Decorate(predicate, gn.layer.getTaggedKey(value))
}

// StartQuery starts a new query on the graph layer, with its origin being the current node.
func (gn *GraphNode) StartQuery() *GraphQuery {
	return gn.layer.StartQuery(string(gn.NodeId))
}

// GetAsInt returns the value of the given predicate found on this node as an integer.
func (gn *GraphNode) GetInt(predicateName string) int64 {
	strValue := gn.Get(predicateName)
	i, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Could not convert predicate %v on node %v to an int: %v", predicateName, gn.NodeId, strValue))
	}
	return i
}

// GetTagged returns the value of the given predicate found on this node, "cast" to the type of the
// given tagged value.
func (gn *GraphNode) GetTagged(predicateName string, example TaggedValue) interface{} {
	strValue := gn.Get(predicateName)
	return gn.layer.parseTaggedKey(strValue, example)
}

// Get returns the value of the given predicate found on this node and panics otherwise.
func (gn *GraphNode) Get(predicateName string) string {
	value, found := gn.TryGet(predicateName)
	if !found {
		panic(fmt.Sprintf("Could not find value for predicate %s on node %s", predicateName, gn.NodeId))
	}

	return value
}

// TryGet returns the value of the given predicate found on this node (if any).
func (gn *GraphNode) TryGet(predicateName string) (string, bool) {
	return gn.StartQuery().Out(predicateName).GetValue()
}