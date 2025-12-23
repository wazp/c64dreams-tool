package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wazp/c64dreams-tool/internal/executor"
	"github.com/wazp/c64dreams-tool/internal/ingest"
	"github.com/wazp/c64dreams-tool/internal/layout"
	"github.com/wazp/c64dreams-tool/internal/normalize"
	"github.com/wazp/c64dreams-tool/pkg/model"
)

func newBuildCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Run ingest → normalize → collide → layout → execute",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOptions(opts); err != nil {
				return err
			}

			if opts.sheet == "" {
				return fmt.Errorf("--sheet is required")
			}

			if opts.input == "" {
				return fmt.Errorf("--input is required")
			}

			if opts.output == "" {
				return fmt.Errorf("--output is required")
			}

			games, err := ingest.LoadCSV(cmd.Context(), opts.sheet)
			if err != nil {
				return err
			}

			normOpts := normalize.Options{Target: opts.target, MaxNameLen: opts.maxNameLen}
			var normalized []model.NormalizedGame
			for _, g := range games {
				ng, err := normalize.NormalizeGame(g, normOpts)
				if err != nil {
					return err
				}
				normalized = append(normalized, ng)
			}

			normalized = normalize.ResolveCollisions(normalized, normOpts)

			layoutOpts := layout.Options{
				GroupByMedia:    opts.groupMedia,
				GroupByAlpha:    opts.groupAlpha,
				AlphaBucketSize: opts.alphaSize,
			}

			planned, err := layout.Plan(normalized, layoutOpts)
			if err != nil {
				return err
			}

			execOpts := executor.Options{
				InputRoot:  opts.input,
				OutputRoot: opts.output,
				DryRun:     opts.dryRun,
				Overwrite:  opts.overwrite,
			}

			results, execErr := executor.Apply(planned, execOpts)

			if opts.json {
				payload := struct {
					Plan    []layout.PlannedFile `json:"plan"`
					Results []jsonResult         `json:"results"`
					Error   string               `json:"error,omitempty"`
				}{
					Plan:    planned,
					Results: flattenResults(results),
				}
				if execErr != nil {
					payload.Error = execErr.Error()
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(payload)
			}

			if execErr != nil {
				return execErr
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Planned %d files\n", len(planned))
			for _, r := range results {
				fmt.Fprintf(cmd.OutOrStdout(), "%s -> %s (%s)\n", r.Source, r.Dest, r.Action)
			}
			if opts.dryRun {
				fmt.Fprintln(cmd.OutOrStdout(), "Dry-run enabled: no changes written")
			}

			return execErr
		},
	}

	return cmd
}

type jsonResult struct {
	Source string `json:"source"`
	Dest   string `json:"dest"`
	Action string `json:"action"`
	Error  string `json:"error,omitempty"`
}

func flattenResults(results []executor.Result) []jsonResult {
	out := make([]jsonResult, 0, len(results))
	for _, r := range results {
		jr := jsonResult{Source: r.Source, Dest: r.Dest, Action: r.Action}
		if r.Error != nil {
			jr.Error = r.Error.Error()
		}
		out = append(out, jr)
	}
	return out
}
