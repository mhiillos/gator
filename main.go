package main

import (
	"fmt"
	"os"

	"github.com/mhiillos/go-blog-aggregator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", cfg)
	cfg.SetUser("mhiillos")
	cfg2, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", cfg2)
}
