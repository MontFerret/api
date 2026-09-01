// Package diagnostics defines the portable structured reporting model shared
// by Ferret runtimes and tooling.
//
// Diagnostics contain universal source values and user-facing reporting
// details. Runtime and compiler implementations may retain richer internal
// error causes, but those implementation-specific causes are intentionally not
// part of this package.
package diagnostics
