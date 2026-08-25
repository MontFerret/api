package api

import "fmt"

type OptimizationLevel int

const (
	OptimizationNone OptimizationLevel = iota
	OptimizationBasic
	OptimizationFull
	OptimizationAggressive
)

func ParseOptimizationLevel(level int) (OptimizationLevel, error) {
	switch level {
	case 0:
		return OptimizationNone, nil
	case 1:
		return OptimizationBasic, nil
	case 2:
		return OptimizationFull, nil
	case 3:
		return OptimizationAggressive, nil
	default:
		return OptimizationNone, fmt.Errorf("invalid optimization level: %d", level)
	}
}
