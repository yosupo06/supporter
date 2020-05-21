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

func execSubmitCmd() {
	src, err := toSource(*submitProblem)
	if err != nil {
		log.Fatal(err)
	}
	if config.BundleSource != nil {
		ext := path.Ext(src)
		newSrc := strings.TrimSuffix(src, ext) + ".out" + ext
		commandStr := new(bytes.Buffer)
		if err := config.BundleSource.Execute(commandStr, map[string]string{
			"Source": src,
			"Output": newSrc,
		}); err != nil {
			log.Fatal(err)
		}
		command := strings.Fields(commandStr.String())

		log.Infof("Bundle: %v -> %v", src, newSrc)
		log.WithField("Command", command).Debug("Compile Command")

		cmd := exec.Command(command[0], command[1:]...)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		src = newSrc
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

	config, err := getProblemConfig(path.Join(*submitProblem, "info.toml"))
	if err != nil {
		log.Fatal(err)
	}
	url, err := getProblemURL(config)
	if err != nil {
		log.Fatal(err)
	}
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
