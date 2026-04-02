package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/polunzh/mailbox-cli/internal/release"
)

func main() {
	var data release.HomebrewFormulaData

	flag.StringVar(&data.Repo, "repo", "", "GitHub repo in owner/name form")
	flag.StringVar(&data.Version, "version", "", "release version tag")
	flag.StringVar(&data.DarwinAMD64SHA, "darwin-amd64-sha", "", "sha256 for darwin amd64 archive")
	flag.StringVar(&data.DarwinARM64SHA, "darwin-arm64-sha", "", "sha256 for darwin arm64 archive")
	flag.Parse()

	formula, err := release.RenderHomebrewFormula(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render homebrew formula: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(formula)
}
