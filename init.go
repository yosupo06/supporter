package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/BurntSushi/toml"

	"github.com/otiai10/copy"
)

var (
	initCmd      = app.Command("i", "Init contest")
	initURL      = initCmd.Arg("url", "Contest URL").Required().String()
	initProblems = initCmd.Arg("problems", "Contest Problems").Required().Strings()
)

func initProblem(dir string, contestURL string, id string, config *Config) error {
	pdir := path.Join(dir, id)
	if err := os.Mkdir(pdir, 0755); err != nil {
		return err
	}
	log.Printf("Init problem: %v", id)

	mainPath := path.Join(pdir, "main.cpp")
	if err := copyFile(config.TemplateSrc, mainPath); err != nil {
		return fmt.Errorf("Failed to copy from %v to %v : %v", config.TemplateSrc, mainPath, err)
	}

	if err := os.Mkdir(path.Join(pdir, "ourtest"), 0755); err != nil {
		return err
	}

	file, err := os.Create(path.Join(dir, id, "config_supporter.toml"))
	if err != nil {
		return err
	}

	if err := toml.NewEncoder(file).Encode(Config{
		ProblemID:  id,
		ContestURL: contestURL,
	}); err != nil {
		return err
	}

	return nil
}

func execInitCmd() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	config, err := readConfig(cwd)
	if err != nil {
		log.Fatal(err)
	}

	info, err := getContestInfo(*initURL)
	if err != nil {
		log.Fatalf("Failed to fetch info %v: %v", *initURL, err)
	}
	log.Printf("Contest info: ID(%v)", info.ID)
	dir := info.ID
	if err := os.Mkdir(dir, 0755); err != nil {
		log.Fatalf("Failed to create dir: %v", err)
	}
	log.WithField("files", config.ContestTemplate).Debug("Copy contest template")
	for _, path := range config.ContestTemplate {
		name := filepath.Base(path)
		if err := copy.Copy(path, filepath.Join(dir, name)); err != nil {
			log.Fatal(err)
		}
	}
	for _, problem := range *initProblems {
		if err := initProblem(dir, *initURL, problem, config); err != nil {
			log.Fatal(err)
		}
	}
}
