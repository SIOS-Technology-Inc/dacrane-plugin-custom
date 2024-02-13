package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	dacranepdk "github.com/SIOS-Technology-Inc/dacrane-pdk"
)

func main() {
	config := dacranepdk.NewDefaultPluginConfig()
	dockerHost := "/var/run/docker.sock"
	config.DockerHost = &dockerHost
	dacranepdk.ExecPluginJob(dacranepdk.Plugin{
		Config: config,
		Resources: dacranepdk.MapToFunc(map[string]dacranepdk.Resource{
			"shell": ShellResource,
		}),
	})
}

var ShellResource = dacranepdk.Resource{
	Create: func(parameter any, meta dacranepdk.PluginMeta) (any, error) {
		params := parameter.(map[string]any)
		image := params["image"].(string)
		env := params["env"].([]any)
		tag := params["tag"].(string)
		shell := params["shell"].(string)
		script, ok := params["create"].(string)

		if !ok {
			return parameter, nil
		}

		envOpts := []string{}
		for _, e := range env {
			name := e.(map[string]any)["name"].(string)
			value := e.(map[string]any)["value"].(string)
			opt := fmt.Sprintf(`-e "%s=%s"`, name, value)
			envOpts = append(envOpts, opt)
		}

		netOpt := ""
		if network, ok := params["network"].(string); ok {
			netOpt = fmt.Sprintf("--net %s", network)
		}

		cmd := fmt.Sprintf(
			`docker run --rm -v $HOST_WORKING_DIR:/work %s %s %s:%s %s -c "%s"`,
			strings.Join(envOpts, " "), netOpt, image, tag, shell, script)

		_, err := RunOnSh(cmd, meta)
		if err != nil {
			panic(err)
		}

		return parameter, nil
	},
	Delete: func(parameter any, meta dacranepdk.PluginMeta) error {
		params := parameter.(map[string]any)
		image := params["image"].(string)
		env := params["env"].([]any)
		tag := params["tag"].(string)
		shell := params["shell"].(string)
		script, ok := params["delete"].(string)

		if !ok {
			return nil
		}

		envOpts := []string{}
		for _, e := range env {
			name := e.(map[string]any)["name"].(string)
			value := e.(map[string]any)["value"].(string)
			opt := fmt.Sprintf(`-e "%s=%s"`, name, value)
			envOpts = append(envOpts, opt)
		}

		netOpt := ""
		if network, ok := params["network"].(string); ok {
			netOpt = fmt.Sprintf("--net %s", network)
		}

		cmd := fmt.Sprintf(
			`docker run --rm -v $HOST_WORKING_DIR:/work %s %s %s:%s %s -c "%s"`,
			strings.Join(envOpts, " "), netOpt, image, tag, shell, script)

		_, err := RunOnSh(cmd, meta)
		if err != nil {
			panic(err)
		}

		return nil
	},
}

func RunOnSh(script string, m dacranepdk.PluginMeta) ([]byte, error) {
	m.Log(fmt.Sprintf("> %s\n", script))
	cmd := exec.Command("sh", "-c", script)
	writer := new(bytes.Buffer)
	cmd.Stdout = io.MultiWriter(os.Stderr, writer)
	cmd.Stderr = io.MultiWriter(os.Stderr, writer)
	err := cmd.Run()
	return writer.Bytes(), err
}
