package main

import (
	"fmt"
	"jumpstartcli/internal/examples"
)

func main() {
	result := examples.GetExamples("js.arcbox.list").FormatExamples()
	fmt.Printf("Examples output:\n%s\n", result)
	fmt.Printf("Length: %d\n", len(result))
}
