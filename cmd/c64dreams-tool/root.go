package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wazp/c64dreams-tool/pkg/model"
)

type options struct {
	input      string
	output     string
	sheet      string
	target     model.TargetDevice
	maxNameLen int
	region     string
	groupBy    string
	dryRun     bool
	overwrite  bool
	groupMedia bool
	groupAlpha bool
	alphaSize  int
	json       bool
}

func newRootCmd() *cobra.Command {
	opts := &options{
		target:    model.TargetSD2IEC,
		region:    "both",
		groupBy:   "letter",
		dryRun:    true,
		alphaSize: 1,
	}

	cmd := &cobra.Command{
		Use:           "c64dreams",
		Short:         "C64 Dreams tooling CLI",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().StringVar(&opts.input, "input", "", "Path to the C64 Dreams directory")
	cmd.PersistentFlags().StringVar(&opts.output, "output", "", "Destination path for generated files")
	cmd.PersistentFlags().StringVar(&opts.sheet, "sheet", "", "Path to the spreadsheet CSV with metadata")
	cmd.PersistentFlags().StringVar((*string)(&opts.target), "target", string(opts.target), "Target device: sd2iec, pi1541, kungfuflash, or ultimate")
	cmd.PersistentFlags().IntVar(&opts.maxNameLen, "max-name-len", 0, "Maximum filename length; uses target profile when zero")
	cmd.PersistentFlags().StringVar(&opts.region, "region", opts.region, "Region filter: pal, ntsc, or both")
	cmd.PersistentFlags().StringVar(&opts.groupBy, "group-by", opts.groupBy, "Grouping strategy: letter or none")
	cmd.PersistentFlags().BoolVar(&opts.dryRun, "dry-run", opts.dryRun, "Preview actions without writing files")
	cmd.PersistentFlags().BoolVar(&opts.overwrite, "overwrite", false, "Allow overwriting existing files when applying layout")
	cmd.PersistentFlags().BoolVar(&opts.groupMedia, "group-media", false, "Group output by media type (disks/tape/cart)")
	cmd.PersistentFlags().BoolVar(&opts.groupAlpha, "group-alpha", false, "Group output alphabetically")
	cmd.PersistentFlags().IntVar(&opts.alphaSize, "alpha-bucket-size", opts.alphaSize, "Alphabetical bucket size when grouping")
	cmd.PersistentFlags().BoolVar(&opts.json, "json", false, "Emit JSON output for automation")

	cmd.AddCommand(newNormalizeCmd(opts))
	cmd.AddCommand(newScanCmd(opts))
	cmd.AddCommand(newIngestCmd(opts))

	return cmd
}

func validateOptions(opts *options) error {
	switch opts.target {
	case model.TargetSD2IEC, model.TargetPi1541, model.TargetKungFuFlash, model.TargetUltimate:
	default:
		return fmt.Errorf("invalid target %q (expected one of: %s)", opts.target, strings.Join(allowedTargets(), ", "))
	}

	switch opts.region {
	case "pal", "ntsc", "both":
	default:
		return fmt.Errorf("invalid region %q (expected pal, ntsc, or both)", opts.region)
	}

	switch opts.groupBy {
	case "letter", "none":
	default:
		return fmt.Errorf("invalid group-by %q (expected letter or none)", opts.groupBy)
	}

	if opts.maxNameLen < 0 {
		return fmt.Errorf("max-name-len must be zero or positive")
	}

	if opts.alphaSize < 0 {
		return fmt.Errorf("alpha-bucket-size must be zero or positive")
	}

	return nil
}

func allowedTargets() []string {
	return []string{
		string(model.TargetSD2IEC),
		string(model.TargetPi1541),
		string(model.TargetKungFuFlash),
		string(model.TargetUltimate),
	}
}
