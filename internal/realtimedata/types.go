package realtimedata

type RealTimeData struct {
	Metrics []RealTimeDataMetric
}

type RealTimeDataMetric struct {
	Name        string
	Description string
	Type        string
}
