/*
Copyright The Helm Authors.

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

package barrelman

import (
	"os"
	"text/template"
	"time"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/timeconv"

	"github.com/cirrocloud/structured/log"
)

var printReleaseTemplate = `REVISION: {{.Release.Version}}
RELEASED: {{.ReleaseDate}}
CHART: {{.Release.Chart.Metadata.Name}}-{{.Release.Chart.Metadata.Version}}
USER-SUPPLIED VALUES:
{{.Release.Config.Raw}}
COMPUTED VALUES:
{{.ComputedValues}}
HOOKS:
{{- range .Release.Hooks }}
---
# {{.Name}}
{{.Manifest}}
{{- end }}
MANIFEST:
{{.Release.Manifest}}
`

func printRelease(rel *release.Release) error {
	if rel == nil {
		return nil
	}

	cfg, err := chartutil.CoalesceValues(rel.Chart, rel.Config)
	if err != nil {
		return err
	}
	cfgStr, err := cfg.YAML()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"Release":        rel,
		"ComputedValues": cfgStr,
		"ReleaseDate":    timeconv.Format(rel.Info.LastDeployed, time.ANSIC),
	}
	return tpl(printReleaseTemplate, data)
}

func tpl(t string, vals map[string]interface{}) error {
	tt, err := template.New("_").Parse(t)
	if err != nil {
		return err
	}
	return tt.Execute(os.Stdout, vals)
}

func debug(format string, args ...interface{}) {
	log.Debugf(format, args...)
}
