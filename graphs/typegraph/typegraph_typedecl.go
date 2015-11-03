// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typegraph

import (
	"fmt"

	"github.com/serulian/compiler/compilergraph"
)

// TypeKind defines the various supported kinds of types in the TypeGraph.
type TypeKind int

const (
	ClassType TypeKind = iota
	InterfaceType
	GenericType
)

// TGTypeDeclaration represents a type declaration (class, interface or generic) in the type graph.
type TGTypeDecl struct {
	compilergraph.GraphNode
	tdg *TypeGraph
}

// GetTypeForSRGNode returns the TypeGraph type decl for the given SRG type node, if any.
func (g *TypeGraph) GetTypeForSRGNode(node compilergraph.GraphNode) (TGTypeDecl, bool) {
	typeNode, found := g.tryGetMatchingTypeGraphNode(node, NodeTypeClass, NodeTypeInterface)
	if !found {
		return TGTypeDecl{}, false
	}

	return TGTypeDecl{typeNode, g}, true
}

// Name returns the name of the underlying type.
func (tn TGTypeDecl) Name() string {
	if tn.GraphNode.Kind == NodeTypeGeneric {
		return tn.GraphNode.Get(NodePredicateGenericName)
	}

	return tn.GraphNode.Get(NodePredicateTypeName)
}

// Node returns the underlying node in this declaration.
func (tn TGTypeDecl) Node() compilergraph.GraphNode {
	return tn.GraphNode
}

// HasGenerics returns whether this type has generics defined.
func (tn TGTypeDecl) HasGenerics() bool {
	_, isGeneric := tn.GraphNode.TryGet(NodePredicateTypeGeneric)
	return isGeneric
}

// Generics returns the generics on this type.
func (tn TGTypeDecl) Generics() []TGGeneric {
	if tn.GraphNode.Kind == NodeTypeGeneric {
		return make([]TGGeneric, 0)
	}

	it := tn.GraphNode.StartQuery().
		Out(NodePredicateTypeGeneric).
		BuildNodeIterator()

	var generics = make([]TGGeneric, 0)
	for it.Next() {
		generics = append(generics, TGGeneric{it.Node(), tn.tdg})
	}

	return generics
}

// GetTypeReference returns a new type reference to this type.
func (tn TGTypeDecl) GetTypeReference() TypeReference {
	return tn.tdg.NewInstanceTypeReference(tn.GraphNode)
}

// GetStaticMember returns the static member with the given name under this type, if any.
func (tn TGTypeDecl) GetStaticMember(name string) (TGMember, bool) {
	member, found := tn.GetMember(name)
	if !found || !member.IsStatic() {
		return TGMember{}, false
	}

	return member, true
}

// GetMember returns the member with the given name under this type, if any.
func (tn TGTypeDecl) GetMember(name string) (TGMember, bool) {
	node, found := tn.GraphNode.
		StartQuery().
		Out(NodePredicateMember).
		Has(NodePredicateMemberName, name).
		TryGetNode()

	if !found {
		return TGMember{}, false
	}

	return TGMember{node, tn.tdg}, true
}

// Members returns the type graph members for this type node.
func (tn TGTypeDecl) Members() []TGMember {
	it := tn.GraphNode.StartQuery().
		Out(NodePredicateMember).
		BuildNodeIterator()

	var members = make([]TGMember, 0)
	for it.Next() {
		members = append(members, TGMember{it.Node(), tn.tdg})
	}

	return members
}

// TypeKind returns the kind of the type node.
func (tn TGTypeDecl) TypeKind() TypeKind {
	nodeType := tn.Kind.(NodeType)

	switch nodeType {
	case NodeTypeClass:
		return ClassType

	case NodeTypeInterface:
		return InterfaceType

	case NodeTypeGeneric:
		return GenericType

	default:
		panic(fmt.Sprintf("Unknown kind of type %s for node %s", nodeType, tn.NodeId))
		return ClassType
	}
}
