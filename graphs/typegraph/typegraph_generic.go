// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typegraph

import (
	"github.com/serulian/compiler/compilergraph"
)

// TGGeneric represents a generic in the type graph.
type TGGeneric struct {
	compilergraph.GraphNode
	tdg *TypeGraph
}

// Name returns the name of the underlying generic.
func (tn TGGeneric) Name() string {
	return tn.GraphNode.Get(NodePredicateGenericName)
}

// DescriptiveName returns a nice human-readable name for the generic.
func (tn TGGeneric) DescriptiveName() string {
	return tn.AsType().DescriptiveName()
}

// GetTypeReference returns a new type reference to this generic.
func (tn TGGeneric) GetTypeReference() TypeReference {
	return tn.AsType().GetTypeReference()
}

// Node returns the underlying node in this declaration.
func (tn TGGeneric) Node() compilergraph.GraphNode {
	return tn.GraphNode
}

// Constraint returns the type constraint on this generic.
func (tn TGGeneric) Constraint() TypeReference {
	constraint, hasConstraint := tn.GraphNode.TryGetTagged(NodePredicateGenericSubtype, tn.tdg.AnyTypeReference())
	if hasConstraint {
		return constraint.(TypeReference)
	}

	return tn.tdg.AnyTypeReference()
}

// Title returns a nice title for the generic.
func (tn TGGeneric) Title() string {
	return "generic"
}

// AsType returns the generic as a TGTypeDecl.
func (tn TGGeneric) AsType() TGTypeDecl {
	return TGTypeDecl{tn.GraphNode, tn.tdg}
}
