// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import "fmt"

// Useful for debugging.
var _ = fmt.Printf

type typeMemberOption int

const (
	typeMemberDeclaration typeMemberOption = iota
	typeMemberDefinition
)

type typeReferenceOption int

const (
	typeReferenceWithVoid typeReferenceOption = iota
	typeReferenceNoVoid
)

type statementBlockOption int

const (
	statementBlockWithTerminator statementBlockOption = iota
	statementBlockWithoutTerminator
)

type matchCaseOption int

const (
	matchCaseWithExpression matchCaseOption = iota
	matchCaseWithoutExpression
)

// consumeTopLevel attempts to consume the top-level constructs of a Serulian source file.
func (p *sourceParser) consumeTopLevel() AstNode {
	rootNode := p.startNode(NodeTypeFile)
	defer p.finishNode()

	// Start at the first token.
	p.consumeToken()

	// Once we've seen a non-import, no further imports are allowed.
	seenNonImport := false

Loop:
	for {
		switch {

		// imports.
		case p.isKeyword("import") || p.isKeyword("from"):
			if seenNonImport {
				p.emitError("Imports must precede all definitions")
				break Loop
			}

			p.currentNode().Connect(NodePredicateChild, p.consumeImport())

		// type definitions.
		case p.isKeyword("class") || p.isKeyword("interface"):
			seenNonImport = true
			p.currentNode().Connect(NodePredicateChild, p.consumeTypeDefinition())
			p.tryConsumeStatementTerminator()

		// functions.
		case p.isKeyword("function"):
			seenNonImport = true
			p.currentNode().Connect(NodePredicateChild, p.consumeFunction(typeMemberDefinition))
			p.tryConsumeStatementTerminator()

		// variables.
		case p.isKeyword("var"):
			seenNonImport = true
			p.currentNode().Connect(NodePredicateChild, p.consumeVar(NodeTypeVariableStatement))
			p.tryConsumeStatementTerminator()

		// EOF.
		case p.isToken(tokenTypeEOF):
			// If we hit the end of the file, then we're done but not because of an expected
			// rule.
			p.emitError("Unexpected EOF at root level: %v", p.currentToken.position)
			break Loop

		case p.isToken(tokenTypeError):
			break Loop

		default:
			p.emitError("Unexpected token at root level: %v", p.currentToken.kind)
			break Loop

		}

		if p.isToken(tokenTypeEOF) {
			break Loop
		}
	}

	return rootNode
}

// consumeImport attempts to consume an import statement.
//
// Supported forms (all must be terminated by \n or EOF):
// from something import foobar
// from something import foobar as barbaz
// import something
// import something as foobar
// import "somestring" as barbaz
func (p *sourceParser) consumeImport() AstNode {
	importNode := p.startNode(NodeTypeImport)
	defer p.finishNode()

	// from ...
	if p.tryConsumeKeyword("from") {
		// Decorate the node with its source.
		token, ok := p.consume(tokenTypeIdentifer, tokenTypeStringLiteral)
		if !ok {
			return importNode
		}

		importNode.Decorate(NodeImportPredicateLocation, p.reportImport(token.value))
		importNode.Decorate(NodeImportPredicateSource, token.value)
		p.consumeImportSource(importNode, NodeImportPredicateSubsource, NodeImportPredicateName, tokenTypeIdentifer)
		return importNode
	}

	p.consumeImportSource(importNode, NodeImportPredicateSource, NodeImportPredicatePackageName, tokenTypeIdentifer, tokenTypeStringLiteral)
	return importNode
}

func (p *sourceParser) consumeImportSource(importNode AstNode, sourcePredicate string, namePredicate string, allowedValues ...tokenType) {
	// import ...
	if !p.consumeKeyword("import") {
		return
	}

	// "something" or something
	token, ok := p.consume(allowedValues...)
	if !ok {
		return
	}

	if sourcePredicate == NodeImportPredicateSource {
		importNode.Decorate(NodeImportPredicateLocation, p.reportImport(token.value))
	}

	importNode.Decorate(sourcePredicate, token.value)

	// as something (optional)
	if p.tryConsumeKeyword("as") {
		named, ok := p.consumeIdentifier()
		if !ok {
			return
		}

		importNode.Decorate(namePredicate, named)
	} else {
		// If the import was a string value, then an 'as' is required.
		if token.kind == tokenTypeStringLiteral {
			p.emitError("Import from SCM URL requires an 'as' clause")
		} else {
			// Otherwise, literal imports receive the name of the package source as their own package name.
			importNode.Decorate(namePredicate, token.value)
		}
	}

	// end of the statement
	p.consumeStatementTerminator()
}

// consumeTypeDefinition attempts to consume a type definition.
func (p *sourceParser) consumeTypeDefinition() AstNode {
	if p.isKeyword("class") {
		return p.consumeClassDefinition()
	} else if p.isKeyword("interface") {
		return p.consumeInterfaceDefinition()
	} else {
		return p.createErrorNode("Expected 'class' or 'interface', Found: %s", p.currentToken.value)
	}
}

// consumeClassDefinition consumes a class definition.
//
// class Identifier { ... }
// class Identifier : BaseClass.Path + AnotherBaseClass.Path { ... }
// class Identifier<Generic> { ... }
// class Identifier<Generic> : BaseClass.Path { ... }
func (p *sourceParser) consumeClassDefinition() AstNode {
	classNode := p.startNode(NodeTypeClass)
	defer p.finishNode()

	// class ...
	if !p.consumeKeyword("class") {
		return classNode
	}

	// Identifier
	className, ok := p.consumeIdentifier()
	if !ok {
		return classNode
	}

	classNode.Decorate(NodeClassPredicateName, className)

	// Generics (optional).
	p.consumeGenerics(classNode, NodeTypeDefinitionGeneric)

	// Inheritance.
	if _, ok := p.tryConsume(tokenTypeColon); ok {
		// Consume identifier paths until we don't find a plus.
		for {
			classNode.Connect(NodeClassPredicateBaseType, p.consumeIdentifierPath())
			if _, ok := p.tryConsume(tokenTypePlus); !ok {
				break
			}
		}
	}

	// Open bracket.
	if _, ok := p.consume(tokenTypeLeftBrace); !ok {
		return classNode
	}

	// Consume class members.
	p.consumeClassMembers(classNode)

	// Close bracket.
	p.consume(tokenTypeRightBrace)

	return classNode
}

// consumeInterfaceDefinition consumes an interface definition.
//
// interface Identifier { ... }
// interface Identifier<Generic> { ... }
func (p *sourceParser) consumeInterfaceDefinition() AstNode {
	interfaceNode := p.startNode(NodeTypeInterface)
	defer p.finishNode()

	// interface ...
	if !p.consumeKeyword("interface") {
		return interfaceNode
	}

	// Identifier
	interfaceName, ok := p.consumeIdentifier()
	if !ok {
		return interfaceNode
	}

	interfaceNode.Decorate(NodeInterfacePredicateName, interfaceName)

	// Generics (optional).
	p.consumeGenerics(interfaceNode, NodeTypeDefinitionGeneric)

	// Open bracket.
	if _, ok := p.consume(tokenTypeLeftBrace); !ok {
		return interfaceNode
	}

	// Consume interface members.
	p.consumeInterfaceMembers(interfaceNode)

	// Close bracket.
	p.consume(tokenTypeRightBrace)
	return interfaceNode
}

// consumeClassMembers consumes the member definitions of a class.
func (p *sourceParser) consumeClassMembers(typeNode AstNode) {
	for {
		switch {
		case p.isKeyword("var"):
			typeNode.Connect(NodeTypeDefinitionMember, p.consumeVar(NodeTypeField))
			p.consumeStatementTerminator()

		case p.isKeyword("function"):
			typeNode.Connect(NodeTypeDefinitionMember, p.consumeFunction(typeMemberDefinition))

		case p.isKeyword("constructor"):
			typeNode.Connect(NodeTypeDefinitionMember, p.consumeConstructor(typeMemberDefinition))

		case p.isKeyword("property"):
			typeNode.Connect(NodeTypeDefinitionMember, p.consumeProperty(typeMemberDefinition))

		case p.isKeyword("operator"):
			typeNode.Connect(NodeTypeDefinitionMember, p.consumeOperator(typeMemberDefinition))

		case p.isToken(tokenTypeRightBrace):
			// End of the class members list
			return

		default:
			p.emitError("Expected class member, found %s", p.currentToken.value)
			return
		}
	}
}

// consumeInterfaceMembers consumes the member definitions of an interface.
func (p *sourceParser) consumeInterfaceMembers(typeNode AstNode) {
	for {
		switch {
		case p.isKeyword("function"):
			typeNode.Connect(NodeTypeDefinitionMember, p.consumeFunction(typeMemberDeclaration))

		case p.isKeyword("constructor"):
			typeNode.Connect(NodeTypeDefinitionMember, p.consumeConstructor(typeMemberDeclaration))

		case p.isKeyword("property"):
			typeNode.Connect(NodeTypeDefinitionMember, p.consumeProperty(typeMemberDeclaration))

		case p.isKeyword("operator"):
			typeNode.Connect(NodeTypeDefinitionMember, p.consumeOperator(typeMemberDeclaration))

		case p.isToken(tokenTypeRightBrace):
			// End of the class members list
			return

		default:
			p.emitError("Expected interface member, found %s", p.currentToken.value)
			p.consumeUntil(tokenTypeNewline, tokenTypeSyntheticSemicolon, tokenTypeEOF)
			return
		}
	}
}

// consumeOperator consumes an operator declaration or definition
//
// Supported forms:
// operator Plus (leftValue SomeType, rightValue SomeType)
func (p *sourceParser) consumeOperator(option typeMemberOption) AstNode {
	operatorNode := p.startNode(NodeTypeOperator)
	defer p.finishNode()

	// operator
	p.consumeKeyword("operator")

	// Operator Name.
	identifier, ok := p.consumeIdentifier()
	if !ok {
		return operatorNode
	}

	operatorNode.Decorate(NodeOperatorName, identifier)

	// Parameters.
	// (
	if _, ok := p.consume(tokenTypeLeftParen); !ok {
		return operatorNode
	}

	// identifier TypeReference (, another)
	for {
		operatorNode.Connect(NodeOperatorParameter, p.consumeParameter())

		if _, ok := p.tryConsume(tokenTypeComma); !ok {
			break
		}
	}

	// )
	if _, ok := p.consume(tokenTypeRightParen); !ok {
		return operatorNode
	}

	// If this is a declaration, then we look for a statement terminator and
	// finish the parse.
	if option == typeMemberDeclaration {
		p.consumeStatementTerminator()
		return operatorNode
	}

	// Otherwise, we need a body.
	operatorNode.Connect(NodeOperatorBody, p.consumeStatementBlock(statementBlockWithTerminator))
	return operatorNode
}

// consumeProperty consumes a property declaration or definition
//
// Supported forms:
// property<SomeType> SomeName
// property<SomeType> SomeName { get }
// property<SomeType> SomeName {
//   get { .. }
//   set { .. }
// }
//
func (p *sourceParser) consumeProperty(option typeMemberOption) AstNode {
	propertyNode := p.startNode(NodeTypeProperty)
	defer p.finishNode()

	// property
	p.consumeKeyword("property")

	// Property type: <Foo>
	if _, ok := p.consume(tokenTypeLessThan); !ok {
		return propertyNode
	}

	propertyNode.Connect(NodePropertyDeclaredType, p.consumeTypeReference(typeReferenceNoVoid))

	if _, ok := p.consume(tokenTypeGreaterThan); !ok {
		return propertyNode
	}

	// Property name.
	identifier, ok := p.consumeIdentifier()
	if !ok {
		return propertyNode
	}

	propertyNode.Decorate(NodePropertyName, identifier)

	// If this is a declaration, then having a brace is optional.
	if option == typeMemberDeclaration {
		// Check for the open brace. If found, then this is the beginning of a
		// read-only declaration.
		if _, ok := p.tryConsume(tokenTypeLeftBrace); !ok {
			p.consumeStatementTerminator()
			return propertyNode
		}

		propertyNode.Decorate(NodePropertyReadOnly, "true")
		if !p.consumeKeyword("get") {
			return propertyNode
		}

		p.consume(tokenTypeRightBrace)
		p.consumeStatementTerminator()
		return propertyNode
	} else {
		// Otherwise, this is a definition. "get" (and optional "set") blocks
		// are required.
		if _, ok := p.consume(tokenTypeLeftBrace); !ok {
			p.consumeStatementTerminator()
			return propertyNode
		}

		// Add the getter (required)
		propertyNode.Connect(NodePropertyGetter, p.consumePropertyBlock("get"))

		// Add the setter (optional)
		if p.isKeyword("set") {
			propertyNode.Connect(NodePropertySetter, p.consumePropertyBlock("set"))
		} else {
			propertyNode.Decorate(NodePropertyReadOnly, "true")
		}

		p.consume(tokenTypeRightBrace)
		p.consumeStatementTerminator()
		return propertyNode
	}
}

// consumePropertyBlock consumes a get or set block for a property definition
func (p *sourceParser) consumePropertyBlock(keyword string) AstNode {
	blockNode := p.startNode(NodeTypePropertyBlock)
	blockNode.Decorate(NodePropertyBlockType, keyword)
	defer p.finishNode()

	// get or set
	if !p.consumeKeyword(keyword) {
		return blockNode
	}

	// Statement block.
	blockNode.Connect(NodePropertyBlockBody, p.consumeStatementBlock(statementBlockWithTerminator))
	return blockNode
}

// consumeConstructor consumes a constructor declaration or definition
//
// Supported forms:
// constructor SomeName()
// constructor SomeName<SomeGeneric>()
// constructor SomeName(someArg int)
//
func (p *sourceParser) consumeConstructor(option typeMemberOption) AstNode {
	constructorNode := p.startNode(NodeTypeConstructor)
	defer p.finishNode()

	// constructor
	p.consumeKeyword("constructor")

	// Name.
	identifier, ok := p.consumeIdentifier()
	if !ok {
		return constructorNode
	}

	constructorNode.Decorate(NodeConstructorName, identifier)

	// Generics (optional).
	p.consumeGenerics(constructorNode, NodeConstructorGeneric)

	// Parameters.
	// (
	if _, ok := p.consume(tokenTypeLeftParen); !ok {
		return constructorNode
	}

	// identifier TypeReference (, another)
	for {
		constructorNode.Connect(NodeConstructorParameter, p.consumeParameter())

		if _, ok := p.tryConsume(tokenTypeComma); !ok {
			break
		}
	}

	// )
	if _, ok := p.consume(tokenTypeRightParen); !ok {
		return constructorNode
	}

	// If this is a declaration, then we look for a statement terminator and
	// finish the parse.
	if option == typeMemberDeclaration {
		p.consumeStatementTerminator()
		return constructorNode
	}

	// Otherwise, we need a body.
	constructorNode.Connect(NodeConstructorBody, p.consumeStatementBlock(statementBlockWithTerminator))
	return constructorNode
}

// consumeFunction consumes a function declaration or definition
//
// Supported forms:
// function<ReturnType> SomeName()
// function<ReturnType> SomeName<SomeGeneric>()
//
func (p *sourceParser) consumeFunction(option typeMemberOption) AstNode {
	functionNode := p.startNode(NodeTypeFunction)
	defer p.finishNode()

	// function
	p.consumeKeyword("function")

	// return type: <Foo>
	if _, ok := p.consume(tokenTypeLessThan); !ok {
		return functionNode
	}

	functionNode.Connect(NodeFunctionReturnType, p.consumeTypeReference(typeReferenceWithVoid))

	if _, ok := p.consume(tokenTypeGreaterThan); !ok {
		return functionNode
	}

	// Name.
	identifier, ok := p.consumeIdentifier()
	if !ok {
		return functionNode
	}

	functionNode.Decorate(NodeFunctionName, identifier)

	// Generics (optional).
	p.consumeGenerics(functionNode, NodeFunctionGeneric)

	// Parameters.
	// (
	if _, ok := p.consume(tokenTypeLeftParen); !ok {
		return functionNode
	}

	// identifier TypeReference (, another)
	for {
		if !p.isToken(tokenTypeIdentifer) {
			break
		}

		functionNode.Connect(NodeFunctionParameter, p.consumeParameter())

		if _, ok := p.tryConsume(tokenTypeComma); !ok {
			break
		}
	}

	// )
	if _, ok := p.consume(tokenTypeRightParen); !ok {
		return functionNode
	}

	// If this is a declaration, then we look for a statement terminator and
	// finish the parse.
	if option == typeMemberDeclaration {
		p.consumeStatementTerminator()
		return functionNode
	}

	// Otherwise, we need a function body.
	functionNode.Connect(NodeFunctionBody, p.consumeStatementBlock(statementBlockWithTerminator))
	return functionNode
}

// consumeParameter consumes a function or other type member parameter definition
func (p *sourceParser) consumeParameter() AstNode {
	parameterNode := p.startNode(NodeTypeParameter)
	defer p.finishNode()

	// Parameter name.
	identifier, ok := p.consumeIdentifier()
	if !ok {
		return parameterNode
	}

	parameterNode.Decorate(NodeParameterType, identifier)

	// Parameter type.
	parameterNode.Connect(NodeParameterType, p.consumeTypeReference(typeReferenceNoVoid))
	return parameterNode
}

// typeReferenceMap contains a map from tokenType to associated node type for the
// specialized type reference modifiers (nullable, stream, etc).
var typeReferenceMap = map[tokenType]NodeType{
	tokenTypeTimes:        NodeTypeStream,
	tokenTypeQuestionMark: NodeTypeNullable,
}

// consumeTypeReference consumes a type reference
func (p *sourceParser) consumeTypeReference(option typeReferenceOption) AstNode {
	// If void is allowed, check for it first.
	if option == typeReferenceWithVoid && p.isKeyword("void") {
		typeRefNode := p.startNode(NodeTypeTypeReference)
		voidNode := p.startNode(NodeTypeVoid)
		p.consumeKeyword("void")
		p.finishNode()
		p.finishNode()

		typeRefNode.Connect(NodeTypeReferencePath, voidNode)
		return typeRefNode
	}

	// Otherwise, left recursively build a type reference.
	rightNodeBuilder := func(leftNode AstNode, operatorToken lexeme) (AstNode, bool) {
		nodeType, ok := typeReferenceMap[operatorToken.kind]
		if !ok {
			panic(fmt.Sprintf("Unknown type reference modifier: %v", operatorToken.kind))
		}

		// Create the node representing the wrapped type reference.
		parentNode := p.createNode(nodeType)
		parentNode.Connect(NodeTypeReferenceInnerType, leftNode)
		return parentNode, true
	}

	found, _ := p.performLeftRecursiveParsing(p.consumeSimpleTypeReference, rightNodeBuilder,
		tokenTypeTimes, tokenTypeQuestionMark)
	return found
}

// consumeSimpleTypeReference consumes a type reference that cannot be void, nullable
// or streamable.
func (p *sourceParser) consumeSimpleTypeReference() (AstNode, bool) {
	typeRefNode := p.startNode(NodeTypeTypeReference)
	defer p.finishNode()

	// Identifier path.
	typeRefNode.Connect(NodeTypeReferencePath, p.consumeIdentifierPath())

	// Optional generics:
	// <
	if _, ok := p.tryConsume(tokenTypeLessThan); !ok {
		return typeRefNode, true
	}

	// Foo, Bar, Baz
	for {
		typeRefNode.Connect(NodeTypeReferenceGeneric, p.consumeTypeReference(typeReferenceNoVoid))

		if _, ok := p.tryConsume(tokenTypeComma); !ok {
			break
		}
	}

	// >
	p.consume(tokenTypeGreaterThan)
	return typeRefNode, true
}

// consumeGenerics attempts to consume generic definitions on a type or function, decorating
// that type node.
//
// Supported Forms:
// <Foo>
// <Foo, Bar>
// <Foo : SomePath>
// <Foo : SomePath, Bar>
func (p *sourceParser) consumeGenerics(parentNode AstNode, predicate string) {
	// <
	if _, ok := p.tryConsume(tokenTypeLessThan); !ok {
		return
	}

	for {
		parentNode.Connect(predicate, p.consumeGeneric())

		// ,
		if _, ok := p.tryConsume(tokenTypeComma); !ok {
			break
		}
	}

	// >
	p.consume(tokenTypeGreaterThan)
}

// consumeGeneric consumes a generic definition found on a type node.
//
// Supported Forms:
// Foo
// Foo : Bar
func (p *sourceParser) consumeGeneric() AstNode {
	genericNode := p.startNode(NodeTypeGeneric)
	defer p.finishNode()

	// Generic name.
	genericName, ok := p.consumeIdentifier()
	if !ok {
		return genericNode
	}

	genericNode.Decorate(NodeGenericPredicateName, genericName)

	// Optional: subtype.
	if _, ok := p.tryConsume(tokenTypeColon); !ok {
		return genericNode
	}

	genericNode.Connect(NodeGenericSubtype, p.consumeIdentifierPath())
	return genericNode
}

// consumeIdentifierPath consumes a path consisting of one (or more identifies)
//
// Supported Forms:
// foo
// foo(.bar)*
func (p *sourceParser) consumeIdentifierPath() AstNode {
	identifierPath := p.startNode(NodeTypeIdentifierPath)
	defer p.finishNode()

	var currentNode AstNode
	for {
		nextNode := p.consumeIdentifierAccess()
		if currentNode != nil {
			nextNode.Connect(NodeIdentifierAccessSource, currentNode)
		}

		currentNode = nextNode

		// Check for additional steps.
		if _, ok := p.tryConsume(tokenTypeDotAccessOperator); !ok {
			break
		}
	}

	identifierPath.Connect(NodeIdentifierPathRoot, currentNode)
	return identifierPath
}

// consumeIdentifierAccess consumes an identifier and returns an IdentifierAccessNode.
func (p *sourceParser) consumeIdentifierAccess() AstNode {
	identifierAccessNode := p.startNode(NodeTypeIdentifierAccess)
	defer p.finishNode()

	// Consume the next step in the path.
	identifier, ok := p.consumeIdentifier()
	if !ok {
		return identifierAccessNode
	}

	identifierAccessNode.Decorate(NodeIdentifierAccessName, identifier)
	return identifierAccessNode
}

// consumeStatementBlock consumes a block of statements
//
// Form:
// { ... statements ... }
func (p *sourceParser) consumeStatementBlock(option statementBlockOption) AstNode {
	statementBlockNode := p.startNode(NodeTypeStatementBlock)
	defer p.finishNode()

	// Consume the start of the block: {
	if _, ok := p.consume(tokenTypeLeftBrace); !ok {
		return statementBlockNode
	}

	// Consume statements.
	for {
		// Check for a label on the statement.
		var statementLabel string

		if p.isToken(tokenTypeIdentifer) && p.isNextToken(tokenTypeColon) {
			statementLabel, _ = p.consumeIdentifier()
			p.consume(tokenTypeColon)
		}

		// Try to consume a statement.
		statementNode, ok := p.tryConsumeStatement()
		if !ok {
			break
		}

		// Add the label to the statement (if any).
		if statementLabel != "" {
			statementNode.Decorate(NodeStatementLabel, statementLabel)
		}

		// Connect the statement to the block.
		statementBlockNode.Connect(NodeStatementBlockStatement, statementNode)

		// Consume the terminator for the statement.
		if p.isToken(tokenTypeRightBrace) {
			break
		}

		p.consumeStatementTerminator()
	}

	// Consume the end of the block: }
	p.consume(tokenTypeRightBrace)
	if option == statementBlockWithTerminator {
		p.consumeStatementTerminator()
	}

	return statementBlockNode
}

// tryConsumeStatement attempts to consume a statement.
func (p *sourceParser) tryConsumeStatement() (AstNode, bool) {
	switch {
	// Match statement.
	case p.isKeyword("match"):
		return p.consumeMatchStatement(), true

	// With statement.
	case p.isKeyword("with"):
		return p.consumeWithStatement(), true

	// For statement.
	case p.isKeyword("for"):
		return p.consumeForStatement(), true

	// Var statement.
	case p.isKeyword("var"):
		return p.consumeVar(NodeTypeVariableStatement), true

	// If statement.
	case p.isKeyword("if"):
		return p.consumeIfStatement(), true

	// Return statement.
	case p.isKeyword("return"):
		return p.consumeReturnStatement(), true

	// Break statement.
	case p.isKeyword("break"):
		return p.consumeJumpStatement("break", NodeTypeBreakStatement, NodeBreakStatementLabel), true

	// Continue statement.
	case p.isKeyword("continue"):
		return p.consumeJumpStatement("continue", NodeTypeContinueStatement, NodeContinueStatementLabel), true

	default:
		// Look for an assignment statement.
		if assignNode, ok := p.tryConsumeAssignStatement(); ok {
			return assignNode, true
		}

		// Look for an expression as a statement.
		if exprNode, ok := p.tryConsumeExpression(); ok {
			return exprNode, true
		}

		return nil, false
	}
}

// tryConsumeAssignStatement attempts to consume an assignment statement.
//
// Forms:
// a = expression
// a, b = expression
func (p *sourceParser) tryConsumeAssignStatement() (AstNode, bool) {
	// To determine if we have an assignment statement, we need to perform
	// a non-insignificant amount of lookahead, as this form can be mistaken for
	// expressions with ease:
	if !p.lookaheadAssignStatement() {
		return nil, false
	}

	assignNode := p.startNode(NodeTypeAssignStatement)
	defer p.finishNode()

	// Consume the identifiers.
	for {
		assignNode.Connect(NodeAssignStatementName, p.consumeIdentifierExpression())

		if _, ok := p.tryConsume(tokenTypeComma); !ok {
			break
		}
	}

	p.consume(tokenTypeEquals)
	assignNode.Connect(NodeAssignStatementValue, p.consumeExpression())
	return assignNode, true
}

// lookaheadAssignStatement determines whether there is an assignment statement
// at the current head of the lexer stream.
func (p *sourceParser) lookaheadAssignStatement() bool {
	t := p.newLookaheadTracker()

	if _, ok := t.matchToken(tokenTypeIdentifer); !ok {
		return false
	}

	if _, ok := t.matchToken(tokenTypeEquals); !ok {
		for {
			if _, ok := t.matchToken(tokenTypeComma); !ok {
				return false
			}

			if _, ok := t.matchToken(tokenTypeIdentifer); !ok {
				return false
			}

			if _, ok := t.matchToken(tokenTypeEquals); ok {
				break
			}
		}
	}

	return true
}

// consumeMatchStatement consumes a match statement.
//
// Forms:
// match somExpr {
//   case someExpr:
//      statements
//
//   case anotherExpr, secondExpr:
//      statements
//
//   default:
//      statements
// }
//
// match {
//   case someExpr:
//      statements
//
//   case anotherExpr:
//      statements
//
//   default:
//      statements
// }
func (p *sourceParser) consumeMatchStatement() AstNode {
	matchNode := p.startNode(NodeTypeMatchStatement)
	defer p.finishNode()

	// match
	p.consumeKeyword("match")

	// Consume a match expression (if any).
	if expression, ok := p.tryConsumeExpression(); ok {
		matchNode.Connect(NodeMatchStatementExpression, expression)
	}

	// Consume the opening of the block.
	if _, ok := p.consume(tokenTypeLeftBrace); !ok {
		return matchNode
	}

	// Consume one (or more) case statements.
	for {
		caseNode, ok := p.tryConsumeMatchCase("case", matchCaseWithExpression)
		if !ok {
			break
		}
		matchNode.Connect(NodeMatchStatementCase, caseNode)
	}

	// Consume a default statement.
	if defaultCaseNode, ok := p.tryConsumeMatchCase("default", matchCaseWithoutExpression); ok {
		matchNode.Connect(NodeMatchStatementCase, defaultCaseNode)
	}

	// Consume the closing of the block.
	if _, ok := p.consume(tokenTypeRightBrace); !ok {
		return matchNode
	}

	return matchNode
}

// tryConsumeMatchCase tries to consume a case block under a match node
// with the given keyword.
func (p *sourceParser) tryConsumeMatchCase(keyword string, option matchCaseOption) (AstNode, bool) {
	// keyword
	if !p.tryConsumeKeyword(keyword) {
		return nil, false
	}

	// Create the case node.
	caseNode := p.startNode(NodeTypeMatchStatementCase)
	defer p.finishNode()

	if option == matchCaseWithExpression {
		caseNode.Connect(NodeMatchStatementCaseExpression, p.consumeExpression())
	}

	// Colon after the expression or keyword.
	if _, ok := p.consume(tokenTypeColon); !ok {
		return caseNode, true
	}

	// Consume one (or more) statements, followed by statement terminators.
	for {
		statementNode, ok := p.tryConsumeStatement()
		if !ok {
			break
		}

		caseNode.Connect(NodeMatchStatementCaseStatement, statementNode)

		if _, ok := p.consumeStatementTerminator(); !ok {
			return caseNode, true
		}
	}

	return caseNode, true
}

// consumeWithStatement consumes a with statement.
//
// Forms:
// with someExpr {}
// with someExpr as someIdentifier {}
func (p *sourceParser) consumeWithStatement() AstNode {
	withNode := p.startNode(NodeTypeWithStatement)
	defer p.finishNode()

	// with
	p.consumeKeyword("with")

	// Scoped expression.
	withNode.Connect(NodeWithStatementExpression, p.consumeExpression())

	// Optional: 'as' and then an identifier.
	if p.tryConsumeKeyword("as") {
		identifier, ok := p.consumeIdentifier()
		if !ok {
			return withNode
		}

		withNode.Decorate(NodeWithStatementExpressionName, identifier)
	}

	// Consume the statement block.
	withNode.Connect(NodeWithStatementBlock, p.consumeStatementBlock(statementBlockWithoutTerminator))
	return withNode
}

// consumeForStatement consumes a loop statement.
//
// Forms:
// for {}
// for someExpr {}
// for varName in someExpr {}
func (p *sourceParser) consumeForStatement() AstNode {
	forNode := p.startNode(NodeTypeLoopStatement)
	defer p.finishNode()

	// for
	p.consumeKeyword("for")

	// If the next two tokens are an identifier and the keyword "in",
	// then we have a variable declaration of the for loop.
	if p.isToken(tokenTypeIdentifer) && p.isNextKeyword("in") {
		forVariableIdentifier, _ := p.consumeIdentifier()
		forNode.Decorate(NodeLoopStatementVariableName, forVariableIdentifier)
		p.consumeKeyword("in")
	}

	// Consume the expression (if any).
	if expression, ok := p.tryConsumeExpression(); ok {
		forNode.Connect(NodeLoopStatementExpression, expression)
	}

	forNode.Connect(NodeLoopStatementBlock, p.consumeStatementBlock(statementBlockWithoutTerminator))
	return forNode
}

// consumeVar consumes a variable field or statement.
//
// Forms:
// var<SomeType> someName
// var<SomeType> someName = someExpr
// var someName = someExpr
func (p *sourceParser) consumeVar(nodeType NodeType) AstNode {
	variableNode := p.startNode(nodeType)
	defer p.finishNode()

	// var
	p.consumeKeyword("var")

	// Type declaration (optional if there is an init expression)
	var hasType bool
	if _, ok := p.tryConsume(tokenTypeLessThan); ok {
		variableNode.Connect(NodeVariableStatementDeclaredType, p.consumeTypeReference(typeReferenceNoVoid))

		if _, ok := p.consume(tokenTypeGreaterThan); !ok {
			return variableNode
		}

		hasType = true
	}

	// Name.
	identifier, ok := p.consumeIdentifier()
	if !ok {
		return variableNode
	}

	variableNode.Decorate(NodeVariableStatementName, identifier)

	// Initializer expression. Optional if a type given, otherwise required.
	if !hasType && !p.isToken(tokenTypeEquals) {
		p.emitError("An initializer is required for variable %s, as it has no declared type", identifier)
	}

	if _, ok := p.tryConsume(tokenTypeEquals); ok {
		variableNode.Connect(NodeVariableStatementExpression, p.consumeExpression())
	}

	return variableNode
}

// consumeIfStatement consumes a conditional statement.
//
// Forms:
// if someExpr { ... }
// if someExpr { ... } else { ... }
// if someExpr { ... } else if { ... }
func (p *sourceParser) consumeIfStatement() AstNode {
	conditionalNode := p.startNode(NodeTypeConditionalStatement)
	defer p.finishNode()

	// if
	p.consumeKeyword("if")

	// Expression.
	conditionalNode.Connect(NodeConditionalStatementConditional, p.consumeExpression())

	// Statement block.
	conditionalNode.Connect(NodeConditionalStatementBlock, p.consumeStatementBlock(statementBlockWithoutTerminator))

	// Optional 'else'.
	if !p.tryConsumeKeyword("else") {
		return conditionalNode
	}

	// After an 'else' can be either another if statement OR a statement block.
	if p.isKeyword("if") {
		conditionalNode.Connect(NodeConditionalStatementElseClause, p.consumeIfStatement())
	} else {
		conditionalNode.Connect(NodeConditionalStatementElseClause, p.consumeStatementBlock(statementBlockWithoutTerminator))
	}

	return conditionalNode
}

// consumeReturnStatement consumes a return statement.
//
// Forms:
// return
// return someExpr
func (p *sourceParser) consumeReturnStatement() AstNode {
	returnNode := p.startNode(NodeTypeReturnStatement)
	defer p.finishNode()

	// return
	p.consumeKeyword("return")

	// Check for an expression following the return.
	if p.isStatementTerminator() {
		return returnNode
	}

	returnNode.Connect(NodeReturnStatementValue, p.consumeExpression())
	return returnNode
}

// consumeJumpStatement consumes a statement that can jump flow, such
// as break or continue.
//
// Forms:
// break
// continue
// continue SomeLabel
func (p *sourceParser) consumeJumpStatement(keyword string, nodeType NodeType, labelPredicate string) AstNode {
	jumpNode := p.startNode(nodeType)
	defer p.finishNode()

	// Keyword.
	p.consumeKeyword(keyword)

	// Check for a label.
	if labelName, ok := p.tryConsumeIdentifier(); ok {
		jumpNode.Decorate(labelPredicate, labelName)
	}

	return jumpNode
}

// consumeExpression consumes an expression.
func (p *sourceParser) consumeExpression() AstNode {
	if exprNode, ok := p.tryConsumeExpression(); ok {
		return exprNode
	}

	return p.createErrorNode("Unsupported expression type!")
}

// tryConsumeExpression attempts to consume an expression. If an expression
// coult not be found, returns false.
func (p *sourceParser) tryConsumeExpression() (AstNode, bool) {
	return p.oneOf(p.tryConsumeLambdaExpression, p.tryConsumeAwaitExpression, p.tryConsumeArrowExpression)
}

// tryConsumeLambdaExpression tries to consume a lambda expression of one of the following forms:
// (arg1, arg2) => expression
// function<ReturnType> (arg1 type, arg2 type) { ... }
func (p *sourceParser) tryConsumeLambdaExpression() (AstNode, bool) {
	// Check for the function keyword. If found, we have a full definition lambda function.
	if p.isKeyword("function") {
		return p.consumeFullLambdaExpression(), true
	}

	// Otherwise, we look for an inline lambda expression. To do so, we need to perform
	// a non-insignificant amount of lookahead, as this form can be mistaken for other
	// expressions with ease:
	//
	// Forms:
	// () => expression
	// (arg1) => expression
	// (arg1, arg2) => expression
	if !p.lookaheadLambdaExpr() {
		return nil, false
	}

	// If we've reached this point, we've found a lambda expression and can start properly
	// consuming it.
	lambdaNode := p.startNode(NodeTypeLambdaExpression)
	defer p.finishNode()

	// (
	p.consume(tokenTypeLeftParen)

	// Optional: arguments.
	if !p.isToken(tokenTypeRightParen) {
		for {
			lambdaNode.Connect(NodeLambdaExpressionParameter, p.consumeIdentifierExpression())
			if _, ok := p.tryConsume(tokenTypeComma); !ok {
				break
			}
		}
	}

	// )
	p.consume(tokenTypeRightParen)

	// =>
	p.consume(tokenTypeLambdaArrowOperator)

	// expression.
	lambdaNode.Connect(NodeLambdaExpressionChildExpr, p.consumeExpression())
	return lambdaNode, true
}

// lookaheadLambdaExpr performs lookahead to determine if there is a lambda expression
// at the head of the lexer stream.
func (p *sourceParser) lookaheadLambdaExpr() bool {
	t := p.newLookaheadTracker()

	// (
	if _, ok := t.matchToken(tokenTypeLeftParen); !ok {
		return false
	}

	// argument identifier or close paren.
	if _, ok := t.matchToken(tokenTypeRightParen); !ok {
		for {
			// argument identifier
			if _, ok := t.matchToken(tokenTypeIdentifer); !ok {
				return false
			}

			// comma
			if _, ok := t.matchToken(tokenTypeComma); !ok {
				break
			}
		}

		// )
		if _, ok := t.matchToken(tokenTypeRightParen); !ok {
			return false
		}
	}

	// =>
	if _, ok := t.matchToken(tokenTypeLambdaArrowOperator); !ok {
		return false
	}

	return true
}

// consumeFullLambdaExpression consumes a fully-defined lambda function.
//
// Form:
// function<ReturnType> (arg1 type, arg2 type) { ... }
func (p *sourceParser) consumeFullLambdaExpression() AstNode {
	funcNode := p.startNode(NodeTypeLambdaExpression)
	defer p.finishNode()

	// function
	p.consumeKeyword("function")

	// return type (optional)
	if _, ok := p.tryConsume(tokenTypeLessThan); ok {
		funcNode.Connect(NodeLambdaExpressionReturnType, p.consumeTypeReference(typeReferenceWithVoid))
		p.consume(tokenTypeGreaterThan)
	}

	// Parameter list.
	if _, ok := p.consume(tokenTypeLeftParen); !ok {
		return funcNode
	}

	if !p.isToken(tokenTypeRightParen) {
		for {
			funcNode.Connect(NodeLambdaExpressionParameter, p.consumeParameter())
			if _, ok := p.tryConsume(tokenTypeComma); !ok {
				break
			}
		}
	}

	if _, ok := p.consume(tokenTypeRightParen); !ok {
		return funcNode
	}

	// Block.
	funcNode.Connect(NodeLambdaExpressionBlock, p.consumeStatementBlock(statementBlockWithoutTerminator))
	return funcNode
}

// tryConsumeNonArrowExpression tries to consume an expression that is found under an arrow.
func (p *sourceParser) tryConsumeNonArrowExpression() (AstNode, bool) {
	// TODO(jschorr): Cache this!
	binaryParser := p.buildBinaryOperatorExpressionFnTree(
		// Nullable operators.
		boe{tokenTypeNullOrValueOperator, NodeNullComparisonExpression},

		// Comparison operators.
		boe{tokenTypeEqualsEquals, NodeComparisonEqualsExpression},
		boe{tokenTypeNotEquals, NodeComparisonNotEqualsExpression},

		boe{tokenTypeLTE, NodeComparisonLTEExpression},
		boe{tokenTypeGTE, NodeComparisonGTEExpression},

		boe{tokenTypeLessThan, NodeComparisonLTExpression},
		boe{tokenTypeGreaterThan, NodeComparisonGTExpression},

		// Boolean operators.
		boe{tokenTypeBooleanOr, NodeBooleanOrExpression},
		boe{tokenTypeBooleanAnd, NodeBooleanAndExpression},

		// Bitwise operators.
		boe{tokenTypePipe, NodeBitwiseOrExpression},
		boe{tokenTypeAnd, NodeBitwiseAndExpression},
		boe{tokenTypeXor, NodeBitwiseXorExpression},
		boe{tokenTypeBitwiseShiftLeft, NodeBitwiseShiftLeftExpression},

		// TODO(jschorr): Find a solution for the >> issue.
		//boe{tokenTypeGreaterThan, NodeBitwiseShiftRightExpression},

		// Numeric operators.
		boe{tokenTypePlus, NodeBinaryAddExpression},
		boe{tokenTypeMinus, NodeBinarySubtractExpression},
		boe{tokenTypeModulo, NodeBinaryModuloExpression},
		boe{tokenTypeTimes, NodeBinaryMultiplyExpression},
		boe{tokenTypeDiv, NodeBinaryDivideExpression},

		// Stream operator.
		boe{tokenTypeEllipsis, NodeDefineRangeExpression})

	return binaryParser()
}

func (p *sourceParser) consumeNonArrowExpression() AstNode {
	if node, ok := p.tryConsumeNonArrowExpression(); ok {
		return node
	}

	p.emitError("Expected expression, found: %s", p.currentToken.kind)
	return nil
}

// tryConsumeAwaitExpression tries to consume an await expression.
//
// Form: <- a
func (p *sourceParser) tryConsumeAwaitExpression() (AstNode, bool) {
	if _, ok := p.tryConsume(tokenTypeArrowPortOperator); !ok {
		return nil, false
	}

	exprNode := p.startNode(NodeTypeAwaitExpression)
	defer p.finishNode()

	exprNode.Connect(NodeAwaitExpressionSource, p.consumeNonArrowExpression())
	return exprNode, true
}

// tryConsumeArrowExpression tries to consumes an arrow expression.
//
// Form: a <- b
func (p *sourceParser) tryConsumeArrowExpression() (AstNode, bool) {
	exprNode := p.startNode(NodeTypeArrowExpression)
	defer p.finishNode()

	destinationNode, ok := p.tryConsumeNonArrowExpression()
	if !ok {
		return nil, false
	}

	if _, ok := p.tryConsume(tokenTypeArrowPortOperator); !ok {
		return destinationNode, true
	}

	exprNode.Connect(NodeArrowExpressionDestination, destinationNode)
	exprNode.Connect(NodeArrowExpressionSource, p.consumeNonArrowExpression())
	return exprNode, true
}

// boe represents information a binary operator token and its associated node type.
type boe struct {
	// The token representing the binary expression's operator.
	binaryOperatorToken tokenType

	// The type of node to create for this expression.
	binaryExpressionNodeType NodeType
}

// buildBinaryOperatorExpressionFnTree builds a tree of functions to try to consume a set of binary
// operator expressions.
func (p *sourceParser) buildBinaryOperatorExpressionFnTree(operators ...boe) tryParserFn {
	// Start with a base expression function.
	var currentParseFn tryParserFn
	currentParseFn = p.tryConsumeCallAccessExpression

	for i := range operators {
		// Note: We have to reverse this to ensure we have proper precedence.
		currentParseFn = func(operatorInfo boe, currentFn tryParserFn) tryParserFn {
			return (func() (AstNode, bool) {
				return p.tryConsumeBinaryExpression(currentFn, operatorInfo.binaryOperatorToken, operatorInfo.binaryExpressionNodeType)
			})
		}(operators[len(operators)-i-1], currentParseFn)
	}

	return currentParseFn
}

// tryConsumeBinaryExpression tries to consume a binary operator expression.
func (p *sourceParser) tryConsumeBinaryExpression(subTryExprFn tryParserFn, binaryTokenType tokenType, nodeType NodeType) (AstNode, bool) {
	rightNodeBuilder := func(leftNode AstNode, operatorToken lexeme) (AstNode, bool) {
		rightNode, ok := subTryExprFn()
		if !ok {
			return nil, false
		}

		// Create the expression node representing the binary expression.
		exprNode := p.createNode(nodeType)
		exprNode.Connect(NodeBinaryExpressionLeftExpr, leftNode)
		exprNode.Connect(NodeBinaryExpressionRightExpr, rightNode)
		return exprNode, true
	}

	return p.performLeftRecursiveParsing(subTryExprFn, rightNodeBuilder, binaryTokenType)
}

// memberAccessExprMap contains a map from the member access token types to their
// associated node types.
var memberAccessExprMap = map[tokenType]NodeType{
	tokenTypeDotAccessOperator:     NodeMemberAccessExpression,
	tokenTypeArrowAccessOperator:   NodeDynamicMemberAccessExpression,
	tokenTypeNullDotAccessOperator: NodeNullableMemberAccessExpression,
	tokenTypeStreamAccessOperator:  NodeStreamMemberAccessExpression,
}

// tryConsumeCallAccessExpression attempts to consume call expressions (function calls or slices) or
// member accesses (dot, nullable, stream, etc.
func (p *sourceParser) tryConsumeCallAccessExpression() (AstNode, bool) {
	rightNodeBuilder := func(leftNode AstNode, operatorToken lexeme) (AstNode, bool) {
		// If this is a member access of some kind, we next look for an identifier.
		if operatorNodeType, ok := memberAccessExprMap[operatorToken.kind]; ok {
			// Consume an identifier.
			identifier, ok := p.consumeIdentifier()
			if !ok {
				return nil, false
			}

			// Create the expression node.
			exprNode := p.createNode(operatorNodeType)
			exprNode.Connect(NodeMemberAccessChildExpr, leftNode)
			exprNode.Decorate(NodeMemberAccessIdentifier, identifier)
			return exprNode, true
		}

		// Handle the other kinds of operators: casts, function calls, slices.
		switch operatorToken.kind {
		case tokenTypeDotCastStart:
			// Cast: a.(b)
			typeReferenceNode := p.consumeTypeReference(typeReferenceNoVoid)

			// Consume the close parens.
			p.consume(tokenTypeRightParen)

			exprNode := p.createNode(NodeCastExpression)
			exprNode.Connect(NodeCastExpressionType, typeReferenceNode)
			exprNode.Connect(NodeCastExpressionChildExpr, leftNode)
			return exprNode, true

		case tokenTypeLeftParen:
			// Function call: a(b)
			exprNode := p.createNode(NodeFunctionCallExpression)
			exprNode.Connect(NodeFunctionCallExpressionChildExpr, leftNode)

			// Consume zero (or more) parameters.
			if !p.isToken(tokenTypeRightParen) {
				for {
					// Consume an expression.
					exprNode.Connect(NodeFunctionCallArgument, p.consumeExpression())

					// Consume an (optional) comma.
					if _, ok := p.tryConsume(tokenTypeComma); !ok {
						break
					}
				}
			}

			// Consume the close parens.
			p.consume(tokenTypeRightParen)
			return exprNode, true

		case tokenTypeLeftBracket:
			// Slice/Indexer:
			// a[b]
			// a[b:c]
			// a[:b]
			// a[b:]
			exprNode := p.createNode(NodeSliceExpression)
			exprNode.Connect(NodeSliceExpressionChildExpr, leftNode)

			// Check for a colon token. If found, this is a right-side-only
			// slice.
			if _, ok := p.tryConsume(tokenTypeColon); ok {
				exprNode.Connect(NodeSliceExpressionRightIndex, p.consumeExpression())
				p.consume(tokenTypeRightBracket)
				return exprNode, true
			}

			// Otherwise, look for the left or index expression.
			indexNode := p.consumeExpression()

			// If we find a right bracket after the expression, then we're done.
			if _, ok := p.tryConsume(tokenTypeRightBracket); ok {
				exprNode.Connect(NodeSliceExpressionIndex, indexNode)
				return exprNode, true
			}

			// Otherwise, a colon is required.
			if _, ok := p.tryConsume(tokenTypeColon); !ok {
				p.emitError("Expected colon in slice, found: %v", p.currentToken.value)
				return exprNode, true
			}

			// Consume the (optional right expression).
			if _, ok := p.tryConsume(tokenTypeRightBracket); ok {
				exprNode.Connect(NodeSliceExpressionLeftIndex, indexNode)
				return exprNode, true
			}

			exprNode.Connect(NodeSliceExpressionLeftIndex, indexNode)
			exprNode.Connect(NodeSliceExpressionRightIndex, p.consumeExpression())
			p.consume(tokenTypeRightBracket)
			return exprNode, true
		}

		return nil, false
	}

	return p.performLeftRecursiveParsing(p.tryConsumeBaseExpression, rightNodeBuilder,
		tokenTypeDotCastStart,
		tokenTypeLeftParen,
		tokenTypeLeftBracket,
		tokenTypeDotAccessOperator,
		tokenTypeArrowAccessOperator,
		tokenTypeNullDotAccessOperator,
		tokenTypeStreamAccessOperator)
}

// tryConsumeBaseExpression attempts to consume base expressions (literals, identifiers, parenthesis).
func (p *sourceParser) tryConsumeBaseExpression() (AstNode, bool) {
	switch {

	// List expression.
	case p.isToken(tokenTypeLeftBracket):
		return p.consumeListExpression(), true

	// Unary: ~
	case p.isToken(tokenTypeTilde):
		p.consume(tokenTypeTilde)

		bitNode := p.startNode(NodeBitwiseNotExpression)
		defer p.finishNode()
		bitNode.Connect(NodeUnaryExpressionChildExpr, p.consumeExpression())
		return bitNode, true

	// Unary: !
	case p.isToken(tokenTypeNot):
		p.consume(tokenTypeNot)

		notNode := p.startNode(NodeBooleanNotExpression)
		defer p.finishNode()
		notNode.Connect(NodeUnaryExpressionChildExpr, p.consumeExpression())
		return notNode, true

	// Nested expression.
	case p.isToken(tokenTypeLeftParen):
		p.consume(tokenTypeLeftParen)
		exprNode := p.consumeExpression()
		p.consume(tokenTypeRightParen)
		return exprNode, true

	// Numeric literal.
	case p.isToken(tokenTypeNumericLiteral):
		literalNode := p.startNode(NodeNumericLiteralExpression)
		defer p.finishNode()

		token, _ := p.consume(tokenTypeNumericLiteral)
		literalNode.Decorate(NodeNumericLiteralExpressionValue, token.value)

		return literalNode, true

	// Boolean literal.
	case p.isToken(tokenTypeBooleanLiteral):
		literalNode := p.startNode(NodeBooleanLiteralExpression)
		defer p.finishNode()

		token, _ := p.consume(tokenTypeBooleanLiteral)
		literalNode.Decorate(NodeBooleanLiteralExpressionValue, token.value)

		return literalNode, true

	// String literal.
	case p.isToken(tokenTypeStringLiteral):
		literalNode := p.startNode(NodeStringLiteralExpression)
		defer p.finishNode()

		token, _ := p.consume(tokenTypeStringLiteral)
		literalNode.Decorate(NodeStringLiteralExpressionValue, token.value)

		return literalNode, true

	// Template string literal.
	case p.isToken(tokenTypeTemplateStringLiteral):
		return p.consumeTemplateString(), true
	}

	return p.tryConsumeIdentifierExpression()
}

// consumeListExpression consumes an inline list expression.
func (p *sourceParser) consumeListExpression() AstNode {
	listNode := p.startNode(NodeListExpression)
	defer p.finishNode()

	// [
	if _, ok := p.consume(tokenTypeLeftBracket); !ok {
		return listNode
	}

	if !p.isToken(tokenTypeRightBracket) {
		// Consume one (or more) values.
		for {
			listNode.Connect(NodeListExpressionValue, p.consumeExpression())

			if p.isToken(tokenTypeRightBracket) {
				break
			}

			if _, ok := p.consume(tokenTypeComma); !ok {
				break
			}
		}
	}

	// ]
	p.consume(tokenTypeRightBracket)
	return listNode
}

// consumeTemplateString consumes a template string literal.
func (p *sourceParser) consumeTemplateString() AstNode {
	literalNode := p.startNode(NodeTemplateStringLiteralExpression)
	defer p.finishNode()

	// TODO(jschorr): We be parsing the contents of this string literal. Yaarr!
	token, _ := p.consume(tokenTypeTemplateStringLiteral)
	literalNode.Decorate(NodeTemplateStringLiteralExpressionValue, token.value)

	return literalNode
}

// tryConsumeIdentifierExpression tries to consume an identifier as an expression.
//
// Form:
// someIdentifier
func (p *sourceParser) tryConsumeIdentifierExpression() (AstNode, bool) {
	if p.isToken(tokenTypeIdentifer) {
		return p.consumeIdentifierExpression(), true
	}

	return nil, false
}

// consumeIdentifierExpression consumes an identifier as an expression.
//
// Form:
// someIdentifier
func (p *sourceParser) consumeIdentifierExpression() AstNode {
	identifierNode := p.startNode(NodeTypeIdentifierExpression)
	defer p.finishNode()

	value, ok := p.consumeIdentifier()
	if !ok {
		return identifierNode
	}

	identifierNode.Decorate(NodeIdentifierExpressionName, value)
	return identifierNode
}
