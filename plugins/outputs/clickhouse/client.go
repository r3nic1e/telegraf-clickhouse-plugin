package clickhouse

import (
	"fmt"
	"strings"
	"time"
	"database/sql"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	_ "github.com/ClickHouse/clickhouse-go"
)

type ClickhouseClient struct {
	URL       string
	SQLs      []string `toml:"create_sql"`

	timeout    time.Duration
	connection *sql.DB
}

func (c *ClickhouseClient) Connect() error {
	conn, err := sql.Open("clickhouse", c.URL)
	if err != nil {
		return err
	}

	if err := conn.Ping(); err != nil {
		return err
	}

	c.connection = conn

	return c.createTables()
}

func (c *ClickhouseClient) createTables() error {
	for _, create_sql := range c.SQLs {
		_, err := c.connection.Exec(create_sql)
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
url = "tcp://127.0.0.1:9000"
# SQLs to create tables
create_sql = ["CREATE TABLE IF NOT EXISTS blablabla""]
# Time shift for timezone
time_shift = -3600`
}

func (c *ClickhouseClient) Write(metrics []telegraf.Metric) (err error) {
	inserts := make(map[string]*clickhouseMetrics)

	for _, metric := range metrics {
		table := metric.Name()

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

		tx, err := c.connection.Begin()
		if err != nil {
			return err
		}

		q := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES
		`, table, strings.Join(columns, ","))
		stmt, err := tx.Prepare(q)
		defer stmt.Close()

		for _, row := range rows {
			if _, err = stmt.Exec(row...); err != nil {
				return err
			}
		}

		if err = tx.Commit(); err != nil {
			return err
		}

	}
	return err
}

func newClickhouse() *ClickhouseClient {
	return &ClickhouseClient{
		timeout:  time.Minute,
	}
}

func init() {
	outputs.Add("clickhouse", func() telegraf.Output {
		return newClickhouse()
	})
}
