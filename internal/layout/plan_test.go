package layout

import (
	"testing"

	"github.com/wazp/c64dreams-tool/pkg/model"
)

func TestMediaGroupFor(t *testing.T) {
	cases := []struct {
		ext      string
		expected string
	}{
		{".d64", "disks"},
		{"d81", "disks"},
		{"tap", "tape"},
		{"crt", "cart"},
	}

	for _, tc := range cases {
		got, err := MediaGroupFor(tc.ext)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", tc.ext, err)
		}
		if got != tc.expected {
			t.Fatalf("expected %s got %s", tc.expected, got)
		}
	}
}

func TestPlanNoGrouping(t *testing.T) {
	game := sampleGame("g1", "Jumpman", "Disk1", model.ContentDisk)
	planned, err := Plan([]model.NormalizedGame{game}, Options{BaseDir: "games"})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}

	expectPath(t, planned, []string{"games/jumpman/disk1.d64"})
}

func TestPlanMediaGrouping(t *testing.T) {
	game := sampleGame("g1", "Jumpman", "Disk1", model.ContentDisk)
	planned, err := Plan([]model.NormalizedGame{game}, Options{BaseDir: "games", GroupByMedia: true})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}

	expectPath(t, planned, []string{"games/disks/jumpman/disk1.d64"})
}

func TestPlanAlphaGrouping(t *testing.T) {
	game := sampleGame("g1", "Jumpman", "Disk1", model.ContentDisk)
	planned, err := Plan([]model.NormalizedGame{game}, Options{BaseDir: "games", GroupByAlpha: true})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}

	expectPath(t, planned, []string{"games/j/jumpman/disk1.d64"})
}

func TestPlanAlphaRangeSizeTwo(t *testing.T) {
	game := sampleGame("g1", "Alpha", "Disk1", model.ContentDisk)
	planned, err := Plan([]model.NormalizedGame{game}, Options{BaseDir: "games", GroupByAlpha: true, AlphaBucketSize: 2})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}

	expectPath(t, planned, []string{"games/ag-al/alpha/disk1.d64"})
}

func TestPlanMediaAndAlpha(t *testing.T) {
	game := sampleGame("g1", "Jumpman", "Disk1", model.ContentDisk)
	planned, err := Plan([]model.NormalizedGame{game}, Options{BaseDir: "games", GroupByMedia: true, GroupByAlpha: true})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}

	expectPath(t, planned, []string{"games/disks/j/jumpman/disk1.d64"})
}

func TestPlanStability(t *testing.T) {
	game1 := sampleGame("g1", "Alpha", "Disk1", model.ContentDisk)
	game2 := sampleGame("g2", "Beta", "Tape", model.ContentTape)
	planned, err := Plan([]model.NormalizedGame{game1, game2}, Options{BaseDir: "games", GroupByAlpha: true})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}

	expectPath(t, planned, []string{
		"games/a/alpha/disk1.d64",
		"games/b/beta/tape.tap",
	})
}

func sampleGame(id, name, label string, ct model.ContentType) model.NormalizedGame {
	return model.NormalizedGame{
		ID:     id,
		Title:  name,
		Name:   model.NormalizedName{Normalized: name},
		Region: model.RegionBoth,
		Target: model.TargetSD2IEC,
		Variants: []model.NormalizedVariant{
			{
				Label:       model.NormalizedName{Normalized: label},
				Region:      model.RegionBoth,
				ContentType: ct,
			},
		},
	}
}

func expectPath(t *testing.T, planned []PlannedFile, expected []string) {
	t.Helper()
	if len(planned) != len(expected) {
		t.Fatalf("expected %d planned files, got %d", len(expected), len(planned))
	}
	for i := range planned {
		if planned[i].Path != expected[i] {
			t.Fatalf("index %d expected %s got %s", i, expected[i], planned[i].Path)
		}
	}
}
