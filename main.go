package main

import (
	"fmt"
	"github.com/waisbrot/tp-link-api/lib"
)

func main() {
	conf := lib.Config{
		Host: "http://192.168.0.1",
	}
	var client, err = lib.Client(&conf)
	if err != nil {
		panic(err)
	}
	fmt.Println(client)
}
