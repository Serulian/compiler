// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typegraph

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/compilerutil"

	"github.com/serulian/compiler/graphs/typegraph/proto"
)

// TypeReference represents a saved type reference in the graph.
type TypeReference struct {
	tdg   *TypeGraph // The type graph.
	value string     // The encoded value of the type reference.
}

// Deserializes a type reference string value into a TypeReference.
func (t *TypeGraph) DeserializieTypeRef(value string) TypeReference {
	return TypeReference{
		tdg:   t,
		value: value,
	}
}

// NewTypeReference returns a new type reference pointing to the given type node and some (optional) generics.
func (t *TypeGraph) NewTypeReference(typeDecl TGTypeDecl, generics ...TypeReference) TypeReference {
	return TypeReference{
		tdg:   t,
		value: buildTypeReferenceValue(typeDecl.GraphNode, false, generics...),
	}
}

// NewInstanceTypeReference returns a new type reference pointing to a type and its generics (if any).
func (t *TypeGraph) NewInstanceTypeReference(typeDecl TGTypeDecl) TypeReference {
	typeNode := typeDecl.GraphNode

	// Fast path for generics.
	if typeNode.Kind == NodeTypeGeneric {
		return TypeReference{
			tdg:   t,
			value: buildTypeReferenceValue(typeNode, false),
		}
	}

	var generics = make([]TypeReference, 0)
	git := typeNode.StartQuery().Out(NodePredicateTypeGeneric).BuildNodeIterator()
	for git.Next() {
		genericType := TGTypeDecl{git.Node(), t}
		generics = append(generics, t.NewTypeReference(genericType))
	}

	return t.NewTypeReference(typeDecl, generics...)
}

// Verify returns an error if the type reference is invalid in some way. Returns nil if it is valid.
func (tr TypeReference) Verify() error {
	if tr.IsAny() || tr.IsVoid() {
		return nil
	}

	// Function type references are properly restricted based on the parser, so no checks to make.
	if tr.HasReferredType(tr.tdg.FunctionType()) {
		return nil
	}

	refGenerics := tr.Generics()
	referredType := tr.ReferredType()
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

	return nil
}

// EqualsOrAny returns true if this type reference is equal to the other given, OR if it is 'any'.
func (tr TypeReference) EqualsOrAny(other TypeReference) bool {
	if tr.IsAny() {
		return true
	}

	return tr == other
}

// ContainsType returns true if the current type reference has a reference to the given type.
func (tr TypeReference) ContainsType(typeDecl TGTypeDecl) bool {
	reference := tr.tdg.NewInstanceTypeReference(typeDecl)
	return strings.Contains(tr.value, reference.value)
}

// ExtractTypeDiff attempts to extract the child type reference from this type reference used in place
// of a reference to the given type in the other reference. For example, if this is a reference
// to SomeClass<int> and the other reference is SomeClass<T>, passing in 'T' will return 'int'.
func (tr TypeReference) ExtractTypeDiff(otherRef TypeReference, diffType TGTypeDecl) (TypeReference, bool) {
	// Only normal type references apply.
	if !tr.isNormal() || !otherRef.isNormal() {
		return TypeReference{}, false
	}

	// If the referred type is not the same as the other ref's referred type, nothing more to do.
	if tr.referredTypeNode() != otherRef.referredTypeNode() {
		return TypeReference{}, false
	}

	// If the other reference doesn't even contain the diff type, nothing more to do.
	if !otherRef.ContainsType(diffType) {
		return TypeReference{}, false
	}

	// Check the generics of the type.
	otherGenerics := otherRef.Generics()
	localGenerics := tr.Generics()

	for index, genericRef := range otherGenerics {
		if !genericRef.isNormal() {
			continue
		}

		// If the type referred to by the generic is the diff type, then return the associated
		// generic type in the local reference.
		if genericRef.HasReferredType(diffType) {
			return localGenerics[index], true
		}

		// Recursively check the generic.
		extracted, found := localGenerics[index].ExtractTypeDiff(genericRef, diffType)
		if found {
			return extracted, true
		}
	}

	// Check the parameters of the type.
	otherParameters := otherRef.Parameters()
	localParameters := tr.Parameters()

	if len(otherParameters) != len(localParameters) {
		return TypeReference{}, false
	}

	for index, parameterRef := range otherParameters {
		if !parameterRef.isNormal() {
			continue
		}

		// If the type referred to by the parameter is the diff type, then return the associated
		// parameter type in the local reference.
		if parameterRef.HasReferredType(diffType) {
			return localParameters[index], true
		}

		// Recursively check the parameter.
		extracted, found := localParameters[index].ExtractTypeDiff(parameterRef, diffType)
		if found {
			return extracted, true
		}
	}

	return TypeReference{}, false
}

// CheckNominalConvertable checks that the current type reference refers to a type that is nominally deriving
// from the given type reference's type or vice versa.
func (tr TypeReference) CheckNominalConvertable(other TypeReference) error {
	if !tr.isNormal() || !other.isNormal() {
		return fmt.Errorf("Type '%v' cannot be converted to type '%v'", tr, other)
	}

	referredType := tr.ReferredType()
	otherType := other.ReferredType()

	if referredType.TypeKind() != NominalType && otherType.TypeKind() != NominalType {
		return fmt.Errorf("Type '%v' cannot be converted to or from type '%v'", tr, other)
	}

	if !tr.checkNominalParent(other) && !other.checkNominalParent(tr) {
		return fmt.Errorf("Type '%v' cannot be converted to or from type '%v'", tr, other)
	}

	return nil
}

func (tr TypeReference) checkNominalParent(other TypeReference) bool {
	if tr == other {
		return true
	}

	// Walk the parent types, comparing as we go along.
	referredType := tr.ReferredType()
	if referredType.TypeKind() != NominalType {
		return false
	}

	var parentType = referredType.ParentTypes()[0]
	for {
		if parentType == other {
			return true
		}

		if !parentType.isNormal() {
			return false
		}

		parentTypeRef := parentType.ReferredType()
		if parentTypeRef.TypeKind() != NominalType {
			return false
		}

		parentType = parentTypeRef.ParentTypes()[0]
	}

	return false
}

// CheckStructuralSubtypeOf checks that the current type reference refers to a type that is structurally deriving
// from the given type reference's type.
func (tr TypeReference) CheckStructuralSubtypeOf(other TypeReference) bool {
	if !tr.isNormal() || !other.isNormal() {
		return false
	}

	referredType := tr.ReferredType()
	for _, parentRef := range referredType.ParentTypes() {
		if parentRef == other {
			return true
		}
	}

	return false
}

// CheckConcreteSubtypeOf checks that the current type reference refers to a type that is a concrete subtype
// of the specified *generic* interface.
func (tr TypeReference) CheckConcreteSubtypeOf(otherType TGTypeDecl) ([]TypeReference, error) {
	if otherType.TypeKind() != ImplicitInterfaceType {
		panic("Cannot use non-interface type in call to CheckImplOfGeneric")
	}

	if !otherType.HasGenerics() {
		panic("Cannot use non-generic type in call to CheckImplOfGeneric")
	}

	if !tr.isNormal() {
		if tr.IsAny() {
			return nil, fmt.Errorf("Any type %v does not implement type %v", tr, otherType.Name())
		}

		if tr.IsVoid() {
			return nil, fmt.Errorf("Void type %v does not implement type %v", tr, otherType.Name())
		}

		if tr.IsNullable() {
			return nil, fmt.Errorf("Nullable type %v cannot match type %v", tr, otherType.Name())
		}

		if tr.IsNull() {
			return nil, fmt.Errorf("null %v cannot match type %v", tr, otherType.Name())
		}
	}

	localType := tr.ReferredType()

	// Fast check: If the referred type is the type expected, return it directly.
	if localType.GraphNode == otherType.GraphNode {
		return tr.Generics(), nil
	}

	// For each of the generics defined on the interface, find at least one type member whose
	// type contains a reference to that generic. We'll then search for the same member in the
	// current type reference and (if found), infer the generic type for that generic based
	// on the type found in the same position. Once we have concrete types for each of the generics,
	// we can then perform normal subtype checking to verify.
	otherTypeGenerics := otherType.Generics()
	localTypeGenerics := localType.Generics()

	localRefGenerics := tr.Generics()

	resolvedGenerics := make([]TypeReference, len(otherTypeGenerics))

	for index, typeGeneric := range otherTypeGenerics {
		var matchingMember *TGMember = nil

		// Find a member in the interface that uses the generic in its member type.
		for _, member := range otherType.Members() {
			memberType := member.MemberType()
			if !memberType.ContainsType(typeGeneric.AsType()) {
				continue
			}

			matchingMember = &member
			break
		}

		// If there is no matching member, then we assign a type of "any" for this generic.
		if matchingMember == nil {
			resolvedGenerics[index] = tr.tdg.AnyTypeReference()
			continue
		}

		// Otherwise, lookup the member under the current type reference's type.
		localMember, found := localType.GetMember(matchingMember.Name())
		if !found {
			// If not found, this is not a matching type.
			return nil, fmt.Errorf("Type %v cannot be used in place of type %v as it does not implement member %v", tr, otherType.Name(), matchingMember.Name())
		}

		// Now that we have a matching member in the local type, attempt to extract the concrete type
		// used as the generic.
		concreteType, found := localMember.MemberType().ExtractTypeDiff(matchingMember.MemberType(), typeGeneric.AsType())
		if !found {
			// If not found, this is not a matching type.
			return nil, fmt.Errorf("Type %v cannot be used in place of type %v as member %v does not have the same signature", tr, otherType.Name(), matchingMember.Name())
		}

		// Replace any generics from the local type reference with those of the type.
		var replacedConcreteType = concreteType
		if len(localTypeGenerics) > 0 {
			for index, localGeneric := range localTypeGenerics {
				replacedConcreteType = replacedConcreteType.ReplaceType(localGeneric.AsType(), localRefGenerics[index])
			}
		}

		resolvedGenerics[index] = replacedConcreteType
	}

	return resolvedGenerics, tr.CheckSubTypeOf(tr.tdg.NewTypeReference(otherType, resolvedGenerics...))
}

// CheckSubTypeOf returns whether the type pointed to by this type reference is a subtype
// of the other type reference: tr <: other
//
// Subtyping rules in Serulian are as follows:
//   - All types are subtypes of 'any'.
//   - The special "null" type is a subtype of any *nullable* type.
//   - A non-nullable type is a subtype of a nullable type (but not vice versa).
//   - A class is a subtype of itself (and no other class) and only if generics and parameters match.
//   - A class (or interface) is a subtype of an interface if it defines that interface's full signature.
func (tr TypeReference) CheckSubTypeOf(other TypeReference) error {
	if tr.IsVoid() || other.IsVoid() {
		return fmt.Errorf("Void types cannot be used interchangeably")
	}

	if tr.IsNull() {
		if !other.IsAny() && !other.IsNullable() {
			return fmt.Errorf("null cannot be used in place of non-nullable type %v", other)
		}

		return nil
	}

	if other.IsNull() {
		return fmt.Errorf("null cannot be supertype of any other type")
	}

	// If the other is the any type, then we know this to be a subtype.
	if other.IsAny() {
		return nil
	}

	// If this type is the any type, then it cannot be a subtype.
	if tr.IsAny() {
		return fmt.Errorf("Cannot use type 'any' in place of type '%v'", other)
	}

	// Check nullability.
	if !other.IsNullable() && tr.IsNullable() {
		return fmt.Errorf("Nullable type '%v' cannot be used in place of non-nullable type '%v'", tr, other)
	}

	// Directly the same = subtype.
	if other == tr {
		return nil
	}

	// Strip out the nullability from the other type.
	originalOther := other
	if other.IsNullable() {
		other = other.AsNonNullable()
	}

	// Directly the same = subtype.
	if other == tr {
		return nil
	}

	localType := tr.ReferredType()
	otherType := other.ReferredType()

	// If the other reference's type node is not an interface, then this reference cannot be a subtype.
	if otherType.TypeKind() != ImplicitInterfaceType {
		return fmt.Errorf("'%v' cannot be used in place of non-interface '%v'", tr, originalOther)
	}

	localGenerics := tr.Generics()
	otherGenerics := other.Generics()

	// If both types are non-generic, fast path by looking up the signatures on otherType directly on
	// the members of localType. If we don't find exact matches, then we know this is not a subtype.
	if len(localGenerics) == 0 && len(otherGenerics) == 0 {
		oit := otherType.StartQuery().
			Out(NodePredicateMember, NodePredicateTypeOperator).
			BuildNodeIterator(NodePredicateMemberSignature, NodePredicateMemberName)

		for oit.Next() {
			signature := oit.Values()[NodePredicateMemberSignature]
			_, exists := localType.StartQuery().
				Out(NodePredicateMember, NodePredicateTypeOperator).
				Has(NodePredicateMemberSignature, signature).
				TryGetNode()

			if !exists {
				return buildSubtypeMismatchError(tr, originalOther, oit.Values()[NodePredicateMemberName])
			}
		}

		return nil
	}

	// Otherwise, build the list of member signatures to compare. We'll have to deserialize them
	// and replace the generic types in order to properly compare.
	otherSigs := other.buildMemberSignaturesMap()
	localSigs := tr.buildMemberSignaturesMap()

	// Ensure that every signature in otherSigs is under localSigs.
	for memberName, memberSig := range otherSigs {
		localSig, exists := localSigs[memberName]
		if !exists || localSig != memberSig {
			return buildSubtypeMismatchError(tr, originalOther, memberName)
		}
	}

	return nil
}

// buildSubtypeMismatchError returns an error describing the mismatch between the two types for the given
// member name.
func buildSubtypeMismatchError(left TypeReference, right TypeReference, memberName string) error {
	rightMember, rightExists := right.referredTypeNode().
		StartQuery().
		Out(NodePredicateMember, NodePredicateTypeOperator).
		Has(NodePredicateMemberName, memberName).
		TryGetNode()

	if !rightExists {
		// Should never happen... (of course, it will at some point, now that I said this!)
		panic(fmt.Sprintf("Member '%s' doesn't exist under type '%v'", memberName, right))
	}

	var memberKind = "member"
	if rightMember.Kind == NodeTypeOperator {
		memberKind = "operator"
		memberName = rightMember.Get(NodePredicateOperatorName)
	}

	_, leftExists := left.referredTypeNode().
		StartQuery().
		Out(NodePredicateMember, NodePredicateTypeOperator).
		Has(NodePredicateMemberName, memberName).
		TryGetNode()

	if !leftExists {
		return fmt.Errorf("Type '%v' does not define or export %s '%s', which is required by type '%v'", left, memberKind, memberName, right)
	} else {
		// TODO(jschorr): Be nice to have specific errors here, but it'll require a lot of manual checking.
		return fmt.Errorf("%s '%s' under type '%v' does not match that defined in type '%v'", memberKind, memberName, left, right)
	}
}

// buildMemberSignaturesMap returns a map of member name -> member signature, where each signature
// is adjusted by replacing the referred type's generics, with the references found under this
// overall type reference.
func (tr TypeReference) buildMemberSignaturesMap() map[string]string {
	membersMap := map[string]string{}

	mit := tr.referredTypeNode().StartQuery().
		Out(NodePredicateMember, NodePredicateTypeOperator).
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
		return node.StartQuery().In(NodePredicateMember).GetNode() == tr.referredTypeNode()
	}, "Type reference must be parent of member node")

	// Retrieve the generics of the parent type.
	parentNode := tr.referredTypeNode()
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
		genericType := TGTypeDecl{genericNode, tr.tdg}

		// Replace the generic in the member type.
		adjustedType := tr.Build(memberSig.GetMemberType()).(TypeReference).
			ReplaceType(genericType, genericRef).
			Value()

		memberSig.MemberType = &adjustedType

		// Replace the generic in any generic constraints.
		for cindex, constraint := range memberSig.GetGenericConstraints() {
			memberSig.GenericConstraints[cindex] = tr.Build(constraint).(TypeReference).
				ReplaceType(genericType, genericRef).
				Value()
		}

		index = index + 1
	}

	// Reserialize the member signature.
	return memberSig.Value()
}

// isNormal returns whether this type reference refers to a normal type.
func (tr TypeReference) isNormal() bool {
	return tr.getSlot(trhSlotFlagSpecial)[0] == specialFlagNormal
}

// IsAny returns whether this type reference refers to the special 'any' type.
func (tr TypeReference) IsAny() bool {
	return tr.getSlot(trhSlotFlagSpecial)[0] == specialFlagAny
}

// IsVoid returns whether this type reference refers to the special 'void' type.
func (tr TypeReference) IsVoid() bool {
	return tr.getSlot(trhSlotFlagSpecial)[0] == specialFlagVoid
}

// IsNull returns whether this type reference refers to the special 'null' type
// (which is distinct from a nullable type).
func (tr TypeReference) IsNull() bool {
	return tr.getSlot(trhSlotFlagSpecial)[0] == specialFlagNull
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

// HasReferredType returns whether this type references refers to the given type.
func (tr TypeReference) HasReferredType(typeDecl TGTypeDecl) bool {
	if tr.getSlot(trhSlotFlagSpecial)[0] != specialFlagNormal {
		return false
	}

	return tr.referredTypeNode() == typeDecl.GraphNode
}

// ReferredType returns the type decl to which the type reference refers.
func (tr TypeReference) ReferredType() TGTypeDecl {
	return TGTypeDecl{tr.referredTypeNode(), tr.tdg}
}

// referredTypeNode returns the node to which the type reference refers.
func (tr TypeReference) referredTypeNode() compilergraph.GraphNode {
	if tr.getSlot(trhSlotFlagSpecial)[0] != specialFlagNormal {
		panic(fmt.Sprintf("Cannot get referred type for special type references of type %s", tr.getSlot(trhSlotFlagSpecial)))
	}

	return tr.tdg.layer.GetNode(tr.getSlot(trhSlotTypeId))
}

type MemberResolutionKind int

const (
	MemberResolutionOperator MemberResolutionKind = iota
	MemberResolutionStatic
	MemberResolutionInstance
	MemberResolutionInstanceOrStatic
)

// ResolveMember looks for an member with the given name under the referred type and returns it (if any).
func (tr TypeReference) ResolveMember(memberName string, module compilercommon.InputSource, kind MemberResolutionKind) (TGMember, bool) {
	if tr.getSlot(trhSlotFlagSpecial)[0] != specialFlagNormal {
		return TGMember{}, false
	}

	var connectingPredicate = NodePredicateMember
	var namePredicate = NodePredicateMemberName

	if kind == MemberResolutionOperator {
		connectingPredicate = NodePredicateTypeOperator
		namePredicate = NodePredicateOperatorName
	}

	memberNode, found := tr.referredTypeNode().
		StartQuery().
		Out(connectingPredicate).
		Has(namePredicate, memberName).
		TryGetNode()

	if !found {
		return TGMember{}, false
	}

	member := TGMember{memberNode, tr.tdg}

	// If the member is exported, then always return it. Otherwise, only return it if the asking module
	// is the same as the declaring module.
	if !member.IsExported() {
		memberModule := memberNode.Get(NodePredicateModulePath)
		if memberModule != string(module) {
			return TGMember{}, false
		}
	}

	// Check that the member being static matches the resolution option.
	if (kind == MemberResolutionInstance && member.IsStatic()) ||
		(kind == MemberResolutionStatic && !member.IsStatic()) {
		return TGMember{}, false
	}

	return member, true
}

// WithGeneric returns a copy of this type reference with the given generic added.
func (tr TypeReference) WithGeneric(generic TypeReference) TypeReference {
	return tr.withSubReference(subReferenceGeneric, generic)
}

// WithParameter returns a copy of this type reference with the given parameter added.
func (tr TypeReference) WithParameter(parameter TypeReference) TypeReference {
	return tr.withSubReference(subReferenceParameter, parameter)
}

// AsValueOfStream returns a type reference to a Stream, with this type reference as the value.
func (tr TypeReference) AsValueOfStream() TypeReference {
	return tr.tdg.NewTypeReference(tr.tdg.StreamType(), tr)
}

// AsNullable returns a copy of this type reference that is nullable.
func (tr TypeReference) AsNullable() TypeReference {
	if tr.IsAny() || tr.IsVoid() || tr.IsNull() {
		return tr
	}

	return tr.withFlag(trhSlotFlagNullable, nullableFlagTrue)
}

// AsNonNullable returns a copy of this type reference that is non-nullable.
func (tr TypeReference) AsNonNullable() TypeReference {
	return tr.withFlag(trhSlotFlagNullable, nullableFlagFalse)
}

// Intersect returns the type common to both type references or any if they are uncommon.
func (tr TypeReference) Intersect(other TypeReference) TypeReference {
	if tr.IsVoid() {
		return other
	}

	if other.IsVoid() {
		return tr
	}

	if tr.IsAny() || other.IsAny() {
		return tr.tdg.AnyTypeReference()
	}

	// Ensure both are nullable or non-nullable.
	var trAdjusted = tr
	var otherAdjusted = other

	if tr.IsNullable() {
		otherAdjusted = other.AsNullable()
	}

	if other.IsNullable() {
		trAdjusted = tr.AsNullable()
	}

	if trAdjusted == otherAdjusted {
		return trAdjusted
	}

	if trAdjusted.CheckSubTypeOf(otherAdjusted) == nil {
		return otherAdjusted
	}

	if otherAdjusted.CheckSubTypeOf(trAdjusted) == nil {
		return trAdjusted
	}

	// TODO: support some sort of union types here if/when we need to?
	return tr.tdg.AnyTypeReference()
}

// Localize returns a copy of this type reference with any references to the specified generics replaced with
// a string that does not reference a specific type node ID, but rather a localized ID instead. This allows
// type references that reference different type and type member generics to be compared.
func (tr TypeReference) Localize(generics ...TGGeneric) TypeReference {
	if tr.getSlot(trhSlotFlagSpecial)[0] != specialFlagNormal {
		return tr
	}

	var currentTypeReference = tr
	for _, generic := range generics {
		genericNode := generic.GraphNode
		replacement := TypeReference{
			value: buildLocalizedRefValue(genericNode),
			tdg:   tr.tdg,
		}

		currentTypeReference = currentTypeReference.ReplaceType(generic.AsType(), replacement)
	}

	return currentTypeReference
}

// TransformUnder replaces any generic references in this type reference with the references found in
// the other type reference.
//
// For example, if this type reference is function<T> and the other is
// SomeClass<int>, where T is the generic of 'SomeClass', this method will return function<int>.
func (tr TypeReference) TransformUnder(other TypeReference) TypeReference {
	// Skip 'any' and 'void' types.
	if tr.IsAny() || other.IsAny() {
		return tr
	}

	if tr.IsVoid() || other.IsVoid() {
		return tr
	}

	// Skip any non-generic types.
	otherRefGenerics := other.Generics()
	if len(otherRefGenerics) == 0 {
		return tr
	}

	// Make sure we have the same number of generics.
	otherType := other.ReferredType()
	if otherType.GraphNode.Kind == NodeTypeGeneric {
		panic(fmt.Sprintf("Cannot transform a reference to a generic: %v", other))
	}

	otherTypeGenerics := otherType.Generics()
	if len(otherRefGenerics) != len(otherTypeGenerics) {
		return tr
	}

	// Replace the generics.
	var currentTypeReference = tr
	for index, generic := range otherRefGenerics {
		currentTypeReference = currentTypeReference.ReplaceType(otherTypeGenerics[index].AsType(), generic)
	}

	return currentTypeReference
}

// ReplaceType returns a copy of this type reference, with the given type node replaced with the
// given type reference.
func (tr TypeReference) ReplaceType(typeDecl TGTypeDecl, replacement TypeReference) TypeReference {
	typeNode := typeDecl.GraphNode

	typeNodeRef := TypeReference{
		tdg:   tr.tdg,
		value: buildTypeReferenceValue(typeNode, false),
	}

	// If the current type reference refers to the type node itself, then just wholesale replace it.
	if tr.value == typeNodeRef.value {
		return replacement
	}

	// Check if we have a direct nullable type as well.
	if tr.AsNullable().value == typeNodeRef.AsNullable().value {
		return replacement.AsNullable()
	}

	// Otherwise, search for the type string (with length prefix) in the subreferences and replace it there.
	searchString := typeNodeRef.lengthPrefixedValue()
	replacementStr := replacement.lengthPrefixedValue()

	updatedStr := strings.Replace(tr.value, searchString, replacementStr, -1)

	// Also replace the nullable version.
	tnNullable := typeNodeRef.AsNullable()
	replacementNullable := replacement.AsNullable()

	nullableSearchString := tnNullable.lengthPrefixedValue()
	nullableReplacementStr := replacementNullable.lengthPrefixedValue()

	nullableUpdatedStr := strings.Replace(updatedStr, nullableSearchString, nullableReplacementStr, -1)

	return TypeReference{
		tdg:   tr.tdg,
		value: nullableUpdatedStr,
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

	if tr.IsVoid() {
		buffer.WriteString("void")
		return
	}

	if tr.IsNull() {
		buffer.WriteString("null")
		return
	}

	if tr.IsLocalRef() {
		buffer.WriteString(tr.getSlot(trhSlotTypeId))
		return
	}

	typeNode := tr.referredTypeNode()

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
