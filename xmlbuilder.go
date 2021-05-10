package main

import (
	"container/list"
	"strings"
)

const defaultIndent = 2

type stack struct {
	data *list.List
}

func newStack() *stack {
	return &stack{list.New()}
}

func (s *stack) Len() int {
	return s.data.Len()
}

func (s *stack) Push(name string) {
	s.data.PushFront(name)
}

func (s *stack) Pop() string {
	if s.Len() == 0 {
		panic("Xml stack is empty")
	}
	fe := s.data.Front()
	val := s.data.Remove(fe)
	return val.(string)
}

var xmlReplacer = strings.NewReplacer(
	"<", "&lt;",
	">", "&gt;",
	"\"", "&quot;",
	"&", "&amp;",
)

type Xmler interface {
	Xml(xb *XmlBuilder)
}

type XmlBuilder struct {
	sb     *strings.Builder
	st     *stack
	indent int
}

func NewXmlBuilder() *XmlBuilder {
	sb := &strings.Builder{}
	st := newStack()
	return &XmlBuilder{sb, st, defaultIndent}
}

func NewXmlBuilderZero() *XmlBuilder {
	sb := &strings.Builder{}
	st := newStack()
	return &XmlBuilder{sb, st, 0}
}

func (xb *XmlBuilder) String() string {
	return xb.sb.String()
}

func (xb *XmlBuilder) Open(name string) {
	spaces := strings.Repeat(" ", xb.st.Len()*xb.indent)
	xb.sb.WriteString(spaces)
	xb.sb.WriteString("<" + name + ">")
	xb.sb.WriteString("\n")
	xb.st.Push(name)
}

func (xb *XmlBuilder) Close() {
	name := xb.st.Pop()
	spaces := strings.Repeat(" ", xb.st.Len()*xb.indent)
	xb.sb.WriteString(spaces)
	xb.sb.WriteString("</" + name + ">")
	xb.sb.WriteString("\n")
}

func (xb *XmlBuilder) WriteNode(tag, val string) {
	spaces := strings.Repeat(" ", xb.st.Len()*xb.indent)
	nv := xmlReplacer.Replace(val)

	if len(spaces) > 0 {
		xb.sb.WriteString(spaces)
	}
	xb.sb.WriteString("<" + tag + ">")
	xb.sb.WriteString(" ")
	xb.sb.WriteString(nv)
	xb.sb.WriteString(" ")
	xb.sb.WriteString("</" + tag + ">")
	xb.sb.WriteString("\n")
}

func (xb *XmlBuilder) WriteToken(tk Token) {
	tk.Xml(xb)
}

func (xb *XmlBuilder) WriteKeyword(v string) {
	tk := NewKeywordToken(v)
	xb.WriteToken(tk)
}

func (xb *XmlBuilder) WriteSymbol(v string) {
	tk := NewSymbolToken(v)
	xb.WriteToken(tk)
}
