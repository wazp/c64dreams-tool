package ingest

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/wazp/c64dreams-tool/pkg/model"
)

var wordCharRegexp = regexp.MustCompile(`[A-Za-z0-9]+`)

// LoadCSV reads a C64 Dreams metadata CSV and converts it into structured game data.
func LoadCSV(ctx context.Context, path string) ([]model.Game, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	headers := map[string]int{}
	var games []model.Game

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv: %w", err)
		}

		// Skip empty rows.
		if isEmpty(rec) {
			continue
		}

		if len(headers) == 0 {
			if hasHeader(rec, "Title") {
				headers = indexHeaders(rec)
			}
			continue
		}

		title := strings.TrimSpace(value(rec, headers, "Title"))
		if title == "" {
			continue
		}

		content := strings.ToLower(strings.TrimSpace(value(rec, headers, "Type")))
		prgName := strings.TrimSpace(value(rec, headers, "PRG Name"))
		version := strings.TrimSpace(value(rec, headers, "Version"))
		source := strings.TrimSpace(value(rec, headers, "Source"))
		customNotes := strings.TrimSpace(value(rec, headers, "Custom Notes"))
		gameNotes := strings.TrimSpace(value(rec, headers, "Game Notes"))
		retroarchNotes := strings.TrimSpace(value(rec, headers, "Retroarch Notes"))

		notes := joinNonEmpty(" | ", gameNotes, retroarchNotes, customNotes, source)

		ct := mapContentType(content)
		sourcePath := chooseSourcePath(prgName, title, ct)

		variant := model.Variant{
			Label:           chooseLabel(version, content),
			Region:          model.RegionBoth,
			PreferredTarget: model.TargetUltimate,
			ContentType:     ct,
			SourcePath:      sourcePath,
			Notes:           notes,
		}

		game := model.Game{
			ID:             slugify(title),
			Title:          title,
			NormalizedName: title,
			Region:         model.RegionBoth,
			Variants:       []model.Variant{variant},
		}

		games = append(games, game)
	}

	return games, nil
}

func hasHeader(rec []string, key string) bool {
	for _, field := range rec {
		if strings.EqualFold(strings.TrimSpace(field), key) {
			return true
		}
	}
	return false
}

func indexHeaders(rec []string) map[string]int {
	indices := make(map[string]int)
	for i, field := range rec {
		name := strings.ToLower(strings.TrimSpace(field))
		if name != "" {
			indices[name] = i
		}
	}
	return indices
}

func value(rec []string, headers map[string]int, name string) string {
	idx, ok := headers[strings.ToLower(name)]
	if !ok || idx >= len(rec) {
		return ""
	}
	return rec[idx]
}

func isEmpty(rec []string) bool {
	for _, field := range rec {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}

func joinNonEmpty(sep string, parts ...string) string {
	var filtered []string
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			filtered = append(filtered, strings.TrimSpace(part))
		}
	}
	return strings.Join(filtered, sep)
}

func chooseLabel(version, content string) string {
	if version != "" {
		return version
	}
	if content != "" {
		return strings.ToUpper(content)
	}
	return "Variant"
}

func chooseSourcePath(prgName, title string, ct model.ContentType) string {
	name := title
	cleanPRG := strings.TrimSpace(prgName)
	if cleanPRG != "" && !strings.EqualFold(cleanPRG, "n/a") && !strings.ContainsAny(cleanPRG, " \\//") {
		name = cleanPRG
	}

	ext := extensionForContent(ct)
	if strings.HasSuffix(strings.ToLower(name), "."+ext) {
		return name
	}

	if ext != "" {
		return name + "." + ext
	}

	return name
}

func extensionForContent(ct model.ContentType) string {
	switch ct {
	case model.ContentDisk:
		return "d64"
	case model.ContentTape:
		return "tap"
	case model.ContentCart:
		return "crt"
	case model.ContentPrg:
		return "prg"
	case model.ContentZip:
		return "zip"
	default:
		return ""
	}
}

func mapContentType(content string) model.ContentType {
	switch strings.ToLower(content) {
	case "d64", "d71", "d81", "d1m", "g64":
		return model.ContentDisk
	case "t64", "tap":
		return model.ContentTape
	case "prg":
		return model.ContentPrg
	case "crt", "ef", "easyflash":
		return model.ContentCart
	case "zip":
		return model.ContentZip
	default:
		return model.ContentUnknown
	}
}

func slugify(title string) string {
	lower := strings.ToLower(title)
	words := wordCharRegexp.FindAllString(lower, -1)
	return strings.Join(words, "-")
}
