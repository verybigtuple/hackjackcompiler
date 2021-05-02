package main

import "testing"

func TestOneLevel(t *testing.T) {
	want := "<level1>\n</level1>\n"

	xb := NewXmlBuilder()
	xb.Open("level1")
	xb.Close()
	got := xb.String()

	if want != got {
		t.Errorf("want:\n%s\ngot:\n%s", want, got)
	}
}

func TestTwoLevels(t *testing.T) {
	want := "<level1>\n  <level2>\n  </level2>\n</level1>\n"

	xb := NewXmlBuilder()
	xb.Open("level1")
	xb.Open("level2")
	xb.Close()
	xb.Close()
	got := xb.String()
	if want != got {
		t.Errorf("want:\n%s\ngot:\n%s", want, got)
	}
}

func TestTokens(t *testing.T) {
	want := "<level1>\n  <level2>\n    <keyword> var </keyword>\n  </level2>\n</level1>\n"

	xb := NewXmlBuilder()
	xb.Open("level1")
	xb.Open("level2")
	xb.WriteKeyword("var")
	xb.Close()
	xb.Close()
	got := xb.String()
	if want != got {
		t.Errorf("want:\n%s\ngot:\n%s", want, got)
	}
}
