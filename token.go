package main

import (
	"bufio"
	"errors"
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

type TokenType int

const (
	TokenKeyword TokenType = iota
	TokenIdentifier
	TokenSymbol
	TokenStringConst
	TokenIntegerConst
)

func (tt TokenType) Type() TokenType {
	return tt
}

func (tt TokenType) String() string {
	switch tt {
	case TokenKeyword:
		return "Keyword"
	case TokenIdentifier:
		return "Identifier"
	case TokenSymbol:
		return "Symbol"
	case TokenStringConst:
		return "String constant"
	case TokenIntegerConst:
		return "Integer constant"
	default:
		return fmt.Sprintf("Undefined token type %d", tt)
	}
}

type Token interface {
	Xmler
	Type() TokenType
	GetValue() string
	Line() int
	Pos() int
}

type defaultToken struct {
	value   string
	xmlNode string
	line    int
	pos     int
}

func (dt *defaultToken) GetValue() string {
	return dt.value
}

func (dt *defaultToken) Xml(xb *XmlBuilder) {
	xb.WriteNode(dt.xmlNode, dt.value)
}

func (dt *defaultToken) Line() int {
	return dt.line
}

func (dt *defaultToken) Pos() int {
	return dt.pos
}

type KeywordToken struct {
	TokenType
	defaultToken
}

func NewKeywordToken(value string, line, pos int) Token {
	return &KeywordToken{TokenKeyword, defaultToken{value, "keyword", line, pos}}
}

type IdentifierToken struct {
	TokenType
	defaultToken
}

func NewIdentifierToken(value string, line, pos int) Token {
	return &IdentifierToken{TokenIdentifier, defaultToken{value, "identifier", line, pos}}
}

type SymbolToken struct {
	TokenType
	defaultToken
}

func NewSymbolToken(value string, line, pos int) Token {
	return &SymbolToken{TokenSymbol, defaultToken{value, "symbol", line, pos}}
}

type StringConstantToken struct {
	TokenType
	defaultToken
}

func NewStringConstantToken(value string, line, pos int) Token {
	return &StringConstantToken{
		TokenStringConst,
		defaultToken{value, "stringConstant", line, pos},
	}
}

type IntegerConstantToken struct {
	TokenType
	defaultToken
}

func NewIntegerConstantToken(value string, line, pos int) Token {
	return &IntegerConstantToken{
		TokenIntegerConst,
		defaultToken{value, "integerConstant", line, pos},
	}
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
	xml    *XmlBuilder
	line   int
	pos    int
}

func NewTokenizer(r *bufio.Reader) *Tokenizer {
	sb := strings.Builder{}
	xb := NewXmlBuilderZero()
	xb.Open("tokens")
	t := Tokenizer{reader: r, buf: sb, xml: xb, line: 1}
	return &t
}

func (t *Tokenizer) ReadToken() (Token, error) {
	first, err := t.skipSpaces()
	if err != nil {
		return nil, err
	}

	first, err = t.skipComment(first)
	if err != nil {
		return nil, err
	}

	var newTk Token
	startPos := t.pos
	switch {
	case symbols[first]:
		newTk = NewSymbolToken(string(first), t.line, startPos)
	case first == '"':
		word := t.readStringToken()
		newTk = NewStringConstantToken(word, t.line, startPos)
	case unicode.IsLetter(rune(first)):
		word, err := t.readWord(first)
		if err != nil {
			return nil, err
		}
		if keywords[word] {
			newTk = NewKeywordToken(word, t.line, startPos)
		} else {
			newTk = NewIdentifierToken(word, t.line, startPos)
		}
	case unicode.IsNumber(rune(first)):
		word, err := t.readWord(first)
		if err != nil {
			return nil, err
		}
		newTk = NewIntegerConstantToken(word, t.line, startPos)
	}

	if newTk != nil {
		t.xml.WriteToken(newTk)
		return newTk, nil
	}
	return nil, fmt.Errorf("Line: %d, Pos: %d undefined token type", t.line, t.pos)
}

func (t *Tokenizer) WriteXml(wr *bufio.Writer) {
	t.xml.Close()
	wr.WriteString(t.xml.String())
}

func (t *Tokenizer) nextByte() (byte, error) {
	b, err := t.reader.ReadByte()
	if err == nil {
		t.pos++
	}
	return b, err
}

func (t *Tokenizer) nextLine() {
	t.line++
	t.pos = 0
}

func (t *Tokenizer) skipSpaces() (ch byte, err error) {
	for {
		ch, err = t.nextByte()
		if isEOL(ch) {
			t.nextLine()
			continue
		}
		if !isSpace(ch) {
			return
		}
	}
}

func (t *Tokenizer) isInlineComment(first byte) bool {
	if first != '/' {
		return false
	}
	next, err := t.reader.Peek(1)
	return err == nil && next[0] == '/'
}

func (t *Tokenizer) isMultiLineComment(first byte) bool {
	if first != '/' {
		return false
	}
	next, err := t.reader.Peek(1)
	return err == nil && next[0] == '*'
}

func (t *Tokenizer) skipInlineComment() (byte, error) {
	_, err := t.reader.ReadBytes('\n')
	if err != nil {
		return 0, err
	}
	t.nextLine()
	return t.skipSpaces()
}

func (t *Tokenizer) skipMultilineComment() (after byte, err error) {
	for {
		first, err := t.nextByte()
		if err != nil {
			return 0, err
		}
		if first == '*' {
			next, err := t.nextByte()
			if err != nil {
				return 0, err
			}
			if next == '/' {
				break
			}
		}
		if first == '\n' {
			t.nextLine()
		}
	}
	return t.skipSpaces()
}

func (t *Tokenizer) skipComment(first byte) (after byte, err error) {
	for {
		if !t.isInlineComment(first) && !t.isMultiLineComment(first) {
			return first, nil
		}

		if t.isInlineComment(first) {
			first, err = t.skipInlineComment()
			if err != nil {
				return first, err
			}
		}

		if t.isMultiLineComment(first) {
			first, err = t.skipMultilineComment()
			if err != nil {
				return first, err
			}
		}
	}
}

func (t *Tokenizer) readWord(fb byte) (string, error) {
	t.buf.Reset()
	t.buf.WriteByte(fb)
	for {
		next, err := t.reader.Peek(1)
		if err == nil && (isSpace(next[0]) || symbols[next[0]] || isEOL(next[0])) {
			break
		}
		// Just break in case of EOF in order to return the word
		if errors.Is(err, io.EOF) {
			break
		}
		// If peak do not return EOF error, than ReadByte will be ok
		if ch, _ := t.nextByte(); ch != 0 {
			t.buf.WriteByte(ch)
		}
	}
	word := t.buf.String()
	return word, nil
}

func (t *Tokenizer) readStringToken() string {
	t.buf.Reset()

	for {
		ch, err := t.nextByte()
		if err != nil || ch == '"' {
			break
		}
		if ch == '\n' {
			t.nextLine()
		}
		t.buf.WriteByte(ch)
	}
	return t.buf.String()
}
