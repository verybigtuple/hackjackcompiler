package main

type NodeType int

func (nt NodeType) Type() NodeType {
	return nt
}

type Node interface {
	Type() NodeType
}

const (
	NodeVarDec NodeType = iota
)

type VarDecNode struct {
	NodeType
	VarType Token
	Ids     []*IdentifierToken
}

func NewVarDecNode(vType Token, id *IdentifierToken) *VarDecNode {
	nt := VarDecNode{NodeType: NodeVarDec, VarType: vType}
	nt.Ids = append(nt.Ids, id)
	return &nt
}

func (t *VarDecNode) IsClass() bool {
	_, ok := t.VarType.(*IdentifierToken)
	return ok
}
