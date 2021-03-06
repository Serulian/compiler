// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scopegraph

import (
	"fmt"

	"github.com/serulian/compiler/compilergraph"
	"github.com/serulian/compiler/graphs/scopegraph/proto"
	"github.com/serulian/compiler/sourceshape"
)

var _ = fmt.Printf

const ANONYMOUS_REFERENCE = "_"

var ALLOWED_ANONYMOUS = []compilergraph.Predicate{sourceshape.NodeArrowStatementDestination, sourceshape.NodeArrowStatementRejection}

// scopeIdentifierExpression scopes an identifier expression in the SRG.
func (sb *scopeBuilder) scopeIdentifierExpression(node compilergraph.GraphNode, context scopeContext) proto.ScopeInfo {
	name, hasName := node.TryGet(sourceshape.NodeIdentifierExpressionName)
	if !hasName {
		return newScope().Invalid().GetScope()
	}

	if name == ANONYMOUS_REFERENCE {
		// Make sure this node is under an assignment of some kind.
		var found = false
		for _, predicate := range ALLOWED_ANONYMOUS {
			if _, ok := node.TryGetIncomingNode(predicate); ok {
				found = true
				break
			}
		}

		if !found {
			sb.decorateWithError(node, "Anonymous identifier '_' cannot be used as a value")
			return newScope().Invalid().GetScope()
		}

		return newScope().ForAnonymousScope(sb.sg.tdg).GetScope()
	}

	// Check the cache.
	namedScope, found := context.lookupLocalScopeName(name)
	if !found {
		// Lookup the name given, starting at the location of the current node.
		var rerr error
		namedScope, rerr = sb.lookupNamedScope(name, node)
		if rerr != nil {
			sb.decorateWithError(node, "%v", rerr)
			return newScope().Invalid().GetScope()
		}
	}

	// Ensure that the named scope has a valid type.
	if !namedScope.IsValid(context) {
		return newScope().Invalid().GetScope()
	}

	// Warn if we are accessing an assignable value under an async function, as it will be executing
	// in a different context.
	if namedScope.IsAssignable() && namedScope.UnderModule() {
		srgImpl, found := context.getParentContainer(sb.sg.srg)
		if found && srgImpl.ContainingMember().IsAsyncFunction() {
			sb.decorateWithWarning(node, "%v '%v' is defined outside the async function and will therefore be unique for each call to this function", namedScope.Title(), name)
		}
	}

	context.staticDependencyCollector.checkNamedScopeForDependency(namedScope)
	return newScope().ForNamedScope(namedScope, context).GetScope()
}
