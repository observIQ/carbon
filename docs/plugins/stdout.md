## `stdout` plugin

The `stdout` plugin will write entries to stdout in JSON format. This is particularly useful for debugging a config file
or running one-time batch processing jobs.

### Configuration Fields

| Field | Default  | Description                        |
| ---   | ---      | ---                                |
| `id`  | `stdout` | A unique identifier for the plugin |


### Example Configurations

#### Simple configuration

Configuration:
```yaml
- type: stdout
```
