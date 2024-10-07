package debug

import "sync/atomic"

// isDebug concurrent safe debug signal holder.
var isDebug atomic.Bool

// IsOn return the debug state, return true if it's on.
func IsOn() bool {
	return isDebug.Load()
}

// EnableDebug concurrent safe helper to enable debug information in the
// stderr.
func EnableDebug() {
	isDebug.Store(true)
}

// DisableDebug concurrent safe helper to disable debug information in the
// stderr.
func DisableDebug() {
	isDebug.Store(false)
}
