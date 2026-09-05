package walkforward

import (
	"errors"
)

// Walk Forward Analysis errors
var (
	// ErrInvalidTrainBars indicates train bars <= 0
	ErrInvalidTrainBars = errors.New("train_bars must be positive")

	// ErrInvalidTestBars indicates test bars <= 0
	ErrInvalidTestBars = errors.New("test_bars must be positive")

	// ErrInsufficientBars indicates not enough bars for a window
	ErrInsufficientBars = errors.New("insufficient bars for train+test window")

	// ErrInvalidWindowID indicates window ID out of range
	ErrInvalidWindowID = errors.New("invalid window ID")

	// ErrNoWindows indicates no windows generated
	ErrNoWindows = errors.New("no windows generated")

	// ErrIncompleteWindows indicates not all windows have results
	ErrIncompleteWindows = errors.New("not all windows have results")

	// ErrNoResults indicates no window results available
	ErrNoResults = errors.New("no window results available")
)
