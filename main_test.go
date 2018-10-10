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
		`docker run --entrypoint /bin/sh -w /home/van nginx:12 -c "ls"
docker run --entrypoint /bin/sh alpine -c "ping google.com"`,
	}, {
		"escape quote",
		[]Step{{
			Image:   "nginx:12",
			Command: `ls "me"`,
		}},
		`docker run --entrypoint /bin/sh nginx:12 -c "ls \"me\""`,
	}, {
		"escape newline",
		[]Step{{
			Image:   "nginx:12",
			Command: `ping google.com
echo "\n1\\n2"`,
		}},
		`docker run --entrypoint /bin/sh nginx:12 -c "ping google.com
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
	}, {
		"./test/run4.yaml",
		[]Step{{
			Image: "alpine",
			Volumes: []string{"$(pwd):/workspace"},
		}},
		nil,
	},{
		"./test/run5.yaml",
		nil,
		fmt.Errorf("invalid volume at step 1"),
	}}

	for _, tc := range tcs {
		steps, err := loadConfig(tc.filename)
		if !compareError(err, tc.err) {
			t.Errorf("[%v] expect %v, got %v", tc.filename, tc.err, err)
		}

		if err != nil {
			continue
		}

		if !compareSteps(steps, tc.expect) {
			t.Errorf("[%v] expect %s, got %s", tc.filename, jsonify(tc.expect), jsonify(steps))
		}
	}
}
