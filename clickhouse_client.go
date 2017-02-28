package telegraf_beget_clickhouse_plugin

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/roistat/go-clickhouse"
)

type ClickhouseClient struct {
	URL        string
	Username   string
	Password   string
	Database   string
	Timeout    time.Duration
	connection *clickhouse.Conn
}

func (c *ClickhouseClient) Connect() error {
	transport := clickhouse.NewHttpTransport()
	transport.Timeout = c.Timeout

	c.connection = clickhouse.NewConn(c.URL, transport)

	err := c.connection.Ping()
	if err != nil {
		return err
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
# Username and password to auth
username = "default"
password = ""
# Database to insert data
database = "default"`
}

func (c *ClickhouseClient) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		table := metric.Name()

		var columns clickhouse.Columns
		var row clickhouse.Row

		for name, value := range metric.Tags() {
			columns = append(columns, name)
			row = append(row, value)
		}

		for name, value := range metric.Fields() {
			columns = append(columns, name)
			row = append(row, value)
		}

		columns = append(columns, "date", "datetime")

		date := metric.Time().Format("2006-01-02")
		datetime := metric.Time().Format("2006-01-02 15:04:05")
		row = append(row, date, datetime)

		query, err := clickhouse.BuildInsert(table, columns, row)
		fmt.Println(query.Stmt)
		if err != nil {
			return err
		}

		err = query.Exec(c.connection)
		if err != nil {
			return err
		}
	}
	return nil
}

func newClickhouse() *ClickhouseClient {
	return &ClickhouseClient{
		Timeout: time.Minute,
	}
}

func init() {
	outputs.Add("clickhouse", func() telegraf.Output {
		return newClickhouse()
	})
}
