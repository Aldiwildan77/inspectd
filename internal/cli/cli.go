package cli

import (
	"fmt"
	"os"

	"github.com/Aldiwildan77/inspectd/internal/goroutines"
	"github.com/Aldiwildan77/inspectd/internal/memory"
	"github.com/Aldiwildan77/inspectd/internal/runtimeinfo"
	"github.com/Aldiwildan77/inspectd/internal/snapshot"
)

func Run() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	command := os.Args[1]

	var output []byte
	var err error

	switch command {
	case "runtime":
		output, err = runtimeinfo.CollectJSON()
	case "memory":
		output, err = memory.CollectJSON()
	case "goroutines":
		output, err = goroutines.CollectJSON()
	case "snapshot":
		output, err = snapshot.CollectJSON()
	default:
		os.Exit(1)
	}

	if err != nil {
		os.Exit(1)
	}

	fmt.Println(string(output))
}
