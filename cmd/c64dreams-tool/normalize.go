package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wazp/c64dreams-tool/internal/ingest"
	"github.com/wazp/c64dreams-tool/internal/normalize"
	"github.com/wazp/c64dreams-tool/pkg/model"
)

func newNormalizeCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "normalize",
		Short: "Normalize the C64 Dreams collection for hardware targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOptions(opts); err != nil {
				return err
			}

			if opts.sheet == "" {
				return fmt.Errorf("--sheet is required")
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

			if opts.json {
				payload := struct {
					Games []model.NormalizedGame `json:"games"`
				}{Games: normalized}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(payload)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Normalized %d games (target=%s max-len=%d)\n", len(normalized), opts.target, normOpts.EffectiveMaxLen())
			for i, ng := range normalized {
				if i >= 5 {
					fmt.Fprintf(cmd.OutOrStdout(), "... (%d more)\n", len(normalized)-i)
					break
				}
				fmt.Fprintf(cmd.OutOrStdout(), "- %s -> %s\n", ng.Title, ng.Name.Normalized)
			}

			return nil
		},
	}

	return cmd
}
