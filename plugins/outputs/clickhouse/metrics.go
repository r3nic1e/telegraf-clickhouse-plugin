package clickhouse

import (
	"github.com/influxdata/telegraf"
)

type clickhouseMetric map[string]interface{}

func (cm *clickhouseMetric) GetColumns() []string {
	columns := make([]string, 0)

	for column := range *cm {
		columns = append(columns, column)
	}
	return columns
}
func (cm *clickhouseMetric) AddData(name string, value interface{}, overwrite bool) {
	if _, exists := (*cm)[name]; !overwrite && exists {
		return
	}

	(*cm)[name] = value
}

func newClickhouseMetric(metric telegraf.Metric) *clickhouseMetric {
	cm := &clickhouseMetric{}

	for name, value := range metric.Fields() {
		cm.AddData(name, value, true)
	}
	for name, value := range metric.Tags() {
		cm.AddData(name, value, true)
	}

	metricTime := metric.Time()
	datetime := metricTime.Format("2006-01-02 15:04:05")
	cm.AddData("datetime", datetime, true)

	return cm
}

type clickhouseMetrics []*clickhouseMetric

func (cms *clickhouseMetrics) GetColumns() []string {
	if len(*cms) == 0 {
		return []string{}
	}

	randomMetric := (*cms)[0] // all previous metrics are same
	return randomMetric.GetColumns()
}
func (cms *clickhouseMetrics) AddMissingColumn(name string, value interface{}) {
	for _, metric := range *cms {
		metric.AddData(name, value, false)
	}
}
func (cms *clickhouseMetrics) AddMetric(metric telegraf.Metric) {
	newMetric := newClickhouseMetric(metric)

	if len(*cms) > 0 {
		randomMetric := (*cms)[0] // all previous metrics are same

		for name := range *newMetric {
			if _, exists := (*randomMetric)[name]; !exists {
				cms.AddMissingColumn(name, 0)
			}
		}

		for name := range *randomMetric {
			if _, exists := (*newMetric)[name]; !exists {
				newMetric.AddData(name, 0, false)
			}
		}
	}

	*cms = append(*cms, newMetric)
}
func (cms *clickhouseMetrics) GetRowsByColumns(columns []string) [][]interface{} {
	rows := make([][]interface{}, len(*cms))

	for i, metric := range *cms {
		rows[i] = make([]interface{}, len(columns))
		for j, column := range columns {
			rows[i][j] = (*metric)[column]
		}
	}

	return rows
}
