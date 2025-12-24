package normalize

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/wazp/c64dreams-tool/pkg/model"
)

var punctuationRegexp = regexp.MustCompile(`[^A-Za-z0-9\s]+`)

var stopWords = map[string]struct{}{
	"the": {}, "of": {}, "and": {}, "a": {}, "an": {}, "for": {}, "to": {}, "in": {}, "on": {}, "at": {}, "by": {}, "with": {}, "from": {},
}

var romanNumerals = map[string]string{
	"i":    "1",
	"ii":   "2",
	"iii":  "3",
	"iv":   "4",
	"v":    "5",
	"vi":   "6",
	"vii":  "7",
	"viii": "8",
	"ix":   "9",
	"x":    "10",
}

// NormalizeGame converts a Game into a NormalizedGame using simple, deterministic rules.
func NormalizeGame(game model.Game, opts Options) (model.NormalizedGame, error) {
	profile := model.ProfileFor(opts.Target)

	ng := model.NormalizedGame{
		ID:     game.ID,
		Title:  game.Title,
		Region: game.Region,
		Target: opts.Target,
	}

	ng.Name = normalizeName(game.Title, opts.EffectiveMaxLen())
	if profile.ForceLowercase {
		ng.Name.Normalized = strings.ToLower(ng.Name.Normalized)
	}

	for _, v := range game.Variants {
		varRegion := v.Region
		if varRegion == "" {
			varRegion = game.Region
		}

		nv := model.NormalizedVariant{
			Label:           normalizeName(v.Label, opts.EffectiveMaxLen()),
			Region:          varRegion,
			PreferredTarget: v.PreferredTarget,
			ContentType:     v.ContentType,
			SourcePath:      v.SourcePath,
			Notes:           v.Notes,
		}
		if profile.ForceLowercase {
			nv.Label.Normalized = strings.ToLower(nv.Label.Normalized)
		}
		ng.Variants = append(ng.Variants, nv)
	}

	return ng, nil
}

func normalizeName(value string, maxLen int) model.NormalizedName {
	name := strings.ReplaceAll(strings.TrimSpace(value), "'", "")
	name = punctuationRegexp.ReplaceAllString(name, " ")
	name = strings.Join(strings.Fields(name), " ")

	words := strings.Fields(name)
	var kept []string
	for _, w := range words {
		lower := strings.ToLower(w)
		if repl, ok := romanNumerals[lower]; ok {
			kept = append(kept, repl)
			continue
		}
		if _, stop := stopWords[lower]; stop {
			continue
		}
		kept = append(kept, preserveCase(w))
	}

	normalized := strings.Join(kept, " ")
	truncated := false

	if maxLen > 0 {
		runes := []rune(normalized)
		if len(runes) > maxLen {
			normalized = string(runes[:maxLen])
			truncated = true
		}
	}

	return model.NormalizedName{
		Original:   value,
		Normalized: normalized,
		Truncated:  truncated,
	}
}

func preserveCase(word string) string {
	if word == "" {
		return word
	}
	runes := []rune(strings.ToLower(word))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
