# Style Guide

This documents describe a set o conventions to be used during development of this project. These rules are not set in stone and should be ignored if you find you need.

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
