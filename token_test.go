package main

import (
	"bufio"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

func compareTokens(t *testing.T, a Token, b Token) {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		t.Fatalf("Got %v %+v; want: %v %+v", reflect.TypeOf(a), a, reflect.TypeOf(b), b)
	}

	if a.GetValue() != b.GetValue() {
		t.Fatalf("Got %v; want: %v", a.GetValue(), b.GetValue())
	}
}

func getAllTokens(t *testing.T, testCase string, want []Token) {
	reader := bufio.NewReader(strings.NewReader(testCase))
	tokenizer := NewTokenizer(reader)

	got := make([]Token, 0, len(want))
	for {
		if tk, err := tokenizer.ReadToken(); err == nil {
			got = append(got, tk)
		} else if !errors.Is(err, io.EOF) {
			t.Fatalf("Unexpected error %v", err)
		} else {
			break
		}
	}
	if len(got) != len(want) {
		t.Fatalf("Got %d items; want %d items", len(got), len(want))
	}
	for i, g := range got {
		compareTokens(t, g, want[i])
	}
}

func TestSimpleTokens(t *testing.T) {
	testCases := [...]struct {
		str  string
		want Token
	}{
		{"class", NewKeywordToken("class", 0, 0)},
		{"a", NewIdentifierToken("a", 0, 0)},
		{"id", NewIdentifierToken("id", 0, 0)},
		{"+", NewSymbolToken("+", 0, 0)},
		{"0", NewIntegerConstantToken("0", 0, 0)},
		{"100", NewIntegerConstantToken("100", 0, 0)},
		{"\"string\"", NewStringConstantToken("string", 0, 0)},
	}

	for _, tc := range testCases {
		t.Run(tc.str, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.str))
			tokenizer := NewTokenizer(reader)
			got, err := tokenizer.ReadToken()
			if err != nil {
				t.Fatalf("Unexpected error %v", err)
			}
			compareTokens(t, got, tc.want)
		})
	}
}

func TestMultiTokens(t *testing.T) {
	testCase := "class MyClass();"

	want := []Token{
		NewKeywordToken("class", 0, 0),
		NewIdentifierToken("MyClass", 0, 0),
		NewSymbolToken("(", 0, 0),
		NewSymbolToken(")", 0, 0),
		NewSymbolToken(";", 0, 0),
	}

	getAllTokens(t, testCase, want)
}

func TestSpaces(t *testing.T) {
	testCases := []struct {
		name string
		str  string
	}{
		{"End lines", "\n\n\nclass\n\n\n;\n\n\n"},
		{"Spaces", "   class;  "},
		{"Tabs", "\t\tclass;\t\t"},
		{"Tabs and spaces", "\t   class   ; \t "},
	}

	want := []Token{
		NewKeywordToken("class", 0, 0),
		NewSymbolToken(";", 0, 0),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getAllTokens(t, tc.str, want)
		})
	}
}

func TestComments(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{"Inline Comments", "// Comment1\n // Comment2\n class //Comment3\n"},
		{"Inline with keywrords", "// class\n class // class"},
		{"Multiline comments", "/* One Line */\n/** *** */\n/*one\ntwo*/\nclass"},
		{"Both comments", "// Inline comment\n /* Multi line\ncomment */ \n class // Inline"},
	}

	want := []Token{NewKeywordToken("class", 0, 0)}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getAllTokens(t, tc.code, want)
		})
	}
}

func TestStringConstant(t *testing.T) {
	testCase := "a = \"String Constant\"; b"
	want := []Token{
		NewIdentifierToken("a", 0, 0),
		NewSymbolToken("=", 0, 0),
		NewStringConstantToken("String Constant", 0, 0),
		NewSymbolToken(";", 0, 0),
		NewIdentifierToken("b", 0, 0),
	}
	getAllTokens(t, testCase, want)
}


func TestOneLinePos(t *testing.T) {
	testCase := "let  a = foo;\nbar;\n"
	reader := bufio.NewReader(strings.NewReader(testCase))
	tokenizer := NewTokenizer(reader)

	want := [...]struct {
		val  string
		line int
		pos  int
	}{
		{"let", 1, 1},
		{"a", 1, 6},
		{"=", 1, 8},
		{"foo", 1, 10},
		{";", 1, 13},
		{"bar", 2, 1},
		{";", 2, 4},
	}

	for _, w := range want {
		t.Run(w.val, func(t *testing.T) {
			tk, err := tokenizer.ReadToken()
			if err != nil && errors.Is(err, io.EOF) {
				t.Fatal("Unexpected EOF")
			}
			if err != nil {
				t.Fatalf("Unexpected error %v", err)
			}
			if w.line != tk.Line() {
				t.Errorf("Wanted line %d; Got %d", w.line, tk.Line())
			}
			if w.pos != tk.Pos() {
				t.Errorf("Wanted pos %d; Got %d", w.pos, tk.Pos())
			}
		})
	}
}
