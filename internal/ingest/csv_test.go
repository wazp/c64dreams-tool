package ingest

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wazp/c64dreams-tool/pkg/model"
)

func TestLoadCSV(t *testing.T) {
	content := "summary,,,,\n" +
		",,Title,Type,Multi-disk,Joystick Port,TrueDrive Enabled,Autowarp,Autoload State,Genre,Manual,Zzap! Review 1,Zzap! Review 2,Zzap! Review 3,Game Notes,Retroarch Notes,PRG Name,Group,Version,Source,Custom Notes\n" +
		"c,,Test Game,d64,,2,No,,,Shmup,Yes,70%,,,Note one,Retro note,PRG1,Remember,Test Game +1,CSDb,Custom note\n" +
		",,Second Game,prg,,2,No,,,Shmup,Yes,70%,,, , ,n/a,Group B,,GB64,\n"

	dir := t.TempDir()
	path := filepath.Join(dir, "games.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	games, err := LoadCSV(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadCSV returned error: %v", err)
	}

	if len(games) != 2 {
		t.Fatalf("expected 2 games, got %d", len(games))
	}

	first := games[0]
	if first.Title != "Test Game" || first.ID != "test-game" {
		t.Fatalf("unexpected first game: %+v", first)
	}
	if len(first.Variants) != 1 {
		t.Fatalf("expected 1 variant, got %d", len(first.Variants))
	}
	variant := first.Variants[0]
	if variant.Label != "Test Game +1" {
		t.Fatalf("unexpected variant label: %s", variant.Label)
	}
	if variant.ContentType != model.ContentDisk {
		t.Fatalf("unexpected content type: %s", variant.ContentType)
	}
	if variant.SourcePath != "PRG1.d64" {
		t.Fatalf("unexpected source path: %s", variant.SourcePath)
	}
	if variant.Notes != "Note one | Retro note | Custom note | CSDb" {
		t.Fatalf("unexpected notes: %q", variant.Notes)
	}

	second := games[1]
	if second.Variants[0].ContentType != model.ContentPrg {
		t.Fatalf("unexpected second content type: %s", second.Variants[0].ContentType)
	}
	if second.Variants[0].SourcePath != "Second Game.prg" { // falls back to title when PRG name is n/a
		t.Fatalf("unexpected fallback source path: %s", second.Variants[0].SourcePath)
	}
}
