// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package es5

import (
	"fmt"

	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/generator/es5/statemachine"
	"github.com/serulian/compiler/graphs/srg"
	"github.com/serulian/compiler/graphs/typegraph"

	"github.com/cevaris/ordered_map"
)

// generateImplementedMembers generates all the members under the given type or module into ES5.
func (gen *es5generator) generateImplementedMembers(typeOrModule typegraph.TGTypeOrModule) *ordered_map.OrderedMap {
	memberMap := ordered_map.NewOrderedMap()
	members := typeOrModule.Members()
	for _, member := range members {
		if !member.HasImplementation() {
			continue
		}

		memberMap.Set(member, gen.generateImplementedMember(member))
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
	return statemachine.GenerateFunctionSource(gm, gm.Generator.templater, gm.Generator.pather, gm.Generator.scopegraph)
}

// GetterSource returns the generated code for the getter for this member.
func (gm generatingMember) GetterSource() string {
	getterNode, _ := gm.SRGMember.Getter()
	getterBodyNode, _ := getterNode.Body()
	getterBody := propertyBodyInfo{getterBodyNode, []string{""}}
	return statemachine.GenerateFunctionSource(getterBody, gm.Generator.templater, gm.Generator.pather, gm.Generator.scopegraph)
}

// SetterSource returns the generated code for the setter for this member.
func (gm generatingMember) SetterSource() string {
	setterNode, _ := gm.SRGMember.Setter()
	setterBodyNode, _ := setterNode.Body()
	setterBody := propertyBodyInfo{setterBodyNode, []string{"val"}}
	return statemachine.GenerateFunctionSource(setterBody, gm.Generator.templater, gm.Generator.pather, gm.Generator.scopegraph)
}

type propertyBodyInfo struct {
	bodyNode       compilergraph.GraphNode
	parameterNames []string
}

func (pbi propertyBodyInfo) BodyNode() compilergraph.GraphNode {
	return pbi.bodyNode
}

func (pbi propertyBodyInfo) Parameters() []string {
	return pbi.parameterNames
}

func (pbi propertyBodyInfo) Generics() []string {
	return []string{}
}

func (pbi propertyBodyInfo) RequiresThis() bool {
	return true
}

// functionTemplateStr defines the template for generating function members.
const functionTemplateStr = `
{{ if .Member.IsStatic }}$static{{ else }}$instance{{ end }}.{{ .Member.Name }} = {{ .FunctionSource }}`

// propertyTemplateStr defines the template for generating properties.
const propertyTemplateStr = `
{{ if .Member.IsStatic }}$static{{ else }}$instance{{ end }}.{{ .Member.Name }} = 
	{{ if .Member.IsReadOnly }}
		{{ .GetterSource }}
	{{ else }}
	function(opt_val) {
		if (arguments.length == 0) {
			return ({{ .GetterSource }}).call(this);
		} else {
			return ({{ .SetterSource }}).call(this, opt_val);
		}
	};
	{{ end }}
`
