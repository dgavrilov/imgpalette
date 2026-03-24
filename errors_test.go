package imgpalette

import "testing"

func TestErrorsAreDefined(t *testing.T) {
	errs := []error{
		ErrNilImage,
		ErrInvalidCount,
		ErrInvalidResize,
		ErrOpenImage,
		ErrDecodeImage,
		ErrNoColors,
	}
	for _, err := range errs {
		if err == nil {
			t.Fatal("expected non-nil error")
		}
	}
}
