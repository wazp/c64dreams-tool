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
	Source    string
	Title     string
	Content   model.ContentType
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

			baseName := v.Label.Normalized
			if v.SourcePath != "" {
				baseName = path.Base(v.SourcePath)
			}
			fileName := sanitizeFile(baseName, ext)

			components = append(components, fileName)

			src := v.SourcePath
			if src == "" {
				src = path.Join(gameDir, fileName)
			}

			planned = append(planned, PlannedFile{
				GameID:    g.ID,
				VariantID: fmt.Sprintf("%s-%d", g.ID, vi),
				Target:    g.Target,
				Source:    src,
				Title:     g.Title,
				Content:   v.ContentType,
				Path:      path.Join(components...),
			})
		}

		// preserve input order
		_ = gi
	}

	return planned, nil
}

func sanitizeName(name string) string {
	runes := []rune(strings.TrimSpace(name))
	var b strings.Builder
	last := rune(0)
	for _, r := range runes {
		if r >= 'A' && r <= 'Z' {
			r = r - 'A' + 'a'
		}
		switch {
		case r == '_':
			r = ' '
		case r == '\'':
			continue
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ':
			// keep
		default:
			r = '-'
		}
		// collapse repeated spaces/dashes
		if (r == ' ' || r == '-') && (last == ' ' || last == '-') {
			continue
		}
		b.WriteRune(r)
		last = r
	}
	return strings.Trim(b.String(), " -")
}

func sanitizeFile(name, ext string) string {
	base := strings.TrimSpace(name)
	actualExt := path.Ext(base)
	base = strings.TrimSuffix(base, actualExt)
	runes := []rune(base)
	var b strings.Builder
	last := rune(0)
	for _, r := range runes {
		switch {
		case r == '_':
			r = ' '
		case r == '\'':
			continue
		case (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ':
			if r >= 'A' && r <= 'Z' {
				r = r - 'A' + 'a'
			}
		default:
			r = '-'
		}
		if (r == ' ' || r == '-') && (last == ' ' || last == '-') {
			continue
		}
		b.WriteRune(r)
		last = r
	}
	clean := strings.Trim(b.String(), " -")
	extUse := strings.ToLower(actualExt)
	if ext != "" {
		extUse = "." + strings.ToLower(ext)
	}
	if clean == "" {
		return strings.TrimLeft(extUse, ".")
	}
	if extUse != "" && !strings.HasSuffix(clean, extUse) {
		return clean + extUse
	}
	return clean
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

func alphaBucket(name string, size int) string {
	clean := strings.ToLower(strings.TrimSpace(name))
	if clean == "" {
		return "misc"
	}
	runes := []rune(clean)
	first := runes[0]
	if first >= '0' && first <= '9' {
		return string(first)
	}
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
