package imgpalette

// ColorSpace selects distance space for palette extraction.
type ColorSpace int

const (
	// SpaceRGB uses RGB color space.
	SpaceRGB ColorSpace = iota
	// SpaceLab uses CIE Lab color space.
	SpaceLab
	// SpaceOKLab uses OKLab color space.
	SpaceOKLab
)

type config struct {
	count         int
	sampleStride  int
	resizeTo      int
	colorSpace    ColorSpace
	filterGray    bool
	minSaturation float64
	minCoverage   float64
}

func defaultConfig() config {
	return config{
		count:         5,
		sampleStride:  1,
		resizeTo:      256,
		colorSpace:    SpaceOKLab,
		minSaturation: 0,
		minCoverage:   0,
	}
}

// Option configures extraction-related behavior (Extract*, Dominant*, Accent, Quantize).
type Option func(*config)

// Count sets the maximum number of colors to return.
func Count(n int) Option {
	return func(cfg *config) {
		cfg.count = n
	}
}

// SampleStride sets pixel sampling step. Values <= 1 are treated as 1.
func SampleStride(n int) Option {
	return func(cfg *config) {
		cfg.sampleStride = n
	}
}

// Resize sets max image side before extraction (never upscales).
func Resize(n int) Option {
	return func(cfg *config) {
		cfg.resizeTo = n
	}
}

// Space sets preferred color space for extraction heuristics.
func Space(space ColorSpace) Option {
	return func(cfg *config) {
		cfg.colorSpace = space
	}
}

// MinSaturation sets minimum saturation threshold for accent-like selection.
func MinSaturation(v float64) Option {
	return func(cfg *config) {
		cfg.minSaturation = v
	}
}

// MinCoverage sets minimum coverage threshold for accent-like selection.
func MinCoverage(v float64) Option {
	return func(cfg *config) {
		cfg.minCoverage = v
	}
}

// FilterGray excludes near-gray colors from accent selection when enabled.
// Gray detection uses the MinSaturation threshold.
func FilterGray(enabled bool) Option {
	return func(cfg *config) {
		cfg.filterGray = enabled
	}
}

func resolveConfig(opts ...Option) (config, error) {
	cfg := defaultConfig()

	for _, option := range opts {
		if option == nil {
			continue
		}
		option(&cfg)
	}

	if cfg.count <= 0 {
		return config{}, ErrInvalidCount
	}
	if cfg.sampleStride <= 1 {
		cfg.sampleStride = 1
	}
	if cfg.resizeTo <= 0 {
		return config{}, ErrInvalidResize
	}
	if cfg.minSaturation < 0 {
		cfg.minSaturation = 0
	}
	if cfg.minSaturation > 1 {
		cfg.minSaturation = 1
	}
	if cfg.minCoverage < 0 {
		cfg.minCoverage = 0
	}
	if cfg.minCoverage > 1 {
		cfg.minCoverage = 1
	}

	return cfg, nil
}
