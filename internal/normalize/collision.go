package normalize

import (
	"fmt"
	"strings"

	"github.com/wazp/c64dreams-tool/pkg/model"
)

// ResolveCollisions applies deterministic suffixes to a set of normalized games using the provided options.
func ResolveCollisions(games []model.NormalizedGame, opts Options) []model.NormalizedGame {
	return resolveCollisions(games, opts.EffectiveMaxLen())
}

// resolveCollisions applies deterministic suffixing to conflicting normalized names.
func resolveCollisions(games []model.NormalizedGame, maxLen int) []model.NormalizedGame {
	groups := make(map[string][]*model.NormalizedName)

	add := func(device model.TargetDevice, region model.Region, name *model.NormalizedName) {
		key := collisionKey(device, region, name.Normalized)
		name.CollisionGroup = key
		groups[key] = append(groups[key], name)
	}

	for i := range games {
		add(games[i].Target, games[i].Region, &games[i].Name)
		for j := range games[i].Variants {
			add(games[i].Target, games[i].Variants[j].Region, &games[i].Variants[j].Label)
		}
	}

	for _, entries := range groups {
		if len(entries) <= 1 {
			continue
		}
		for idx, name := range entries {
			name.Collision = true
			name.CollisionIndex = idx
			if idx == 0 {
				continue
			}
			suffix := fmt.Sprintf("~%d", idx)
			base := name.Normalized
			safeLen := maxLen - len([]rune(suffix))
			if safeLen < 0 {
				safeLen = 0
			}
			runes := []rune(base)
			if len(runes) > safeLen {
				base = string(runes[:safeLen])
			}
			name.Normalized = base + suffix
		}
	}

	return games
}

func collisionKey(device model.TargetDevice, region model.Region, name string) string {
	return strings.ToLower(string(device) + "|" + string(region) + "|" + name)
}
