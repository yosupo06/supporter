package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	buildCmd     = app.Command("b", "Build contest")
	buildProblem = buildCmd.Arg("problem", "Problem Directory").Required().String()
	buildOpt     = buildCmd.Flag("opt", "Opt").Short('O').Bool()
)

func toSource(problem string) (string, error) {
	stat, err := os.Stat(problem)
	if err != nil {
		return "", errors.New("File is not exists")
	}
	if stat.IsDir() {
		problem = path.Join(problem, "main.cpp")
	}
	return problem, nil
}

func compile(problem string, opt bool) {
	src, err := toSource(problem)
	if err != nil {
		log.Fatal(err)
	}
	output := strings.TrimSuffix(src, path.Ext(src))

	commandStr := new(bytes.Buffer)
	if !opt {
		config.CompileDebug.Execute(commandStr, map[string]string{
			"Source": src,
			"Output": output,
		})
	} else {
		config.CompileOPT.Execute(commandStr, map[string]string{
			"Source": src,
			"Output": output,
		})
	}
	command := strings.Fields(commandStr.String())

	log.Infof("Compile: %v", src)
	log.WithField("Command", command).Debug("Compile Command")

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal("Failed to build: ", err)
	}
}

func execBuildCmd() {
	compile(*buildProblem, *buildOpt)
}
