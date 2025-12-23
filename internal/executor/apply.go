package executor

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wazp/c64dreams-tool/internal/layout"
)

// Result captures the outcome of applying a single planned file.
type Result struct {
	Source string
	Dest   string
	Action string // copy, skip, mkdir, error
	Error  error
}

// Apply executes a planned layout onto the filesystem with safety and dry-run support.
func Apply(planned []layout.PlannedFile, opts Options) ([]Result, error) {
	results := make([]Result, 0, len(planned))

	if opts.OutputRoot == "" {
		return nil, errors.New("output root is required")
	}

	dry := opts.DryRun || opts.VerifyOnly

	sorted := make([]layout.PlannedFile, len(planned))
	copy(sorted, planned)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})

	outputAbs, err := filepath.Abs(opts.OutputRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve output root: %w", err)
	}

	for _, p := range sorted {
		cleanRel := path.Clean(p.Path)
		if path.IsAbs(cleanRel) {
			res := Result{Dest: cleanRel, Action: "error", Error: errors.New("destination path must be relative")}
			results = append(results, res)
			return results, res.Error
		}

		destFull := filepath.Join(outputAbs, filepath.FromSlash(cleanRel))
		destFull = filepath.Clean(destFull)

		if !strings.HasPrefix(destFull, outputAbs) {
			res := Result{Dest: destFull, Action: "error", Error: errors.New("destination escapes output root")}
			results = append(results, res)
			return results, res.Error
		}

		srcFull := filepath.Join(opts.InputRoot, filepath.FromSlash(cleanRel))

		dir := filepath.Dir(destFull)
		if !dry {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				res := Result{Source: srcFull, Dest: destFull, Action: "error", Error: fmt.Errorf("mkdir: %w", err)}
				results = append(results, res)
				return results, res.Error
			}
		} else {
			results = append(results, Result{Source: srcFull, Dest: dir, Action: "mkdir"})
		}

		srcInfo, err := os.Stat(srcFull)
		if err != nil {
			res := Result{Source: srcFull, Dest: destFull, Action: "error", Error: fmt.Errorf("source missing: %w", err)}
			results = append(results, res)
			return results, res.Error
		}
		if srcInfo.IsDir() {
			res := Result{Source: srcFull, Dest: destFull, Action: "error", Error: errors.New("source is directory")}
			results = append(results, res)
			return results, res.Error
		}

		if opts.VerifyOnly {
			results = append(results, Result{Source: srcFull, Dest: destFull, Action: "skip"})
			continue
		}

		destInfo, err := os.Stat(destFull)
		exists := err == nil
		if exists && destInfo.IsDir() {
			res := Result{Source: srcFull, Dest: destFull, Action: "error", Error: errors.New("destination is directory")}
			results = append(results, res)
			return results, res.Error
		}

		if exists && !opts.Overwrite {
			results = append(results, Result{Source: srcFull, Dest: destFull, Action: "skip"})
			continue
		}

		if dry {
			results = append(results, Result{Source: srcFull, Dest: destFull, Action: "copy"})
			continue
		}

		if err := copyFile(srcFull, destFull); err != nil {
			res := Result{Source: srcFull, Dest: destFull, Action: "error", Error: err}
			results = append(results, res)
			return results, res.Error
		}

		results = append(results, Result{Source: srcFull, Dest: destFull, Action: "copy"})
	}

	return results, nil
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	return nil
}
