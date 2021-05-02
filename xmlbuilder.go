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

type XmlBuilder struct {
	sb *strings.Builder
	st *stack
}

func NewXmlBuilder() *XmlBuilder {
	sb := &strings.Builder{}
	st := newStack()
	return &XmlBuilder{sb, st}
}

func (xb *XmlBuilder) String() string {
	return xb.sb.String()
}

func (xb *XmlBuilder) Open(name string) {
	spaces := strings.Repeat(" ", xb.st.Len()*defaultIndent)
	xb.sb.WriteString(spaces)
	xb.sb.WriteString("<" + name + ">")
	xb.sb.WriteString("\n")
	xb.st.Push(name)
}

func (xb *XmlBuilder) Close() {
	name := xb.st.Pop()
	spaces := strings.Repeat(" ", xb.st.Len()*defaultIndent)
	xb.sb.WriteString(spaces)
	xb.sb.WriteString("</" + name + ">")
	xb.sb.WriteString("\n")
}

func (xb *XmlBuilder) WriteToken(tk Token) {
	ident := xb.st.Len() * defaultIndent
	spaces := strings.Repeat(" ", ident)
	xb.sb.WriteString(spaces)
	xb.sb.WriteString(tk.GetXml())
	xb.sb.WriteString("\n")
}

func (xb *XmlBuilder) WriteKeyword(v string) {
	tk := NewKeywordToken(v)
	xb.WriteToken(tk)
}

func (xb *XmlBuilder) WriteSymbol(v string) {
	tk := NewSymbolToken(v)
	xb.WriteToken(tk)
}
