package main

import (
	"bufio"
	"strings"
	"testing"
)

type testCase struct {
	name     string
	code     string
	hasError bool
}

func simpleTest(t *testing.T, start func(*ParseTree) Node, cases []testCase) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.code))
			tz := NewTokenizer(reader)
			pt := NewPasreTree(tz)
			pt.root = start
			_, err := pt.Parse()
			if !tc.hasError && err != nil {
				t.Errorf("Got error: %v", err)
			}
		})
	}
}

func TestDecNode(t *testing.T) {
	varDecCases := []testCase{
		{"One int", "var int a;", false},
		{"One char", "var char a;", false},
		{"One bool", "var boolean a;", false},
		{"One class", "var MyClass a;", false},
		{"Many ints", "var int a, b, c;", false},
		//Errors
		{"Wrong comma", "var int a,b, ;", true},
		{"No ;", "var int a, b, c", true},
		{"Unexpected finish", "var int a, b, c:", true},
		{"Wrong type", "var bool a;", true},
		{"Wrong type keyword", "var class a;", true},
		{"Wrong keyord", "vara int a;", true},
	}
	start := func(t *ParseTree) Node { return t.varDec() }
	simpleTest(t, start, varDecCases)
}

func TestTermNode(t *testing.T) {
	termTestCases := []testCase{
		{"Integer", "0", false},
		{"String", "\"String const\"", false},
		{"True const", "true", false},
		{"False const", "false", false},
		{"Null const", "null", false},
		{"This const", "this", false},
		{"Var name", "a", false},
		{"Array", "a[1]", false},
		{"Array with expression", "a[1+1]", false},
		{"Expression", "(a + b)", false},
		{"Unary minus", "-a", false},
		{"Unary Not", "~1", false},
		//Errors
		{"Symbol", "+", true},
		{"End of string", ";", true},
		{"No end for array", "a[1;", true},
		{"Worng end of array", "a[1);", true},
		{"Worng keword", "class", true},
		{"Wrong unary", "+a", true},
	}
	start := func(t *ParseTree) Node { return t.term() }
	simpleTest(t, start, termTestCases)
}

func TestExprNode(t *testing.T) {
	exprCases := []testCase{
		{"Two operands", "a + 0", false},
		{"More operands", "a + 0 - 1 / 3", false},
		//Errors
		{"Wrong operands", "a~1", true},
		{"Wrong keyword", "a+class", true},
	}

	start := func(p *ParseTree) Node { return p.expression() }
	simpleTest(t, start, exprCases)
}

func TestCallNode(t *testing.T) {
	subRoutineCases := []testCase{
		{"Call without params", "foo()", false},
		{"Call with one param", "foo(1)", false},
		{"Call with many params", "foo(1, a, d)", false},
		{"Call nested", "foo(bar(a))", false},
		{"Class Mamber without params", "MyClass.Foo()", false},
		{"Class Mamber with params", "MyClass.Foo(1, 2, a + 3)", false},
		{"Class Mamber with nester", "MyClass.Foo(bar(a) + 1)", false},

		//Errors
		{"Wrong parenthesis", "Foo[]", true},
		{"Wrong commas", "Foo(a + b,)", true},
	}

	start := func(t *ParseTree) Node { return t.subroutineCall() }
	simpleTest(t, start, subRoutineCases)
}

func TestIfNode(t *testing.T) {
	ifStatementCases := []testCase{
		{"Empty if", "if(a) {}", false},
		{"If with statement", "if(a) { if (1) {} }", false},
		{"If with else", "if(a) {} else {}", false},
		{"If with else with statements", "if(a) {if (1) {} } else { if (2) {} }", false},
		{"If with let", "if(a) {let b = 0;}", false},

		//errors
		{"If without expression", "if {a}", true},
		{"else without if", "else {}", true},
		{"else with expression", "else (a) {}", true},
		{"if else if", "if (a) else if (a) {}", true},
	}
	start := func(p *ParseTree) Node { return p.ifStatement() }
	simpleTest(t, start, ifStatementCases)
}

func TestLetNode(t *testing.T) {
	letCases := []testCase{
		{"Regular", "let a = 0;", false},
		{"Array", "let a[0] = 0;", false},
		{"Expressions", "let a[foo(1)] = bar(2);", false},
		// errors
		{"Wrong end", "let a = 0", true},
		{"No expression", "let a;", true},
		{"No expression for array", "let a[0];", true},
	}
	start := func(p *ParseTree) Node { return p.letStatement() }
	simpleTest(t, start, letCases)
}

func TestWhileNode(t *testing.T) {
	letCases := []testCase{
		{"Regular empty statement", "while (true) {}", false},
		{"Regular with statement", "while (true) {let a = b; do foo();}", false},
		{"Regular with expression", "while (a | b) {let a = b; do foo();}", false},
		// errors
		{"Empty expression", "while ()", true},
	}
	start := func(p *ParseTree) Node { return p.whileStatement() }
	simpleTest(t, start, letCases)
}

func TestDoNode(t *testing.T) {
	doCases := []testCase{
		{"Regular do", "do foo();", false},
		{"Regular class do", "do MyClass.foo();", false},
	}
	start := func(p *ParseTree) Node { return p.doStatement() }
	simpleTest(t, start, doCases)
}

func TestReturnNode(t *testing.T) {
	returnCases := []testCase{
		{"Regualr empty", "return;", false},
		{"With return", "return a+b;", false},
	}
	start := func(p *ParseTree) Node { return p.returnStatement() }
	simpleTest(t, start, returnCases)
}

func TestSubroutineBody(t *testing.T) {
	suroutineTest := []testCase{
		{"Void Subroutine", "function void Foo() { return; }", false},
		{"Int return Subroutine", "function int Foo() { return; }", false},
		{"Class return Subroutine", "function MyClass Foo() { return; }", false},
		{"Constructor", "constructor MyClass Foo() { return this; }", false},
		{"Method", "method int Foo() { return this; }", false},
		{"Parameters", "method int Foo(int a, int b) { return this; }", false},
		{"Many statements", "method int Foo(int a, int b) { var int a; return this; }", false},
	}
	start := func(p *ParseTree) Node { return p.subroutineDec() }
	simpleTest(t, start, suroutineTest)
}
