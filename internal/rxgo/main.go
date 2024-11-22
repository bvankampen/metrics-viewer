package rxgo

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bvankampen/metrics-viewer/internal/realtimedata"
	"github.com/bvankampen/metrics-viewer/internal/scraper"
	"github.com/bvankampen/metrics-viewer/internal/ui"
	"github.com/reactivex/rxgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func Run(ctx *cli.Context) {
	scraper := scraper.Scraper{}
	scraper.Init(ctx)

	timer := rxgo.Interval(rxgo.WithDuration(time.Duration(scraper.ScrapeInterval()) * time.Second)).
		Map(func(ctx context.Context, _ interface{}) (interface{}, error) {
			return time.Now().Unix(), nil
		})

	ui := ui.NewAppUI(ctx)

	filterChan := make(chan rxgo.Item)
	sortChan := make(chan rxgo.Item)
	filterObservable := rxgo.FromChannel(filterChan)
	sortObservable := rxgo.FromChannel(sortChan)

	go func() {
		filterChan <- rxgo.Of("")
		sortChan <- rxgo.Of(map[string]interface{}{
			"column":    0,
			"ascending": true,
		})
	}()

	dataSource := rxgo.Create([]rxgo.Producer{
		func(ctx context.Context, ch chan<- rxgo.Item) {
			for {
				data, err := scraper.Scrape()
				if err != nil {
					ch <- rxgo.Error(err)
					logrus.Errorf("Error scraping metrics: %v", err)
					continue
				}
				ch <- rxgo.Of(data)
				time.Sleep(1 * time.Second)
			}
		},
	})

	pipeline := rxgo.CombineLatest(
		func(i ...interface{}) interface{} {
			return map[string]interface{}{
				"data":        i[0],
				"currentTime": i[1],
				"filterState": i[2],
				"sortState":   i[3],
			}
		},
		[]rxgo.Observable{
			dataSource,
			timer,
			filterObservable,
			sortObservable,
		},
	).Map(func(ctx context.Context, item interface{}) (interface{}, error) {
		vMap := item.(map[string]interface{})
		originalData := vMap["data"].(realtimedata.RealTimeData)
		filter := vMap["filterState"].(string)
		sortConfig := vMap["sortState"].(map[string]interface{})
		sortColumn := sortConfig["column"].(int)
		sortAsc := sortConfig["ascending"].(bool)

		filteredData := applyFilter(originalData, filter)
		filteredSortedData := applySort(filteredData, sortColumn, sortAsc)

		vMap["filteredSortedData"] = filteredSortedData
		return vMap, nil
	}).Map(func(ctx context.Context, item interface{}) (interface{}, error) {
		vMap := item.(map[string]interface{})
		filteredSortedData := vMap["filteredSortedData"].(realtimedata.RealTimeData)

		uiData := convertToTableRows(filteredSortedData)
		vMap["uiData"] = uiData
		return vMap, nil
	})

	observeChan := make(chan interface{})
	go func() {
		for item := range pipeline.Observe() {
			if item.E != nil {
				logrus.Errorf("Error in pipeline.Observe(): %v", item.E)
				continue
			}
			observeChan <- item.V
		}
		close(observeChan)
	}()

	ui.SetFilterHandler(func(newFilter string) {
		filterChan <- rxgo.Of(newFilter)
	})
	ui.SetSortHandler(func(column int, ascending bool) {
		sortChan <- rxgo.Of(map[string]interface{}{
			"column":    column,
			"ascending": ascending,
		})
	})

	ui.Run(observeChan)
}

func applyFilter(data realtimedata.RealTimeData, filter string) realtimedata.RealTimeData {
	if filter == "" {
		return data
	}
	regex, err := regexp.Compile(filter)
	if err != nil {
		logrus.Errorf("Invalid filter regex: %v", err)
		return data
	}

	filteredMetrics := []realtimedata.RealTimeDataMetric{}
	for _, metric := range data.Metrics {
		filteredValues := []realtimedata.RealTimeDataMetricValue{}
		for _, value := range metric.Values {
			labelMatches := false
			for _, label := range value.Labels {
				if regex.MatchString(label.Label) || regex.MatchString(label.Value) {
					labelMatches = true
					break
				}
			}
			if regex.MatchString(metric.Name) || regex.MatchString(value.Value) || labelMatches {
				filteredValues = append(filteredValues, value)
			}
		}
		if len(filteredValues) > 0 {
			filteredMetrics = append(filteredMetrics, realtimedata.RealTimeDataMetric{
				Name:        metric.Name,
				Description: metric.Description,
				Type:        metric.Type,
				Values:      filteredValues,
			})
		}
	}

	return realtimedata.RealTimeData{Metrics: filteredMetrics}
}

func applySort(data realtimedata.RealTimeData, column int, ascending bool) realtimedata.RealTimeData {
	sortedMetrics := []realtimedata.RealTimeDataMetric{}
	for _, metric := range data.Metrics {
		values := metric.Values
		sort.Slice(values, func(i, j int) bool {
			var a, b string
			switch column {
			case 0:
				a, b = metric.Name, metric.Name
			case 1:
				aLabels := labelsToString(values[i].Labels)
				bLabels := labelsToString(values[j].Labels)
				a, b = aLabels, bLabels
			case 2:
				a, b = values[i].Value, values[j].Value
			}
			if ascending {
				return a < b
			}
			return a > b
		})
		sortedMetrics = append(sortedMetrics, realtimedata.RealTimeDataMetric{
			Name:        metric.Name,
			Description: metric.Description,
			Type:        metric.Type,
			Values:      values,
		})
	}

	return realtimedata.RealTimeData{Metrics: sortedMetrics}
}

func labelsToString(labels []realtimedata.RealTimeDataMetricLabel) string {
	var builder strings.Builder
	for _, label := range labels {
		builder.WriteString(fmt.Sprintf("%s=%s ", label.Label, label.Value))
	}
	return strings.TrimSpace(builder.String())
}

func convertToTableRows(data realtimedata.RealTimeData) []ui.TableRow {
	tableRows := []ui.TableRow{}
	labelKeys := getUniqueLabelKeys(data) // Get all unique label keys sorted

	for _, metric := range data.Metrics {
		for _, value := range metric.Values {
			// Initialize labels with empty values for consistent columns
			labels := make(map[string]string)
			for _, key := range labelKeys {
				labels[key] = "" // Default empty value
			}

			// Populate labels with actual data
			for _, label := range value.Labels {
				labels[label.Label] = label.Value
			}

			// Create and append the TableRow
			tableRows = append(tableRows, ui.TableRow{
				MetricName: metric.Name,
				Type:       metric.Type,
				Labels:     labels,
				Value:      value.Value,
			})
		}
	}
	return tableRows
}

// Helper function to get all unique label keys sorted
func getUniqueLabelKeys(data realtimedata.RealTimeData) []string {
	labelSet := make(map[string]struct{})
	for _, metric := range data.Metrics {
		for _, value := range metric.Values {
			for _, label := range value.Labels {
				labelSet[label.Label] = struct{}{}
			}
		}
	}

	// Extract keys and sort them
	labelKeys := make([]string, 0, len(labelSet))
	for key := range labelSet {
		labelKeys = append(labelKeys, key)
	}
	sort.Strings(labelKeys)
	return labelKeys
}
