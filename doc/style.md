# Style Guide

This documents describe a set of conventions to be used during development of this project. These rules are not set in stone and if you feel like they should be changed please open a new PR so we can have a documented discussion.

## Javascript

### Recommended JavaScript Snippets

To create a new styled React component (with typescript):

```json
{
  "Export Default React Component": {
    "prefix": "tsx",
    "body": [
      "import * as React from 'react';",
      "import styled from 'styled-components'",
      "",
      "type Props = {",
      "  className?: string",
      "}",
      "",
      "function ${1:} ({ className }: Props) {",
      "  return (",
      "    <div className={className}>",
      "      ${0}",
      "    </div>",
      "  );",
      "}",
      "",
      "export default styled(${1:}).attrs({ className: ${1:}.name })``"
    ],
    "description": "Create a default-exported, styled React Component."
  }
}
```

## Golang

### Error names

Error variables should be called just `err` unless you need to return multiple errors.

Good:

```go
err := client.Update()
if err := nil {
    return err
}
```

Bad:

```go
clientErr := client.Update()
if clientErr := nil {
    return err
}
```

### Logging

Weave Gitops uses structured logging. The Kubernetes community has a [great guide](https://github.com/kubernetes/community/blob/d7ab088079c96891cee9bbcc88a78649acdf49f1/contributors/devel/sig-instrumentation/migration-to-structured-logging.md) on doing this well.

Here are some terse highlights but the whole thing is worth reading.

#### [On logging style](https://github.com/kubernetes/community/blob/d7ab088079c96891cee9bbcc88a78649acdf49f1/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#remove-string-formatting-from-log-message)

> - Start from a capital letter.
> - Do not end the message with a period.
> - Use active voice. Use complete sentences when there is an acting subject ("A could not do B") or omit the subject if the subject would be the program itself ("Could not do B").
> - Use past tense ("Could not delete B" instead of "Cannot delete B")
> - When referring to an object, state what type of object it is. ("Deleted pod" instead of "Deleted")
>
> For example
>
> ```go
> klog.Infof("delete pod %s with propagation policy %s", ...)
> ```
>
> should be changed to
>
> ```go
> klog.InfoS("Deleted pod", ...)
> ```

#### [On naming arguments in structured logging](https://github.com/kubernetes/community/blob/d7ab088079c96891cee9bbcc88a78649acdf49f1/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments)

> When deciding on names of arguments you should:
>
> - Always use [lowerCamelCase], for example use `containerName` and not `container name` or `container_name`.
> - Use [alphanumeric] characters: no special characters like `%$*`, non-latin, or unicode characters.
> - Use object kind when referencing Kubernetes objects, for example `deployment`, `pod` and `node`.
> - Describe the type of value stored under the key and use normalized labels:
>   - Don't include implementation-specific details in the labels. Don't use `directory`, do use `path`.
>   - Do not provide additional context for how value is used. Don't use `podIP`, do use `IP`.
>   - With the exception of acronyms like "IP" and the standard "err", don't shorten names. Don't use `addr`, do use `address`.
>   - When names are very ambiguous, try to include context in the label. For example, instead of
>     `key` use `cacheKey` or instead of `version` use `dockerVersion`.
> - Be consistent, for example when logging file path we should always use `path` and not switch between
>   `hostPath`, `path`, `file`.

#### [On Good practice for passing values in structured logging](https://github.com/kubernetes/community/blob/d7ab088079c96891cee9bbcc88a78649acdf49f1/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#good-practice-for-passing-values-in-structured-logging)

> When passing a value for a key-value pair, please use following rules:
>
> - Prefer using Kubernetes objects and log them using `klog.KObj` or `klog.KObjSlice`
>   - When the original object is not available, use `klog.KRef` instead
>   - when only one object (for example `*v1.Pod`), we use`klog.KObj`
>   - When type is object slice (for example `[]*v1.Pod`), we use `klog.KObjSlice`
> - Pass structured values directly (avoid calling `.String()` on them first)
> - When the goal is to log a `[]byte` array as string, explicitly convert with `string(<byte array>)`.
