// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package srg

//go:generate stringer -type=TypeMemberKind

import (
	"fmt"

	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/parser"
)

// SRGTypeMember wraps a type memeber declaration or definition in the SRG.
type SRGTypeMember struct {
	compilergraph.GraphNode
	srg *SRG // The parent SRG.
}

// TypeMemberKind defines the various supported kinds of type members in the SRG.
type TypeMemberKind int

const (
	ConstructorTypeMember TypeMemberKind = iota
	VarTypeMember
	FunctionTypeMember
	PropertyTypeMember
	OperatorTypeMember
)

// Name returns the name of this type member.
func (m SRGTypeMember) Name() string {
	if m.GraphNode.Kind == parser.NodeTypeOperator {
		return m.GraphNode.Get(parser.NodeOperatorName)
	}

	return m.GraphNode.Get(parser.NodeFunctionName)
}

// Node returns the underlying type member node for this type member.
func (m SRGTypeMember) Node() compilergraph.GraphNode {
	return m.GraphNode
}

// Location returns the source location for this type member.
func (m SRGTypeMember) Location() compilercommon.SourceAndLocation {
	return salForNode(m.GraphNode)
}

// TypeMemberKind returns the kind matching the type member definition/declaration node type.
func (m SRGTypeMember) TypeMemberKind() TypeMemberKind {
	switch m.GraphNode.Kind {
	case parser.NodeTypeConstructor:
		return ConstructorTypeMember

	case parser.NodeTypeFunction:
		return FunctionTypeMember

	case parser.NodeTypeProperty:
		return PropertyTypeMember

	case parser.NodeTypeOperator:
		return OperatorTypeMember

	case parser.NodeTypeField:
		return VarTypeMember

	default:
		panic(fmt.Sprintf("Unknown kind of type member %s", m.GraphNode.Kind))
		return ConstructorTypeMember
	}
}

// ReturnType returns a type reference to the declared type of this type member, if any.
func (m SRGTypeMember) DeclaredType() (SRGTypeRef, bool) {
	// TODO(jschorr): Remove this conditional by unifying the predicates.
	var predicate = parser.NodePropertyDeclaredType
	if m.GraphNode.Kind == parser.NodeTypeField {
		predicate = parser.NodeVariableStatementDeclaredType
	}

	typeRefNode, found := m.GraphNode.TryGetNode(predicate)
	if !found {
		return SRGTypeRef{}, false
	}

	return SRGTypeRef{typeRefNode, m.srg}, true
}

// ReturnType returns a type reference to the return type of this type member, if any.
func (m SRGTypeMember) ReturnType() (SRGTypeRef, bool) {
	typeRefNode, found := m.GraphNode.TryGetNode(parser.NodeFunctionReturnType)
	if !found {
		return SRGTypeRef{}, false
	}

	return SRGTypeRef{typeRefNode, m.srg}, true
}

// Generics returns the generics on this type member.
func (m SRGTypeMember) Generics() []SRGGeneric {
	it := m.GraphNode.StartQuery().
		Out(parser.NodeFunctionGeneric).
		BuildNodeIterator()

	var generics = make([]SRGGeneric, 0)
	for it.Next() {
		generics = append(generics, SRGGeneric{it.Node(), m.srg})
	}

	return generics
}
