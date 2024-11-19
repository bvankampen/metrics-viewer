package realtimedata

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/sirupsen/logrus"
)

func (d *RealTimeData) AddDescription(metric string, line string) {
	description := strings.ReplaceAll(line, "# HELP "+metric, "")
	description = strings.TrimSpace(description)
	i := d.findMetricByName(metric)
	if i >= 0 {
		d.Metrics[i].Description = description
	} else {
		d.Metrics = append(d.Metrics, RealTimeDataMetric{
			Name:        metric,
			Description: description,
		})
	}
}

func (d *RealTimeData) AddType(metric string, line string) {
	metrictype := strings.ReplaceAll(line, "# TYPE "+metric, "")
	metrictype = strings.TrimSpace(metrictype)
	i := d.findMetricByName(metric)
	if i >= 0 {
		d.Metrics[i].Type = metrictype
	} else {
		d.Metrics = append(d.Metrics, RealTimeDataMetric{
			Name: metric,
			Type: metrictype,
		})
	}
}

func (d *RealTimeData) AddValue(metric string, line string) {
	l := strings.Split(line, " ")
	if len(l) < 2 { // needs better errorhandling?
		logrus.Errorf("metric line is not compliant %s", line)
		return
	}
	labels := strings.ReplaceAll(l[0], metric, " ")
	labels = strings.TrimSpace(labels)

	labels = labels[strings.Index(labels, "{")+1 : strings.Index(labels, "}")-1]

	value := l[1]

	i := d.findMetricByName(metric)
	if i == -1 { // if metric doesn't exists create metric and get index
		d.Metrics = append(d.Metrics, RealTimeDataMetric{
			Name: metric,
		})
		i = d.findMetricByName(metric)
	} else {
		// if type is histogram we only want the sum (?) not the buckets.
		if d.Metrics[i].Type == "histogram" {
			if !strings.HasPrefix(line, metric+"_sum") {
				return
			}
		}
	}
	hash := getHash(labels)
	vi := d.findValueByHash(i, hash)
	if vi >= 0 {
		d.Metrics[i].Values[vi].Value = value
	} else {
		newLabels := []RealTimeDataMetricLabel{}
		for _, ll := range strings.Split(labels, ",") {
			label := strings.Split(ll, "=")
			newLabels = append(newLabels, RealTimeDataMetricLabel{
				Label: label[0],
				Value: label[1],
			})
		}
		d.Metrics[i].Values = append(d.Metrics[i].Values, RealTimeDataMetricValue{
			SHA256: hash,
			Value:  value,
			Labels: newLabels,
		})
	}
}

func (d *RealTimeData) findValueByHash(index int, hash string) int {
	for i, m := range d.Metrics[index].Values {
		if m.SHA256 == hash {
			return i
		}
	}
	return -1
}

func (d *RealTimeData) findMetricByName(name string) int {
	for i, m := range d.Metrics {
		if name == m.Name {
			return i
		}
	}
	return -1
}

func getHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
