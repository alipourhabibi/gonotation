# GoNotation
this repo is the implementation of the [notation](https://github.com/onury/notation) package in golang.

## Example:
```go
array := []glob.Glob{
		glob.NewGlob("dog"),
		glob.NewGlob("car"),
		glob.NewGlob("!car.brand"),
}
notations := notation.New(`{ "car": { "brand": "Dodge", "model": "Charger" }, "dog": { "breed": "Akita" } }`)
filtered, _ := notations.Filter(array, false)
fmt.Println(filtered)
```
in the example above our globs are ["dog", "car", "!car.brand"] \
they will be normalize with the glob package and then filtered with the notation package. \
this repo uses [sjson](https://github.com/tidwall/sjson) and [gjson](https://github.com/tidwall/gjson) under the hood for json manipulation.
