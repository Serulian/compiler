// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package statemachine

// generatingItem wraps some data with an additional generator field.
type generatingItem struct {
	Item      interface{}
	generator *stateGenerator
}

// Snippets returns a helper type for generating small snippets of state-related ES code.
func (gi generatingItem) Snippets() snippets {
	return gi.generator.snippets()
}
