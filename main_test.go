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
		`docker run --entrypoint /bin/sh -w /home/van --rm nginx:12 -c "ls"
docker run --entrypoint /bin/sh --rm alpine -c "ping google.com"`,
	}, {
		"escape quote",
		[]Step{{
			Image:   "nginx:12",
			Command: `ls "me"`,
		}},
		`docker run --entrypoint /bin/sh --rm nginx:12 -c "ls \"me\""`,
	}, {
		"escape newline",
		[]Step{{
			Image: "nginx:12",
			Command: `ping google.com
echo "\n1\\n2"`,
		}},
		`docker run --entrypoint /bin/sh --rm nginx:12 -c "ping google.com
echo \"\\n1\\\\n2\""`,
	}, {
		"volumes",
		[]Step{{
			Image:   "alpine",
			Volumes: []string{"/a:/a"},
		}},
		`docker run --entrypoint /bin/sh -v /a:/a --rm alpine -c ""`,
	}, {
		"env",
		[]Step{{
			Image: "alpine",
			Env:   []string{"a=a", "b=$b"},
		}},
		`docker run --entrypoint /bin/sh -e a=a -e b=$b --rm alpine -c ""`,
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
			Shell:   "/bin/bash",
		}},
		nil,
	}, {
		"./test/run4.yaml",
		[]Step{{
			Image:   "alpine",
			Volumes: []string{"$(pwd):/workspace"},
		}},
		nil,
	}, {
		"./test/run5.yaml",
		nil,
		fmt.Errorf("invalid volume at step 1"),
	}, {
		"./test/run6.yaml",
		[]Step{{
			Image: "alpine",
			Env:   []string{"A=4", "B=$B"},
		}},
		nil,
	}, {
		"./test/run7.yaml",
		[]Step{{
			Image: "alpine",
			Env:   []string{"C=5", "A=4", "C=6"},
		}},
		nil,
	}, {
		"./test/run8.yaml", // invalid env
		nil,
		fmt.Errorf("invalid env at step 1"),
	}, {
		"./test/run9.yaml", //missing version
		nil,
		fmt.Errorf("should be version 1, got version: '2'"),
	}}

	for _, tc := range tcs {
		steps, err := loadConfig(tc.filename)
		if !compareError(err, tc.err) {
			t.Errorf("[%v] expect %v, got %v.", tc.filename, tc.err, err)
		}

		if err != nil {
			continue
		}

		if !compareSteps(steps, tc.expect) {
			t.Errorf("[%v] expect %s, got %s", tc.filename, jsonify(tc.expect), jsonify(steps))
		}
	}
}
