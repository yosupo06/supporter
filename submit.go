package main

import (
	"bytes"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	submitCmd     = app.Command("s", "Submit source")
	submitProblem = submitCmd.Arg("problem", "Problem Dir").Required().String()
	submitClip    = submitCmd.Flag("clip", "Clip to clipboard(OS X only?)").Short('c').Bool()
)

func bundle(config *Config, problem string) (string, error) {
	src, err := toSource(problem)
	if err != nil {
		log.Fatal(err)
	}

	if config.BundleSourceStr == "" {
		return src, nil
	}

	ext := path.Ext(src)
	newSrc := strings.TrimSuffix(src, ext) + ".out" + ext
	log.Infof("Bundle: %v -> %v", src, newSrc)

	commandBuff := new(bytes.Buffer)
	if err := config.BundleSource.Execute(commandBuff, map[string]string{
		"Source": src,
		"Output": newSrc,
	}); err != nil {
		return "", err
	}
	command := strings.Fields(commandBuff.String())
	log.WithField("Command", command).Debug("Bundle Command")

	cmd := exec.Command("bash", "-c", strings.Join(command, " "))
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	src = newSrc

	return newSrc, nil
}

func execSubmitCmd() {
	problem := *submitProblem
	config, err := readConfig(problem)
	if err != nil {
		log.Fatal(err)
	}

	src, err := bundle(config, problem)
	if err != nil {
		log.Fatal(err)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	if *submitClip {
		log.Info("Copy to clipboard")
		cmd := exec.Command("pbcopy")
		cmd.Stdin = srcFile
		if err := cmd.Run(); err != nil {
			log.Error("Failed to copy: ", err)
		}
	}

	url, err := getProblemURL(config)
	if err != nil {
		log.Fatal(err)
	}
	if url != "" {
		absDir, err := filepath.Abs(*submitProblem)
		if err != nil {
			log.Fatal(err)
		}
		cmd := exec.Command("oj", "submit", url, path.Base(src), "--no-open", "-w", "0")
		cmd.Dir = absDir
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}
}
