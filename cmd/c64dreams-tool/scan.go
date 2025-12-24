package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func newScanCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan the C64 Dreams directory and report contents",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOptions(opts); err != nil {
				return err
			}

			if opts.input == "" {
				return fmt.Errorf("--input is required")
			}

			files, err := scanC64Files(opts.input)
			if err != nil {
				return err
			}

			if opts.json {
				payload := struct {
					Input string        `json:"input"`
					Count int           `json:"count"`
					Files []scannedFile `json:"files"`
				}{Input: opts.input, Count: len(files), Files: files}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(payload)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Found %d C64-related files under %s\n", len(files), opts.input)
			for i, f := range files {
				if i >= 10 {
					fmt.Fprintf(cmd.OutOrStdout(), "... (%d more)\n", len(files)-i)
					break
				}
				fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", f.Path)
			}

			return nil
		},
	}

	return cmd
}

type scannedFile struct {
	Path string `json:"path"`
	Ext  string `json:"ext"`
}

func scanC64Files(root string) ([]scannedFile, error) {
	allowed := allC64Exts()
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var files []scannedFile
	err = filepath.WalkDir(rootAbs, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(d.Name())), ".")
		if !hasExt(ext, allowed) {
			return nil
		}
		rel, relErr := filepath.Rel(rootAbs, p)
		if relErr != nil {
			return relErr
		}
		files = append(files, scannedFile{Path: filepath.ToSlash(rel), Ext: ext})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sortScanned(files)
	return files, nil
}

func sortScanned(files []scannedFile) {
	less := func(i, j int) bool { return files[i].Path < files[j].Path }
	sort.Slice(files, less)
}

func hasExt(ext string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(ext, a) {
			return true
		}
	}
	return false
}

func allC64Exts() []string {
	return []string{"d64", "d71", "d81", "g64", "tap", "t64", "crt", "ef", "prg", "zip"}
}
