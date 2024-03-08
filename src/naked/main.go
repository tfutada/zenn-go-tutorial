package main

import "log"

func main() {
	printInfo("foo", true, true)
}

// create printInfo function
func printInfo(name string, isFoo bool, isBar bool) {
	log.Println("name:", name)
	log.Println("isFoo:", isFoo)
	log.Println("isBar:", isBar)
}
