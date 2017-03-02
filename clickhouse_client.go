package telegraf_beget_clickhouse_plugin

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/roistat/go-clickhouse"
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

	metricTime := metric.Time().Add(- 6 * time.Hour)
	date := metricTime.Format("2006-01-02")
	datetime := metricTime.Format("2006-01-02 15:04:05")
	cm.AddData("date", date, true)
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
func (cms *clickhouseMetrics) GetRowsByColumns(columns []string) clickhouse.Rows {
	rows := make(clickhouse.Rows, 0)

	for _, metric := range *cms {
		row := make(clickhouse.Row, 0)
		for _, column := range columns {
			row = append(row, (*metric)[column])
		}
		rows = append(rows, row)
	}

	return rows
}

type ClickhouseClient struct {
	URL      string
	Database string
	SQLs     []string `toml:"create_sql"`

	timeout    time.Duration
	connection *clickhouse.Conn
}

func (c *ClickhouseClient) Connect() error {
	transport := clickhouse.NewHttpTransport()
	transport.Timeout = c.timeout

	c.connection = clickhouse.NewConn(c.URL, transport)

	err := c.connection.Ping()
	if err != nil {
		return err
	}

	for _, create_sql := range c.SQLs {
		query := clickhouse.NewQuery(create_sql)
		err = query.Exec(c.connection)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ClickhouseClient) Close() error {
	return nil
}

func (c *ClickhouseClient) Description() string {
	return "Configuration for clickhouse server to send metrics to"
}

func (c *ClickhouseClient) SampleConfig() string {
	return `
# URL to connect
url = "http://localhost:8123"
# Database to use
database = "default"
# SQLs to create tables
create_sql = ["CREATE TABLE IF NOT EXISTS blablabla""]`
}

func (c *ClickhouseClient) Write(metrics []telegraf.Metric) (err error) {
	err = nil
	inserts := make(map[string]*clickhouseMetrics)

	for _, metric := range metrics {
		table := c.Database + "." + metric.Name()

		if _, exists := inserts[table]; !exists {
			inserts[table] = &clickhouseMetrics{}
		}

		inserts[table].AddMetric(metric)
	}

	for table, insert := range inserts {
		if len(*insert) == 0 {
			continue
		}

		columns := insert.GetColumns()
		rows := insert.GetRowsByColumns(columns)

		var query clickhouse.Query
		query, err = clickhouse.BuildMultiInsert(table, columns, rows)
		if err != nil {
			continue
		}

		err = query.Exec(c.connection)
		if err != nil {
			continue
		}

	}
	return err
}

func newClickhouse() *ClickhouseClient {
	return &ClickhouseClient{
		Database: "default",
		timeout:  time.Minute,
	}
}

func init() {
	outputs.Add("clickhouse", func() telegraf.Output {
		return newClickhouse()
	})
}
