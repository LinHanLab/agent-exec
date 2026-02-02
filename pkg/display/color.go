package display

// ANSI color codes
const (
	Bold    = "\033[1m"
	Reset   = "\033[0m"
	Cyan    = "\033[36m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Red     = "\033[31m"
	Magenta = "\033[35m"

	// Text effects
	ReverseVideo = "\033[7m"

	// Combined styles
	BoldCyan   = "\033[1;36m"
	BoldYellow = "\033[1;33m"
	BoldGreen  = "\033[1;32m"
	BoldRed    = "\033[1;31m"
)
