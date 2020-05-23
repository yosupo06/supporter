package main

import (
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	runCmd     = app.Command("r", "Run command")
	runProblem = runCmd.Arg("problem", "Problem Directory").Required().String()
	runOpt     = runCmd.Flag("opt", "Opt").Short('O').Bool()
)

func execRunCmd() {
	compile(*runProblem, *runOpt)
	source, err := toSource(*runProblem)
	if err != nil {
		log.Fatal(err)
	}
	bin := strings.TrimSuffix(source, path.Ext(source))
	cmd := exec.Command(bin)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
