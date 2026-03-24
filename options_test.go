package imgpalette

import (
	"errors"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	if cfg.count != 5 || cfg.resizeTo != 256 || cfg.colorSpace != SpaceOKLab {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestResolveConfigInvalid(t *testing.T) {
	_, err := resolveConfig(Count(0))
	if !errors.Is(err, ErrInvalidCount) {
		t.Fatalf("expected ErrInvalidCount, got %v", err)
	}

	_, err = resolveConfig(Resize(0))
	if !errors.Is(err, ErrInvalidResize) {
		t.Fatalf("expected ErrInvalidResize, got %v", err)
	}
}

func TestResolveConfigSpace(t *testing.T) {
	cfg, err := resolveConfig(
		Space(SpaceLab),
		Resize(100),
		Count(3),
		SampleStride(2),
		MinSaturation(0.25),
		MinCoverage(0.15),
	)
	if err != nil {
		t.Fatalf("resolveConfig error: %v", err)
	}
	if cfg.colorSpace != SpaceLab ||
		cfg.resizeTo != 100 ||
		cfg.count != 3 ||
		cfg.sampleStride != 2 ||
		cfg.minSaturation != 0.25 ||
		cfg.minCoverage != 0.15 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestResolveConfigClampsThresholds(t *testing.T) {
	cfg, err := resolveConfig(MinSaturation(2), MinCoverage(-1))
	if err != nil {
		t.Fatalf("resolveConfig error: %v", err)
	}
	if cfg.minSaturation != 1 {
		t.Fatalf("expected minSaturation=1, got %v", cfg.minSaturation)
	}
	if cfg.minCoverage != 0 {
		t.Fatalf("expected minCoverage=0, got %v", cfg.minCoverage)
	}
}

// resolveConfig: minSaturation < 0 is clamped to 0.
func TestResolveConfigClampsSaturationBelow(t *testing.T) {
	cfg, err := resolveConfig(MinSaturation(-0.5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.minSaturation != 0 {
		t.Fatalf("expected minSaturation=0, got %v", cfg.minSaturation)
	}
}

// resolveConfig: minCoverage > 1 is clamped to 1.
func TestResolveConfigClampsCoverageAbove(t *testing.T) {
	cfg, err := resolveConfig(MinCoverage(1.5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.minCoverage != 1 {
		t.Fatalf("expected minCoverage=1, got %v", cfg.minCoverage)
	}
}

// resolveConfig: nil option is skipped without panicking.
func TestResolveConfigNilOption(t *testing.T) {
	_, err := resolveConfig(nil)
	if err != nil {
		t.Fatalf("unexpected error with nil option: %v", err)
	}
}
