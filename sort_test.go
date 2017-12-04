package main

import "testing"

func Test_Sort(t *testing.T) {
	dp1 := []dep{
		{name: "a", deps: []dep{ndep("b"), ndep("c")}},
		{name: "b", deps: []dep{}},
		{name: "c", deps: []dep{ndep("b")}},
	}

	if inorder(dp1) {
		t.Errorf("dp1 should not be in order")
	}

	srt1, err := sortDeps(dp1)
	if err != nil {
		t.Errorf("error on dp1")
	}

	if !inorder(srt1) {
		t.Errorf("sort didn't put in order")
	}
}
