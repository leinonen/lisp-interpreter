package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/leinonen/lisp-interpreter/pkg/types"
)

type Parser struct {
	tokens   []types.Token
	position int
	current  types.Token
}

func NewParser(tokens []types.Token) *Parser {
	p := &Parser{
		tokens: tokens,
	}
	p.readToken()
	return p
}

func (p *Parser) readToken() {
	if p.position >= len(p.tokens) {
		p.current = types.Token{Type: types.TokenType(-1), Value: ""} // EOF token
	} else {
		p.current = p.tokens[p.position]
	}
	p.position++
}

func (p *Parser) Parse() (types.Expr, error) {
	if len(p.tokens) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	// Check for remaining tokens (like unmatched closing parentheses)
	if p.current.Type != types.TokenType(-1) { // Not EOF
		return nil, fmt.Errorf("unexpected token after expression: %v", p.current)
	}

	return expr, nil
}

func (p *Parser) parseExpr() (types.Expr, error) {
	switch p.current.Type {
	case types.NUMBER:
		return p.parseNumber()
	case types.STRING:
		return p.parseString()
	case types.BOOLEAN:
		return p.parseBoolean()
	case types.SYMBOL:
		return p.parseSymbol()
	case types.KEYWORD:
		return p.parseKeyword()
	case types.LPAREN:
		return p.parseList()
	case types.LBRACKET:
		return p.parseBracket()
	case types.QUOTE:
		return p.parseQuote()
	case types.RPAREN:
		return nil, fmt.Errorf("unexpected closing parenthesis")
	case types.RBRACKET:
		return nil, fmt.Errorf("unexpected closing bracket")
	default:
		return nil, fmt.Errorf("unexpected token: %v", p.current)
	}
}

func (p *Parser) parseNumber() (types.Expr, error) {
	// First try to parse as a regular float64
	value, err := strconv.ParseFloat(p.current.Value, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number: %s", p.current.Value)
	}

	// Check if this number might lose precision when converted to float64
	// If the string representation is a large integer, use BigNumberExpr
	originalStr := p.current.Value

	// Check if it's an integer without decimal point
	if !strings.Contains(originalStr, ".") && !strings.Contains(originalStr, "e") && !strings.Contains(originalStr, "E") {
		// Check if it's too large for safe float64 integer representation
		if len(originalStr) > 15 || (len(originalStr) == 16 && originalStr[0] > '1') {
			// This is a large integer that might lose precision in float64
			expr := &types.BigNumberExpr{Value: originalStr}
			p.readToken()
			return expr, nil
		}
	}

	// Use regular NumberExpr for smaller numbers or floating point numbers
	expr := &types.NumberExpr{Value: value}
	p.readToken()
	return expr, nil
}

func (p *Parser) parseString() (types.Expr, error) {
	expr := &types.StringExpr{Value: p.current.Value}
	p.readToken()
	return expr, nil
}

func (p *Parser) parseBoolean() (types.Expr, error) {
	var value bool
	switch p.current.Value {
	case "true":
		value = true
	case "false":
		value = false
	default:
		return nil, fmt.Errorf("invalid boolean value: %s", p.current.Value)
	}
	expr := &types.BooleanExpr{Value: value}
	p.readToken()
	return expr, nil
}

func (p *Parser) parseSymbol() (types.Expr, error) {
	expr := &types.SymbolExpr{Name: p.current.Value}
	p.readToken()
	return expr, nil
}

func (p *Parser) parseKeyword() (types.Expr, error) {
	expr := &types.KeywordExpr{Value: p.current.Value}
	p.readToken()
	return expr, nil
}

func (p *Parser) parseList() (types.Expr, error) {
	p.readToken() // consume '('

	elements := make([]types.Expr, 0)

	for p.current.Type != types.RPAREN {
		if p.current.Type == types.TokenType(-1) { // EOF
			return nil, fmt.Errorf("unmatched opening parenthesis")
		}

		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		elements = append(elements, expr)
	}

	p.readToken() // consume ')'

	// Check for special module forms
	if len(elements) > 0 {
		if symbolExpr, ok := elements[0].(*types.SymbolExpr); ok {
			switch symbolExpr.Name {
			case "module":
				return p.parseModuleFromElements(elements)
			case "import":
				return p.parseImportFromElements(elements)
			case "load":
				return p.parseLoadFromElements(elements)
			case "require":
				return p.parseRequireFromElements(elements)
			}
		}
	}

	return &types.ListExpr{Elements: elements}, nil
}

func (p *Parser) parseModuleFromElements(elements []types.Expr) (types.Expr, error) {
	// (module name (export sym1 sym2...) body...)
	if len(elements) < 4 {
		return nil, fmt.Errorf("module requires at least name, export list, and body")
	}

	// Get module name
	nameExpr, ok := elements[1].(*types.SymbolExpr)
	if !ok {
		return nil, fmt.Errorf("module name must be a symbol")
	}

	// Get export list
	exportListExpr, ok := elements[2].(*types.ListExpr)
	if !ok {
		return nil, fmt.Errorf("module export list must be a list")
	}

	// Check export list format: (export symbol1 symbol2...)
	if len(exportListExpr.Elements) < 1 {
		return nil, fmt.Errorf("export list cannot be empty")
	}

	exportKeyword, ok := exportListExpr.Elements[0].(*types.SymbolExpr)
	if !ok || exportKeyword.Name != "export" {
		return nil, fmt.Errorf("export list must start with 'export'")
	}

	// Parse exported symbols
	exports := make([]string, len(exportListExpr.Elements)-1)
	for i, expr := range exportListExpr.Elements[1:] {
		symExpr, ok := expr.(*types.SymbolExpr)
		if !ok {
			return nil, fmt.Errorf("exported names must be symbols")
		}
		exports[i] = symExpr.Name
	}

	// Get body expressions
	body := elements[3:]

	return &types.ModuleExpr{
		Name:    nameExpr.Name,
		Exports: exports,
		Body:    body,
	}, nil
}

func (p *Parser) parseImportFromElements(elements []types.Expr) (types.Expr, error) {
	// (import module-name)
	if len(elements) != 2 {
		return nil, fmt.Errorf("import requires exactly one module name")
	}

	moduleNameExpr, ok := elements[1].(*types.SymbolExpr)
	if !ok {
		return nil, fmt.Errorf("import module name must be a symbol")
	}

	return &types.ImportExpr{ModuleName: moduleNameExpr.Name}, nil
}

func (p *Parser) parseLoadFromElements(elements []types.Expr) (types.Expr, error) {
	// (load "filename.lisp")
	if len(elements) != 2 {
		return nil, fmt.Errorf("load requires exactly one filename")
	}

	filenameExpr, ok := elements[1].(*types.StringExpr)
	if !ok {
		return nil, fmt.Errorf("load filename must be a string")
	}

	return &types.LoadExpr{Filename: filenameExpr.Value}, nil
}

func (p *Parser) parseRequireFromElements(elements []types.Expr) (types.Expr, error) {
	// Supported syntaxes:
	// (require "filename.lisp")
	// (require "filename.lisp" :as alias)
	// (require "filename.lisp" :only [symbol1 symbol2])

	if len(elements) < 2 {
		return nil, fmt.Errorf("require requires at least a filename")
	}

	filenameExpr, ok := elements[1].(*types.StringExpr)
	if !ok {
		return nil, fmt.Errorf("require filename must be a string")
	}

	requireExpr := &types.RequireExpr{
		Filename: filenameExpr.Value,
	}

	// Parse optional modifiers
	if len(elements) > 2 {
		if len(elements) < 4 {
			return nil, fmt.Errorf("require modifier requires an argument")
		}

		keywordExpr, ok := elements[2].(*types.KeywordExpr)
		if !ok {
			return nil, fmt.Errorf("require modifier must be a keyword (:as or :only)")
		}

		switch keywordExpr.Value {
		case "as":
			// (require "file.lisp" :as alias)
			if len(elements) != 4 {
				return nil, fmt.Errorf("require :as expects exactly one alias symbol")
			}
			aliasExpr, ok := elements[3].(*types.SymbolExpr)
			if !ok {
				return nil, fmt.Errorf("require :as alias must be a symbol")
			}
			requireExpr.AsAlias = aliasExpr.Name

		case "only":
			// (require "file.lisp" :only [symbol1 symbol2])
			if len(elements) != 4 {
				return nil, fmt.Errorf("require :only expects exactly one symbol list")
			}
			listExpr, ok := elements[3].(*types.ListExpr)
			if !ok {
				return nil, fmt.Errorf("require :only expects a list of symbols")
			}

			onlyList := make([]string, len(listExpr.Elements))
			for i, elem := range listExpr.Elements {
				symbolExpr, ok := elem.(*types.SymbolExpr)
				if !ok {
					return nil, fmt.Errorf("require :only list must contain only symbols")
				}
				onlyList[i] = symbolExpr.Name
			}
			requireExpr.OnlyList = onlyList

		default:
			return nil, fmt.Errorf("require modifier must be :as or :only, got %s", keywordExpr.Value)
		}
	}

	return requireExpr, nil
}

func (p *Parser) parseQuote() (types.Expr, error) {
	// Consume the quote token
	p.readToken()

	// Parse the expression that follows the quote
	expr, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("error parsing quoted expression: %v", err)
	}

	// Convert 'expr to (quote expr)
	quoteSymbol := &types.SymbolExpr{Name: "quote"}
	return &types.ListExpr{Elements: []types.Expr{quoteSymbol, expr}}, nil
}

func (p *Parser) parseBracket() (types.Expr, error) {
	p.readToken() // consume '['

	elements := make([]types.Expr, 0)

	for p.current.Type != types.RBRACKET {
		if p.current.Type == types.TokenType(-1) { // EOF
			return nil, fmt.Errorf("unmatched opening bracket")
		}

		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		elements = append(elements, expr)
	}

	p.readToken() // consume ']'

	return &types.BracketExpr{Elements: elements}, nil
}
