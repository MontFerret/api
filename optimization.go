package api

type OptimizationLevel int

const (
	OptimizationNone OptimizationLevel = iota
	OptimizationBasic
	OptimizationFull
	OptimizationAggressive
)
