package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"

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
		return "", err
	}
	if stat.IsDir() {
		problem = path.Join(problem, "main.cpp")
	}
	return problem, nil
}

func toSourceDir(problem string) (string, error) {
	source, err := toSource(problem)
	if err != nil {
		return "", err
	}
	return filepath.Dir(source), nil
}

func compile(config *Config, problem string, opt bool) error {
	src, err := toSource(problem)
	if err != nil {
		return err
	}
	log.Infof("Compile: %v", src)

	var compileComm *template.Template
	if opt {
		compileComm = config.CompileOPT
	} else {
		compileComm = config.CompileDebug
	}

	cxx := os.Getenv("CXX")
	if cxx == "" {
		cxx = "g++"
	}
	output := strings.TrimSuffix(src, path.Ext(src))

	commandBuff := new(bytes.Buffer)
	compileComm.Execute(commandBuff, map[string]string{
		"CXX":    cxx,
		"Source": src,
		"Output": output,
	})
	command := strings.Fields(commandBuff.String())

	log.WithField("Command", command).Debug("Compile Command")

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to build: %v", err)
	}

	return nil
}

func execBuildCmd() {
	config, err := readConfig(*buildProblem)
	if err != nil {
		log.Fatal(err)
	}
	if err := compile(config, *buildProblem, *buildOpt); err != nil {
		log.Fatal(err)
	}
}
