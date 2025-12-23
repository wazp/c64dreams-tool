package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wazp/c64dreams-tool/internal/ingest"
)

func newIngestCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Load and echo metadata from the C64 Dreams spreadsheet",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.sheet == "" {
				return fmt.Errorf("--sheet is required")
			}

			games, err := ingest.LoadCSV(cmd.Context(), opts.sheet)
			if err != nil {
				return err
			}

			if opts.json {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(games)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Loaded %d games\n", len(games))
			for i, g := range games {
				if i >= 5 {
					fmt.Fprintf(cmd.OutOrStdout(), "... (%d more)\n", len(games)-i)
					break
				}
				label := ""
				if len(g.Variants) > 0 {
					label = g.Variants[0].Label
				}
				fmt.Fprintf(cmd.OutOrStdout(), "- %s (id=%s variant=%s)\n", g.Title, g.ID, label)
			}

			return nil
		},
	}

	return cmd
}
