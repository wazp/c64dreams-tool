package layout

// Options defines how output paths should be organized.
type Options struct {
	BaseDir         string
	GroupByMedia    bool
	GroupByAlpha    bool
	AlphaBucketSize int
}

// defaultAlphaBucketSize is used when GroupByAlpha is true but size is not set.
const defaultAlphaBucketSize = 1
