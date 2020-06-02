package builtin

import (
	"context"
	"os"
	"sync"

	"github.com/bluemedora/bplogagent/entry"
	"github.com/bluemedora/bplogagent/plugin"
	"github.com/bluemedora/bplogagent/plugin/helper"
	jsoniter "github.com/json-iterator/go"
)

func init() {
	plugin.Register("stdout", &StdoutConfig{})
}

// StdoutConfig is the configuration of the Stdout plugin
type StdoutConfig struct {
	helper.BasicPluginConfig `yaml:",inline"`
}

// Build will build a stdout plugin.
func (c StdoutConfig) Build(context plugin.BuildContext) (plugin.Plugin, error) {
	basicPlugin, err := c.BasicPluginConfig.Build(context.Logger)
	if err != nil {
		return nil, err
	}

	return &StdoutPlugin{
		BasicPlugin: basicPlugin,
		encoder:     jsoniter.ConfigFastest.NewEncoder(os.Stdout),
	}, nil
}

// LoggerOutput is a plugin that logs entries using the internal logger.
type StdoutPlugin struct {
	helper.BasicPlugin
	helper.BasicLifecycle
	helper.BasicOutput
	encoder *jsoniter.Encoder
	mux     sync.Mutex
}

// Process will log entries received.
func (o *StdoutPlugin) Process(ctx context.Context, entry *entry.Entry) error {
	o.mux.Lock()
	err := o.encoder.Encode(entry)
	if err != nil {
		o.mux.Unlock()
		return err
	}
	o.mux.Unlock()
	return nil
}