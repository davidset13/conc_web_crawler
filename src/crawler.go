package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println(seedUrls)
	fmt.Println("Number of CPUs:", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())
}
