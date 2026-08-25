package linguist

import (
	"strings"
	"testing"
)

func TestParseColors(t *testing.T) {
	input := `
---
Dart:
  type: programming
  color: "#00B4AB"
Go:
  type: programming
  color: "#00ADD8"
'Quoted Language':
  color: '#123abc'
No Color:
  type: data
`

	colors, err := ParseColors(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"Dart":            "#00B4AB",
		"Go":              "#00ADD8",
		"Quoted Language": "#123abc",
	}
	for language, color := range want {
		if colors[language] != color {
			t.Fatalf("%s color = %q, want %q", language, colors[language], color)
		}
	}
}
