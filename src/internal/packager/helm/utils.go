// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package helm contins operations for working with helm charts
package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/defenseunicorns/zarf/src/pkg/message"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"

	"helm.sh/helm/v3/pkg/chart/loader"
)

// loadChartFromTarball returns a helm chart from a tarball
func (h *Helm) loadChartFromTarball() (*chart.Chart, error) {
	// Get the path the temporary helm chart tarball
	sourceFile := StandardName(filepath.Join(h.BasePath, "charts"), h.Chart) + ".tgz"
	if h.ChartLoadOverride != "" {
		sourceFile = h.ChartLoadOverride
	}

	// Load the loadedChart tarball
	loadedChart, err := loader.Load(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("unable to load helm chart archive: %w", err)
	}

	if err = loadedChart.Validate(); err != nil {
		return nil, fmt.Errorf("unable to validate loaded helm chart: %w", err)
	}

	return loadedChart, nil
}

// parseChartValues reads the context of the chart values into an interface if it exists
func (h *Helm) parseChartValues() (map[string]any, error) {
	valueOpts := &values.Options{}

	for idx, file := range h.Chart.ValuesFiles {
		path := StandardName(filepath.Join(h.BasePath, "values"), h.Chart) + "-" + strconv.Itoa(idx)
		// If we are overriding the chart path, assuming this is for zarf prepare
		if h.ChartLoadOverride != "" {
			path = file
		}
		valueOpts.ValueFiles = append(valueOpts.ValueFiles, path)
	}

	httpProvider := getter.Provider{
		Schemes: []string{"http", "https"},
		New:     getter.NewHTTPGetter,
	}

	providers := getter.Providers{httpProvider}
	return valueOpts.MergeValues(providers)
}

func (h *Helm) createActionConfig(namespace string, spinner *message.Spinner) error {
	// OMG THIS IS SOOOO GROSS PPL... https://github.com/helm/helm/issues/8780
	_ = os.Setenv("HELM_NAMESPACE", namespace)

	// Initialize helm SDK
	actionConfig := new(action.Configuration)
	settings := cli.New()

	// Setup K8s connection
	err := actionConfig.Init(settings.RESTClientGetter(), namespace, "", spinner.Updatef)

	// Set the actionConfig is the received Helm pointer
	h.actionConfig = actionConfig

	return err
}
