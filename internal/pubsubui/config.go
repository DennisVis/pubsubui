// Copyright 2022 Dennis Vis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pubsubui

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	envKeyHost     = "PUBSUBUI_HOST"
	envKeyPort     = "PUBSUBUI_PORT"
	envKeyConfig   = "PUBSUBUI_CONFIG"
	envKeyProjects = "GOOGLE_CLOUD_PROJECTS"
)

const (
	flagNameHost     = "host"
	flagNamePort     = "port"
	flagNameConfig   = "config"
	flagNameProjects = "projects"
)

var (
	defaultValueHost     = "0.0.0.0"
	defaultValuePort     = uint(8080)
	defaultValueConfig   = ""
	defaultValueProjects = ""
)

var (
	flagHost     = flag.String(flagNameHost, defaultValueHost, "The host to which to bind the service")
	flagPort     = flag.Uint(flagNamePort, defaultValuePort, "The port to which to bind the service")
	flagConfig   = flag.String(flagNameConfig, defaultValueConfig, "The path to the topics config file")
	flagProjects = flag.String(
		flagNameProjects,
		defaultValueProjects,
		"The Google Cloud Platform projects to target (if not set in the config file)",
	)
)

type config struct {
	host           string
	port           uint
	configFilePath string
	projectIDs     []string
}

func parseString(v string) (string, error) {
	return v, nil
}

func parseUint(v string) (uint, error) {
	pv, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "invalid uint: %s", v)
	}

	return uint(pv), nil
}

func foo[T any](envKey, flagName string, flagVal *T, defVal *T, parseFn func(string) (T, error)) (T, error) {
	// Try to get the value from the environment first.
	envVal, ok := os.LookupEnv(envKey)
	if ok {
		parsedEnvVal, err := parseFn(envVal)
		if err != nil {
			return *defVal, errors.Wrapf(err, "invalid value for env var %q: %q", envKey, envVal)
		}

		return parsedEnvVal, nil
	}

	// If not set in the environment we check if a flag was provided.
	switch {
	case flagVal != nil:
		return *flagVal, nil
	case defVal != nil:
		return *defVal, nil
	default:
		return *defVal, errors.Errorf("missing value for env var %q or flag \"-%s\"", envKey, flagName)
	}
}

func newConfig() (*config, error) {
	logWithPrefix("application: config: creating")

	flag.Parse()

	host, err := foo(envKeyHost, flagNameHost, flagHost, &defaultValueHost, parseString)
	if err != nil {
		return nil, errors.Wrap(err, "config: could not configure host")
	}

	port, err := foo(envKeyPort, flagNameHost, flagPort, &defaultValuePort, parseUint)
	if err != nil {
		return nil, errors.Wrap(err, "config: could not configure port")
	}

	configFilePath, err := foo(envKeyConfig, flagNameConfig, flagConfig, &defaultValueConfig, parseString)
	if err != nil {
		return nil, errors.Wrap(err, "config: could not configure config file path")
	}

	projectIDsStr, err := foo(envKeyProjects, flagNameProjects, flagProjects, nil, parseString)
	if err != nil {
		return nil, errors.Wrap(err, "config: could not configure GCP projects")
	}
	projectIDs := filterEmptyStrings(strings.Split(projectIDsStr, ","))

	cfg := config{
		host:           host,
		port:           uint(port),
		configFilePath: configFilePath,
		projectIDs:     projectIDs,
	}

	logWithPrefix("application: config: created")

	return &cfg, nil
}
