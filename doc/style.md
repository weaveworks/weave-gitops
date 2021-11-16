# Style Guide

This documents describe a set of conventions to be used during development of this project. These rules are not set in stone and if you feel like they should be changed please open a new PR so we can have a documented discussion.

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
