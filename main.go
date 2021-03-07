package main

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/alecthomas/kingpin"

	log "github.com/sirupsen/logrus"
)

var (
	app = kingpin.New("supporter", "Contest support tool")
)

func init() {
	// set log level
	if _, ok := os.LookupEnv("DEBUG"); ok {
		log.SetLevel(log.DebugLevel)
		log.Info("DEBUG MODE")
	}

	// check oj
	go func() {
		if err := exec.Command("oj", "--version").Run(); err != nil {
			log.Warn("oj is not installed, some functions cannot be used")
		}
	}()
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
	case downloadCmd.FullCommand():
		execDownloadCmd()
	case runCmd.FullCommand():
		execRunCmd()
	case testCmd.FullCommand():
		execTestCmd()
	case submitCmd.FullCommand():
		execSubmitCmd()
	}
}
