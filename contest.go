package main

import (
	"errors"
	"net/url"
	"strings"
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
