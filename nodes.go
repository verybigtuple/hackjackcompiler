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
	NodeLetStatement
	NodeExpression
	NodeTerm
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

type LetStatementNode struct {
	NodeType
	VarName  Token
	ArrayExp *ExpressionNode
	ValueExp *ExpressionNode
}

func NewLetStatementNode(varName Token, arrExp *ExpressionNode, valExp *ExpressionNode) *LetStatementNode {
	lsn := LetStatementNode{
		NodeType: NodeLetStatement,
		VarName:  varName,
		ArrayExp: arrExp,
		ValueExp: valExp,
	}
	return &lsn
}

func (lsn *LetStatementNode) Xml(xb *XmlBuilder) {
	xb.Open("letStatement")
	defer xb.Close()

	xb.WriteKeyword("let")
	xb.WriteToken(lsn.VarName)
	if lsn.ArrayExp != nil {
		xb.WriteSymbol("[")
		lsn.ArrayExp.Xml(xb)
		xb.WriteSymbol("]")
	}
	xb.WriteSymbol("=")
	lsn.ValueExp.Xml(xb)
}

type ExpressionNode struct {
	NodeType
	term    *TermNode
	ops     []Token
	opTerms []*TermNode
}

func NewExpressionNode(term *TermNode) *ExpressionNode {
	return &ExpressionNode{NodeType: NodeExpression, term: term}
}

func (en *ExpressionNode) AddOpTerm(op Token, term *TermNode) {
	en.ops = append(en.ops, op)
	en.opTerms = append(en.opTerms, term)
}

func (en *ExpressionNode) Xml(xb *XmlBuilder) {
	xb.Open("expression")
	defer xb.Close()

	en.term.Xml(xb)
	if len(en.ops) > 0 {
		if len(en.ops) != len(en.opTerms) {
			panic("Expression node is build wrong in operations and terms")
		}
		for i, op := range en.ops {
			xb.WriteToken(op)
			en.opTerms[i].Xml(xb)
		}
	}
}

type termNodeType int

const (
	termNodeConst termNodeType = iota
	termNodeVar
	termNodeArray
	termNodeExpr
	termNodeCall
	termNodeUnary
)

type TermNode struct {
	NodeType
	termType  termNodeType
	jConst    Token
	jVar      Token
	arrayIdx  *ExpressionNode
	exp       *ExpressionNode
	unaryOp   Token
	unaryTerm *TermNode
}

func NewConstTermNode(jConst Token) *TermNode {
	return &TermNode{NodeType: NodeTerm, termType: termNodeConst, jConst: jConst}
}

func NewVarTermNode(jVar Token) *TermNode {
	return &TermNode{NodeType: NodeTerm, termType: termNodeVar, jVar: jVar}
}

func NewArrayTermNode(jVar Token, idx *ExpressionNode) *TermNode {
	return &TermNode{NodeType: NodeTerm, termType: termNodeArray, jVar: jVar, arrayIdx: idx}
}

func NewExpressionTermNode(exp *ExpressionNode) *TermNode {
	return &TermNode{NodeType: NodeTerm, termType: termNodeExpr, exp: exp}
}

// func NewCallTermNode(jVar Token, idx *TermNode) *TermNode {
// }

func NewUnaryTermNode(op Token, term *TermNode) *TermNode {
	return &TermNode{NodeType: NodeTerm, termType: termNodeUnary, unaryOp: op, unaryTerm: term}
}

func (tn *TermNode) Xml(xb *XmlBuilder) {
	xb.Open("term")
	defer xb.Close()

	switch tn.termType {
	case termNodeConst:
		xb.WriteToken(tn.jConst)
	case termNodeVar:
		xb.WriteToken(tn.jVar)
	case termNodeArray:
		xb.WriteToken(tn.jVar)
		xb.WriteSymbol("[")
		tn.arrayIdx.Xml(xb)
		xb.WriteSymbol("]")
	case termNodeExpr:
		xb.WriteSymbol("(")
		tn.exp.Xml(xb)
		xb.WriteSymbol(")")
	case termNodeUnary:
		xb.WriteToken(tn.unaryOp)
		tn.unaryTerm.Xml(xb)
	default:
		panic("Xml is not defined for the type of node")
	}
}
