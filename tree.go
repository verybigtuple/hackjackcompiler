package main

import (
	"errors"
	"fmt"
	"io"
	"runtime"
)

type ParseTree struct {
	tz      *Tokenizer
	current Token
	peeked  [2]Token // buffer for peeked values
	root    func(*ParseTree) Node
}

func NewPasreTree(tz *Tokenizer) *ParseTree {
	pt := ParseTree{tz: tz}
	pt.root = func(pt *ParseTree) Node { return pt.letStatement() }
	return &pt
}

func (t *ParseTree) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		*errp = e.(error)
	}
}

func (t *ParseTree) errorf(format string, args ...interface{}) {
	panic(fmt.Errorf(format, args...))
}

func (t *ParseTree) Parse() (root Node, err error) {
	defer t.recover(&err)
	root = t.root(t)
	return
}

func (t *ParseTree) next() Token {
	if t.peeked[1] != nil {
		t.current, t.peeked[0], t.peeked[1] = t.peeked[0], t.peeked[1], nil
		return t.current
	}

	if t.peeked[0] != nil {
		t.current, t.peeked[0] = t.peeked[0], nil
		return t.current
	}

	var err error
	t.current, err = t.tz.ReadToken()
	if err != nil && !errors.Is(err, io.EOF) {
		t.errorf("Unexpected token: %v", err)
	}
	if err != nil && errors.Is(err, io.EOF) {
		t.errorf("Unexpected EOF")
	}
	return t.current
}

func (t *ParseTree) peek(fw int) Token {
	if fw < 0 || fw > 1 {
		panic("Can peak only for 0 or 1")
	}

	if t.peeked[fw] != nil {
		return t.peeked[fw]
	}

	p, err := t.tz.ReadToken()
	if err != nil && !errors.Is(err, io.EOF) {
		t.errorf("Unknown Token: %v", err)
	}
	t.peeked[fw] = p

	return p
}

func isTokenType(tk Token, tt TokenType) bool {
	return tk != nil && tk.Type() == tt
}

func isTokenOne(tk Token, tt TokenType, val string) bool {
	return isTokenType(tk, tt) && tk.GetValue() == val
}

func isTokenAny(tk Token, tt TokenType, vals ...string) bool {
	if isTokenType(tk, tt) {
		for _, v := range vals {
			if tk.GetValue() == v {
				return true
			}
		}
	}
	return false
}

func (t *ParseTree) feed() Token {
	return t.next()
}

func (t *ParseTree) feedToken(tt TokenType, val string) Token {
	tk := t.next()
	if tk.Type() != tt {
		t.errorf("Unexpexted token. Expected %v %s", tt, tk.GetValue())
	}
	if val != "" && tk.GetValue() != val {
		t.errorf("Unexpected token value. Expected %s; Got: %s", val, tk.GetValue())
	}
	return tk
}

// type:'int'|'char'|'boolean'|className
func (t *ParseTree) feedType() Token {
	tk := t.next()
	switch tk.Type() {
	case TokenKeyword:
		val := tk.GetValue()
		if val != "int" && val != "char" && val != "boolean" {
			t.errorf("Unexpected type")
		}
	case TokenIdentifier:
		break
	default:
		t.errorf("Unexpected Token for Var type")
	}

	return tk
}

// varDec:'var' type varName (','varName)*';'
func (t *ParseTree) varDec() *VarDecNode {
	t.feedToken(TokenKeyword, "var")
	vd := NewVarDecNode(t.feedType(), t.feedToken(TokenIdentifier, ""))
	for !isTokenOne(t.peek(0), TokenSymbol, ";") {
		t.feedToken(TokenSymbol, ",")
		vd.AddId(t.feedToken(TokenIdentifier, ""))
	}
	t.feedToken(TokenSymbol, ";")
	return vd
}

func (t *ParseTree) statements() *StatementsNode {
	st := NewStatementsNode()
	p := t.peek(0)
	for !isTokenOne(p, TokenSymbol, "}") {
		var newSt Node

		switch p.GetValue() {
		case "let":
			newSt = t.letStatement()
		case "if":
			newSt = t.ifStatement()
		case "while":
			newSt = t.whileStatement()
		default:
			t.errorf("Expected one of statement keywords let/if/while/return")
		}

		st.AddSt(newSt)
		p = t.peek(0)
	}
	return st
}

// 'let'varName ('['expression']')?'='expression';'
func (t *ParseTree) letStatement() *LetStatementNode {
	t.feedToken(TokenKeyword, "let")
	varNameToken := t.feedToken(TokenIdentifier, "")

	var arrExpr *ExpressionNode
	if p := t.peek(0); isTokenOne(p, TokenSymbol, "[") {
		t.feed() //feed [
		arrExpr = t.expression()
		t.feedToken(TokenSymbol, "]")
	}
	t.feedToken(TokenSymbol, "=")
	valExpr := t.expression()
	t.feedToken(TokenSymbol, ";")
	return NewLetStatementNode(varNameToken, arrExpr, valExpr)
}

// 'if''('expression')''{'statements'}'('else''{'statements'}')?
func (t *ParseTree) ifStatement() *IfStatementNode {
	t.feedToken(TokenKeyword, "if")
	t.feedToken(TokenSymbol, "(")
	ifExpr := t.expression()
	t.feedToken(TokenSymbol, ")")
	t.feedToken(TokenSymbol, "{")
	ifSt := t.statements()
	t.feedToken(TokenSymbol, "}")

	var elseSt *StatementsNode
	if isTokenOne(t.peek(0), TokenKeyword, "else") {
		t.feed()
		t.feedToken(TokenSymbol, "{")
		elseSt = t.statements()
		t.feedToken(TokenSymbol, "}")
	}
	return NewIfStatementNode(ifExpr, ifSt, elseSt)
}

func (t *ParseTree) whileStatement() *WhileStatementNode {
	t.feedToken(TokenKeyword, "while")
	t.feedToken(TokenSymbol, "(")
	expr := t.expression()
	t.feedToken(TokenSymbol, ")")
	t.feedToken(TokenSymbol, "{")
	st := t.statements()
	t.feedToken(TokenSymbol, "}")
	return NewWhileStatementNode(expr, st)
}

// term (op term)*
func (t *ParseTree) expression() *ExpressionNode {
	term := t.term()
	en := NewExpressionNode(term)

	p := t.peek(0)
	for isTokenAny(p, TokenSymbol, "+", "-", "*", "/", "&", "|", "<", ">", "=") {
		opToken := t.feed()
		nextTerm := t.term()
		en.AddOpTerm(opToken, nextTerm)
		p = t.peek(0)
	}
	return en
}

// term:  integerConstant | stringConstant | keywordConstant | varName | varName'['expression']'|
// subroutineCall |'('expression')'| unaryOp term
func (t *ParseTree) term() *TermNode {
	pFirst := t.peek(0)
	var tn *TermNode
	switch {
	// integerConstant
	case isTokenType(pFirst, TokenIntegerConst):
		intToken := t.feed()
		tn = NewConstTermNode(intToken)
	// stringConstant
	case isTokenType(pFirst, TokenStringConst):
		strToken := t.feed()
		tn = NewConstTermNode(strToken)
	// keywordConstant
	case isTokenAny(pFirst, TokenKeyword, "true", "false", "null", "this"):
		kwToken := t.feed()
		tn = NewConstTermNode(kwToken)
	// unaryOp term
	case isTokenAny(pFirst, TokenSymbol, "-", "~"):
		unOpTk := t.feed()
		childTerm := t.term()
		tn = NewUnaryTermNode(unOpTk, childTerm)
	case isTokenOne(pFirst, TokenSymbol, "("):
		t.feed()
		expr := t.expression()
		tn = NewExpressionTermNode(expr)
		t.feedToken(TokenSymbol, ")")
	//varName | varName'['expression']' | subroutineCall
	case isTokenType(pFirst, TokenIdentifier):
		pSecond := t.peek(1) // peek one more
		if isTokenOne(pSecond, TokenSymbol, "[") {
			ident := t.feed() // feed Identifier (peek(0))
			t.feed()          // feed [
			expr := t.expression()
			tn = NewArrayTermNode(ident, expr)
			t.feedToken(TokenSymbol, "]")
		} else if isTokenOne(pSecond, TokenSymbol, "(") {
			call := t.subroutineCall() // Call will feed Identifier and ( itself
			tn = NewCallTermNode(call)
		} else {
			ident := t.feed()
			tn = NewVarTermNode(ident)
		}
	default:
		t.errorf("Token is not a term: %v %s", pFirst.Type(), pFirst.GetValue())
	}

	return tn
}

// subroutineName '(' expressionList ')' | (className |varName) '.' subroutineName '(' expressionList ')'
func (t *ParseTree) subroutineCall() *SubroutineCallNode {
	name := t.feedToken(TokenIdentifier, "")

	var className Token // can be nil
	var sbrName Token
	if isTokenOne(t.peek(0), TokenSymbol, ".") {
		t.feed()
		className = name
		sbrName = t.feedToken(TokenIdentifier, "")
	} else {
		sbrName = name
	}
	t.feedToken(TokenSymbol, "(")
	params := t.expressionList()
	t.feedToken(TokenSymbol, ")")
	return NewSubroutineCallNode(className, sbrName, params)
}

// (expression (','expression)* )?
func (t *ParseTree) expressionList() *ExpressionListNode {
	eln := NewExpressionListNode()

	if !isTokenOne(t.peek(0), TokenSymbol, ")") {
		firstExpr := t.expression()
		eln.AddExpr(firstExpr)
	}

	p := t.peek(0)
	for isTokenOne(p, TokenSymbol, ",") {
		t.feed() // feed ","
		addExpr := t.expression()
		eln.AddExpr(addExpr)
		p = t.peek(0)
	}
	return eln
}
