## `add` operator

The `add` operator adds a value to an `entry`'s `record`, `labels`, or `resource`.

### Configuration Fields

| Field      | Default          | Description                                                                                                                                                                                                                              |
| ---        | ---              | ---                                                                                                                                                                                                                                      |
| `id`       | `add`    | A unique identifier for the operator                                                                                                                                                                                                     |
| `output`   | Next in pipeline | The connected operator(s) that will receive all outbound entries                                                                                                                                                                         |
| `field`      | required       | The [field](/docs/types/field.md) to be added.    
| `value`      | required       | `value` is either a static value or an [expression](https://github.com/open-telemetry/opentelemetry-log-collection/blob/main/docs/types/expression.md). If a value is specified, it will be added to each entry at the field defined by `field`. If an expression is specified, it will be evaluated for each entry and added at the field defined by `field`
| `on_error` | `send`           | The behavior of the operator if it encounters an error. See [on_error](/docs/types/on_error.md)                                                                                                                                          |
| `if`       |                  | An [expression](/docs/types/expression.md) that, when set, will be evaluated to determine whether this operator should be used for the given entry. This allows you to do easy conditional parsing without branching logic with routers. |


Example usage:
 
<hr>
Add a string to the record

```yaml
- type: add
  field: key2
  value: val2
```

<table>
<tr><td> Input entry </td> <td> Output entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "labels": { },  
  "record": {
    "key1": "val1",
  }
}
```

</td>
<td>

```json
{
  "resource": { },
  "labels": { },  
  "record": {
    "key1": "val1",
    "key2": "val2"
  }
}
```

</td>
</tr>
</table>

<hr>
Add a value to the record using an expression

```yaml
- type: add
  field: key2
  value: EXPR($.key1 + "_suffix")
```

<table>
<tr><td> Input entry </td> <td> Output entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "labels": { },  
  "record": {
    "key1": "val1",
  }
}
```

</td>
<td>

```json
{
  "resource": { },
  "labels": { },  
  "record": {
    "key1": "val1",
    "key2": "val1_suffix"
  }
}
```

</td>
</tr>
</table>

<hr>
Add an object to the record

```yaml
- type: add
  field: key2
  value:
    nestedkey: nestedvalue
```

<table>
<tr><td> Input entry </td> <td> Output entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "labels": { },  
  "record": {
    "key1": "val1",
  }
}
```

</td>
<td>

```json
{
  "resource": { },
  "labels": { },  
  "record": {
    "key1": "val1",
    "key2": {
      "nestedkey":"nested value"
    }
  }
}
```

</td>
</tr>
</table>

<hr>
Add a value to labels

```yaml
- type: add
  field: $labels.key2
  value: val2
```

<table>
<tr><td> Input entry </td> <td> Output entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "labels": { },  
  "record": {
    "key1": "val1",
  }
}
```

</td>
<td>

```json
{
  "resource": { },
  "labels": {
     "key2": "val2"
  },  
  "record": {
    "key1": "val1"
  }
}
```

</td>
</tr>
</table>

<hr>
Add a value to resource

```yaml
- type: add
  field: $resource.key2
  value: val2
```

<table>
<tr><td> Input entry </td> <td> Output entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "labels": { },  
  "record": {
    "key1": "val1",
  }
}
```

</td>
<td>

```json
{
  "resource": { 
    "key2": "val2"
  },
  "labels": { },  
  "record": {
    "key1": "val1"
  }
}
```

</td>
</tr>
</table>

Add a value to resource using an expression

```yaml
- type: add
  field: $resource.key2
  value: EXPR($.key1 + "_suffix")
```

<table>
<tr><td> Input entry </td> <td> Output entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "labels": { },  
  "record": {
    "key1": "val1",
  }
}
```

</td>
<td>

```json
{
  "resource": { 
    "key2": "val_suffix"
  },
  "labels": { },  
  "record": {
    "key1": "val1",
  }
}
```

</td>
</tr>
</table>