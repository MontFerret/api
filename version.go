package api

// Version is an opaque, implementation-provided runtime version value.
// It preserves the original string without normalization or validation and
// does not require any particular versioning scheme.
type Version string

// String returns the implementation-provided version string unchanged.
func (v Version) String() string {
	return string(v)
}
