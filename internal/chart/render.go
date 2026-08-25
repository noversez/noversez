package chart

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"math"

	"github.com/noversez/noversez/internal/langstats"
)

const (
	width       = 620
	height      = 300
	centerX     = 155
	centerY     = 150
	radius      = 85
	strokeWidth = 28
)

var (
	//go:embed languages.svg.tmpl
	templateSource string

	svgTemplate = template.Must(template.New("languages.svg").Parse(templateSource))

	fallbackColor = "#8b949e"
)

type segment struct {
	Name       string
	Color      string
	Dash       string
	Gap        string
	Offset     string
	LegendY    int
	DotY       int
	Percentage string
}

type view struct {
	Width       int
	Height      int
	CenterX     int
	CenterY     int
	Radius      int
	StrokeWidth int
	Empty       bool
	Segments    []segment
	TopName     string
	TopPercent  string
}

func Render(languages []langstats.Language, colors map[string]string) ([]byte, error) {
	data := view{
		Width:       width,
		Height:      height,
		CenterX:     centerX,
		CenterY:     centerY,
		Radius:      radius,
		StrokeWidth: strokeWidth,
	}

	var total int64
	for _, language := range languages {
		total += language.Bytes
	}
	if total == 0 {
		data.Empty = true
		return execute(data)
	}

	circumference := 2 * math.Pi * radius
	offset := 0.0
	for i, language := range languages {
		percentage := float64(language.Bytes) / float64(total)
		length := circumference * percentage
		legendY := 55 + i*28
		color := colors[language.Name]
		if color == "" {
			color = fallbackColor
		}

		data.Segments = append(data.Segments, segment{
			Name:       language.Name,
			Color:      color,
			Dash:       fmt.Sprintf("%.2f", length),
			Gap:        fmt.Sprintf("%.2f", circumference-length),
			Offset:     fmt.Sprintf("%.2f", -offset),
			LegendY:    legendY,
			DotY:       legendY - 5,
			Percentage: fmt.Sprintf("%.1f%%", percentage*100),
		})
		offset += length
	}

	data.TopName = languages[0].Name
	data.TopPercent = data.Segments[0].Percentage
	return execute(data)
}

func execute(data view) ([]byte, error) {
	var output bytes.Buffer
	if err := svgTemplate.Execute(&output, data); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
