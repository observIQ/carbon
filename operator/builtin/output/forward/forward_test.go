package forward

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/observiq/stanza/entry"
	"github.com/observiq/stanza/operator/buffer"
	"github.com/observiq/stanza/operator/helper"
	"github.com/observiq/stanza/testutil"
	"github.com/stretchr/testify/require"
)

func TestForwardOutput(t *testing.T) {
	received := make(chan []byte, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, _ := ioutil.ReadAll(req.Body)
		received <- body
	}))

	cfg := NewForwardOutputConfig("test")
	memoryCfg := buffer.NewMemoryBufferConfig()
	memoryCfg.MaxChunkDelay = helper.NewDuration(50 * time.Millisecond)
	cfg.BufferConfig = buffer.Config{
		Builder: memoryCfg,
	}
	cfg.Address = srv.URL

	ops, err := cfg.Build(testutil.NewBuildContext(t))
	require.NoError(t, err)
	forwardOutput := ops[0].(*ForwardOutput)

	e := entry.New()
	e.Record = "test"
	e.Timestamp = e.Timestamp.Round(time.Second)
	require.NoError(t, forwardOutput.Start())
	defer forwardOutput.Stop()
	require.NoError(t, forwardOutput.Process(context.Background(), e))

	select {
	case <-time.After(time.Second):
		require.FailNow(t, "Timed out waiting for server to receive entry")
	case body := <-received:
		var entries []*entry.Entry
		require.NoError(t, json.Unmarshal(body, &entries))
		require.Equal(t, []*entry.Entry{e}, entries)
	}

}