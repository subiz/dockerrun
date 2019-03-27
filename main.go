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
	Volumes []string
	Env     []string
}

func main() {
	app := cli.NewApp()
	app.Name = "dockerun"
	app.Usage = "dockerun"
	app.Version = "1.1.4"
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
	return parseConfigMap(obj)
}

func getEnv(envis []interface{}) ([]string, error) {
	env := make([]string, 0)
	for _, gei := range envis {
		e, ok := gei.(string)
		if !ok {
			return nil, fmt.Errorf("invalid env")
		}
		env = append(env, strings.TrimSpace(e))
	}
	return env, nil
}

func checkVersion(obj map[interface{}]interface{}) error {
	ver, _ := obj["version"].(interface{})
	version := toString(ver)

	if strings.TrimSpace(version) != "1" {
		return fmt.Errorf("should be version 1, got version: '%s'", version)
	}
	return nil
}

func parseConfigMap(obj map[interface{}]interface{}) ([]Step, error) {
	if err := checkVersion(obj); err != nil {
		return nil, err
	}

	genvis, _ := obj["env"].([]interface{})
	genv, err := getEnv(genvis)
	if err != nil {
		return nil, err
	}
	stepsints, _ := obj["steps"].([]interface{})
	steps := make([]Step, 0)
	for i, stepint := range stepsints {
		stepi, ok := stepint.(map[interface{}]interface{})
		if !ok {
			continue
		}

		trim := strings.TrimSpace
		img, _ := stepi["image"].(string)
		cmd, _ := stepi["command"].(string)
		dir, _ := stepi["dir"].(string)
		shell, _ := stepi["shell"].(string)

		envis, _ := stepi["env"].([]interface{})
		env, err := getEnv(envis)
		if err != nil {
			return nil, fmt.Errorf("%v at step %d", err, i+1)
		}

		volis, _ := stepi["volumes"].([]interface{})
		vols := make([]string, 0)
		for _, vi := range volis {
			v, ok := vi.(string)
			if !ok {
				return nil, fmt.Errorf("invalid volume at step %d", i+1)
			}
			if len(strings.Split(v, ":")) != 2 {
				return nil, fmt.Errorf("invalid volume at step %d", i+1)
			}
			vols = append(vols, strings.TrimSpace(v))
		}
		img, cmd, dir, shell = trim(img), trim(cmd), trim(dir), trim(shell)
		if img == "" {
			continue
		}
		steps = append(steps, Step{
			Image:   img,
			Command: cmd,
			Dir:     dir,
			Shell:   shell,
			Volumes: vols,
			Env:     append(genv, env...),
		})
	}
	return steps, nil
}

func stepToCommand(step Step) string {
	dir := strings.TrimSpace(step.Dir)
	if len(dir) != 0 {
		dir = " -w " + dir
	}

	env := strings.Join(step.Env, " -e ")
	if len(env) > 0 {
		env = " -e " + env
	}

	if step.Shell == "" {
		step.Shell = "/bin/sh"
	}
	cmd := strings.Replace(step.Command, `\`, `\\`, -1)
	cmd = strings.Replace(cmd, `"`, `\"`, -1)
	cmd = strings.Replace(cmd, `\\$`, `\$`, -1)
	vol := strings.Join(step.Volumes, " -v ")
	if len(vol) > 0 {
		vol = " -v " + vol
	}
	return fmt.Sprintf(`docker run --privileged --entrypoint %s%s%s%s --rm %s -c "%s"`, step.Shell, env, dir, vol, step.Image, cmd)
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
