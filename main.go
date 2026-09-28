package main

import (
	"automarket/internal/router"
)

func main() {
	r := router.SetupRouter()

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
