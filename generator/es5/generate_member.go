// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package es5

import (
	"fmt"

	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/compilerutil"
	"github.com/serulian/compiler/generator/es5/statemachine"
	"github.com/serulian/compiler/graphs/srg"
	"github.com/serulian/compiler/graphs/typegraph"
)

// generateImplementedMembers generates all the members under the given type or module into ES5.
func (gen *es5generator) generateImplementedMembers(typeOrModule typegraph.TGTypeOrModule) map[typegraph.TGMember]string {
	// Queue all the members to be generated.
	members := typeOrModule.Members()
	generatedSource := make([]string, len(members))
	queue := compilerutil.Queue()
	for index, member := range members {
		fn := func(key interface{}, value interface{}) bool {
			generatedSource[key.(int)] = gen.generateImplementedMember(value.(typegraph.TGMember))
			return true
		}

		if member.HasImplementation() {
			queue.Enqueue(index, member, fn)
		}
	}

	// Generate the full source tree for each member.
	queue.Run()

	// Build a map from member to source tree.
	memberMap := map[typegraph.TGMember]string{}
	for index, member := range members {
		memberMap[member] = generatedSource[index]
	}

	return memberMap
}

// generateImplementedMember generates the given member into ES5.
func (gen *es5generator) generateImplementedMember(member typegraph.TGMember) string {
	srgMember, _ := member.SRGMember()

	generating := generatingMember{member, srgMember, gen}

	switch srgMember.MemberKind() {
	case srg.ConstructorMember:
		fallthrough

	case srg.FunctionMember:
		fallthrough

	case srg.OperatorMember:
		return gen.templater.Execute("function", functionTemplateStr, generating)

	case srg.PropertyMember:
		return gen.templater.Execute("property", propertyTemplateStr, generating)

	default:
		panic(fmt.Sprintf("Unknown kind of member %s", srgMember.MemberKind()))
	}
}

// generatingMember represents a member being generated.
type generatingMember struct {
	Member    typegraph.TGMember
	SRGMember srg.SRGMember
	Generator *es5generator // The parent generator.
}

// IsStatic returns whether the generating member is static.
func (gm generatingMember) IsStatic() bool {
	return gm.Member.IsStatic()
}

// RequiresThis returns whether the generating member is requires the "this" var.
func (gm generatingMember) RequiresThis() bool {
	return !gm.Member.IsStatic()
}

// Generics returns the names of the generics for this member, if any.
func (gm generatingMember) Generics() []string {
	generics := gm.Member.Generics()
	genericNames := make([]string, len(generics))
	for index, generic := range generics {
		genericNames[index] = generic.Name()
	}

	return genericNames
}

// Parameters returns the names of the parameters for this member, if any.
func (gm generatingMember) Parameters() []string {
	parameters := gm.SRGMember.Parameters()
	parameterNames := make([]string, len(parameters))
	for index, parameter := range parameters {
		parameterNames[index] = parameter.Name()
	}

	return parameterNames
}

func (gm generatingMember) BodyNode() compilergraph.GraphNode {
	bodyNode, _ := gm.SRGMember.Body()
	return bodyNode
}

// FunctionSource returns the generated code for the implementation for this member.
func (gm generatingMember) FunctionSource() string {
	return statemachine.GenerateFunctionSource(gm, gm.Generator.templater, gm.Generator.scopegraph)
}

// functionTemplateStr defines the template for generating function members.
const functionTemplateStr = `
{{ if .Member.IsStatic }}$static{{ else }}$instance{{ end }}.{{ .Member.Name }} = {{ .FunctionSource }}`

// propertyTemplateStr defines the template for generating properties.
const propertyTemplateStr = `
	// THIS
`
