package clickhouse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var (
	m telegraf.Metric
	TAGS = map[string]string{"name": "cpu1"}
	FIELDS = map[string]interface{}{"idle": 50, "sys": 30}
)

func init() {
	metric, err := metric.New(
		"cpu",
		TAGS,
		FIELDS,
		time.Now(),
	)
	if err != nil {
		panic(err)
	}

	m = metric
}

func TestCreateMetric(t *testing.T) {
	metric := newClickhouseMetric(m)
	assert.NotNil(t, metric)

	a := []string{"datetime"}
	for _, tag := range m.TagList() {
		assert.Equal(t, (*metric)[tag.Key], tag.Value)
		a = append(a, tag.Key)
	}
	for _, field := range m.FieldList() {
		assert.Equal(t, (*metric)[field.Key], field.Value)
		a = append(a, field.Key)
	}

	assert.ElementsMatch(t, metric.GetColumns(), a)
	assert.Equal(t, (*metric)["datetime"], m.Time().Format("2006-01-02 15:04:05"))
}

