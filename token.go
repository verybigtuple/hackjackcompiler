package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// All symbols of Jack Language
var symbols = map[byte]bool{
	'(': true, ')': true,
	'[': true, ']': true,
	'{': true, '}': true,
	',': true, ';': true, '=': true, '.': true,
	'+': true, '-': true, '*': true, '/': true,
	'&': true, '|': true, '~': true,
	'<': true, '>': true,
}

// All reserved words of Jack Language
var keywords = map[string]bool{
	"class": true, "constructor": true, "method": true, "function": true,
	"int": true, "boolean": true, "char": true, "void": true,
	"var": true, "static": true, "field": true,
	"let": true, "do": true, "if": true, "else": true, "while": true, "return": true,
	"true": true, "false": true, "null": true,
	"this": true,
}

var xmlReplacer = strings.NewReplacer(
	"<", "&lt;",
	">", "&gt;",
	"\"", "&quot;",
	"&", "&amp;",
)

type Token interface {
	GetValue() string
	GetXml() string
}

type defaultToken struct {
	value   string
	xmlNode string
}

func (dt *defaultToken) GetValue() string {
	return dt.value
}

func (dt *defaultToken) GetXml() string {
	nv := xmlReplacer.Replace(dt.value)
	return fmt.Sprintf("<%s> %s </%s>", dt.xmlNode, nv, dt.xmlNode)
}

type KeywordToken struct {
	Token
}

func NewKeywordToken(value string) Token {
	return &defaultToken{value: value, xmlNode: "keyword"}
}

type IdentifierToken struct {
	Token
}

func NewIdentifierToken(value string) Token {
	return &defaultToken{value: value, xmlNode: "identifier"}
}

type SymbolToken struct {
	Token
}

func NewSymbolToken(value string) Token {
	return &defaultToken{value: value, xmlNode: "symbol"}
}

type StringConstantToken struct {
	Token
}

func NewStringConstantToken(value string) Token {
	return &defaultToken{value: value, xmlNode: "stringConstant"}
}

type IntegerConstantToken struct {
	Token
}

func NewIntegerConstantToken(value string) Token {
	return &defaultToken{value: value, xmlNode: "integerConstant"}
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r'
}

func isEOL(ch byte) bool {
	return ch == '\n'
}

type Tokenizer struct {
	reader *bufio.Reader
	buf    strings.Builder
	line   int
}

func (t *Tokenizer) ReadToken() (Token, error) {
	t.buf.Reset()
	first, err := t.skipSpaces()
	if err != nil {
		return nil, err
	}

	if first == '/' {
		err := t.skipComment()
		if err != nil {
			return nil, err
		}
	}

	switch {
	case symbols[first]:
		return NewSymbolToken(string(first)), nil
	case first == '"':
		t.readStringToken()
		word := t.buf.String()
		return NewKeywordToken(word), nil
	case unicode.IsLetter(rune(first)):
		t.readWord(first)
		word := t.buf.String()
		if keywords[word] {
			return NewKeywordToken(word), nil
		}
		return NewIdentifierToken(word), nil
	case unicode.IsNumber(rune(first)):
		t.readWord(first)
		word := t.buf.String()
		return NewIntegerConstantToken(word), nil
	}

	return nil, nil

}

func (t *Tokenizer) skipSpaces() (ch byte, err error) {
	for {
		ch, err = t.reader.ReadByte()
		if ch == '\n' {
			t.line++
		}
		if !isSpace(ch) {
			return
		}
	}
}

func (t *Tokenizer) skipComment() (err error) {
	next, err := t.reader.Peek(1)
	if err != nil {
		return err
	}
	// one line comment - skip everythin until next line
	if next[0] == '/' {
		for {
			ch, err := t.reader.ReadByte()
			if err != nil {
				return err
			}
			if isEOL(ch) {
				t.line++
				return nil
			}
		}
	}

	// Multiline comment
	if next[0] == '*' {
		twoBuf := make([]byte, 2)

		for {
			_, err := io.ReadFull(t.reader, twoBuf)
			if bytes.Equal(twoBuf, []byte{'*', '/'}) || err != nil {
				twoBuf = nil
				return nil
			}
			if bytes.Contains(twoBuf, []byte{'\n'}) {
				t.line++
			}
		}
	}

	return nil
}

func (t *Tokenizer) readWord(fb byte) (err error) {
	t.buf.Reset()
	t.buf.WriteByte(fb)
	var next []byte
	for {
		next, err = t.reader.Peek(1)
		if err != nil {
			return err
		}
		if isSpace(next[0]) || symbols[next[0]] {
			return nil
		}
		ch, err := t.reader.ReadByte()
		if err != nil {
			return err
		}
		t.buf.WriteByte(ch)
	}
}

func (t *Tokenizer) readStringToken() error {
	for {
		ch, err := t.reader.ReadByte()
		if err != nil {
			return err
		}
		if ch == '"' {
			return nil
		}
		t.buf.WriteByte(ch)
	}
}
