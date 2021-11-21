package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ContestSite int

const (
	AtCoder ContestSite = iota
	Codeforces
	Other
)

type ContestInfo struct {
	Site    ContestSite
	ID      string
	Offline bool
}

func getContestInfo(contestURL string) (ContestInfo, error) {
	if contestURL == "" {
		return ContestInfo{}, errors.New("contest URL is empty")
	}
	u, err := url.Parse(contestURL)
	if err != nil {
		return ContestInfo{}, err
	}
	if u.Scheme == "" {
		return ContestInfo{
			Site:    Other,
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
			Site:    AtCoder,
			ID:      path[2],
			Offline: false,
		}, nil
	}
	if u.Host == "codeforces.com" {
		if path[1] != "contest" {
			return ContestInfo{}, errors.New("Invalid URL of Codeforces")
		}
		return ContestInfo{
			Site:    Codeforces,
			ID:      "codeforces-" + path[2],
			Offline: false,
		}, nil
	}
	return ContestInfo{}, errors.New("Unknown URL")
}

func predictOrder(problemID string) (int, error) {
	if val, err := strconv.Atoi(problemID); err == nil {
		return val - 1, nil
	}

	// normalize problemId
	arr := strings.Split(problemID, "_")
	problemID = strings.ToLower(arr[len(arr)-1])

	r := regexp.MustCompile("^[a-z]+$")

	if r.MatchString(problemID) {
		ord := 0
		for _, c := range problemID {
			ord *= 26
			ord += int(c) - 'a'
		}
		base := 1
		for i := 0; i < len(problemID); i++ {
			ord += base
			base *= 26
		}
		return ord - 1, nil
	}

	return -1, fmt.Errorf("failed to predict order: %v", problemID)
}

func predictProblemURL(contestURL string, problemID string) (string, error) {
	contest, err := getContestInfo(contestURL)
	if err != nil {
		return "", err
	}

	if contest.Site == Other {
		// unknown site
		return "", nil
	}
	cmd := exec.Command("oj-api", "get-contest", contestURL)
	cmd.Stderr = os.Stderr
	list, err := cmd.Output()
	if err != nil {
		return "", err
	}
	response := OjAPIGetContest{}
	if err := json.Unmarshal(list, &response); err != nil {
		return "", err
	}
	log.WithFields(log.Fields{
		"rawResponse": string(list),
		"response":    response,
	}).Debug("OJ Response")

	// complete same
	for _, problem := range response.Result.Problems {
		if strings.EqualFold(problemID, problem.URL) {
			return problem.URL, nil
		}
	}

	// order
	expectOrd, err := predictOrder(problemID)
	if err == nil {
		for _, problem := range response.Result.Problems {
			if problemOrd, err := predictOrder(problem.URL); err == nil && expectOrd == problemOrd {
				return problem.URL, nil
			}
		}
	}

	return "", fmt.Errorf("cannot find URL contest(%s), problem(%s)", contestURL, problemID)
}
