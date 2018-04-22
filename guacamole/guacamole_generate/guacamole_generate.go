package main

import (
	"os"

	"hack.systems/random/guacamole"
)

func main() {
	g := guacamole.New()
	buf := make([]byte, 1024*1024)
	for {
		g.Fill(buf)
		os.Stdout.Write(buf)
	}
}
