package result

import "errors"

var (
	// ErrInUse indicates that another Consume or Collect operation already owns
	// consumption and stopping has not begun. The rejected call leaves it unaffected.
	// Implementations may wrap or join this error; callers use errors.Is.
	ErrInUse = errors.New("output is already being consumed")

	// ErrClosed indicates that closure or finalization has begun or completed and
	// the output no longer accepts consumption. It also identifies active consumption
	// interrupted by explicit Close when closure wins before outcome commitment.
	// Automatic finalization does not add it to the original consumption result,
	// and repeated Close returns the cleanup outcome rather than this lifecycle error.
	// Implementations may wrap or join this error; callers use errors.Is.
	ErrClosed = errors.New("output is closed")
)
