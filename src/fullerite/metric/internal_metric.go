package metric

// InternalMetrics holds the key:value pairs for counters/gauges
type InternalMetrics struct {
	Counters   map[string]float64
	Gauges     map[string]float64
	Dimensions map[string]string
}

// NewInternalMetrics initializes the internal components of InternalMetrics
func NewInternalMetrics() *InternalMetrics {
	inst := new(InternalMetrics)
	inst.Counters = make(map[string]float64)
	inst.Gauges = make(map[string]float64)
	inst.Dimensions = make(map[string]string)
	return inst
}
