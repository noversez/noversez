package langstats

import "sort"

type Language struct {
	Name  string
	Bytes int64
}

type Totals map[string]int64

func NewTotals() Totals {
	return make(Totals)
}

func (t Totals) Add(languages map[string]int64) {
	for name, bytes := range languages {
		t[name] += bytes
	}
}

func (t Totals) Top(limit int) []Language {
	if limit < 1 {
		return nil
	}

	languages := make([]Language, 0, len(t))
	for name, bytes := range t {
		languages = append(languages, Language{Name: name, Bytes: bytes})
	}

	sort.Slice(languages, func(i, j int) bool {
		if languages[i].Bytes == languages[j].Bytes {
			return languages[i].Name < languages[j].Name
		}
		return languages[i].Bytes > languages[j].Bytes
	})

	if len(languages) <= limit {
		return languages
	}

	var other int64
	for _, language := range languages[limit:] {
		other += language.Bytes
	}

	return append(
		languages[:limit],
		Language{Name: "Other", Bytes: other},
	)
}
