# Style Guide

This documents describe a set of conventions to be used during development of this project. These rules are not set in stone and if you feel like they should be changed please open a new PR so we can have a documented discussion.

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
