package output

import (
	"context"

	"github.com/observiq/carbon/entry"
	"github.com/observiq/carbon/plugin"
	"github.com/observiq/carbon/plugin/helper"
)

func init() {
	plugin.Register("drop_output", func() plugin.Builder { return NewDropOutputConfig("") })
}

func NewDropOutputConfig(pluginID string) *DropOutputConfig {
	return &DropOutputConfig{
		OutputConfig: helper.NewOutputConfig(pluginID, "drop_output"),
	}
}

// DropOutputConfig is the configuration of a drop output plugin.
type DropOutputConfig struct {
	helper.OutputConfig `yaml:",inline"`
}

// Build will build a drop output plugin.
func (c DropOutputConfig) Build(context plugin.BuildContext) (plugin.Plugin, error) {
	outputPlugin, err := c.OutputConfig.Build(context)
	if err != nil {
		return nil, err
	}

	dropOutput := &DropOutput{
		OutputPlugin: outputPlugin,
	}

	return dropOutput, nil
}

// DropOutput is a plugin that consumes and ignores incoming entries.
type DropOutput struct {
	helper.OutputPlugin
}

// Process will drop the incoming entry.
func (p *DropOutput) Process(ctx context.Context, entry *entry.Entry) error {
	return nil
}
