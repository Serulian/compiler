// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typegraph

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/compilerutil"
	"github.com/serulian/compiler/graphs/typegraph/proto"
)

// TypeReference represents a saved type reference in the graph.
type TypeReference struct {
	tdg   *TypeGraph // The type graph.
	value string     // The encoded value of the type reference.
}

// NewTypeReference returns a new type reference pointing to the given type node and some (optional) generics.
func (t *TypeGraph) NewTypeReference(typeNode compilergraph.GraphNode, generics ...TypeReference) TypeReference {
	return TypeReference{
		tdg:   t,
		value: buildTypeReferenceValue(typeNode, false, generics...),
	}
}

// NewInstanceTypeReference returns a new type reference pointing to a type and its generic (if any).
func (t *TypeGraph) NewInstanceTypeReference(typeNode compilergraph.GraphNode) TypeReference {
	var generics = make([]TypeReference, 0)

	git := typeNode.StartQuery().Out(NodePredicateTypeGeneric).BuildNodeIterator()
	for git.Next() {
		generics = append(generics, t.NewTypeReference(git.Node()))
	}

	return t.NewTypeReference(typeNode, generics...)
}

// Verify returns an error if the type reference is invalid in some way. Returns nil if it is valid.
func (tr TypeReference) Verify() error {
	if tr.IsAny() {
		return nil
	}

	refGenerics := tr.Generics()
	referredType := TGTypeDecl{tr.ReferredType(), tr.tdg}
	typeGenerics := referredType.Generics()

	// Check generics count.
	if len(typeGenerics) != len(refGenerics) {
		return fmt.Errorf("Expected %v generics on type '%s', found: %v", len(typeGenerics), referredType.Name(), len(refGenerics))
	}

	// Check generics constraints.
	if len(typeGenerics) > 0 {
		for index, typeGeneric := range typeGenerics {
			refGeneric := refGenerics[index]
			err := refGeneric.CheckSubTypeOf(typeGeneric.Constraint())
			if err != nil {
				return fmt.Errorf("Generic '%s' (#%v) on type '%s' has constraint '%v'. Specified type '%v' does not match: %v", typeGeneric.Name(), index+1, referredType.Name(), typeGeneric.Constraint(), refGeneric, err)
			}
		}
	}

	// Check parameters.
	if tr.HasParameters() && referredType.GraphNode != tr.tdg.FunctionType() {
		return fmt.Errorf("Only function types can have parameters. Found on type: %v", tr)
	}

	return nil
}

// CheckSubTypeOf returns whether the type pointed to by this type reference is a subtype
// of the other type reference: tr <: other
//
// Subtyping rules in Serulian are as follows:
//   - All types are subtypes of 'any'.
//   - A non-nullable type is a subtype of a nullable type (but not vice versa).
//   - A class is a subtype of itself (and no other class) and only if generics and parameters match.
//   - A class (or interface) is a subtype of an interface if it defines that interface's full signature.
func (tr TypeReference) CheckSubTypeOf(other TypeReference) error {
	// If the other is the any type, then we know this to be a subtype.
	if other.IsAny() {
		return nil
	}

	// Check nullability.
	if !other.IsNullable() && tr.IsNullable() {
		return fmt.Errorf("Nullable type '%v' cannot be used in place of non-nullable type '%v'", tr, other)
	}

	// Directly the same = subtype.
	if other == tr {
		return nil
	}

	localType := TGTypeDecl{tr.ReferredType(), tr.tdg}
	otherType := TGTypeDecl{other.ReferredType(), tr.tdg}

	// If the other reference's type node is not an interface, then this reference cannot be a subtype.
	if otherType.TypeKind() != InterfaceType {
		return fmt.Errorf("'%v' cannot be used in place of non-interface '%v'", tr, other)
	}

	localGenerics := tr.Generics()
	otherGenerics := other.Generics()

	// If both types are non-generic, fast path by looking up the signatures on otherType directly on
	// the members of localType. If we don't find exact matches, then we know this is not a subtype.
	if len(localGenerics) == 0 && len(otherGenerics) == 0 {
		oit := otherType.StartQuery().
			Out(NodePredicateTypeMember, NodePredicateTypeOperator).
			BuildNodeIterator(NodePredicateMemberSignature)

		for oit.Next() {
			signature := oit.Values()[NodePredicateMemberSignature]
			_, exists := localType.StartQuery().
				Out(NodePredicateTypeMember, NodePredicateTypeOperator).
				Has(NodePredicateMemberSignature, signature).
				TryGetNode()

			if !exists {
				memberName := oit.Node().Get(NodePredicateMemberName)
				return fmt.Errorf("'%v' cannot be used in place of '%v' as it does not define member '%s' with matching signature", tr, other, memberName)
			}
		}

		return nil
	}

	// Otherwise, build the list of member signatures to compare. We'll have to deserialize them
	// and replace the generic types properly in order to compare.
	otherSigs := other.buildMemberSignaturesMap()
	localSigs := tr.buildMemberSignaturesMap()

	// Ensure that every signature in otherSigs is under localSigs.
	for memberName, memberSig := range otherSigs {
		localSig, exists := localSigs[memberName]
		if !exists {
			return fmt.Errorf("'%v' cannot be used in place of '%v' as it does not define member '%s'", tr, other, memberName)
		}

		if localSig != memberSig {
			return fmt.Errorf("'%v' cannot be used in place of '%v' as member '%s' does not have the same signature in both types", tr, other, memberName)
		}
	}

	return nil
}

// buildMemberSignaturesMap returns a map of member name -> member signature, where each signature
// is adjusted by replacing the referred type's generics, with the references found under this
// overall type reference.
func (tr TypeReference) buildMemberSignaturesMap() map[string]string {
	membersMap := map[string]string{}

	mit := tr.ReferredType().StartQuery().
		Out(NodePredicateTypeMember, NodePredicateTypeOperator).
		BuildNodeIterator(NodePredicateMemberName)

	for mit.Next() {
		// Get the current member's signature, adjusted for the type's generics.
		adjustedMemberSig := tr.adjustedMemberSignature(mit.Node())
		membersMap[mit.Values()[NodePredicateMemberName]] = adjustedMemberSig
	}

	return membersMap
}

// adjustedMemberSignature returns the member signature found on the given node, adjusted for
// the parent type's generics, as specified in this type reference. Will panic if the type reference
// does not refer to the node's parent type.
func (tr TypeReference) adjustedMemberSignature(node compilergraph.GraphNode) string {
	compilerutil.DCHECK(func() bool {
		return node.StartQuery().In(NodePredicateTypeMember).GetNode() == tr.ReferredType()
	}, "Type reference must be parent of member node")

	// Retrieve the generics of the parent type.
	parentNode := tr.ReferredType()
	pgit := parentNode.StartQuery().Out(NodePredicateTypeGeneric).BuildNodeIterator()

	// Parse the member signature.
	esig := &proto.MemberSig{}
	memberSig := node.GetTagged(NodePredicateMemberSignature, esig).(*proto.MemberSig)

	// Replace the generics of the parent type in the signature with those of the type reference.
	generics := tr.Generics()

	var index = 0
	for pgit.Next() {
		genericNode := pgit.Node()
		genericRef := generics[index]

		// Replace the generic in the member type.
		adjustedType := tr.Build(memberSig.GetMemberType()).(TypeReference).
			ReplaceType(genericNode, genericRef).
			Value()

		memberSig.MemberType = &adjustedType

		// Replace the generic in any generic constraints.
		for cindex, constraint := range memberSig.GetGenericConstraints() {
			memberSig.GenericConstraints[cindex] = tr.Build(constraint).(TypeReference).
				ReplaceType(genericNode, genericRef).
				Value()
		}

		index = index + 1
	}

	// Reserialize the member signature.
	return memberSig.Value()
}

// IsAny returns whether this type reference refers to the special 'any' type.
func (tr TypeReference) IsAny() bool {
	return tr.getSlot(trhSlotFlagSpecial)[0] == specialFlagAny
}

// IsLocalRef returns whether this type reference is a localized reference.
func (tr TypeReference) IsLocalRef() bool {
	return tr.getSlot(trhSlotFlagSpecial)[0] == specialFlagLocal
}

// HasGenerics returns whether the type reference has generics.
func (tr TypeReference) HasGenerics() bool {
	return tr.GenericCount() > 0
}

// HasParameters returns whether the type reference has parameters.
func (tr TypeReference) HasParameters() bool {
	return tr.ParameterCount() > 0
}

// GenericCount returns the number of generics on this type reference.
func (tr TypeReference) GenericCount() int {
	return tr.getSlotAsInt(trhSlotGenericCount)
}

// ParameterCount returns the number of parameters on this type reference.
func (tr TypeReference) ParameterCount() int {
	return tr.getSlotAsInt(trhSlotParameterCount)
}

// Generics returns the generics defined on this type reference, if any.
func (tr TypeReference) Generics() []TypeReference {
	return tr.getSubReferences(subReferenceGeneric)
}

// Parameters returns the parameters defined on this type reference, if any.
func (tr TypeReference) Parameters() []TypeReference {
	return tr.getSubReferences(subReferenceParameter)
}

// IsNullable returns whether the type reference refers to a nullable type.
func (tr TypeReference) IsNullable() bool {
	return tr.getSlot(trhSlotFlagNullable)[0] == nullableFlagTrue
}

// ReferredType returns the node to which the type reference refers.
func (tr TypeReference) ReferredType() compilergraph.GraphNode {
	if tr.getSlot(trhSlotFlagSpecial)[0] != specialFlagNormal {
		panic(fmt.Sprintf("Cannot get referred type for special type references of type %s", tr.getSlot(trhSlotFlagSpecial)))
	}

	return tr.tdg.layer.GetNode(tr.getSlot(trhSlotTypeId))
}

// WithGeneric returns a copy of this type reference with the given generic added.
func (tr TypeReference) WithGeneric(generic TypeReference) TypeReference {
	return tr.withSubReference(subReferenceGeneric, generic)
}

// WithParameter returns a copy of this type reference with the given parameter added.
func (tr TypeReference) WithParameter(parameter TypeReference) TypeReference {
	return tr.withSubReference(subReferenceParameter, parameter)
}

// AsNullable returns a copy of this type reference that is nullable.
func (tr TypeReference) AsNullable() TypeReference {
	return tr.withFlag(trhSlotFlagNullable, nullableFlagTrue)
}

// Localize returns a copy of this type reference with any references to the specified generics replaced with
// a string that does reference a specific type node ID, but a localized ID instead. This allows
// type references that reference different type and type member generics to be compared.
func (tr TypeReference) Localize(generics ...compilergraph.GraphNode) TypeReference {
	if tr.getSlot(trhSlotFlagSpecial)[0] != specialFlagNormal {
		return tr
	}

	var currentTypeReference = tr
	for _, genericNode := range generics {
		replacement := TypeReference{
			value: buildLocalizedRefValue(genericNode),
			tdg:   tr.tdg,
		}

		currentTypeReference = currentTypeReference.ReplaceType(genericNode, replacement)
	}

	return currentTypeReference
}

// TransformUnder replaces any generic references in this type reference with the references found in
// the other type reference.
//
// For example, if this type reference is function<T> and the other is
// SomeClass<int>, where T is the generic of 'SomeClass', this method will return function<int>.
func (tr TypeReference) TransformUnder(other TypeReference) TypeReference {
	// Skip 'any' types.
	if tr.IsAny() || other.IsAny() {
		return tr
	}

	// Skip any non-generic types.
	generics := other.Generics()
	if len(generics) == 0 {
		return tr
	}

	// Make sure we have the same number of generics.
	otherTypeNode := other.ReferredType()
	if otherTypeNode.Kind == NodeTypeGeneric {
		panic(fmt.Sprintf("Cannot transform a reference to a generic: %v", other))
	}

	otherType := TGTypeDecl{otherTypeNode, tr.tdg}
	otherTypeGenerics := otherType.Generics()
	if len(generics) != len(otherTypeGenerics) {
		return tr
	}

	// Replace the generics.
	var currentTypeReference = tr
	for index, generic := range generics {
		currentTypeReference = currentTypeReference.ReplaceType(otherTypeGenerics[index].GraphNode, generic)
	}

	return currentTypeReference
}

// ReplaceType returns a copy of this type reference, with the given type node replaced with the
// given type reference.
func (tr TypeReference) ReplaceType(typeNode compilergraph.GraphNode, replacement TypeReference) TypeReference {
	typeNodeRef := TypeReference{
		tdg:   tr.tdg,
		value: buildTypeReferenceValue(typeNode, false),
	}

	// If the current type reference refers to the type node itself, then just wholesale replace it.
	if tr.value == typeNodeRef.value {
		return replacement
	}

	// Otherwise, search for the type string (with length prefix) in the subreferences and replace it there.
	searchString := typeNodeRef.lengthPrefixedValue()
	replacementStr := replacement.lengthPrefixedValue()

	return TypeReference{
		tdg:   tr.tdg,
		value: strings.Replace(tr.value, searchString, replacementStr, -1),
	}
}

// String returns a human-friendly string.
func (tr TypeReference) String() string {
	var buffer bytes.Buffer
	tr.appendHumanString(&buffer)
	return buffer.String()
}

// appendHumanString appends the human-readable version of this type reference to
// the given buffer.
func (tr TypeReference) appendHumanString(buffer *bytes.Buffer) {
	if tr.IsAny() {
		buffer.WriteString("any")
		return
	}

	if tr.IsLocalRef() {
		buffer.WriteString(tr.getSlot(trhSlotTypeId))
		return
	}

	typeNode := tr.ReferredType()

	if typeNode.Kind == NodeTypeGeneric {
		buffer.WriteString(typeNode.Get(NodePredicateGenericName))
	} else {
		buffer.WriteString(typeNode.Get(NodePredicateTypeName))
	}

	if tr.HasGenerics() {
		buffer.WriteRune('<')
		for index, generic := range tr.Generics() {
			if index > 0 {
				buffer.WriteString(", ")
			}

			generic.appendHumanString(buffer)
		}

		buffer.WriteByte('>')
	}

	if tr.HasParameters() {
		buffer.WriteRune('(')
		for index, parameter := range tr.Parameters() {
			if index > 0 {
				buffer.WriteString(", ")
			}

			parameter.appendHumanString(buffer)
		}

		buffer.WriteByte(')')
	}

	if tr.IsNullable() {
		buffer.WriteByte('?')
	}
}

func (tr TypeReference) Name() string {
	return "TypeReference"
}

func (tr TypeReference) Value() string {
	return tr.value
}

func (tr TypeReference) Build(value string) interface{} {
	return TypeReference{
		tdg:   tr.tdg,
		value: value,
	}
}
