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

func TestStepsToCommand(t *testing.T) {
	tcs := []struct {
		desc   string
		steps  []Step
		expect string
	}{{
		"nil",
		nil,
		"",
	}, {
		"normal",
		[]Step{{
			Image:   "nginx:12",
			Command: "ls",
			Dir:     "/home/van",
		}, {
			Image:   "alpine",
			Command: "ping google.com",
		}},
		`docker run -v $(pwd):/home/van --entrypoint /bin/sh nginx:12 -c "ls"
docker run -v $(pwd):/workspace --entrypoint /bin/sh alpine -c "ping google.com"`,
	}, {
		"escape quote",
		[]Step{{
			Image:   "nginx:12",
			Command: `ls "me"`,
		}},
		`docker run -v $(pwd):/workspace --entrypoint /bin/sh nginx:12 -c "ls \"me\""`,
	}, {
		"escape newline",
		[]Step{{
			Image:   "nginx:12",
			Command: `ping google.com
echo "\n1\\n2"`,
		}},
		`docker run -v $(pwd):/workspace --entrypoint /bin/sh nginx:12 -c "ping google.com
echo \"\\n1\\\\n2\""`,
	}}

	for _, tc := range tcs {
		out := stepsToCommand(tc.steps)
		if out != tc.expect {
			t.Errorf("[%s] expect %s, got %s", tc.desc, tc.expect, out)
		}
	}
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
			Command: "ping google.com\nping subiz.com",
		}},
		nil,
	}, {
		"./test/run3.yaml",
		[]Step{{
			Image:   "alpine",
			Command: "ping -c 4 google.com",
			Shell: "/bin/bash",
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
