## `tcp_input` operator

The `tcp_input` operator listens for logs on one or more TCP connections. The operator assumes that logs are newline separated.

### Configuration Fields

| Field            | Default  | Description                                                           |
| ---              | ---      | ---                                                                   |
| `id`             | required | A unique identifier for the operator                                  |
| `output`         | required | The connected operator(s) that will receive all outbound entries      |
| `listen_address` | required | A listen address of the form `<ip>:<port>`                            |
| `write_to`       | $        | A [field](/docs/types/field.md) that will be set to the log message   |

### Example Configurations

#### Simple

Configuration:
```yaml
- id: my_tcp_input
  type: tcp_input
  listen_adress: "0.0.0.0:54525"
  output: tcp_receiver
```

Send a log:
```bash
$ nc localhost 54525 <<EOF
heredoc> message1
heredoc> message2
heredoc> EOF
```

Generated entries:
```json
{
  "timestamp": "2020-04-30T12:10:17.656726-04:00",
  "record": "message1"
},
{
  "timestamp": "2020-04-30T12:10:17.657143-04:00",
  "record": "message2"
}
```
