package parser

import (
        "fmt"
        "strconv"

        "github.com/miu0x/miuscript/ast"
        "github.com/miu0x/miuscript/lexer"
        "github.com/miu0x/miuscript/token"
)

type Parser struct {
        l *lexer.Lexer

        errors []string

        currToken      token.Token
        peekToken      token.Token
        prefixParseFns map[token.TokenType]prefixParseFn
        infixParseFns  map[token.TokenType]infixParseFn
}

const (
        _ int = iota
        LOWEST
        EQUALS
        LESSGREATER
        SUM
        PRODUCT
        PREFIX // -x or !x
        CALL   // foobar(x)
        INDEX  // array[index]
)

var precedences = map[token.TokenType]int{
        token.EQ:       EQUALS,
        token.NOT_EQ:   EQUALS,
        token.LT:       LESSGREATER,
        token.GT:       LESSGREATER,
        token.PLUS:     SUM,
        token.MINUS:    SUM,
        token.SLASH:    PRODUCT,
        token.ASTERICK: PRODUCT,
        token.LPAREN:   CALL,
        token.LBRACKET: INDEX,
}

type (
        prefixParseFn func() ast.Expression
        infixParseFn  func(ast.Expression) ast.Expression
)

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
        p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
        p.infixParseFns[tokenType] = fn
}

func New(l *lexer.Lexer) *Parser {
        p := &Parser{l: l, errors: []string{}}
        p.nextToken()
        p.nextToken()

        p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
        p.registerPrefix(token.IDENT, p.parseIdentifier)
        p.registerPrefix(token.INT, p.parseIntegerLiteral)
        p.registerPrefix(token.BANG, p.parsePrefixExpressions)
        p.registerPrefix(token.MINUS, p.parsePrefixExpressions)
        p.registerPrefix(token.TRUE, p.parseBooleans)
        p.registerPrefix(token.FALSE, p.parseBooleans)
        p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
        p.registerPrefix(token.IF, p.parseIfExpression)
        p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
        p.registerPrefix(token.STRING, p.parseStringLiteral)
        p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
        p.registerPrefix(token.LBRACE, p.parseHashLiteral)

        p.infixParseFns = make(map[token.TokenType]infixParseFn)
        p.registerInfix(token.PLUS, p.parseInfixExpressions)
        p.registerInfix(token.MINUS, p.parseInfixExpressions)
        p.registerInfix(token.SLASH, p.parseInfixExpressions)
        p.registerInfix(token.ASTERICK, p.parseInfixExpressions)
        p.registerInfix(token.EQ, p.parseInfixExpressions)
        p.registerInfix(token.NOT_EQ, p.parseInfixExpressions)
        p.registerInfix(token.LT, p.parseInfixExpressions)
        p.registerInfix(token.GT, p.parseInfixExpressions)
        p.registerInfix(token.LPAREN, p.parseCallExpression)
        p.registerInfix(token.LBRACKET, p.parseIndexExpression)

        return p
}

func (p *Parser) Errors() []string {
        return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
        msg := fmt.Sprintf("expected next token to be %s, got %s", t, p.peekToken.Type)
        p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
        p.currToken = p.peekToken
        p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
        program := &ast.Program{}
        program.Statements = []ast.Statement{}

        for p.currToken.Type != token.EOF {
                statement := p.parseStatement()
                if statement != nil {
                        program.Statements = append(program.Statements, statement)
                }
                p.nextToken()

        }
        return program
}

func (p *Parser) parseStatement() ast.Statement {
        switch p.currToken.Type {
        case token.LET:
                return p.parseLetStatement()
        case token.RETURN:
                return p.parseReturnStatement()
        default:
                return p.parseExpressionStatement()
        }
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
        statement := &ast.LetStatement{Token: p.currToken}

        if !p.expectPeek(token.IDENT) {
                return nil
        }

        statement.Name = &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}

        if !p.expectPeek(token.ASSIGN) {
                return nil
        }

        p.nextToken()

        statement.Value = p.parseExpression(LOWEST)

        if p.peekTokenIs(token.SEMICOLON) {
                p.nextToken()
        }

        return statement
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
        statement := &ast.ReturnStatement{Token: p.currToken}

        p.nextToken()

        statement.ReturnValue = p.parseExpression(LOWEST)

        for !p.currTokenIs(token.SEMICOLON) {
                p.nextToken()
        }

        return statement
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
        stmt := &ast.ExpressionStatement{Token: p.currToken}

        stmt.Expression = p.parseExpression(LOWEST)

        if p.peekTokenIs(token.SEMICOLON) {
                p.nextToken()
        }

        return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
        prefix := p.prefixParseFns[p.currToken.Type]
        if prefix == nil {
                p.noPrefixParseFnError(p.currToken.Type)
                return nil
        }
        leftExp := prefix()

        for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
                infix := p.infixParseFns[p.peekToken.Type]
                if infix == nil {
                        return leftExp
                }
                p.nextToken()
                leftExp = infix(leftExp)
        }
        return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
        return &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
        lit := &ast.IntegerLiteral{Token: p.currToken}
        value, err := strconv.ParseInt(p.currToken.Literal, 0, 64)
        if err != nil {
                msg := fmt.Sprintf("count not parse string to int64, got=%s", p.currToken.Literal)
                p.errors = append(p.errors, msg)
                return nil
        }

        lit.Value = value
        return lit
}

func (p *Parser) parsePrefixExpressions() ast.Expression {
        expression := &ast.PrefixExpression{
                Token:    p.currToken,
                Operator: p.currToken.Literal,
        }
        p.nextToken()
        expression.Right = p.parseExpression(PREFIX)
        return expression
}

func (p *Parser) parseInfixExpressions(left ast.Expression) ast.Expression {
        expression := &ast.InfixExpression{
                Token:    p.currToken,
                Operator: p.currToken.Literal,
                Left:     left,
        }
        precedence := p.currPrecedence()
        p.nextToken()
        expression.Right = p.parseExpression(precedence)

        return expression
}

func (p *Parser) parseBooleans() ast.Expression {
        return &ast.Boolean{Token: p.currToken, Value: p.currTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
        p.nextToken()

        exp := p.parseExpression(LOWEST)

        if !p.expectPeek(token.RPAREN) {
                return nil
        }

        return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
        expression := &ast.IfExpression{Token: p.currToken}

        if !p.expectPeek(token.LPAREN) {
                return nil
        }

        p.nextToken()

        expression.Condition = p.parseExpression(LOWEST)
        if !p.expectPeek(token.RPAREN) {
                return nil
        }

        if !p.expectPeek(token.LBRACE) {
                return nil
        }

        expression.Consequence = p.parseBlockStatement()

        if p.peekTokenIs(token.ELSE) {
                p.nextToken()

                if !p.expectPeek(token.LBRACE) {
                        return nil
                }

                expression.Alternative = p.parseBlockStatement()

        }

        return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
        block := &ast.BlockStatement{Token: p.currToken}
        block.Statements = []ast.Statement{}

        p.nextToken()

        for !p.currTokenIs(token.RBRACE) && !p.currTokenIs(token.EOF) {
                stmt := p.parseStatement()
                if stmt != nil {
                        block.Statements = append(block.Statements, stmt)
                }
                p.nextToken()
        }
        return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
        lit := &ast.FunctionLiteral{Token: p.currToken}

        if !p.expectPeek(token.LPAREN) {
                return nil
        }

        lit.Parameters = p.parseFunctionParamters()

        if !p.expectPeek(token.LBRACE) {
                return nil
        }
        lit.Body = p.parseBlockStatement()
        return lit
}

func (p *Parser) parseFunctionParamters() []*ast.Identifier {
        identifers := []*ast.Identifier{}

        if p.peekTokenIs(token.RPAREN) {
                p.nextToken()
                return identifers
        }
        p.nextToken()

        ident := &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
        identifers = append(identifers, ident)
        for p.peekTokenIs(token.COMMA) {
                p.nextToken()
                p.nextToken()

                ident := &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
                identifers = append(identifers, ident)
        }

        if !p.expectPeek(token.RPAREN) {
                return nil
        }

        return identifers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
        exp := &ast.CallExpression{Token: p.currToken, Function: function}
        exp.Arguments = p.parseExpressionList(token.RPAREN)
        return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
        args := []ast.Expression{}

        if p.peekTokenIs(token.RPAREN) {
                p.nextToken()
                return args
        }

        p.nextToken()

        args = append(args, p.parseExpression(LOWEST))

        for p.peekTokenIs(token.COMMA) {
                p.nextToken()
                p.nextToken()
                args = append(args, p.parseExpression(LOWEST))
        }

        if !p.expectPeek(token.RPAREN) {
                return nil
        }

        return args
}

func (p *Parser) parseStringLiteral() ast.Expression {
        return &ast.StringLiteral{Token: p.currToken, Value: p.currToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
        array := &ast.ArrayLiteral{
                Token: p.currToken,
        }
        array.Elements = p.parseExpressionList(token.RBRACKET)

        return array
}

func (p *Parser) parseHashLiteral() ast.Expression {
        hash := &ast.HashLiteral{
                Token: p.currToken,
        }

        hash.Pairs = make(map[ast.Expression]ast.Expression)

        for !p.peekTokenIs(token.RBRACE) {
                p.nextToken()
                key := p.parseExpression(LOWEST)

                if !p.expectPeek(token.COLON) {
                        return nil
                }

                p.nextToken()

                value := p.parseExpression(LOWEST)

                hash.Pairs[key] = value
                if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
                        return nil
                }

        }

        if !p.expectPeek(token.RBRACE) {
                return nil
        }

        return hash
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
        list := []ast.Expression{}
        if p.peekTokenIs(end) {
                p.nextToken()
                return list
        }

        p.nextToken()
        list = append(list, p.parseExpression(LOWEST))

        for p.peekTokenIs(token.COMMA) {
                p.nextToken()
                p.nextToken()
                list = append(list, p.parseExpression(LOWEST))
        }

        if !p.expectPeek(end) {
                return nil
        }

        return list
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
        exp := &ast.IndexExpression{Token: p.currToken, Left: left}
        p.nextToken()
        exp.Index = p.parseExpression(LOWEST)
        if !p.expectPeek(token.RBRACKET) {
                return nil
        }
        return exp
}

func (p *Parser) currTokenIs(t token.TokenType) bool {
        return p.currToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
        return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
        if p.peekTokenIs(t) {
                p.nextToken()
                return true
        } else {
                p.peekError(t)
                return false
        }
}

func (p *Parser) peekPrecedence() int {
        if p, ok := precedences[p.peekToken.Type]; ok {
                return p
        }
        return LOWEST
}

func (p *Parser) currPrecedence() int {
        if p, ok := precedences[p.currToken.Type]; ok {
                return p
        }
        return LOWEST
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
        msg := fmt.Sprintf("no prefix function for %s found.", t)
        p.errors = append(p.errors, msg)
}
