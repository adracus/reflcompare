# reflcompare

[![Go Report Card](https://goreportcard.com/badge/github.com/adracus/reflcompare)](https://goreportcard.com/report/github.com/adracus/reflcompare)
[![Go Reference](https://pkg.go.dev/badge/github.com/adracus/reflcompare.svg)](https://pkg.go.dev/github.com/adracus/reflcompare)

reflcompare is what `reflect.DeepEqual` is for equality:
It recursively traverses two given values and compares their fields against each other.
In contrast to `reflect.DeepEqual`, the values *have* to be of identical / comparable
types, otherwise they are not comparable and a panic is thrown.

## Installation

To add this to your module, just run

```shell
go get github.com/adracus/reflcompare
```

