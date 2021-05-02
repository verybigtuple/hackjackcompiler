package main

type NodeType int

func (nt NodeType) Type() NodeType {
	return nt
}

type Node interface {
	Type() NodeType
	Xml(xb *XmlBuilder)
}

const (
	NodeVarDec NodeType = iota
)

type VarDecNode struct {
	NodeType
	VarType Token
	Ids     []Token
}

func NewVarDecNode(vType Token, id Token) *VarDecNode {
	nt := VarDecNode{NodeType: NodeVarDec, VarType: vType}
	nt.Ids = append(nt.Ids, id)
	return &nt
}

func (vdn *VarDecNode) AddId(tk Token) {
	vdn.Ids = append(vdn.Ids, tk)
}

func (vdn *VarDecNode) IsClass() bool {
	return vdn.VarType.Type() == TokenIdentifier
}

func (vdn *VarDecNode) Xml(xb *XmlBuilder) {
	xb.Open("varDec")
	defer xb.Close()

	xb.WriteKeyword("var")
	xb.WriteToken(vdn.VarType)
	xb.WriteToken(vdn.Ids[0])
	if len(vdn.Ids) > 1 {
		for _, id := range vdn.Ids[1:] {
			xb.WriteSymbol(",")
			xb.WriteToken(id)
		}
	}
	xb.WriteSymbol(";")
}
