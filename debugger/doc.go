// Package debugger defines portable contracts and values for controlling and
// inspecting Ferret debug sessions, including events, breakpoints, frames, and
// values.
// Retained event output is detached result.Content, preserving absent versus
// present empty data and remaining readable without consuming another observer's
// output or retaining a live result.Output handle.
package debugger
