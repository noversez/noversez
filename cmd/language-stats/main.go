package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/noversez/noversez/internal/app"
	"github.com/noversez/noversez/internal/githubapi"
	"github.com/noversez/noversez/internal/linguist"
)

const (
	outputPath   = "assets/languages.svg"
	topLanguages = 7
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "language stats:", err)
		os.Exit(1)
	}
}

func run() error {
	token := strings.TrimSpace(os.Getenv("LANG_STATS_TOKEN"))
	if token == "" {
		return errors.New("LANG_STATS_TOKEN is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client := githubapi.New(token)
	colors, err := linguist.New().Colors(ctx)
	if err != nil {
		return fmt.Errorf("load GitHub language colors: %w", err)
	}

	processed, err := app.Generate(ctx, client, colors, outputPath, topLanguages)
	if err != nil {
		return err
	}

	fmt.Printf("generated %s from %d repositories\n", outputPath, processed)
	return nil
}
