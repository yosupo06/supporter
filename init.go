package main

import (
	"log"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

func execInitCmd() {
	info, err := getContestInfo(*initURL)
	if err != nil {
		log.Fatalf("Failed to fetch info %v: %v", *initURL, err)
	}
	log.Print("Contest info:", info)
	dir := info.ID
	if err := os.Mkdir(dir, 0755); err != nil {
		log.Fatal("error: Failed to create dir")
	}
	for _, problem := range *initProblems {
		pdir := path.Join(dir, problem)
		if err := os.Mkdir(pdir, 0755); err != nil {
			log.Fatal("Failed to create dir")
		}
		log.Print("Init problem: ", problem)
		if err := copyFile(config.TemplateSrc, path.Join(pdir, "main.cpp")); err != nil {
			log.Fatalf("Failed to copy from %v to %v : %v", config.TemplateSrc, path.Join(pdir, "main.cpp"), err)
		}
		if err := os.Mkdir(path.Join(pdir, "ourtest"), 0755); err != nil {
			log.Fatal("Failed to create dir")
		}
		file, err := os.Create(path.Join(dir, problem, "info.toml"))
		if err != nil {
			log.Fatal(err)
		}

		if err := toml.NewEncoder(file).Encode(ProblemConfig{
			ProblemID:  problem,
			ContestURL: *initURL,
		}); err != nil {
			log.Fatal(err)
		}
	}
}
