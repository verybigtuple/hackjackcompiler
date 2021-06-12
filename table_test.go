package main

import "testing"

func TestSimpleCase(t *testing.T) {
	tbl := NewSymbolTable("test")
	err := tbl.AddVar(Local, "int", "varName")
	if err != nil {
		t.Fatalf("Unexpected error occured: %v", err)
	}

	want := VarInfo{Local, "int", 0}
	got, err := tbl.GetVarInfo("varName")
	if err != nil {
		t.Fatal("Var not found")
	}
	if got != want {
		t.Fatalf("Got: %+v; want: %+v", got, want)
	}
}

type addVarTest struct {
	Idx   int
	Kind  VarKind
	VType string
	Name  string
}

func TestCounters(t *testing.T) {
	vars := [...]addVarTest{
		{0, Field, "int", "f0"},
		{1, Field, "string", "f1"},
		{2, Field, "bool", "f2"},

		{0, Static, "int", "s0"},
		{1, Static, "string", "s1"},
		{2, Static, "bool", "s2"},

		{0, Arg, "int", "arg0"},
		{1, Arg, "string", "arg1"},
		{2, Arg, "bool", "arg2"},

		{0, Local, "int", "local0"},
		{1, Local, "string", "local1"},
		{2, Local, "bool", "local2"},
	}

	tbl := NewSymbolTable("test")

	for _, v := range vars {
		err := tbl.AddVar(v.Kind, v.VType, v.Name)
		if err != nil {
			t.Fatalf("Unexpected error while adding var %+v: %v", v, err)
		}
	}

	for _, v := range vars {
		if vi, err := tbl.GetVarInfo(v.Name); err == nil {
			if vi.Offset != v.Idx {
				t.Errorf("%s: Got %d; want %d", v.Name, vi.Offset, v.Idx)
			}
		} else {
			t.Errorf("Error while finding %s: %v", v.Name, err)
		}
	}
}

func TestListCreation(t *testing.T) {
	tblList := NewSymbolTableList()
	tblList.CreateTable("root")
	tblList.CreateTable("child")

	if tblList.Len() != 2 {
		t.Error("Expected len 2")
	}
}

func TestListVarInfo(t *testing.T) {
	tblList := NewSymbolTableList()
	tblList.CreateTable("root")
	tblList.AddVar(Field, "string", "Root0")
	tblList.CreateTable("child")
	tblList.AddVar(Local, "string", "Child0")

	tblList.GetVarInfo("Child0")
	tblList.GetVarInfo("Root0")

	// Close Child
	tblList.GetVarInfo("Root0")
}
