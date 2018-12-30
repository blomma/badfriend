package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(time.Now().Local().Format(time.RFC3339))
}
