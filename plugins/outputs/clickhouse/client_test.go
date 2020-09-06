package clickhouse

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TEST_TABLE = "default.test"
)

func TestConnect(t *testing.T) {
	ch := newClickhouse()
	ch.URL = "tcp://127.0.0.1:9000"

	assert.NoError(t, ch.Connect())
}

func TestFailedConnect(t *testing.T) {
	ch := newClickhouse()
	ch.URL = "tcp://127.1.1.1:9999"

	assert.Error(t, ch.Connect())
}

func TestCreateTable(t *testing.T) {
	ch := newClickhouse()
	ch.URL = "tcp://127.0.0.1:9000"

	require.NoError(t, ch.Connect())

	_, err := ch.connection.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", TEST_TABLE))
	require.NoError(t, err)

	sql := fmt.Sprintf("CREATE TABLE %s ( `test` String ) ENGINE = Null", TEST_TABLE)
	ch.SQLs = []string{sql}

	require.NoError(t, ch.Connect())

	row := ch.connection.QueryRow(fmt.Sprintf("SHOW CREATE TABLE %s", TEST_TABLE))
	var create_sql string
	require.NoError(t, row.Scan(&create_sql))
	assert.Equal(t, sql, strings.Join(strings.Fields(create_sql), " "))
}
