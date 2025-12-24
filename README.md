# C64 Dreams Tool

Command-line pipeline to reshape the C64 Dreams collection into layouts suitable for real hardware. It reads the C64 Dreams CSV, normalizes names per target rules, plans output paths, and applies them with a dry-run-first executor.

## Quickstart (prebuilt binary)

1) Place the C64 Dreams CSV locally (e.g., `C64 Dreams - C64 Dreams.csv`).
2) Run a dry-run (no writes by default):
```bash
./c64dreams-tool build \
  --sheet "C64 Dreams - C64 Dreams.csv" \
  --input /path/to/c64dreams \
  --output /path/to/output \
  --target sd2iec \
  --group-media \
  --group-alpha \
  --alpha-bucket-size 1
```
3) Review the planned actions; nothing is written unless `--dry-run=false`.
4) Write outputs when ready:
```bash
./c64dreams-tool build ... --dry-run=false [--overwrite]
```
5) Automation/JSON output:
```bash
./c64dreams-tool build ... --json
```

## Flags (build command)

**Input/output**
- `--sheet <path>`: CSV export of C64 Dreams (required).
- `--input <dir>`: Source root containing the C64 Dreams files (required).
- `--output <dir>`: Destination root for generated layout (required).

**Target & naming**
- `--target {sd2iec|pi1541|kungfuflash|ultimate}`: Hardware profile (defaults to sd2iec).
- `--max-name-len <n>`: Override target filename length; 0 uses target default.

**Layout**
- `--group-media`: Group by media type (`disks`, `tape`, `cart`, `prg`, `zip`).
- `--group-alpha`: Group alphabetically; digits go under their leading digit.
- `--alpha-bucket-size <n>`: Bucket size for alpha grouping (default 1).

**Execution**
- `--dry-run`: Default true; list actions without writing.
- `--overwrite`: Allow overwriting existing files.
- `--json`: Emit plan and results as JSON (suppresses human logs).

## Utility commands
- `ingest --sheet <path> [--json]`: Preview CSV metadata as structured games/variants.
- `normalize --sheet <path> --target <device> [--max-name-len] [--json]`: Preview normalized names and collision resolution without writing files.
- `scan --input <dir> [--json]`: List C64-relevant files under the source root.

## Behavior
- Filenames are lowercased; apostrophes removed; underscores → spaces; other specials → dashes; extensions lowercased.
- If a source resolves to a directory, **all** C64-relevant files inside (disk/tape/cart/prg/zip) are copied to the destination directory.
- Media grouping is based on the variant’s content type, but sibling C64 files are also copied alongside (e.g., a cart variant with companion disks).
- Executor uses target-aware extension sets and slug/case-insensitive matching to locate sources; dry-run lists intended actions.

## Build from source

Prereqs: Go 1.25+

Clone and build:
```bash
git clone <repo-url>
cd c64dreams-tool
go build -o c64dreams-tool ./cmd/c64dreams-tool
```

Run tests:
```bash
go test ./...
```
