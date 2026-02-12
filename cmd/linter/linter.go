package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	panicsAnalyzer "github.com/hydra13/shortify/pkg/panics_analyzer"
)

func main() {
	singlechecker.Main(panicsAnalyzer.PanicsAnalyzer)
}
