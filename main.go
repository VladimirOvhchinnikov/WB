package main

import (
	"fmt"
)

func main() {

	var cache Cache = *NewCache(2)

	fmt.Println(cache.Len(), cache.Cap())

	cache.Add(1, "a")
	cache.Add(2, "b")

	fmt.Println(cache.Get(1))
	fmt.Println(cache.Get(2))

	cache.Add(3, "c")
	fmt.Println("------------------")

	fmt.Println(cache.Get(1))
	fmt.Println(cache.Get(2))
	fmt.Println(cache.Get(3))
	fmt.Println(cache.Get(4))

}
