package main

import (
	"errors"
	"io"

	"gopkg.in/yaml.v2"
)

type targetMetric struct {
	Name                string    `yaml:"name"`
	Selector            *selector `yaml:"selector,omitempty"`
	MinimumHealthyCount *int      `yaml:"minimum-healthy-count,omitempty"`
}

type selector struct {
	Labels      map[string]string `yaml:"labels,omitempty"`
	LabelPrefix map[string]string `yaml:"label-prefix,omitempty"`
}

type machineType struct {
	Name             string         `yaml:"name"`
	MetricsCheckList []targetMetric `yaml:"metrics,omitempty"`
	GracePeriod      string         `yaml:"grace-period-of-setting-problematic-state,omitempty"`
}

type config struct {
	MachineTypes []machineType `yaml:"machine-types"`
}

func parseConfig(reader io.Reader) (*config, error) {
	cfg := new(config)
	err := yaml.NewDecoder(reader).Decode(cfg)
	if err != nil {
		return nil, err
	}

	if len(cfg.MachineTypes) == 0 {
		return nil, errors.New("machine-types are not defined")
	}
	for _, t := range cfg.MachineTypes {
		if len(t.GracePeriod) == 0 {
			t.GracePeriod = "1m"
		}
	}
	return cfg, nil
}
