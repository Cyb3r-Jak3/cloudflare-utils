package main

import "testing"

func Test_FileExists(t *testing.T) {
	if FileExists("utils.go") == false {
		t.Errorf("FileExists returned false for utils.go")
	}
	if FileExists("missing") == true {
		t.Errorf("FileExists returned true for missing")
	}
}
