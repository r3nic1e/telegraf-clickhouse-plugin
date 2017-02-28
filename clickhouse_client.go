package telegraf_beget_clickhouse_plugin

import (
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/roistat/go-clickhouse"
)

type ClickhouseClient struct {
	URL        string
	Username   string
	Password   string
	Database   string
	Timeout    internal.Duration
	connection *clickhouse.Conn
}

func (c *ClickhouseClient) Connect() error {
	transport := clickhouse.NewHttpTransport()
	transport.Timeout = c.Timeout.Duration

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
	return ""
}
func (c *ClickhouseClient) Write(metrics []telegraf.Metric) error {
	spew.Dump(metrics)
	return nil
}

func newClickhouse() *ClickhouseClient {
	return &ClickhouseClient{
		Timeout: internal.Duration{Duration: time.Minute},
	}
}

func init() {
	outputs.Add("clickhouse_client", func() telegraf.Output {
		return newClickhouse()
	})
}
