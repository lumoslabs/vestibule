package main

import (
	"fmt"
	"runtime"
	"strings"
)

const author = "Lumos Labs"

var (
	version      = "dev"
	commit, date string
)

func appVersion() string {
	return fmt.Sprintf(
		`%s (%s on %s/%s; %s)`,
		strings.Join([]string{version, commit, date}, "/"),
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Compiler,
	)
}
