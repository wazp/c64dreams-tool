package normalize

import (
	"fmt"
	"testing"

	"github.com/wazp/c64dreams-tool/pkg/model"
)

func TestNormalizeGameRomanNumerals(t *testing.T) {
	game := model.Game{
		ID:     "impossible-mission-2",
		Title:  "Impossible Mission II",
		Region: model.RegionBoth,
		Variants: []model.Variant{{
			Label:           "Disk",
			PreferredTarget: model.TargetSD2IEC,
		}},
	}

	ng, err := NormalizeGame(game, Options{MaxNameLen: 0, Target: model.TargetUltimate})
	if err != nil {
		t.Fatalf("NormalizeGame returned error: %v", err)
	}

	if ng.Name.Normalized != "Impossible Mission 2" {
		t.Fatalf("unexpected normalized name: %q", ng.Name.Normalized)
	}
}

func TestNormalizeGameTruncation(t *testing.T) {
	game := model.Game{Title: "The Great Giana Sisters"}

	ng, err := NormalizeGame(game, Options{MaxNameLen: 8})
	if err != nil {
		t.Fatalf("NormalizeGame returned error: %v", err)
	}

	if !ng.Name.Truncated {
		t.Fatalf("expected name to be truncated")
	}
	if len([]rune(ng.Name.Normalized)) != 8 {
		t.Fatalf("expected normalized length 8, got %d", len([]rune(ng.Name.Normalized)))
	}
}

func TestNormalizeGameCollisionFlagging(t *testing.T) {
	game := model.Game{
		Title: "Test Title",
		Variants: []model.Variant{
			{Label: "Disk A"},
			{Label: "disk a"},
		},
	}

	ng, err := NormalizeGame(game, Options{MaxNameLen: 0, Target: model.TargetSD2IEC})
	if err != nil {
		t.Fatalf("NormalizeGame returned error: %v", err)
	}

	games := ResolveCollisions([]model.NormalizedGame{ng}, Options{Target: model.TargetSD2IEC})

	if len(games[0].Variants) != 2 {
		t.Fatalf("expected two variants")
	}

	if !games[0].Variants[0].Label.Collision || !games[0].Variants[1].Label.Collision {
		t.Fatalf("expected both variants to be marked as collisions")
	}
	if games[0].Variants[0].Label.Normalized != "Disk" {
		t.Fatalf("expected first variant unchanged, got %s", games[0].Variants[0].Label.Normalized)
	}
	if games[0].Variants[1].Label.Normalized != "Disk~1" {
		t.Fatalf("expected second variant suffixed, got %s", games[0].Variants[1].Label.Normalized)
	}
}

func TestTargetSpecificMaxLengths(t *testing.T) {
	cases := []struct {
		name        string
		target      model.TargetDevice
		expectedMax int
	}{
		{"sd2iec", model.TargetSD2IEC, 16},
		{"pi1541", model.TargetPi1541, 16},
		{"kungfuflash", model.TargetKungFuFlash, 255},
		{"ultimate", model.TargetUltimate, 255},
	}

	longTitle := "ABCDEFGHIJKLMNOPQRSTUVWXABCDEFGHIJKLMNOPQRSTUVWX"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			game := model.Game{Title: longTitle}
			ng, err := NormalizeGame(game, Options{Target: tc.target})
			if err != nil {
				t.Fatalf("NormalizeGame returned error: %v", err)
			}

			if len([]rune(ng.Name.Normalized)) > tc.expectedMax {
				t.Fatalf("name length %d exceeds expected max %d", len([]rune(ng.Name.Normalized)), tc.expectedMax)
			}

			if tc.expectedMax == 16 && len([]rune(ng.Name.Normalized)) != tc.expectedMax {
				t.Fatalf("expected truncation to 16 for %s", tc.name)
			}

			if tc.target == model.TargetKungFuFlash && len([]rune(ng.Name.Normalized)) <= 16 {
				t.Fatalf("expected kungfuflash to allow longer than 16 chars")
			}

			if tc.target == model.TargetUltimate && len([]rune(ng.Name.Normalized)) <= 16 {
				t.Fatalf("expected ultimate to allow longer than 16 chars")
			}
		})
	}
}

func TestOverrideMaxNameLenWins(t *testing.T) {
	game := model.Game{Title: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"}

	ng, err := NormalizeGame(game, Options{Target: model.TargetUltimate, MaxNameLen: 10})
	if err != nil {
		t.Fatalf("NormalizeGame returned error: %v", err)
	}

	if len([]rune(ng.Name.Normalized)) != 10 {
		t.Fatalf("expected override to truncate to 10, got %d", len([]rune(ng.Name.Normalized)))
	}
}

func TestCollisionResolutionAcrossTargets(t *testing.T) {
	cases := []struct {
		name     string
		opts     Options
		titles   []string
		expected []string
	}{
		{
			name:   "sd2iec-16",
			opts:   Options{Target: model.TargetSD2IEC},
			titles: []string{"Impossible Mission II", "Impossible Mission III", "Impossible Mission IV"},
			expected: []string{
				"Impossible Missi",
				"Impossible Mis~1",
				"Impossible Mis~2",
			},
		},
		{
			name:   "override-8",
			opts:   Options{Target: model.TargetSD2IEC, MaxNameLen: 8},
			titles: []string{"Impossible Mission II", "Impossible Mission III", "Impossible Mission IV"},
			expected: []string{
				"Impossib",
				"Imposs~1",
				"Imposs~2",
			},
		},
		{
			name:   "ultimate-long",
			opts:   Options{Target: model.TargetUltimate},
			titles: []string{"Very Long Adventure Name", "Very Long Adventure Name"},
			expected: []string{
				"Very Long Adventure Name",
				"Very Long Adventure Name~1",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var normalized []model.NormalizedGame
			for i, title := range tc.titles {
				game := model.Game{ID: fmt.Sprintf("g%d", i), Title: title}
				ng, err := NormalizeGame(game, tc.opts)
				if err != nil {
					t.Fatalf("NormalizeGame returned error: %v", err)
				}
				normalized = append(normalized, ng)
			}

			normalized = ResolveCollisions(normalized, tc.opts)

			if len(normalized) != len(tc.expected) {
				t.Fatalf("expected %d games", len(tc.expected))
			}
			for i := range normalized {
				if normalized[i].Name.Normalized != tc.expected[i] {
					t.Fatalf("index %d got %q expected %q", i, normalized[i].Name.Normalized, tc.expected[i])
				}
				if !normalized[i].Name.Collision {
					t.Fatalf("expected collision flag for index %d", i)
				}
				if normalized[i].Name.CollisionIndex != i {
					t.Fatalf("expected collision index %d got %d", i, normalized[i].Name.CollisionIndex)
				}
			}
		})
	}
}
