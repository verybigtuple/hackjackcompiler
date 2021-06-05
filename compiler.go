package main

type Compiler struct {
	SymbolTblList *SymbolTableList
}

func NewCompiler() *Compiler {
	stList := NewSymbolTableList()
	return &Compiler{stList}
}
