package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func jsonify(a interface{}) string {
	b, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func compareStep(a, b []Step) bool {
	return jsonify(a) == jsonify(b)
}

func compareError(a, b error) bool {
	if a == nil {
		return b == nil
	}

	if b == nil {
		return false
	}

	return a.Error() == b.Error()
}

func TestLoadConfig(t *testing.T) {
	tcs := []struct {
		filename string
		expect   []Step
		err      error
	}{{
		"./test/run0.yaml",
		[]Step{},
		fmt.Errorf("open ./test/run0.yaml: no such file or directory"),
	}, {
		"./test/run1.yaml",
		[]Step{{
			Image:   "nginx:1.3",
			Command: "echo 123",
			Dir:     "/workspace",
		}, {
			Image:   "alpine:3.8",
			Command: "ping google.com",
		}},
		nil,
	}}

	for _, tc := range tcs {
		steps, err := loadConfig(tc.filename)
		if !compareError(err, tc.err) {
			t.Errorf("[%v] expect %v, got %v", tc.filename, tc.err, err)
		}

		if err != nil {
			continue
		}

		if !compareStep(steps, tc.expect) {
			t.Errorf("[%v] expect %s, got %s", tc.filename, jsonify(tc.expect), jsonify(steps))
		}
	}
}
