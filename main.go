package main

import "github.com/radulucut/cleed/cmd/cleed"

var (
	version = "dev"
)

func main() {
	cleed.Version = version
	cleed.Execute()
}
