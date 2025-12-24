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
	"github.com/wazp/c64dreams-tool/pkg/model"
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

		srcRel := p.Source
		if srcRel == "" {
			srcRel = cleanRel
		}
		srcRelClean := path.Clean(srcRel)
		if path.IsAbs(srcRelClean) {
			res := Result{Dest: destFull, Source: srcRelClean, Action: "error", Error: errors.New("source path must be relative")}
			results = append(results, res)
			return results, res.Error
		}
		srcFull := filepath.Join(opts.InputRoot, filepath.FromSlash(srcRelClean))

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

		allowedExts := aliasesForExt(p.Content, strings.TrimPrefix(strings.ToLower(filepath.Ext(cleanRel)), "."))

		srcInfo, err := os.Stat(srcFull)
		if err != nil {
			base := filepath.Base(srcRelClean)
			dirSlug := slug(filepath.Base(filepath.Dir(cleanRel)))
			titleSlug := slug(p.Title)
			fileSlug := slug(strings.TrimSuffix(base, filepath.Ext(base)))
			plannedSlug := slug(strings.TrimSuffix(filepath.Base(cleanRel), filepath.Ext(cleanRel)))

			match, matchErr := findBySlug(opts.InputRoot, []string{dirSlug, titleSlug}, []string{fileSlug, plannedSlug}, allowedExts)
			if matchErr != nil {
				match, matchErr = findInMatchingDir(opts.InputRoot, filepath.Base(filepath.Dir(cleanRel)), allowedExts)
			}
			if matchErr != nil {
				match, matchErr = findCaseInsensitive(opts.InputRoot, base)
			}
			if matchErr != nil {
				match, matchErr = findCaseInsensitiveNoExt(opts.InputRoot, base)
			}
			if matchErr != nil {
				match, matchErr = findAnyBySlug(opts.InputRoot, []string{titleSlug, fileSlug, plannedSlug}, allowedExts)
			}
			if matchErr != nil {
				match, matchErr = findAnyByExt(opts.InputRoot, allowedExts)
			}
			if matchErr != nil {
				res := Result{Source: srcFull, Dest: destFull, Action: "error", Error: fmt.Errorf("source missing: %w", err)}
				results = append(results, res)
				return results, res.Error
			}
			srcFull = match
			srcInfo, err = os.Stat(srcFull)
			if err != nil {
				res := Result{Source: srcFull, Dest: destFull, Action: "error", Error: fmt.Errorf("source missing after search: %w", err)}
				results = append(results, res)
				return results, res.Error
			}
		}
		if srcInfo.IsDir() {
			files, pickErr := pickFilesInDir(srcFull, allC64Exts())
			if pickErr != nil {
				res := Result{Source: srcFull, Dest: destFull, Action: "error", Error: pickErr}
				results = append(results, res)
				return results, res.Error
			}
			destDir := filepath.Dir(destFull)
			for _, f := range files {
				fileName := sanitizeFileName(filepath.Base(f))
				destPath := filepath.Join(destDir, fileName)
				if opts.VerifyOnly {
					results = append(results, Result{Source: f, Dest: destPath, Action: "skip"})
					continue
				}
				if dry {
					results = append(results, Result{Source: f, Dest: destPath, Action: "copy"})
					continue
				}
				if err := copyFile(f, destPath); err != nil {
					res := Result{Source: f, Dest: destPath, Action: "error", Error: err}
					results = append(results, res)
					return results, res.Error
				}
				results = append(results, Result{Source: f, Dest: destPath, Action: "copy"})
			}
			continue
		}

		// include siblings in the same directory that match C64 extensions
		dirFiles, _ := pickFilesInDir(filepath.Dir(srcFull), allC64Exts())
		fileSet := make(map[string]struct{})
		var toCopy []string
		for _, f := range append([]string{srcFull}, dirFiles...) {
			if _, ok := fileSet[f]; ok {
				continue
			}
			fileSet[f] = struct{}{}
			toCopy = append(toCopy, f)
		}

		destDir := filepath.Dir(destFull)
		for _, f := range toCopy {
			fileName := sanitizeFileName(filepath.Base(f))
			destPath := filepath.Join(destDir, fileName)

			if opts.VerifyOnly {
				results = append(results, Result{Source: f, Dest: destPath, Action: "skip"})
				continue
			}

			destInfo, err := os.Stat(destPath)
			exists := err == nil
			if exists && destInfo.IsDir() {
				res := Result{Source: f, Dest: destPath, Action: "error", Error: errors.New("destination is directory")}
				results = append(results, res)
				return results, res.Error
			}

			if exists && !opts.Overwrite {
				results = append(results, Result{Source: f, Dest: destPath, Action: "skip"})
				continue
			}

			if dry {
				results = append(results, Result{Source: f, Dest: destPath, Action: "copy"})
				continue
			}

			if err := copyFile(f, destPath); err != nil {
				res := Result{Source: f, Dest: destPath, Action: "error", Error: err}
				results = append(results, res)
				return results, res.Error
			}

			results = append(results, Result{Source: f, Dest: destPath, Action: "copy"})
		}
	}

	return results, nil
}

func findCaseInsensitive(root, base string) (string, error) {
	var matches []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), base) {
			matches = append(matches, p)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no case-insensitive match for %s", base)
	}
	sort.Strings(matches)
	return matches[0], nil
}

func findCaseInsensitiveNoExt(root, base string) (string, error) {
	baseNoExt := strings.TrimSuffix(base, filepath.Ext(base))
	var matches []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		nameNoExt := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		if strings.EqualFold(nameNoExt, baseNoExt) {
			matches = append(matches, p)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no case-insensitive match (no ext) for %s", base)
	}
	sort.Strings(matches)
	return matches[0], nil
}

func findInMatchingDir(root, dirName string, exts []string) (string, error) {
	if dirName == "." || dirName == "" {
		return "", fmt.Errorf("no dir name provided")
	}
	var candidates []string
	walkErr := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), dirName) {
			entries, err := os.ReadDir(p)
			if err != nil {
				return err
			}
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				if len(exts) == 0 || hasExt(strings.TrimPrefix(filepath.Ext(e.Name()), "."), exts) {
					candidates = append(candidates, filepath.Join(p, e.Name()))
				}
			}
		}
		return nil
	})
	if walkErr != nil {
		return "", walkErr
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no match in dir %s", dirName)
	}
	sort.Strings(candidates)
	return candidates[0], nil
}

func findAnyBySlug(root string, fileSlugs []string, exts []string) (string, error) {
	var candidates []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if len(exts) > 0 && !hasExt(strings.TrimPrefix(filepath.Ext(d.Name()), "."), exts) {
			return nil
		}
		ns := slug(strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())))
		if slugMatch(ns, fileSlugs) {
			candidates = append(candidates, p)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no slug match anywhere")
	}
	sort.Strings(candidates)
	return candidates[0], nil
}

func findAnyByExt(root string, exts []string) (string, error) {
	if len(exts) == 0 {
		return "", fmt.Errorf("no extension provided")
	}
	var matches []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if hasExt(strings.TrimPrefix(filepath.Ext(d.Name()), "."), exts) {
			matches = append(matches, p)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no files with allowed extensions")
	}
	sort.Strings(matches)
	return matches[0], nil
}

func allC64Exts() []string {
	return []string{"d64", "d71", "d81", "g64", "tap", "t64", "crt", "ef", "prg", "zip"}
}

func pickFilesInDir(dir string, exts []string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var matches []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if len(exts) > 0 && !hasExt(strings.TrimPrefix(filepath.Ext(e.Name()), "."), exts) {
			continue
		}
		matches = append(matches, filepath.Join(dir, e.Name()))
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no files in directory match allowed extensions")
	}
	sort.Strings(matches)
	return matches, nil
}

func sanitizeFileName(name string) string {
	base := strings.TrimSpace(name)
	ext := filepath.Ext(base)
	base = strings.TrimSuffix(base, ext)
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
	if ext != "" {
		return clean + strings.ToLower(ext)
	}
	return clean
}

func findBySlug(root string, dirSlugs []string, fileSlugs []string, exts []string) (string, error) {
	if len(dirSlugs) == 0 || len(fileSlugs) == 0 {
		return "", fmt.Errorf("missing slug")
	}

	var candidates []string
	var extOnly []string

	walkErr := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if !slugMatch(slug(d.Name()), dirSlugs) {
			return nil
		}

		entries, err := os.ReadDir(p)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ns := slug(strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())))
			if slugMatch(ns, fileSlugs) {
				if len(exts) == 0 || hasExt(strings.TrimPrefix(filepath.Ext(e.Name()), "."), exts) {
					candidates = append(candidates, filepath.Join(p, e.Name()))
				}
				continue
			}
			if len(exts) == 0 || hasExt(strings.TrimPrefix(filepath.Ext(e.Name()), "."), exts) {
				extOnly = append(extOnly, filepath.Join(p, e.Name()))
			}
		}

		return nil
	})
	if walkErr != nil {
		return "", walkErr
	}
	if len(candidates) == 0 {
		if len(extOnly) == 0 {
			return "", fmt.Errorf("no slug match")
		}
		sort.Strings(extOnly)
		return extOnly[0], nil
	}

	sort.Strings(candidates)
	return candidates[0], nil
}

func slugIn(s string, arr []string) bool {
	for _, a := range arr {
		if a != "" && s == a {
			return true
		}
	}
	return false
}

func slugMatch(s string, arr []string) bool {
	for _, a := range arr {
		if a == "" {
			continue
		}
		if s == a || strings.Contains(s, a) || strings.Contains(a, s) {
			return true
		}
	}
	return false
}

func hasExt(ext string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(ext, a) {
			return true
		}
	}
	return false
}

func aliasesForExt(ct model.ContentType, primary string) []string {
	set := map[string]struct{}{}
	add := func(s string) {
		if s == "" {
			return
		}
		set[strings.ToLower(s)] = struct{}{}
	}

	switch ct {
	case model.ContentDisk:
		add("d64")
		add("d71")
		add("d81")
		add("g64")
	case model.ContentTape:
		add("tap")
		add("t64")
	case model.ContentCart:
		add("crt")
		add("ef")
	case model.ContentPrg:
		add("prg")
	case model.ContentZip:
		add("zip")
	default:
		// unknown: keep primary if provided
	}

	add(primary)

	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func slug(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r = r - 'A' + 'a'
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
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
