package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"runtime"
)

type ParseTree struct {
	tz             *Tokenizer
	current        Token
	peeked         [2]Token // buffer for peeked values
	rootNodeParser func(*ParseTree) Node
	root           Node
}

func NewPasreTree(tz *Tokenizer) *ParseTree {
	pt := ParseTree{tz: tz}
	pt.rootNodeParser = func(pt *ParseTree) Node { return pt.class() }
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

func (t *ParseTree) error(msg string) {
	t.errorf("%s", msg)
}

func (t *ParseTree) errorf(format string, args ...interface{}) {
	panic(fmt.Errorf(format, args...))
}

func (t *ParseTree) Parse() (rootNode Node, err error) {
	defer t.recover(&err)
	t.root = t.rootNodeParser(t)
	rootNode = t.root
	return
}

func (t *ParseTree) WriteXml(wb *bufio.Writer) {
	if t.root != nil {
		xb := NewXmlBuilder()
		t.root.Xml(xb)
		wb.WriteString(xb.String())
	}
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
		t.errorf("Unexpected EOF Ln %d", t.tz.Line)
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
		t.errorf("Unexpected token: %v", err)
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
	if tk.Type() != tt || (val != "" && tk.GetValue() != val) {
		t.errorf("Unexpexted token %v. Expected value %s", tk, val)
	}
	return tk
}

// type:'int'|'char'|'boolean'|className
func (t *ParseTree) varType() Token {
	tk := t.next()
	switch tk.Type() {
	case TokenKeyword:
		val := tk.GetValue()
		if val != "int" && val != "char" && val != "boolean" {
			t.errorf("Unexpected Jack builin type: %v", tk)
		}
	case TokenIdentifier:
		break
	default:
		t.errorf("Unexpected Token for Var type: %v", tk)
	}

	return tk
}

func (t *ParseTree) class() *ClassNode {
	t.feedToken(TokenKeyword, "class")
	clName := t.feedToken(TokenIdentifier, "")
	t.feedToken(TokenSymbol, "{")

	cln := NewClassNode(clName)
	p := t.peek(0)
	for isTokenAny(p, TokenKeyword, "static", "field") {
		clv := t.classVarDec()
		cln.AddVarDecs(clv)
		p = t.peek(0)
	}

	for isTokenAny(p, TokenKeyword, "constructor", "function", "method") {
		sbr := t.subroutineDec()
		cln.AddSbrDecs(sbr)
		p = t.peek(0)
	}

	t.feedToken(TokenSymbol, "}")
	return cln
}

func (t *ParseTree) classVarDec() *ClassVarDecNode {
	p := t.peek(0)
	var varClass Token
	if isTokenAny(p, TokenKeyword, "static", "field") {
		varClass = t.feed()
	}
	varType := t.varType()
	varName := t.feedToken(TokenIdentifier, "")
	vd := NewClassVarDecNode(varClass, varType, varName)
	for !isTokenOne(t.peek(0), TokenSymbol, ";") {
		t.feedToken(TokenSymbol, ",")
		vd.AddVarNames(t.feedToken(TokenIdentifier, ""))
	}
	t.feedToken(TokenSymbol, ";")
	return vd
}

func (t *ParseTree) subroutineDec() *SubroutineDecNode {
	p := t.peek(0)
	if !isTokenAny(p, TokenKeyword, "constructor", "function", "method") {
		t.errorf("Expected method class: constructor, function or method")
	}
	sbrClass := t.feed()

	var returnType Token
	p = t.peek(0)
	if isTokenOne(p, TokenKeyword, "void") {
		returnType = t.feed()
	} else {
		returnType = t.varType()
	}

	sbrName := t.feedToken(TokenIdentifier, "")
	t.feedToken(TokenSymbol, "(")
	paramList := t.parameterList()
	t.feedToken(TokenSymbol, ")")
	sbrBody := t.subroutineBody()
	return NewSubroutineDecNode(sbrClass, returnType, sbrName, paramList, sbrBody)
}

func (t *ParseTree) parameterList() *ParameterListNode {
	pln := NewParameterListNode()

	p := t.peek(0)
	if !isTokenOne(p, TokenSymbol, ")") {
		firtsType := t.varType()
		firstVarName := t.feedToken(TokenIdentifier, "")
		pln.AddParameter(firtsType, firstVarName)

		p = t.peek(0)
		for isTokenOne(p, TokenSymbol, ",") {
			t.feed()
			nextType := t.varType()
			nextVarName := t.feedToken(TokenIdentifier, "")
			pln.AddParameter(nextType, nextVarName)
			p = t.peek(0)
		}
	}
	return pln
}

func (t *ParseTree) subroutineBody() *SubroutineBodyNode {
	t.feedToken(TokenSymbol, "{")
	p := t.peek(0)

	var varDecs []*VarDecNode
	for isTokenOne(p, TokenKeyword, "var") {
		vd := t.varDec()
		varDecs = append(varDecs, vd)
		p = t.peek(0)
	}
	st := t.statements()
	t.feedToken(TokenSymbol, "}")

	sbn := NewSubroutineBodyNode(st)
	sbn.AddVarDec(varDecs...)
	return sbn
}

// varDec:'var' type varName (','varName)*';'
func (t *ParseTree) varDec() *VarDecNode {
	t.feedToken(TokenKeyword, "var")
	vd := NewVarDecNode(t.varType(), t.feedToken(TokenIdentifier, ""))
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
		case "do":
			newSt = t.doStatement()
		case "return":
			newSt = t.returnStatement()
		default:
			t.errorf("Unexpected statement begin: %v", p)
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

	lsn := NewLetStatementNode(varNameToken, valExpr)
	lsn.AddArrayExpr(arrExpr)
	return lsn
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

	ifs := NewIfStatementNode(ifExpr, ifSt)

	if isTokenOne(t.peek(0), TokenKeyword, "else") {
		t.feed()
		t.feedToken(TokenSymbol, "{")
		elseSt := t.statements()
		t.feedToken(TokenSymbol, "}")
		ifs.AddElse(elseSt)
	}
	return ifs
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

func (t *ParseTree) doStatement() *DoStatementNode {
	t.feedToken(TokenKeyword, "do")
	call := t.subroutineCall()
	t.feedToken(TokenSymbol, ";")
	return NewDoStatementNode(call)
}

func (t *ParseTree) returnStatement() *ReturnStatementNode {
	t.feedToken(TokenKeyword, "return")

	rsn := NewReturnNode()
	if !isTokenOne(t.peek(0), TokenSymbol, ";") {
		expr := t.expression()
		rsn.AddExpr(expr)
	}
	t.feedToken(TokenSymbol, ";")
	return rsn
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
		tn = NewIntConstTermNode(intToken)
	// stringConstant
	case isTokenType(pFirst, TokenStringConst):
		strToken := t.feed()
		tn = NewConstTermNode(strToken)
	// keywordConstant
	case isTokenAny(pFirst, TokenKeyword, "true", "false", "null", "this"):
		kwToken := t.feed()
		tn = NewKeyWordConstTermNode(kwToken)
	case isTokenOne(pFirst, TokenKeyword, "this"):
		kwToken := t.feed()
		tn = NewThisConstTermNode(kwToken)
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
		} else if isTokenAny(pSecond, TokenSymbol, "(", ".") {
			call := t.subroutineCall() // Call will feed Identifier and ( itself
			tn = NewCallTermNode(call)
		} else {
			ident := t.feed()
			tn = NewVarTermNode(ident)
		}
	default:
		t.errorf("Token is not a term: %v", pFirst)
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

	if className != nil {
		return NewClassSubroutineCallNode(className, sbrName, params)
	}
	return NewSubroutineCallNode(sbrName, params)
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
