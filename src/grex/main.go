package main

import (
	"ericaro.net/gogrex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)
// small util that turn a regexp into a graphiz .dot file and then into a png
func main() {
	exp := os.Args[1]
	fmt.Printf("Parsing %s\n", exp)

	var m gogrex.StringManager
	g, err := gogrex.ParseGrex(&m, exp)
	if err != nil {
		fmt.Printf("error %s", err)
	}

	fmt.Printf("\n%s\n", g.String())
	err = ioutil.WriteFile("graph.dot", []byte(g.String()), 0644)
	if err != nil {
		panic(err)
	}
	dot("graph.dot")
}

func dot(dot string) error{

	cmd := exec.Command("dot", "-Tpng", "-ograph.png", dot)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
