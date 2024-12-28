---
title: Supported Templating Languages

---


# Supported Templating Languages ~ENTERPRISE~

The following templating languages are supported:
- envsubst (default)
- templating

Declare the templating language to be used to render the template by setting `spec.renderType`.

## Envsubst

`envsubst`, which is short for 'environment substitution', uses [envsubst](https://github.com/a8m/envsubst)
for rendering.
This templating format is used by [clusterctl](https://cluster-api.sigs.k8s.io/clusterctl/overview.html).

Variables can be set for rendering into the template in the `${VAR_NAME}`
syntax.

### Supported Functions

| __Expression__                | __Meaning__                                                     |
| -----------------             | --------------                                                  |
| `${var}`                      | Value of `$var`
| `${#var}`                     | String length of `$var`
| `${var^}`                     | Uppercase first character of `$var`
| `${var^^}`                    | Uppercase all characters in `$var`
| `${var,}`                     | Lowercase first character of `$var`
| `${var,,}`                    | Lowercase all characters in `$var`
| `${var:n}`                    | Offset `$var` `n` characters from start
| `${var:n:len}`                | Offset `$var` `n` characters with max length of `len`
| `${var#pattern}`              | Strip shortest `pattern` match from start
| `${var##pattern}`             | Strip longest `pattern` match from start
| `${var%pattern}`              | Strip shortest `pattern` match from end
| `${var%%pattern}`             | Strip longest `pattern` match from end
| `${var-default}`               | If `$var` is not set, evaluate expression as `$default`
| `${var:-default}`              | If `$var` is not set or is empty, evaluate expression as `$default`
| `${var=default}`               | If `$var` is not set, evaluate expression as `$default`
| `${var:=default}`              | If `$var` is not set or is empty, evaluate expression as `$default`
| `${var/pattern/replacement}`  | Replace as few `pattern` matches as possible with `replacement`
| `${var//pattern/replacement}` | Replace as many `pattern` matches as possible with `replacement`
| `${var/#pattern/replacement}` | Replace `pattern` match with `replacement` from `$var` start
| `${var/%pattern/replacement}` | Replace `pattern` match with `replacement` from `$var` end

## Templating

Templating uses text/templating for rendering, using go-templating style syntax `{{ .params.CLUSTER_NAME }}`
where params are provided by the `.params` variable.
Template functions can also be used with the syntax `{{ .params.CLUSTER_NAME | FUNCTION }}`.

### Supported Functions

As taken (from the [Sprig library](http://masterminds.github.io/sprig/))

| __Function Type__                   | __Functions__                                                     |
| -----------------                   | --------------                                                  |
| String Functions                    | *trim*, *wrap*, *randAlpha*, *plural*
| String List Functions               | *splitList*, *sortAlpha*
| Integer Math Functions              | *add*, *max*, *mul*
| Integer Slice Functions             | *until*, untilStep
| Float Math Functions                | *addf*, *maxf*, *mulf*
| Date Functions                      | *now*, *date*
| Defaults Functions                  | *default*, *empty*, *coalesce*, *fromJson*, *toJson*, *toPrettyJson*, *toRawJson*, ternary
| Encoding Functions                  | *b64enc*, *b64dec*
| Lists and List Functions            | *list*, *first*, *uniq*
| Dictionaries and Dict Functions     | *get*, *set*, *dict*, *hasKey*, *pluck*, *dig*, *deepCopy*
| Type Conversion Functions           | *atoi*, *int64*, *toString*
| Flow Control Functions              | *fail*
| UUID Functions                      | *uuidv4*
| Version Comparison Functions        | *semver*, semverCompare
| Reflection                          | *typeOf*, *kindIs*, *typeIsLike*

### Custom Delimiters

The default delimiters for `renderType: templating` are `{{` and `}}`.
These can be changed by setting the `templates.weave.works/delimiters` annotation
on the template. For example:

- `templates.weave.works/delimiters: "{{,}}"` - default
- `templates.weave.works/delimiters: "${{,}}"`
  - Use `${{` and `}}`, for example `"${{ .params.CLUSTER_NAME }}"`
  - Useful as `{{` in yaml is invalid syntax and needs to be quoted. If you need to provide a un-quoted number value like `replicas: 3` you should use these delimiters.
		- :x: `replicas: {{ .params.REPLICAS }}` Invalid yaml
		- :x: `replicas: "{{ .params.REPLICAS }}"` Valid yaml, incorrect type. The type is a `string` not a `number` and will fail validation.
		- :white_check_mark: `replicas: ${{ .params.REPLICAS }}` Valid yaml and correct `number` type.
- `templates.weave.works/delimiters: "<<,>>" `
  - Use `<<` and `>>`, for example `<< .params.CLUSTER_NAME >>`
  - Useful if you are nesting templates and need to differentiate between the delimiters used in the inner and outer templates.

