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
		{"class", NewKeywordToken("class")},
		{"a", NewIdentifierToken("a")},
		{"id", NewIdentifierToken("id")},
		{"+", NewSymbolToken("+")},
		{"0", NewIntegerConstantToken("0")},
		{"100", NewIntegerConstantToken("100")},
		{"\"string\"", NewStringConstantToken("string")},
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
		NewKeywordToken("class"),
		NewIdentifierToken("MyClass"),
		NewSymbolToken("("),
		NewSymbolToken(")"),
		NewSymbolToken(";"),
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
		NewKeywordToken("class"),
		NewSymbolToken(";"),
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

	want := []Token{NewKeywordToken("class")}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getAllTokens(t, tc.code, want)
		})
	}
}
