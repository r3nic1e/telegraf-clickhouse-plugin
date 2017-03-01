package telegraf_beget_clickhouse_plugin

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/roistat/go-clickhouse"
)

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

func (c *ClickhouseClient) Write(metrics []telegraf.Metric) error {
	inserts := make(map[string](map[string][]interface{}))

	for _, metric := range metrics {
		table := c.Database + "." + metric.Name()

		if _, exists := inserts[table]; !exists {
			insert := make(map[string][]interface{})
			for name := range metric.Tags() {
				insert[name] = make([]interface{}, 0)
			}

			for name := range metric.Fields() {
				insert[name] = make([]interface{}, 0)
			}
			insert["date"] = make([]interface{}, 0)
			insert["datetime"] = make([]interface{}, 0)

			inserts[table] = insert
		}

		insert := inserts[table]
		for name := range insert {
			if value, ok := metric.Fields()[name]; ok {
				insert[name] = append(insert[name], value)
			} else if value, ok := metric.Tags()[name]; ok {
				insert[name] = append(insert[name], value)
			}
		}

		date := metric.Time().Format("2006-01-02")
		datetime := metric.Time().Format("2006-01-02 15:04:05")
		insert["date"] = append(insert["date"], date)
		insert["datetime"] = append(insert["datetime"], datetime)
	}

	for table, insert := range inserts {
		var columns clickhouse.Columns
		var rows clickhouse.Rows

		for name, values := range insert {
			columns = append(columns, name)

			length := len(values)
			if len(rows) == 0 {
				rows = make(clickhouse.Rows, length)
			}

			for i, value := range values {
				rows[i] = append(rows[i], value)
			}
		}

		query, err := clickhouse.BuildMultiInsert(table, columns, rows)
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
		Database: "default",
		timeout:  time.Minute,
	}
}

func init() {
	outputs.Add("clickhouse", func() telegraf.Output {
		return newClickhouse()
	})
}
