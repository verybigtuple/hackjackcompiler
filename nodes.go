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
	NodeClass NodeType = iota
	NodeClassVarDec
	NodeSubroutineDec
	NodeParameterList
	NodeSubroutineBody
	NodeVarDec
	NodeStatements
	NodeLetStatement
	NodeIfStatement
	NodeWhileStatement
	NodeDoStatement
	NodeReturnStatement
	NodeExpression
	NodeTerm
	NodeSubroutineCall
	NodeExpressionList
)

type ClassNode struct {
	NodeType
	Name   Token
	VarDec []*ClassVarDecNode
	SbrDec []*SubroutineDecNode
}

func NewClassNode(name Token) *ClassNode {
	return &ClassNode{NodeType: NodeClass, Name: name}
}

func (cn *ClassNode) AddVarDecs(vd ...*ClassVarDecNode) {
	cn.VarDec = append(cn.VarDec, vd...)
}

func (cn *ClassNode) AddSbrDecs(sbrd ...*SubroutineDecNode) {
	cn.SbrDec = append(cn.SbrDec, sbrd...)
}

func (cn *ClassNode) Xml(xb *XmlBuilder) {
	xb.Open("class")
	defer xb.Close()

	xb.WriteKeyword("class")
	xb.WriteToken(cn.Name)
	xb.WriteKeyword("{")
	for _, vd := range cn.VarDec {
		vd.Xml(xb)
	}
	for _, sd := range cn.SbrDec {
		sd.Xml(xb)
	}
	xb.WriteKeyword("}")
}

type ClassVarDecNode struct {
	NodeType
	VarClass Token
	VarType  Token
	VarNames []Token
}

func NewClassVarDecNode(vc Token, vt Token, name Token) *ClassVarDecNode {
	cvd := ClassVarDecNode{NodeType: NodeClassVarDec, VarClass: vc, VarType: vt}
	cvd.VarNames = append(cvd.VarNames, name)
	return &cvd
}

func (cvd *ClassVarDecNode) AddVarNames(names ...Token) {
	cvd.VarNames = append(cvd.VarNames, names...)
}

func (cvd *ClassVarDecNode) Xml(xb *XmlBuilder) {
	xb.Open("classVarDec")
	defer xb.Close()

	xb.WriteToken(cvd.VarClass)
	xb.WriteToken(cvd.VarType)
	xb.WriteToken(cvd.VarNames[0])
	if len(cvd.VarNames) > 1 {
		for _, n := range cvd.VarNames {
			xb.WriteSymbol(",")
			xb.WriteToken(n)
		}
	}
	xb.WriteSymbol(";")
}

type SubroutineDecNode struct {
	NodeType
	SbrClass   Token
	ReturnType Token
	Name       Token
	ParamList  *ParameterListNode
	Body       *SubroutineBodyNode
}

func NewSubroutineDecNode(sc Token, rt Token, name Token, param *ParameterListNode, b *SubroutineBodyNode) *SubroutineDecNode {
	return &SubroutineDecNode{NodeSubroutineDec, sc, rt, name, param, b}
}

func (sdn *SubroutineDecNode) Xml(xb *XmlBuilder) {
	xb.Open("subroutineDec")
	defer xb.Close()

	xb.WriteToken(sdn.SbrClass)
	xb.WriteToken(sdn.ReturnType)
	xb.WriteToken(sdn.Name)
	xb.WriteSymbol("(")
	sdn.ParamList.Xml(xb)
	xb.WriteSymbol(")")
	sdn.Body.Xml(xb)
}

type ParameterListNode struct {
	NodeType
	varTypes []Token
	varNames []Token
}

func NewParameterListNode() *ParameterListNode {
	return &ParameterListNode{NodeType: NodeParameterList}
}

func (pln *ParameterListNode) AddParameter(varType Token, varName Token) {
	pln.varTypes = append(pln.varTypes, varType)
	pln.varNames = append(pln.varNames, varName)
}

func (pln *ParameterListNode) Xml(xb *XmlBuilder) {
	if len(pln.varTypes) != len(pln.varNames) {
		panic("ParameterListNode is built wrong")
	}

	xb.Open("parameterList")
	defer xb.Close()

	if len(pln.varTypes) > 0 {
		xb.WriteToken(pln.varTypes[0])
		xb.WriteToken(pln.varNames[0])
		for i, t := range pln.varTypes[1:] {
			xb.WriteSymbol(",")
			xb.WriteToken(t)
			xb.WriteToken(pln.varNames[i+1])
		}
	}
}

type SubroutineBodyNode struct {
	NodeType
	VarDec []*VarDecNode
	Statm  *StatementsNode
}

func NewSubroutineBodyNode(statm *StatementsNode) *SubroutineBodyNode {
	return &SubroutineBodyNode{NodeType: NodeSubroutineBody, Statm: statm}
}

func (sbn *SubroutineBodyNode) AddVarDec(vd ...*VarDecNode) {
	sbn.VarDec = append(sbn.VarDec, vd...)
}

func (sbn *SubroutineBodyNode) Xml(xb *XmlBuilder) {
	xb.Open("subroutineBody")
	defer xb.Close()

	xb.WriteSymbol("{")
	for _, vd := range sbn.VarDec {
		vd.Xml(xb)
	}
	sbn.Statm.Xml(xb)
	xb.WriteSymbol("}")
}

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

func NewLetStatementNode(varName Token, valExp *ExpressionNode) *LetStatementNode {
	lsn := LetStatementNode{
		NodeType: NodeLetStatement,
		VarName:  varName,
		ValueExp: valExp,
	}
	return &lsn
}

func (lsn *LetStatementNode) AddArrayExpr(arrExp *ExpressionNode) {
	lsn.ArrayExp = arrExp
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

type StatementsNode struct {
	NodeType
	StList []Node
}

func NewStatementsNode() *StatementsNode {
	return &StatementsNode{NodeType: NodeStatements}
}

func (sn *StatementsNode) AddSt(stat Node) {
	sn.StList = append(sn.StList, stat)
}

func (sn *StatementsNode) Xml(xb *XmlBuilder) {
	xb.Open("statements")
	defer xb.Close()

	for _, s := range sn.StList {
		s.Xml(xb)
	}
}

type IfStatementNode struct {
	NodeType
	IfExpr   *ExpressionNode
	IfStat   *StatementsNode
	ElseStat *StatementsNode // can be nil
}

func NewIfStatementNode(ifExpr *ExpressionNode, ifSt *StatementsNode) *IfStatementNode {
	return &IfStatementNode{NodeType: NodeIfStatement, IfExpr: ifExpr, IfStat: ifSt}
}

func (ifn *IfStatementNode) AddElse(elseSt *StatementsNode) {
	ifn.ElseStat = elseSt
}

func (ifn *IfStatementNode) Xml(xb *XmlBuilder) {
	xb.Open("ifStatement")
	defer xb.Close()

	xb.WriteKeyword("if")
	xb.WriteSymbol("(")
	ifn.IfExpr.Xml(xb)
	xb.WriteSymbol(")")
	xb.WriteSymbol("{")
	ifn.IfStat.Xml(xb)
	xb.WriteSymbol("}")
	if ifn.ElseStat != nil {
		xb.WriteKeyword("else")
		xb.WriteSymbol("{")
		ifn.ElseStat.Xml(xb)
		xb.WriteSymbol("}")
	}
}

type WhileStatementNode struct {
	NodeType
	Expr *ExpressionNode
	Stat *StatementsNode
}

func NewWhileStatementNode(expr *ExpressionNode, stat *StatementsNode) *WhileStatementNode {
	return &WhileStatementNode{NodeWhileStatement, expr, stat}
}

func (wsn *WhileStatementNode) Xml(xb *XmlBuilder) {
	xb.Open("whileStatement")
	defer xb.Close()

	xb.WriteKeyword("while")
	xb.WriteSymbol("(")
	wsn.Expr.Xml(xb)
	xb.WriteSymbol(")")
	xb.WriteSymbol("{")
	wsn.Stat.Xml(xb)
	xb.WriteSymbol("}")
}

type DoStatementNode struct {
	NodeType
	Call *SubroutineCallNode
}

func NewDoStatementNode(call *SubroutineCallNode) *DoStatementNode {
	return &DoStatementNode{NodeDoStatement, call}
}

func (ds *DoStatementNode) Xml(xb *XmlBuilder) {
	xb.Open("doStatement")
	defer xb.Close()
	ds.Call.Xml(xb)
	xb.WriteSymbol(";")
}

type ReturnStatementNode struct {
	NodeType
	Expr *ExpressionNode
}

func NewReturnNode() *ReturnStatementNode {
	return &ReturnStatementNode{NodeType: NodeReturnStatement}
}

func (rsn *ReturnStatementNode) AddExpr(expr *ExpressionNode) {
	rsn.Expr = expr
}

func (rsn *ReturnStatementNode) Xml(xb *XmlBuilder) {
	xb.WriteKeyword("return")
	if rsn.Expr != nil {
		rsn.Expr.Xml(xb)
	}
	xb.WriteSymbol(";")
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

type ExpressionListNode struct {
	NodeType
	Exprs []*ExpressionNode
}

func NewExpressionListNode() *ExpressionListNode {
	return &ExpressionListNode{NodeType: NodeExpressionList}
}

func (eln *ExpressionListNode) AddExpr(expr *ExpressionNode) {
	eln.Exprs = append(eln.Exprs, expr)
}

func (eln *ExpressionListNode) Xml(xb *XmlBuilder) {
	xb.Open("expressionList")
	defer xb.Close()

	for _, expr := range eln.Exprs {
		expr.Xml(xb)
	}
}

type SubroutineCallNode struct {
	NodeType
	ClassName      Token
	SubroutineName Token
	Params         *ExpressionListNode
}

func NewClassSubroutineCallNode(clsName Token, sbrName Token, params *ExpressionListNode) *SubroutineCallNode {
	return &SubroutineCallNode{NodeSubroutineCall, clsName, sbrName, params}
}

func NewSubroutineCallNode(sbrName Token, params *ExpressionListNode) *SubroutineCallNode {
	return &SubroutineCallNode{NodeType: NodeSubroutineCall, SubroutineName: sbrName, Params: params}
}

func (sc *SubroutineCallNode) Xml(xb *XmlBuilder) {
	xb.Open("subroutineCall")
	defer xb.Close()

	if sc.ClassName != nil {
		xb.WriteToken(sc.ClassName)
		xb.WriteSymbol(".")
	}
	xb.WriteToken(sc.SubroutineName)
	xb.WriteSymbol("(")
	sc.Params.Xml(xb)
	xb.WriteSymbol(")")
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
	call      *SubroutineCallNode
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

func NewCallTermNode(call *SubroutineCallNode) *TermNode {
	return &TermNode{NodeType: NodeTerm, termType: termNodeCall, call: call}
}

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
	case termNodeCall:
		tn.call.Xml(xb)
	default:
		panic("Xml is not defined for the type of node")
	}
}
