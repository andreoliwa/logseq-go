// A native curl fixture keeps upstream-check tests independent of shell syntax.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) == 3 && os.Args[1] == "sleep" {
		seconds, err := strconv.Atoi(os.Args[2])
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Duration(seconds) * time.Second)
		return
	}
	file, err := os.OpenFile(os.Getenv("COUNTER"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	_, err = file.WriteString("lookup\n")
	if err != nil {
		panic(err)
	}
	if err = file.Close(); err != nil {
		panic(err)
	}
	if delay := os.Getenv("CHECK_SLEEP"); delay != "" {
		// Keep a descendant holding the output pipes so timeout tests detect
		// killing only the parent rather than the complete lookup tree.
		child := exec.Command(os.Args[0], "sleep", delay)
		child.Stdout, child.Stderr = os.Stdout, os.Stderr
		if err := child.Run(); err != nil {
			panic(err)
		}
	}
	if os.Getenv("CHECK_FAIL") != "" {
		os.Exit(1)
	}
	fmt.Print(os.Getenv("RESPONSE"))
}
