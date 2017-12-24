// Copyright 2017 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package integration

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/phayes/permbits"
)

// integrationSuffix is the suffix for all integrations.
const integrationSuffix = ".int"

func getIntegrationSubDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	return dir
}

// LoadIntegrations loads all the integration found for the current toolkit.
func LoadIntegrations() ([]IntegrationInformation, error) {
	return loadIntegrationsUnderPath(getIntegrationSubDirectory())
}

func loadIntegrationsUnderPath(dirPath string) ([]IntegrationInformation, error) {
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return []IntegrationInformation{}, nil
	}

	if err != nil {
		return []IntegrationInformation{}, err
	}

	// Iterate the directory, finding all binaries and trying to load the integrations found within.
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return []IntegrationInformation{}, err
	}

	if len(files) == 0 {
		return []IntegrationInformation{}, nil
	}

	integrations := make([]IntegrationInformation, 0, len(files))
	for _, f := range files {
		if strings.HasSuffix(f.Name(), integrationSuffix) {
			fullPath := path.Join(dirPath, f.Name())
			permissions, err := permbits.Stat(fullPath)
			if err != nil {
				return []IntegrationInformation{}, err
			}

			if permissions.UserExecute() || permissions.GroupExecute() || permissions.OtherExecute() {
				integrationInfo, err := loadIntegrationAtPath(fullPath)
				if err != nil {
					return []IntegrationInformation{}, err
				}

				integrations = append(integrations, integrationInfo)
			}
		}
	}

	return integrations, nil
}

func loadIntegrationAtPath(fullPath string) (IntegrationInformation, error) {
	p, err := plugin.Open(fullPath)
	if err != nil {
		return IntegrationInformation{}, err
	}

	integrationSymbol, err := p.Lookup(IntegrationConstName)
	if err != nil {
		return IntegrationInformation{}, err
	}

	integration, castOk := integrationSymbol.(Integration)
	if !castOk {
		return IntegrationInformation{}, fmt.Errorf("Could find integration in integration binary `%s`", fullPath)
	}

	return IntegrationInformation{fullPath, integration}, nil
}
