package debugger

// Reason identifies why a debug execution stopped.
type Reason string

const (
	ReasonEntry        Reason = "entry"
	ReasonBreakpoint   Reason = "breakpoint"
	ReasonStep         Reason = "step"
	ReasonPause        Reason = "pause"
	ReasonRuntimeError Reason = "runtime-error"
	ReasonCompleted    Reason = "completed"
	ReasonTerminated   Reason = "terminated"
)

func ReasonFromString(s string) Reason {
	switch s {
	case "entry":
		return ReasonEntry
	case "breakpoint":
		return ReasonBreakpoint
	case "step":
		return ReasonStep
	case "pause":
		return ReasonPause
	case "runtime-error":
		return ReasonRuntimeError
	case "completed":
		return ReasonCompleted
	case "terminated":
		return ReasonTerminated
	default:
		return ""
	}
}
