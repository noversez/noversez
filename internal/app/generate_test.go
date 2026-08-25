package app

import (
	"bytes"
	"context"
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"

	"github.com/noversez/noversez/internal/githubapi"
)

type fakeSource struct {
	repositories []githubapi.Repository
	languages    map[string]map[string]int64
}

func (f fakeSource) ListRepositories(context.Context) ([]githubapi.Repository, error) {
	return f.repositories, nil
}

func (f fakeSource) Languages(_ context.Context, fullName string) (map[string]int64, error) {
	return f.languages[fullName], nil
}

func TestGenerateWritesCompleteChart(t *testing.T) {
	source := fakeSource{
		repositories: []githubapi.Repository{
			{FullName: "me/app"},
			{FullName: "me/service"},
			{FullName: "me/fork", Fork: true},
			{FullName: "me/archive", Archived: true},
		},
		languages: map[string]map[string]int64{
			"me/app":     {"Dart": 70, "Go": 10},
			"me/service": {"Go": 20},
		},
	}

	outputPath := filepath.Join(t.TempDir(), "assets", "languages.svg")
	colors := map[string]string{"Dart": "#00B4AB", "Go": "#00ADD8"}
	processed, err := Generate(context.Background(), source, colors, outputPath, 7)
	if err != nil {
		t.Fatal(err)
	}
	if processed != 2 {
		t.Fatalf("processed = %d, want 2", processed)
	}

	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	var root struct {
		XMLName xml.Name
	}
	if err := xml.Unmarshal(output, &root); err != nil {
		t.Fatalf("invalid SVG: %v", err)
	}
	if root.XMLName.Local != "svg" {
		t.Fatalf("root element = %q, want svg", root.XMLName.Local)
	}
	if !bytes.Contains(output, []byte("Dart")) || !bytes.Contains(output, []byte("Go")) {
		t.Fatal("generated SVG does not contain aggregated languages")
	}
	if !bytes.Contains(output, []byte("#00B4AB")) || !bytes.Contains(output, []byte("#00ADD8")) {
		t.Fatal("generated SVG does not contain GitHub language colors")
	}
	if bytes.Contains(output, []byte("pending")) {
		t.Fatal("generated SVG contains placeholder text")
	}
}
