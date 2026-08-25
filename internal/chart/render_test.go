package chart

import (
	"bytes"
	"encoding/xml"
	"testing"

	"github.com/noversez/noversez/internal/langstats"
)

func TestRenderProducesValidEscapedSVG(t *testing.T) {
	got, err := Render([]langstats.Language{
		{Name: "Go & Tools", Bytes: 75},
		{Name: "Dart", Bytes: 25},
	}, map[string]string{
		"Go & Tools": "#00ADD8",
		"Dart":       "#00B4AB",
	})
	if err != nil {
		t.Fatal(err)
	}

	var root struct {
		XMLName xml.Name
	}
	if err := xml.Unmarshal(got, &root); err != nil {
		t.Fatalf("invalid SVG: %v", err)
	}
	if root.XMLName.Local != "svg" {
		t.Fatalf("root element = %q, want svg", root.XMLName.Local)
	}
	if !bytes.Contains(got, []byte("Go &amp; Tools")) {
		t.Fatal("language name was not XML-escaped")
	}
	if bytes.Contains(got, []byte("ZgotmplZ")) {
		t.Fatal("template rejected a chart value")
	}
	if !bytes.Contains(got, []byte("#00ADD8")) || !bytes.Contains(got, []byte("#00B4AB")) {
		t.Fatal("GitHub language colors are missing")
	}
}

func TestRenderHandlesEmptyData(t *testing.T) {
	got, err := Render(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(got, []byte("No language data")) {
		t.Fatal("empty-state message is missing")
	}
}

func TestRenderUsesFallbackForUnknownLanguage(t *testing.T) {
	got, err := Render([]langstats.Language{{Name: "Unknown", Bytes: 1}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(got, []byte(fallbackColor)) {
		t.Fatal("fallback color is missing")
	}
}
