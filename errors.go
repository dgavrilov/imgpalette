package imgpalette

import "errors"

var (
	// ErrNilImage reports a nil image input.
	ErrNilImage = errors.New("imgpalette: nil image")
	// ErrInvalidCount reports invalid requested color count.
	ErrInvalidCount = errors.New("imgpalette: invalid color count")
	// ErrInvalidResize reports invalid resize value.
	ErrInvalidResize = errors.New("imgpalette: invalid resize value")
	// ErrOpenImage reports image file open failure.
	ErrOpenImage = errors.New("imgpalette: open image failed")
	// ErrDecodeImage reports image decode failure.
	ErrDecodeImage = errors.New("imgpalette: decode image failed")
	// ErrNoColors reports absence of extractable colors.
	ErrNoColors = errors.New("imgpalette: no colors found")
)
