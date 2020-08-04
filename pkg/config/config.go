/*
Copyright 2020 Smorgasbord Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	Path    string `json:"-"`
	BaseURL string `json:"baseURL"`
	Token   string `json:"token"`
}

func FromRaw(data []byte) (*Config, error) {
	c := &Config{}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return c, nil
}

func FromFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c, err := FromRaw(data)
	if err != nil {
		return nil, err
	}
	c.Path = path
	return c, nil
}

func (c *Config) SaveTo(path string) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

func (c *Config) Save() error {
	if c.Path == "" {
		return fmt.Errorf("Config was not loaded via FromFile, use SaveTo instead or set c.Path")
	}
	return c.SaveTo(c.Path)
}
