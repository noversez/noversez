package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/noversez/noversez/internal/chart"
	"github.com/noversez/noversez/internal/githubapi"
	"github.com/noversez/noversez/internal/langstats"
)

type Source interface {
	ListRepositories(context.Context) ([]githubapi.Repository, error)
	Languages(context.Context, string) (map[string]int64, error)
}

func Generate(
	ctx context.Context,
	source Source,
	colors map[string]string,
	outputPath string,
	limit int,
) (int, error) {
	repos, err := source.ListRepositories(ctx)
	if err != nil {
		return 0, err
	}

	totals := langstats.NewTotals()
	processed := 0

	for _, repo := range repos {
		if repo.Fork || repo.Archived || repo.Disabled {
			continue
		}

		languages, err := source.Languages(ctx, repo.FullName)
		if err != nil {
			return 0, fmt.Errorf("read languages for %s: %w", repo.FullName, err)
		}

		totals.Add(languages)
		processed++
	}

	svg, err := chart.Render(totals.Top(limit), colors)
	if err != nil {
		return 0, fmt.Errorf("render chart: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return 0, fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(outputPath, svg, 0o644); err != nil {
		return 0, fmt.Errorf("write chart: %w", err)
	}

	return processed, nil
}
