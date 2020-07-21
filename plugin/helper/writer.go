package helper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/observiq/carbon/entry"
	"github.com/observiq/carbon/plugin"
)

// WriterConfig is the configuration of a writer plugin.
type WriterConfig struct {
	BasicConfig `yaml:",inline"`
	OutputIDs   OutputIDs `json:"output" yaml:"output"`
}

// Build will build a writer plugin from the config.
func (c WriterConfig) Build(context plugin.BuildContext) (WriterPlugin, error) {
	basicPlugin, err := c.BasicConfig.Build(context)
	if err != nil {
		return WriterPlugin{}, err
	}

	writer := WriterPlugin{
		OutputIDs:   c.OutputIDs,
		BasicPlugin: basicPlugin,
	}
	return writer, nil
}

// SetNamespace will namespace the output ids of the writer.
func (c *WriterConfig) SetNamespace(namespace string, exclusions ...string) {
	c.BasicConfig.SetNamespace(namespace, exclusions...)
	for i, outputID := range c.OutputIDs {
		if CanNamespace(outputID, exclusions) {
			c.OutputIDs[i] = AddNamespace(outputID, namespace)
		}
	}
}

// WriterPlugin is a plugin that can write to other plugins.
type WriterPlugin struct {
	BasicPlugin
	OutputIDs     OutputIDs
	OutputPlugins []plugin.Plugin
}

// Write will write an entry to the outputs of the plugin.
func (w *WriterPlugin) Write(ctx context.Context, e *entry.Entry) {
	for i, plugin := range w.OutputPlugins {
		if i == len(w.OutputPlugins)-1 {
			_ = plugin.Process(ctx, e)
			return
		}
		plugin.Process(ctx, e.Copy())
	}
}

// CanOutput always returns true for a writer plugin.
func (w *WriterPlugin) CanOutput() bool {
	return true
}

// Outputs returns the outputs of the writer plugin.
func (w *WriterPlugin) Outputs() []plugin.Plugin {
	return w.OutputPlugins
}

// SetOutputs will set the outputs of the plugin.
func (w *WriterPlugin) SetOutputs(plugins []plugin.Plugin) error {
	outputPlugins := make([]plugin.Plugin, 0)

	for _, pluginID := range w.OutputIDs {
		plugin, ok := w.findPlugin(plugins, pluginID)
		if !ok {
			return fmt.Errorf("plugin '%s' does not exist", pluginID)
		}

		if !plugin.CanProcess() {
			return fmt.Errorf("plugin '%s' can not process entries", pluginID)
		}

		outputPlugins = append(outputPlugins, plugin)
	}

	// No outputs have been set, so use the next configured plugin
	if len(w.OutputIDs) == 0 {
		currentPluginIndex := -1
		for i, plugin := range plugins {
			if plugin.ID() == w.ID() {
				currentPluginIndex = i
				break
			}
		}
		if currentPluginIndex == -1 {
			return fmt.Errorf("unexpectedly could not find self in array of plugins")
		}
		nextPluginIndex := currentPluginIndex + 1
		if nextPluginIndex == len(plugins) {
			return fmt.Errorf("cannot omit output for the last plugin in the pipeline")
		}
		nextPlugin := plugins[nextPluginIndex]
		if !nextPlugin.CanProcess() {
			return fmt.Errorf("plugin '%s' cannot process entries, but it was selected as a receiver because 'output' was omitted", nextPlugin.ID())
		}
		outputPlugins = append(outputPlugins, nextPlugin)
	}

	w.OutputPlugins = outputPlugins
	return nil
}

// FindPlugin will find a plugin matching the supplied id.
func (w *WriterPlugin) findPlugin(plugins []plugin.Plugin, pluginID string) (plugin.Plugin, bool) {
	for _, plugin := range plugins {
		if plugin.ID() == pluginID {
			return plugin, true
		}
	}
	return nil, false
}

// OutputIDs is a collection of plugin IDs used as outputs.
type OutputIDs []string

// UnmarshalJSON will unmarshal a string or array of strings to OutputIDs.
func (o *OutputIDs) UnmarshalJSON(bytes []byte) error {
	var value interface{}
	err := json.Unmarshal(bytes, &value)
	if err != nil {
		return err
	}

	ids, err := o.fromInterface(value)
	if err != nil {
		return err
	}

	*o = ids
	return nil
}

// UnmarshalYAML will unmarshal a string or array of strings to OutputIDs.
func (o *OutputIDs) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value interface{}
	err := unmarshal(&value)
	if err != nil {
		return err
	}

	ids, err := o.fromInterface(value)
	if err != nil {
		return err
	}

	*o = ids
	return nil
}

// fromInterface will parse OutputIDs from a raw interface.
func (o *OutputIDs) fromInterface(value interface{}) (OutputIDs, error) {
	if str, ok := value.(string); ok {
		return OutputIDs{str}, nil
	}

	if array, ok := value.([]interface{}); ok {
		return o.fromArray(array)
	}

	return nil, fmt.Errorf("value is not of type string or string array")
}

// fromArray will parse OutputIDs from a raw array.
func (o *OutputIDs) fromArray(array []interface{}) (OutputIDs, error) {
	ids := OutputIDs{}
	for _, rawValue := range array {
		strValue, ok := rawValue.(string)
		if !ok {
			return nil, fmt.Errorf("value in array is not of type string")
		}
		ids = append(ids, strValue)
	}
	return ids, nil
}
