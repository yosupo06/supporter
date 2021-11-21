package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
)

var (
	testCmd     = app.Command("t", "Test source")
	testProblem = testCmd.Arg("problem", "Problem").Required().String()
	testOpt     = testCmd.Flag("opt", "Opt").Short('O').Bool()
	testLocal   = testCmd.Flag("local", "Lpt").Short('l').Bool()
)

func checkDiff(actual, expect string) bool {
	return cmp.Equal(strings.Fields(actual), strings.Fields(expect))
}

type OjAPIGetContest struct {
	Result struct {
		Problems []struct {
			URL string `json:"url"`
		} `json:"problems"`
	} `json:result`
}

func singleTest(binary, input, expectFile string) {
	log.Infof("Test: %v", input)
	cmd := exec.Command(binary)
	inFile, err := os.Open(input)
	if err != nil {
		log.Fatal("Failed to read input file: ", err)
	}
	var expect *string
	if expectFile != "" {
		expectBin, err := ioutil.ReadFile(expectFile)
		if err != nil {
			log.Fatal("Failed to read output file: ", err)
		}
		expectStr := string(expectBin)
		expect = &expectStr
	}

	cmd.Stdin = inFile
	cmd.Stderr = os.Stderr
	now := time.Now()
	actualBin, err := cmd.Output()
	elapsed := time.Now().Sub(now)
	log.Infof("Time: %v ms", elapsed.Milliseconds())
	actual := string(actualBin)
	if err != nil {
		log.Warn("RE: ", err)
	}

	if expect != nil {
		if !checkDiff(actual, *expect) {
			log.Warn("WA")
			fmt.Fprintln(os.Stderr, "=== output ===")
			fmt.Fprint(os.Stderr, actual)
			fmt.Fprintln(os.Stderr, "=== expect ===")
			fmt.Fprint(os.Stderr, *expect)
		} else {
			log.Info("AC")
		}
	} else {
		fmt.Fprintln(os.Stderr, "=== output ===")
		fmt.Fprint(os.Stderr, actual)
		log.Info("No answer file")
	}
}

func execTestCmd() {
	problem := *testProblem
	config, err := readConfig(problem)
	if err != nil {
		log.Fatal(err)
	}

	// fetch sample
	if !*testLocal {
		if err := fetchSample(problem); err != nil {
			log.Fatal("failed to fetch sample: ", err)
		}
	}

	// compile
	if err := compile(config, problem, *testOpt); err != nil {
		log.Fatal(err)
	}

	dir, err := toSourceDir(problem)
	if err != nil {
		log.Fatal(err)
	}
	inFiles, err := filepath.Glob(dir + "/**/*.in")
	if err != nil {
		log.Fatal(err)
	}

	source, err := toSource(problem)
	if err != nil {
		log.Fatal(err)
	}
	bin := strings.TrimSuffix(source, path.Ext(source))
	for _, inFile := range inFiles {
		outFile := strings.TrimSuffix(inFile, path.Ext(inFile)) + ".out"
		if _, err := os.Stat(outFile); err != nil {
			outFile = ""
		}
		singleTest(bin, inFile, outFile)
	}
}
