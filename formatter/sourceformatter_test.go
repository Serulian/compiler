// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package formatter

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/serulian/compiler/compilercommon"
	"github.com/serulian/compiler/parser"
	"github.com/stretchr/testify/assert"
)

type goldenTest struct {
	name     string
	filename string
}

func (ft *goldenTest) input() []byte {
	b, err := ioutil.ReadFile(fmt.Sprintf("tests/%s.in.seru", ft.filename))
	if err != nil {
		panic(err)
	}

	return b
}

func (ft *goldenTest) output() []byte {
	b, err := ioutil.ReadFile(fmt.Sprintf("tests/%s.out.seru", ft.filename))
	if err != nil {
		panic(err)
	}

	return b
}

func (ft *goldenTest) writeFormatted(value []byte) {
	err := ioutil.WriteFile(fmt.Sprintf("tests/%s.out.seru", ft.filename), value, 0644)
	if err != nil {
		panic(err)
	}
}

var goldenTests = []goldenTest{
	{"basic test", "basic"},
	{"comments test", "comments"},
	{"unary precedence test", "unary"},
	{"binary precedence test", "binary"},
	{"imports test", "imports"},
	{"relative imports test", "relative"},
	{"template strings test", "templatestrings"},
}

func conductParsing(t *testing.T, test goldenTest, source []byte) (*parseTree, formatterNode) {
	parseTree := newParseTree(source)
	inputSource := compilercommon.InputSource(test.filename)
	rootNode := parser.Parse(parseTree.createAstNode, nil, inputSource, string(source))
	if !assert.Empty(t, parseTree.errors, "Expected no parse errors for test %s", test.name) {
		return nil, formatterNode{}
	}

	return parseTree, rootNode.(formatterNode)
}

func TestGolden(t *testing.T) {
	for _, test := range goldenTests {
		if os.Getenv("FILTER") != "" {
			if !strings.Contains(test.name, os.Getenv("FILTER")) {
				continue
			} else {
				fmt.Printf("Matched Test: %v\n", test.name)
			}
		}

		parseTree, rootNode := conductParsing(t, test, test.input())
		if parseTree == nil {
			continue
		}

		formatted := buildFormattedSource(parseTree, rootNode, importHandlingInfo{})
		if os.Getenv("REGEN") == "true" {
			test.writeFormatted(formatted)
		} else {
			expected := string(test.output())

			// Make sure the output is equal to that expected.
			if !assert.Equal(t, expected, string(formatted), test.name) {
				t.Log(string(formatted))
			}

			// Process the formatted source again and ensure it doesn't change.
			reparsedTree, reparsedRootNode := conductParsing(t, test, formatted)
			if reparsedTree == nil {
				continue
			}

			formattedAgain := buildFormattedSource(reparsedTree, reparsedRootNode, importHandlingInfo{})
			if !assert.Equal(t, string(formatted), string(formattedAgain), test.name) {
				t.Log(string(formattedAgain))
			}
		}
	}
}
