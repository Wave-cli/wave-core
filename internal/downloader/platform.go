package downloader

import "runtime"

// os_ and arch_ provide indirection points for runtime values.
// In production these return runtime.GOOS and runtime.GOARCH.
var (
	os_   = func() string { return runtime.GOOS }
	arch_ = func() string { return runtime.GOARCH }
)
