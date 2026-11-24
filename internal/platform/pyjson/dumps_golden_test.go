package pyjson

import (
	"encoding/json"
	"io/fs"
	"os"
	"testing"
)

func TestDumpsMatchesPythonJSON(t *testing.T) {
	raw, err := fs.ReadFile(os.DirFS("../../../testdata/pyjson"), "dumps_golden.json")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var cases []struct {
		Input  string `json:"input"`
		Output string `json:"output"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("decode golden: %v", err)
	}
	for _, testCase := range cases {
		value, err := Decode([]byte(testCase.Input))
		if err != nil {
			t.Errorf("Decode(%s): %v", testCase.Input, err)
			continue
		}
		if got, err := Dumps(value); err != nil || got != testCase.Output {
			t.Errorf("Dumps(%s)\n got %s, %v\nwant %s", testCase.Input, got, err, testCase.Output)
		}
	}
}
