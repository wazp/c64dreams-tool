package model

// Region identifies the television standard a title supports.
// "both" indicates PAL and NTSC are equally supported.
type Region string

const (
	RegionPAL  Region = "pal"
	RegionNTSC Region = "ntsc"
	RegionBoth Region = "both"
)

// ContentType captures the storage format for a variant.
type ContentType string

const (
	ContentUnknown ContentType = "unknown"
	ContentDisk    ContentType = "disk"
	ContentTape    ContentType = "tape"
	ContentPrg     ContentType = "prg"
	ContentZip     ContentType = "zip"
)

// Game represents a single C64 title and its playable variants.
type Game struct {
	ID             string // Stable identifier from the ingest source
	Title          string // Display title from metadata
	NormalizedName string // Canonical name used for output layout
	Region         Region // Primary region for the game
	Variants       []Variant
}

// Variant captures a specific playable build of a game (disk, tape, crack, etc.).
type Variant struct {
	Label           string       // Human-friendly label like "Disk", "Tape", or crack info
	Region          Region       // Variant-specific region if it differs from the game
	PreferredTarget TargetDevice // Best-fit target for this variant
	ContentType     ContentType
	SourcePath      string // Path to archive or file inside the source tree
	Notes           string // Free-form context for rules/normalization
}

// NormalizedName captures the result of a name normalization pass.
type NormalizedName struct {
	Original       string
	Normalized     string
	Truncated      bool
	Collision      bool
	CollisionGroup string
	CollisionIndex int
}

// NormalizedVariant describes a variant ready for layout decisions without filesystem details.
type NormalizedVariant struct {
	Label           NormalizedName
	Region          Region
	PreferredTarget TargetDevice
	ContentType     ContentType
	Notes           string
}

// NormalizedGame is the normalized representation of a Game with collision metadata.
type NormalizedGame struct {
	ID       string
	Title    string
	Name     NormalizedName
	Region   Region
	Target   TargetDevice
	Variants []NormalizedVariant
}
