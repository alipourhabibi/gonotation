package main

import (
	"fmt"

	"github.com/alipourhabibi/gonotation/glob"
	"github.com/alipourhabibi/gonotation/notation"
)

func main() {
	_ = []glob.Glob{
		glob.NewGlob("*"),
		glob.NewGlob("!id"),
		glob.NewGlob("name"),
		glob.NewGlob("car.model"),
		glob.NewGlob("!car.*"),
		glob.NewGlob("id"),
		glob.NewGlob("name"),
		glob.NewGlob("age"),
	}

	array := []glob.Glob{
		glob.NewGlob("dog"),
		glob.NewGlob("car"),
		glob.NewGlob("!car.brand"),
	}
	not := notation.New(`{ "car": { "brand": "Dodge", "model": "Charger" }, "dog": { "breed": "Akita" } }`)
	fil, err := not.Filter(array, false)
	fmt.Println(glob.Normalize(array, false))
	fmt.Println(fil, err)
}
