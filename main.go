package main

import (
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/kingpin"

	log "github.com/sirupsen/logrus"
)

var (
	app          = kingpin.New("supporter", "Contest support tool")
	initCmd      = app.Command("i", "Init contest")
	initURL      = initCmd.Arg("url", "Contest URL").Required().String()
	initProblems = initCmd.Arg("problems", "Contest Problems").Required().Strings()
)

var config struct {
	TemplateSrc     string `toml:"template_src"`
	CompileDebugStr string `toml:"compile_debug"`
	CompileDebug    *template.Template
	CompileOPTStr   string `toml:"compile_opt"`
	CompileOPT      *template.Template
	BundleSourceStr string `toml:"bundle_source"`
	BundleSource    *template.Template
}

func init() {
	if _, ok := os.LookupEnv("DEBUG"); ok {
		log.SetLevel(log.DebugLevel)
		log.Info("DEBUG MODE")
	}

	// check oj
	if err := exec.Command("oj", "--version").Run(); err != nil {
		log.Warn("oj is not installed, some functions cannot be used")
	}

	// read config.toml
	tomlPath := os.Getenv("SUPPORTER_CONFIG")
	if tomlPath == "" {
		log.Fatal("Please set $SUPPORTER_CONFIG")
	}
	if _, err := toml.DecodeFile(tomlPath, &config); err != nil {
		log.Fatalf("Failed to load config from %v: err", tomlPath)
	}
	if _, err := os.Stat(config.TemplateSrc); err != nil {
		log.WithField("templateSrc", config.TemplateSrc).Fatalf("config.templateSrc is not exist")
	}

	var err error
	config.CompileDebug, err = template.New("CompileDebug").Parse(config.CompileDebugStr)
	if err != nil {
		log.Fatal("Template parse failed")
	}
	config.CompileOPT, err = template.New("CompileOPT").Parse(config.CompileOPTStr)
	if err != nil {
		log.Fatal("Template parse failed")
	}

	if config.BundleSourceStr != "" {
		config.BundleSource, err = template.New("BundleSource").Parse(config.BundleSourceStr)
		if err != nil {
			log.Fatal("Template parse failed")
		}
	}
}

type ContestInfo struct {
	ID      string
	Offline bool
}

type ProblemConfig struct {
	ProblemID  string `toml:"problem_id"`
	ContestURL string `toml:"contest_url"`
}

func getContestInfo(contestURL string) (ContestInfo, error) {
	u, err := url.Parse(contestURL)
	if err != nil {
		return ContestInfo{}, err
	}
	if u.Scheme == "" {
		return ContestInfo{
			ID:      contestURL,
			Offline: true,
		}, nil
	}
	path := strings.Split(u.Path, "/")
	if u.Host == "atcoder.jp" {
		if path[1] != "contests" {
			return ContestInfo{}, errors.New("Invalid URL of AtCoder")
		}
		return ContestInfo{
			ID:      path[2],
			Offline: false,
		}, nil
	}
	if u.Host == "codeforces.com" {
		if path[1] != "contest" {
			return ContestInfo{}, errors.New("Invalid URL of Codeforces")
		}
		return ContestInfo{
			ID:      "codeforces-" + path[2],
			Offline: false,
		}, nil
	}
	return ContestInfo{}, errors.New("Unknown URL")
}

func getProblemConfig(tomlPath string) (ProblemConfig, error) {
	tomlFile, err := os.Open(tomlPath)
	if err != nil {
		return ProblemConfig{}, errors.New("Nothing info.toml")
	}
	config := ProblemConfig{}
	if _, err := toml.DecodeReader(tomlFile, &config); err != nil {
		return ProblemConfig{}, err
	}
	return config, nil
}

func copyFile(src, trg string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(trg, input, 0644); err != nil {
		return err
	}
	return nil
}

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case initCmd.FullCommand():
		execInitCmd()
	case buildCmd.FullCommand():
		execBuildCmd()
	case runCmd.FullCommand():
		execRunCmd()
	case testCmd.FullCommand():
		execTestCmd()
	case submitCmd.FullCommand():
		execSubmitCmd()
	}
}
