package main

import (
	"automarket/internal/router"
)

func main() {
	r := router.SetupRouter()

	if err := r.Run(":8081"); err != nil {
		panic(err)
	}
}
