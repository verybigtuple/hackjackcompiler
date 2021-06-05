package main

import (
	"errors"
	"fmt"
)

type VarKind int

const (
	Field VarKind = iota
	Static
	Arg
	Local
)

type VarInfo struct {
	Kind   VarKind
	Type   string
	Offset int
}

type SymbolTable struct {
	Name    string
	table   map[string]VarInfo
	counter map[VarKind]int
}

func NewSymbolTable(name string) *SymbolTable {
	table := make(map[string]VarInfo)
	counter := map[VarKind]int{
		Field:  0,
		Static: 0,
		Arg:    0,
		Local:  0,
	}

	return &SymbolTable{name, table, counter}
}

func (st *SymbolTable) AddVar(kind VarKind, vType, name string) error {
	if _, ok := st.table[name]; ok {
		return fmt.Errorf("Var named %s already in the symbol table %s", name, st.Name)
	}
	st.table[name] = VarInfo{kind, vType, st.counter[kind]}
	st.counter[kind] = st.counter[kind] + 1
	return nil
}

func (st *SymbolTable) GetVarInfo(name string) (vi VarInfo, err error) {
	var ok bool
	if vi, ok = st.table[name]; !ok {
		err = errors.New("Not found")
	}
	return
}

type SymbolTableList struct {
	list []*SymbolTable
}

func NewSymbolTableList() *SymbolTableList {
	list := make([]*SymbolTable, 0)
	return &SymbolTableList{list}
}

func (stl *SymbolTableList) CreateTable(name string) {
	tbl := NewSymbolTable(name)
	stl.list = append(stl.list, tbl)
}

func (stl *SymbolTableList) CloseTable() {
	if len(stl.list) > 0 {
		stl.list = stl.list[:len(stl.list)-1]
	}
}

func (stl *SymbolTableList) Len() int {
	return len(stl.list)
}

func (stl *SymbolTableList) AddVar(kind VarKind, vType, name string) error {
	if len(stl.list) == 0 {
		panic("Symbol table list is empty")
	}
	tbl := stl.list[len(stl.list)-1]
	err := tbl.AddVar(kind, vType, name)
	return err
}

func (stl *SymbolTableList) GetVarInfo(name string) (VarInfo, error) {
	for i := len(stl.list) - 1; i >= 0; i-- {
		tbl := stl.list[i]
		if vi, err := tbl.GetVarInfo(name); err == nil {
			return vi, nil
		}
	}
	return VarInfo{}, fmt.Errorf("Variable %s is not declared", name)
}
