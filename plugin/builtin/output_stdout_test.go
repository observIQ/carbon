package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/bluemedora/bplogagent/entry"
	"github.com/bluemedora/bplogagent/plugin/helper"
	"github.com/bluemedora/bplogagent/plugin/testutil"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
)

func TestStdoutPlugin(t *testing.T) {
	cfg := StdoutConfig{
		BasicPluginConfig: helper.BasicPluginConfig{
			PluginID:   "test_plugin_id",
			PluginType: "stdout",
		},
	}

	plugin, err := cfg.Build(testutil.NewTestBuildContext(t))
	require.NoError(t, err)

	var buf bytes.Buffer
	plugin.(*StdoutPlugin).encoder = jsoniter.ConfigFastest.NewEncoder(&buf)

	ts := time.Unix(1591042864, 0)
	e := &entry.Entry{
		Timestamp: ts,
		Record:    "test record",
	}
	err = plugin.Process(context.Background(), e)
	require.NoError(t, err)

	marshalledTimestamp, err := json.Marshal(ts)
	require.NoError(t, err)

	expected := `{"timestamp":` + string(marshalledTimestamp) + `,"record":"test record"}` + "\n"
	require.Equal(t, expected, buf.String())
}