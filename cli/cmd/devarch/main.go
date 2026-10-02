// Command devarch runs the DevArch service library by name.
package main

import (
	"os"

	"github.com/prospect-ogujiuba/devarch/cli/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
