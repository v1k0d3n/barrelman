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

package cmd

import (
	"fmt"
	"github.com/charter-oss/barrelman/cmd/util"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/chartutil"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/cirrocloud/structured/log"
)

var (
	// defaultKubeVersion is the default value of --kube-version flag
	defaultKubeVersion = fmt.Sprintf("%s.%s", chartutil.DefaultKubeVersion.Major, chartutil.DefaultKubeVersion.Minor)
)

func newTemplateCmd(cmd *barrelman.TemplateCmd) *cobra.Command {

	longDesc := strings.TrimSpace(dedent.Dedent(`
	Render chart templates locally and display the output.

	This does not require Tiller. However, any values that would normally be
	looked up or retrieved in-cluster will be faked locally. Additionally, none
	of the server-side testing of chart validity (e.g. whether an API is supported)
	is done.`))

	shortDesc := `Locally render templates.`

	examples := strings.TrimSpace(dedent.Dedent(`
	To render just one template in a chart, use '-x':

		barrelman template mychart -x templates/deployment.yaml
		
	To render all charts to individual files in a directory, use '--output-dir':

		barrelman template templates/deployment.yaml --output-dir output/
	`))

	cobraCmd := &cobra.Command{
		Use:     "template [flags] CHART",
		Short:   shortDesc,
		Long:    longDesc,
		Example: examples,
		Args:    cobra.ExactArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {

			cmd.Options.ManifestFile = args[0]

			cobraCmd.SilenceUsage = true
			cobraCmd.SilenceErrors = true
			log.Configure(util.LogSettings(cmd.LogOptions)...)
			if err := cmd.Run(); err != nil {
				return err
			}
			return nil
		},
	}

	f := cobraCmd.Flags()
	cmd.LogOptions = f.StringSliceP("log", "l", nil, "log options (e.g. --log=debug,JSON")
	f.BoolVar(&cmd.ShowNotes, "notes", false, "show the computed NOTES.txt file as well")
	f.StringVarP(&cmd.ReleaseName, "name", "n", "release-name", "release name")
	f.BoolVar(&cmd.ReleaseIsUpgrade, "is-upgrade", false, "set .Release.IsUpgrade instead of .Release.IsInstall")
	f.StringArrayVarP(&cmd.RenderFiles, "execute", "x", []string{}, "only execute the given templates")
	f.VarP(&cmd.ValueFiles, "values", "f", "specify values in a YAML file (can specify multiple)")
	f.StringVar(&cmd.Namespace, "namespace", "", "namespace to install the release into")
	f.StringArrayVar(&cmd.Values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&cmd.StringValues, "set-string", []string{}, "set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&cmd.FileValues, "set-file", []string{}, "set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)")
	f.StringVar(&cmd.NameTemplate, "name-template", "", "specify template used to name the release")
	f.StringVar(&cmd.KubeVersion, "kube-version", defaultKubeVersion, "kubernetes version used as Capabilities.KubeVersion.Major/Minor")
	f.StringVar(&cmd.OutputDir, "output-dir", "", "writes the executed templates to files in output-dir instead of stdout")

	return cobraCmd
}
