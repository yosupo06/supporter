package main

import "testing"

func TestPredictOrder(t *testing.T) {
	if ord, err := predictOrder("https://atcoder.jp/contests/typical90/tasks/typical90_a"); err != nil || ord != 0 {
		t.Fatalf("order of typical90_a is invalid: %v %v", ord, err)
	}
	if ord, err := predictOrder("https://atcoder.jp/contests/typical90/tasks/typical90_aa"); err != nil || ord != 26 {
		t.Fatalf("order of typical90_aa is invalid: %v %v", ord, err)
	}
	if ord, err := predictOrder("https://atcoder.jp/contests/typical90/tasks/typical90_ba"); err != nil || ord != 52 {
		t.Fatalf("order of typical90_ba is invalid: %v %v", ord, err)
	}

	if ord, err := predictOrder("1"); err != nil || ord != 0 {
		t.Fatalf("order of 1 is invalid: %v %v", ord, err)
	}
	if ord, err := predictOrder("10"); err != nil || ord != 9 {
		t.Fatalf("order of 10 is invalid: %v %v", ord, err)
	}

	if ord, err := predictOrder("a"); err != nil || ord != 0 {
		t.Fatalf("order of a is invalid: %v %v", ord, err)
	}
	if ord, err := predictOrder("B"); err != nil || ord != 1 {
		t.Fatalf("order of B is invalid: %v %v", ord, err)
	}
}
