package main

import (
	"os"
	"runtime"
	"strconv"

	"go.uber.org/automaxprocs/maxprocs"
)

func main() {
	maxprocs.Set(maxprocs.Logger(func(string, ...any) {}))
	os.Stdout.Write([]byte(strconv.FormatInt(int64(runtime.GOMAXPROCS(0)), 10)))
}
