# reflcompare

reflcompare is what `reflect.DeepEqual` is for equality:
It recursively traverses two given values and compares their fields against each other.
In contrast to `reflect.DeepEqual`, the values *have* to be of identical / comparable
types, otherwise they are not comparable and a panic is thrown.

## Installation

To add this to your module, just run

```shell
go get github.com/adracus/reflcompare
```

