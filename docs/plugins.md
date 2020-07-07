# Plugins

Plugins can be defined by using a file that contains a templated set of builtin plugins.

For example, a very simple plugin for monitoring Apache Tomcat access logs could look like this:
`tomcat.yaml`:
```yaml
---
pipeline:
  - id: tomcat_access_reader
    type: file_input
    include:
      - {{ .path }}
    output: tomcat_regex_parser

  - id: tomcat_regex_parser
    type: regex_parser
    output: {{ .output }}
    regex: '(?P<remote_host>[^\s]+) - (?P<remote_user>[^\s]+) \[(?P<timestamp>[^\]]+)\] "(?P<http_method>[A-Z]+) (?P<path>[^\s]+)[^"]+" (?P<http_status>\d+) (?P<bytes_sent>[^\s]+)'
```

Once a plugin config has been defined, it can be used in the log agent's config file with a `type` matching the filename of the plugin.

`config.yaml`:
```yaml
---
pipeline:
  - id: tomcat_access
    type: tomcat
    output: stdout
    path: /var/log/tomcat/access.log

  - id: stdout
    type: stdout
```

The `tomcat_access` plugin is replaced with the builtin plugins from the rendered config in `tomcat.yaml`.

## Building a plugin

Building a plugin is as easy as pulling out a set of builtin plugins in a working configuration file, then templatizing it with
any parts of the config that need to be treated as variable. In the example of the Tomcat access log plugin above, that just means
adding variables for `path` and `output`.

Plugins use Go's [`text/template`](https://golang.org/pkg/text/template/) package for template rendering. All fields from
the plugin configuration are available as variables in the templates except the `type` field.

For the log agent to discover a plugin, it needs to be in the log agent's `plugin` directory. This can be set with the
`--plugin_dir` argument. For a default installation, the plugin directory is located at `$BPLOGAGENT_HOME/plugins`.