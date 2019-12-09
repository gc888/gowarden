package main

import "flag"

import "fmt"

func main() {
	var ip = flag.Int("flagname", 1234, "help message for flagname")
	flag.Parse()
	fmt.Println("ip has value ", *ip)
}
