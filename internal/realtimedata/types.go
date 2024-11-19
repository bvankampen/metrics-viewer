package realtimedata

type RealTimeData struct {
	Metrics []RealTimeDataMetric
}

type RealTimeDataMetric struct {
	Name        string
	Description string
	Type        string
	Values      []RealTimeDataMetricValue
}

type RealTimeDataMetricValue struct {
	Labels []RealTimeDataMetricLabel
	Value  string
	SHA256 string
}

type RealTimeDataMetricLabel struct {
	Label string
	Value string
}
