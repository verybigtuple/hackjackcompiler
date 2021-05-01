package main

import "fmt"

type ParseTree struct {
	tz      *Tokenizer
	current Token
}

func NewPasreTree(tz *Tokenizer) *ParseTree {
	return &ParseTree{tz: tz, current: nil}
}

func (t *ParseTree) Parse() (Node, error) {
	root, err := t.varDec()
	return root, err
}

func (t *ParseTree) feedKeyword(val string) (kw *KeywordToken, err error) {
	if k, ok := t.current.(*KeywordToken); ok && k.GetValue() == val {
		t.current, err = t.tz.ReadToken()
		kw = k
	} else {
		err = fmt.Errorf("Expected Keyword %s", val)
	}
	return
}

func (t *ParseTree) feedIdent() (id *IdentifierToken, err error) {
	if k, ok := t.current.(*IdentifierToken); ok {
		t.current, err = t.tz.ReadToken()
		id = k
	} else {
		err = fmt.Errorf("Expected Identifier")
	}
	return
}

func (t *ParseTree) feedType() (Token, error) {
	switch tk := t.current.(type) {
	case *KeywordToken:
		val := tk.GetValue()
		if val == "int" || val == "char" || val == "bool" {
			var err error
			t.current, err = t.tz.ReadToken()
			return tk, err
		}
	case *IdentifierToken:
		var err error
		t.current, err = t.tz.ReadToken()
		return tk, err
	}
	return nil, fmt.Errorf("Expected type")
}

func (t *ParseTree) varDec() (vd *VarDecNode, err error) {
	t.current, err = t.tz.ReadToken()
	if err != nil {
		return
	}

	_, err = t.feedKeyword("var")
	if err != nil {
		return
	}

	var typeToken Token
	typeToken, err = t.feedType()
	if err != nil {
		return
	}

	var idToken *IdentifierToken
	idToken, err = t.feedIdent()
	if err != nil {
		return
	}

	vd = NewVarDecNode(typeToken, idToken)
	return
}
