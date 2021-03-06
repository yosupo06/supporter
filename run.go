package main

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	runCmd     = app.Command("r", "Run command")
	runProblem = runCmd.Arg("problem", "Problem Directory").Required().String()
	runOpt     = runCmd.Flag("opt", "Opt").Short('O').Bool()
)

func execRunCmd() {
	config, err := readConfig(*runProblem)
	if err != nil {
		log.Fatal(err)
	}

	compile(config, *runProblem, *runOpt)

	source, err := toSource(*runProblem)
	if err != nil {
		log.Fatal(err)
	}
	bin := strings.TrimSuffix(source, path.Ext(source))
	absBin, err := filepath.Abs(bin)
	if err != nil {
		log.Fatal("Failed to convert to absolute path: ", bin)
	}
	bin = absBin
	cmd := exec.Command(bin)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
