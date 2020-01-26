package main

import (
	"flag"
	"fmt"

	. "github.com/TRUMANCFY/DSEProject/voter"
)

var port = flag.String("port", "8080", "please provide UI Port")

func main() {
	flag.Parse()
	fmt.Println(*port)
	v := &Voter{Port: *port}
	v.ListenToGui()
}
