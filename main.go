package main

import (
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
	"strings"
)

type Step struct {
	Image   string
	Command string
	Dir     string
	Shell   string
}

func main() {
	app := cli.NewApp()
	app.Name = "dockerun"
	app.Usage = "dockerun"
	app.Version = "1.0.1"
	app.Action = run
	l := log.New(os.Stderr, "", 0)
	if err := app.Run(os.Args); err != nil {
		l.Fatal(err)
	}
}

func loadConfig(name string) ([]Step, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	obj := make(map[interface{}]interface{})
	dec := yaml.NewDecoder(f)
	for {
		err := dec.Decode(&obj)
		if err == nil {
			continue
		}

		if err == io.EOF {
			break
		}
		return nil, err
	}

	return parseConfigMap(obj), nil
}

func parseConfigMap(obj map[interface{}]interface{}) []Step {
	stepsints, _ := obj["steps"].([]interface{})
	steps := make([]Step, 0)
	for _, stepint := range stepsints {
		stepi, ok := stepint.(map[interface{}]interface{})
		if !ok {
			continue
		}

		trim := strings.TrimSpace
		img, _ := stepi["image"].(string)
		cmd, _ := stepi["command"].(string)
		dir, _ := stepi["dir"].(string)
		shell, _ := stepi["shell"].(string)
		img, cmd, dir, shell = trim(img), trim(cmd), trim(dir), trim(shell)
		if img == "" {
			continue
		}
		steps = append(steps, Step{Image: img, Command: cmd, Dir: dir, Shell: shell})
	}
	return steps
}

func stepToCommand(step Step) string {
	if step.Dir == "" {
		step.Dir = "/workspace"
	}

	if step.Shell == "" {
		step.Shell = "/bin/sh"
	}
	cmd := strings.Replace(step.Command, `\`, `\\`, -1)
	cmd = strings.Replace(cmd, `"`, `\"`, -1)

	return fmt.Sprintf(`docker run -v $(pwd):%s --entrypoint %s %s -c "%s"`, step.Dir, step.Shell, step.Image, cmd)
}

func stepsToCommand(steps []Step) string {
	var cmds []string
	for _, step := range steps {
		cmds = append(cmds, stepToCommand(step))
	}
	return strings.Join(cmds, "\n")
}

func run(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.ShowAppHelp(c)
	}
	name := c.Args().Get(0)

	steps, err := loadConfig(name)
	if err != nil {
		return err
	}
	cmd := stepsToCommand(steps)
	fmt.Println(cmd)
	return nil
}
