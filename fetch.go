package main

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

var (
	downloadCmd      = app.Command("d", "Download testcases")
	downloadProblems = downloadCmd.Arg("problems", "Contest Problems").Required().Strings()
)

func fetchSample(problem string) error {
	config, err := readConfig(problem)
	if err != nil {
		return err
	}

	dir, err := toSourceDir(problem)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path.Join(dir, "test")); err == nil {
		log.Info("Testcases are already fetched")
		return nil
	}

	url, err := getProblemURL(config)
	if err != nil {
		return err
	}
	if url == "" {
		log.Info("Unknown site, skip")
		return nil
	}

	log.Infof("Fetch testcase from %v", url)

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	cmd := exec.Command("oj", "d", url)
	cmd.Dir = absDir
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func execDownloadCmd() {
	for _, problem := range *downloadProblems {
		if err := fetchSample(problem); err != nil {
			log.Fatal(err)
		}
	}
}
