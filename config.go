package main

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

// Config contains contents of supporter_config.toml
type Config struct {
	// Path of template code
	TemplateSrc string `toml:"template_src,omitempty"`

	// Compile command(debug mode)
	CompileDebugStr string `toml:"compile_debug,omitempty"`
	CompileDebug    *template.Template
	// Compile command(release mode)
	CompileOPTStr string `toml:"compile_opt,omitempty"`
	CompileOPT    *template.Template

	// Bundle command
	BundleSourceStr string `toml:"bundle_source,omitempty"`
	BundleSource    *template.Template

	// Contest URL
	ContestURL string `toml:"contest_url,omitempty"`
	// Problem ID
	ProblemID string `toml:"problem_id,omitempty"`
}

func parentDirs(path string) (dirs []string, err error) {
	dirs = make([]string, 0)
	err = nil

	// to absolute path
	if path, err = filepath.Abs(path); err != nil {
		return
	}

	// to parent dir if path points file
	fInfo, err := os.Stat(path)
	if err != nil {
		return
	}
	if !fInfo.IsDir() {
		path = filepath.Dir(path)
	}

	for {
		dirs = append(dirs, path)
		if nextPath := filepath.Dir(path); path == nextPath {
			break
		} else {
			path = nextPath
		}
	}
	return
}

var memoizedConfig = make(map[string]*Config)

func readConfig(path string) (*Config, error) {
	if val, ok := memoizedConfig[path]; ok {
		log.WithField("path", path).Debug("Already read, skip")
		return val, nil
	}

	config := &Config{}
	dirs, err := parentDirs(path)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(dirs); i++ {
		// read config.toml
		tomlPath := filepath.Join(dirs[len(dirs)-1-i], "config_supporter.toml")
		if _, err := os.Stat(tomlPath); err != nil {
			continue
		}
		if _, err := toml.DecodeFile(tomlPath, config); err != nil {
			return nil, err
		}
	}

	config.CompileDebug, err = template.New("CompileDebug").Parse(config.CompileDebugStr)
	if err != nil {
		return nil, err
	}
	config.CompileOPT, err = template.New("CompileOPT").Parse(config.CompileOPTStr)
	if err != nil {
		return nil, err
	}

	config.BundleSource, err = template.New("BundleSource").Parse(config.BundleSourceStr)
	if err != nil {
		return nil, err
	}

	log.WithField("config", config).WithField("path", path).Debug("Read Config")

	memoizedConfig[path] = config

	return config, nil
}
