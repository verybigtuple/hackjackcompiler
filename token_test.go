package main

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

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
	}

	for _, tc := range testCases {
		t.Run(tc.str, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.str))
			tokenizer := NewTokenizer(reader)
			got, err := tokenizer.ReadToken()
			if err != nil {
				t.Errorf("Unexpected error %v", err)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tc.want) {
				t.Errorf("Got %v; want: %v", reflect.TypeOf(got), reflect.TypeOf(tc.want))
				return
			}
			if got.GetValue() != tc.want.GetValue() {
				t.Errorf("Got %v; want: %v", got.GetValue(), tc.want.GetValue())
				return
			}
		})
	}
}
