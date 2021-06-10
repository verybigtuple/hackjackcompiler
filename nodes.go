package main

type NodeType int

func (nt NodeType) Type() NodeType {
	return nt
}

type Node interface {
	Type() NodeType
	Xml(xb *XmlBuilder)
	Compile(c *Compiler)
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
	xb.WriteSymbol("{")
	for _, vd := range cn.VarDec {
		vd.Xml(xb)
	}
	for _, sd := range cn.SbrDec {
		sd.Xml(xb)
	}
	xb.WriteSymbol("}")
}

func (cn *ClassNode) Compile(c *Compiler) {
	c.SymbolTblList.CreateTable(cn.Name.GetValue())
	defer c.SymbolTblList.CloseTable()

	for _, sbr := range cn.SbrDec {
		sbr.Compile(c)
	}
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
	for i, n := range cvd.VarNames {
		if i > 0 {
			xb.WriteSymbol(",")
		}
		xb.WriteToken(n)
	}
	xb.WriteSymbol(";")
}

func (cvd *ClassVarDecNode) Compile(c *Compiler) {}

type SubroutineDecNode struct {
	NodeType
	SbrKind    Token
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

	xb.WriteToken(sdn.SbrKind)
	xb.WriteToken(sdn.ReturnType)
	xb.WriteToken(sdn.Name)
	xb.WriteSymbol("(")
	sdn.ParamList.Xml(xb)
	xb.WriteSymbol(")")
	sdn.Body.Xml(xb)
}

func (sdn *SubroutineDecNode) Compile(c *Compiler) {
	fn := c.SymbolTblList.Name() + "." + sdn.Name.GetValue()
	c.SymbolTblList.CreateTable(fn)
	defer c.SymbolTblList.CloseTable()

	c.Function(fn, sdn.Body.LocalVarLen())
	sdn.ParamList.Compile(c)
	sdn.Body.Compile(c)
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

func (pln *ParameterListNode) Compile(c *Compiler) {
	for i, vt := range pln.varTypes {
		vn := pln.varNames[i]
		c.SymbolTblList.AddVar(Arg, vt.GetValue(), vn.GetValue())
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

func (sbn *SubroutineBodyNode) LocalVarLen() int {
	var lvSum int
	for _, vd := range sbn.VarDec {
		lvSum += vd.Len()
	}
	return lvSum
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

func (sbn *SubroutineBodyNode) Compile(c *Compiler) {
	for _, vd := range sbn.VarDec {
		vd.Compile(c)
	}
	sbn.Statm.Compile(c)
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

func (vdn *VarDecNode) Len() int {
	return len(vdn.Ids)
}

func (vdn *VarDecNode) Xml(xb *XmlBuilder) {
	xb.Open("varDec")
	defer xb.Close()

	xb.WriteKeyword("var")
	xb.WriteToken(vdn.VarType)
	for i, id := range vdn.Ids {
		if i > 0 {
			xb.WriteSymbol(",")
		}
		xb.WriteToken(id)
	}

	xb.WriteSymbol(";")
}

func (vdn *VarDecNode) Compile(c *Compiler) {
	for _, id := range vdn.Ids {
		c.SymbolTblList.AddVar(Local, vdn.VarType.GetValue(), id.GetValue())
	}
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
	xb.WriteSymbol(";")
}

func (lsn *LetStatementNode) Compile(c *Compiler) {
	if lsn.ArrayExp == nil {
		lsn.ValueExp.Compile(c)
		vi, _ := c.SymbolTblList.GetVarInfo(lsn.VarName.GetValue())
		c.Pop(GetSegment(vi.Kind), strconv.Itoa(vi.Offset))
	}
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

func (sn *StatementsNode) Compile(c *Compiler) {
	for _, st := range sn.StList {
		st.Compile(c)
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

func (ifn *IfStatementNode) Compile(c *Compiler) {
	elseLabel, endLabel := c.OpenIf()

	ifn.IfExpr.Compile(c)
	c.UnaryOp("~")
	if ifn.ElseStat != nil {
		c.IfGoto(elseLabel)
	} else {
		c.IfGoto(endLabel)
	}
	ifn.IfStat.Compile(c)
	if ifn.ElseStat != nil {
		c.Goto(endLabel)
		c.Label(elseLabel)
		ifn.ElseStat.Compile(c)
	}
	c.Label(endLabel)
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

func (wsn *WhileStatementNode) Compile(c *Compiler) {
	bLabel, eLabel := c.OpenWhile()

	c.Label(bLabel)
	wsn.Expr.Compile(c)
	c.UnaryOp("~")
	c.IfGoto(eLabel)
	wsn.Stat.Compile(c)
	c.Goto(bLabel)
	c.Label(eLabel)
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
	xb.WriteKeyword("do")
	ds.Call.Xml(xb)
	xb.WriteSymbol(";")
}

func (ds *DoStatementNode) Compile(c *Compiler) {
	ds.Call.Compile(c)
	// Clean return from function
	c.Pop(TempSegm, "0")
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
	xb.Open("returnStatement")
	defer xb.Close()

	xb.WriteKeyword("return")
	if rsn.Expr != nil {
		rsn.Expr.Xml(xb)
	}
	xb.WriteSymbol(";")
}

func (rsn *ReturnStatementNode) Compile(c *Compiler) {
	if rsn.Expr != nil {
		rsn.Expr.Compile(c)
	} else {
		c.Push(ConstSegm, "0")
	}
	c.Return()
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

func (en *ExpressionNode) Compile(c *Compiler) {
	if len(en.ops) != len(en.opTerms) {
		panic("Expression node is build wrong in operations and terms")
	}

	en.term.Compile(c)
	if len(en.ops) > 0 {
		for i, op := range en.ops {
			en.opTerms[i].Compile(c)
			c.BinaryOp(op.GetValue())
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

func (eln *ExpressionListNode) Len() int {
	return len(eln.Exprs)
}

func (eln *ExpressionListNode) Xml(xb *XmlBuilder) {
	xb.Open("expressionList")
	defer xb.Close()

	for i, expr := range eln.Exprs {
		if i > 0 {
			xb.WriteSymbol(",")
		}
		expr.Xml(xb)
	}
}

func (eln *ExpressionListNode) Compile(c *Compiler) {
	for _, expr := range eln.Exprs {
		expr.Compile(c)
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

func (scn *SubroutineCallNode) Xml(xb *XmlBuilder) {
	// Due to some reason  Subrooutine call does not have open/close tag
	if scn.ClassName != nil {
		xb.WriteToken(scn.ClassName)
		xb.WriteSymbol(".")
	}
	xb.WriteToken(scn.SubroutineName)
	xb.WriteSymbol("(")
	scn.Params.Xml(xb)
	xb.WriteSymbol(")")
}

func (scn *SubroutineCallNode) Compile(c *Compiler) {
	scn.Params.Compile(c)

	var name string
	if scn.ClassName != nil {
		name = scn.ClassName.GetValue() + "." + scn.SubroutineName.GetValue()
	} else {
		name = scn.SubroutineName.GetValue()
	}

	ac := scn.Params.Len()
	c.Call(name, ac)
}

type termNodeType int

const (
	termNodeConst termNodeType = iota
	termNodeIntConst
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

func NewIntConstTermNode(intConst Token) *TermNode {
	return &TermNode{NodeType: NodeTerm, termType: termNodeIntConst, jConst: intConst}
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
	case termNodeConst, termNodeIntConst:
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

func (tn *TermNode) Compile(c *Compiler) {
	switch tn.termType {
	case termNodeIntConst:
		c.Push(ConstSegm, tn.jConst.GetValue())
	case termNodeExpr:
		tn.exp.Compile(c)
	}
}
