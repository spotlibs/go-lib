package debug

import "sync/atomic"

// isDebug concurrent safe debug signal holder.
var isDebug atomic.Bool

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
