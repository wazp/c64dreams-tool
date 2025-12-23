package layout

import (
	"fmt"
	"path"
	"strings"

	"github.com/wazp/c64dreams-tool/pkg/model"
)

// PlannedFile represents a variant placed into a relative path.
type PlannedFile struct {
	GameID    string
	VariantID string
	Target    model.TargetDevice
	Path      string
}

// Plan maps normalized games into relative output paths without touching the filesystem.
func Plan(games []model.NormalizedGame, opts Options) ([]PlannedFile, error) {
	alphaSize := opts.AlphaBucketSize
	if alphaSize <= 0 {
		alphaSize = defaultAlphaBucketSize
	}

	var planned []PlannedFile

	for gi, g := range games {
		gameDir := sanitizeName(g.Name.Normalized)
		bucket := ""
		if opts.GroupByAlpha {
			bucket = alphaBucket(g.Name.Normalized, alphaSize)
		}

		for vi, v := range g.Variants {
			ext := extensionForContent(v.ContentType)
			components := make([]string, 0, 5)
			if opts.BaseDir != "" {
				components = append(components, sanitizePathPart(opts.BaseDir))
			}

			if opts.GroupByMedia {
				media, err := MediaGroupFor(ext)
				if err != nil {
					return nil, err
				}
				components = append(components, media)
			}

			if bucket != "" {
				components = append(components, bucket)
			}

			components = append(components, gameDir)

			fileName := sanitizeName(v.Label.Normalized)
			if ext != "" {
				fileName += "." + ext
			}

			components = append(components, fileName)

			planned = append(planned, PlannedFile{
				GameID:    g.ID,
				VariantID: fmt.Sprintf("%s-%d", g.ID, vi),
				Target:    g.Target,
				Path:      path.Join(components...),
			})
		}

		// preserve input order
		_ = gi
	}

	return planned, nil
}

func sanitizeName(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	lower = strings.ReplaceAll(lower, " ", "_")
	return lower
}

func sanitizePathPart(part string) string {
	return strings.Trim(path.Clean(strings.ReplaceAll(part, "\\", "/")), "/")
}

func extensionForContent(ct model.ContentType) string {
	switch ct {
	case model.ContentDisk:
		return "d64"
	case model.ContentTape:
		return "tap"
	case model.ContentPrg:
		return "prg"
	case model.ContentZip:
		return "zip"
	default:
		return ""
	}
}

func alphaBucket(name string, size int) string {
	clean := strings.ToLower(strings.TrimSpace(name))
	if clean == "" {
		return "misc"
	}
	runes := []rune(clean)
	first := runes[0]
	if first < 'a' || first > 'z' {
		return "misc"
	}

	if size <= 1 {
		return string(first)
	}

	second := 'a'
	if len(runes) > 1 && runes[1] >= 'a' && runes[1] <= 'z' {
		second = runes[1]
	}

	groupSize := size * 3
	if groupSize < 1 {
		groupSize = 1
	}
	idx := int(second-'a') / groupSize
	start := rune('a' + idx*groupSize)
	end := start + rune(groupSize-1)
	if end > 'z' {
		end = 'z'
	}

	return fmt.Sprintf("%c%c-%c%c", first, start, first, end)
}
