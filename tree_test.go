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

var varDecCases = []testCase{
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

func TestDecNode(t *testing.T) {
	start := func(t *ParseTree) Node { return t.varDec() }
	simpleTest(t, start, varDecCases)
}

var termTestCases = []testCase{
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

func TestTermNode(t *testing.T) {
	start := func(t *ParseTree) Node { return t.term() }
	simpleTest(t, start, termTestCases)
}

var exprCases = []testCase{
	//{"Simple", "0", false},
	{"Two operands", "a + 0", false},
	{"More operands", "a + 0 - 1 / 3", false},
	//Errors
	{"Wrong operands", "a~1", true},
	{"Wrong keyword", "a+class", true},
}

func TestExprNode(t *testing.T) {
	start := func(t *ParseTree) Node { return t.expression() }
	simpleTest(t, start, exprCases)
}
