package langstats

import (
	"reflect"
	"testing"
)

func TestTotalsTopGroupsRemainder(t *testing.T) {
	totals := NewTotals()
	totals.Add(map[string]int64{"Go": 80, "Dart": 40, "CSS": 10})
	totals.Add(map[string]int64{"Go": 20, "HTML": 5, "Shell": 2})

	got := totals.Top(3)
	want := []Language{
		{Name: "Go", Bytes: 100},
		{Name: "Dart", Bytes: 40},
		{Name: "CSS", Bytes: 10},
		{Name: "Other", Bytes: 7},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Top() = %#v, want %#v", got, want)
	}
}

func TestTotalsTopUsesNameAsTieBreaker(t *testing.T) {
	totals := Totals{"TypeScript": 10, "Dart": 10, "Go": 10}

	got := totals.Top(3)
	want := []Language{
		{Name: "Dart", Bytes: 10},
		{Name: "Go", Bytes: 10},
		{Name: "TypeScript", Bytes: 10},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Top() = %#v, want %#v", got, want)
	}
}

func TestTotalsTopPreservesEveryLanguageByte(t *testing.T) {
	totals := Totals{
		"Dart":       100,
		"Go":         90,
		"Java":       80,
		"C":          70,
		"Shell":      60,
		"HTML":       50,
		"CSS":        40,
		"TypeScript": 30,
		"JavaScript": 20,
	}

	var before int64
	for _, bytes := range totals {
		before += bytes
	}

	var after int64
	for _, language := range totals.Top(7) {
		after += language.Bytes
	}

	if after != before {
		t.Fatalf("Top() preserved %d bytes, want %d", after, before)
	}
}
