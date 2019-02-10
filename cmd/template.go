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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/strvals"
	"k8s.io/helm/pkg/tiller"
	"k8s.io/helm/pkg/timeconv"

	bfest "github.com/charter-se/barrelman/manifest"
	"github.com/charter-se/barrelman/version"
	"github.com/charter-se/structured"
	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
)

const defaultDirectoryPermission = 0755

var (
	whitespaceRegex = regexp.MustCompile(`^\s*$`)

	// defaultKubeVersion is the default value of --kube-version flag
	defaultKubeVersion = fmt.Sprintf("%s.%s", chartutil.DefaultKubeVersion.Major, chartutil.DefaultKubeVersion.Minor)
)

const templateDesc = `
Render chart templates locally and display the output.

This does not require Tiller. However, any values that would normally be
looked up or retrieved in-cluster will be faked locally. Additionally, none
of the server-side testing of chart validity (e.g. whether an API is supported)
is done.

To render just one template in a chart, use '-x':

	$ helm template mychart -x templates/deployment.yaml
`

type templateCmd struct {
	namespace        string
	valueFiles       valueFiles
	chartPath        string
	out              io.Writer
	values           []string
	stringValues     []string
	fileValues       []string
	nameTemplate     string
	showNotes        bool
	releaseName      string
	releaseIsUpgrade bool
	renderFiles      []string
	kubeVersion      string
	outputDir        string
	Options          *cmdOptions
	Config           *Config
	Log              structured.Logger
	LogOptions       *[]string
}

func newTemplateCmd(cmd *templateCmd) *cobra.Command {

	cobraCmd := &cobra.Command{
		Use:   "template [flags] CHART",
		Short: fmt.Sprintf("locally render templates"),
		Long:  templateDesc,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cmd.Options.ManifestFile = args[0]
			}
			cobraCmd.SilenceUsage = true
			cobraCmd.SilenceErrors = true
			log.Configure(logSettings(cmd.LogOptions)...)
			if err := cmd.Run(); err != nil {
				return err
			}
			return nil
		},
	}

	f := cobraCmd.Flags()
	f.BoolVar(&cmd.showNotes, "notes", false, "show the computed NOTES.txt file as well")
	f.StringVarP(&cmd.releaseName, "name", "n", "release-name", "release name")
	f.BoolVar(&cmd.releaseIsUpgrade, "is-upgrade", false, "set .Release.IsUpgrade instead of .Release.IsInstall")
	f.StringArrayVarP(&cmd.renderFiles, "execute", "x", []string{}, "only execute the given templates")
	f.VarP(&cmd.valueFiles, "values", "f", "specify values in a YAML file (can specify multiple)")
	f.StringVar(&cmd.namespace, "namespace", "", "namespace to install the release into")
	f.StringArrayVar(&cmd.values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&cmd.stringValues, "set-string", []string{}, "set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&cmd.fileValues, "set-file", []string{}, "set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)")
	f.StringVar(&cmd.nameTemplate, "name-template", "", "specify template used to name the release")
	f.StringVar(&cmd.kubeVersion, "kube-version", defaultKubeVersion, "kubernetes version used as Capabilities.KubeVersion.Major/Minor")
	f.StringVar(&cmd.outputDir, "output-dir", "", "writes the executed templates to files in output-dir instead of stdout")

	return cobraCmd
}

func (cmd *templateCmd) Run() error {
	var err error
	ver := version.Get()
	log.WithFields(log.Fields{
		"Version": ver.Version,
		"Commit":  ver.Commit,
		"Branch":  ver.Branch,
	}).Info("Barrelman")

	cmd.Config, err = GetConfigFromFile(cmd.Options.ConfigFile)
	if err != nil {
		return errors.Wrap(err, "got error while loading config")
	}
	archives, err := processManifest(&bfest.Config{
		DataDir:      cmd.Options.DataDir,
		ManifestFile: cmd.Options.ManifestFile,
		AccountTable: cmd.Config.Account,
	}, cmd.Options.NoSync)
	if err != nil {
		return errors.Wrap(err, "apply failed")
	}

	for _, v := range archives.List {
		log.WithFields(log.Fields{
			"File":      v.Path,
			"MetaName":  v.MetaName,
			"Namespace": v.Namespace,
			"ChartName": v.ChartName,
		}).Debug("Template")
		if err := cmd.Export(v.Reader); err != nil {
			return errors.WithFields(errors.Fields{
				"file": v.Path,
				"name": v.MetaName,
			}).Wrap(err, "Export failed")
		}
	}
	return nil
}

func (cmd *templateCmd) Export(inChart io.Reader) error {

	// verify that output-dir exists if provided
	if cmd.outputDir != "" {
		_, err := os.Stat(cmd.outputDir)
		if os.IsNotExist(err) {
			return fmt.Errorf("output-dir '%s' does not exist", cmd.outputDir)
		}
	}

	if cmd.namespace == "" {
		cmd.namespace = defaultNamespace()
	}
	// get combined values and create config
	rawVals, err := vals(cmd.valueFiles, cmd.values, cmd.stringValues, cmd.fileValues, "", "", "")
	if err != nil {
		return err
	}
	config := &chart.Config{Raw: string(rawVals), Values: map[string]*chart.Value{}}
	if msgs := validation.IsDNS1123Subdomain(cmd.releaseName); cmd.releaseName != "" && len(msgs) > 0 {
		return fmt.Errorf("release name %s is invalid: %s", cmd.releaseName, strings.Join(msgs, ";"))
	}

	// If template is specified, try to run the template.
	if cmd.nameTemplate != "" {
		cmd.releaseName, err = generateName(cmd.nameTemplate)
		if err != nil {
			return err
		}
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	c, err := chartutil.LoadArchive(inChart)
	if err != nil {
		return errors.Wrap(err, "chart load failed")
	}

	renderOpts := renderutil.Options{
		ReleaseOptions: chartutil.ReleaseOptions{
			Name:      cmd.releaseName,
			IsInstall: !cmd.releaseIsUpgrade,
			IsUpgrade: cmd.releaseIsUpgrade,
			Time:      timeconv.Now(),
			Namespace: cmd.namespace,
		},
		KubeVersion: cmd.kubeVersion,
	}

	renderedTemplates, err := renderutil.Render(c, config, renderOpts)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Name":      cmd.releaseName,
		"Chart":     c,
		"Config":    config,
		"Namespace": cmd.namespace,
		"Info":      &release.Info{LastDeployed: timeconv.Timestamp(time.Now())},
	}).Debug("chart")

	listManifests := manifest.SplitManifests(renderedTemplates)
	var manifestsToRender []manifest.Manifest

	// if we have a list of files to render, then check that each of the
	// provided files exists in the chart.
	if len(cmd.renderFiles) > 0 {
		for _, f := range cmd.renderFiles {
			missing := true
			if !filepath.IsAbs(f) {
				newF, err := filepath.Abs(filepath.Join(cmd.chartPath, f))
				if err != nil {
					return fmt.Errorf("could not turn template path %s into absolute path: %s", f, err)
				}
				f = newF
			}

			for _, manifest := range listManifests {
				// manifest.Name is rendered using linux-style filepath separators on Windows as
				// well as macOS/linux.
				manifestPathSplit := strings.Split(manifest.Name, "/")
				// remove the chart name from the path
				manifestPathSplit = manifestPathSplit[1:]
				toJoin := append([]string{cmd.chartPath}, manifestPathSplit...)
				manifestPath := filepath.Join(toJoin...)

				// if the filepath provided matches a manifest path in the
				// chart, render that manifest
				if f == manifestPath {
					manifestsToRender = append(manifestsToRender, manifest)
					missing = false
				}
			}
			if missing {
				return fmt.Errorf("could not find template %s in chart", f)
			}
		}
	} else {
		// no renderFiles provided, render all manifests in the chart
		manifestsToRender = listManifests
	}

	for _, m := range tiller.SortByKind(manifestsToRender) {
		data := m.Content
		b := filepath.Base(m.Name)
		if !cmd.showNotes && b == "NOTES.txt" {
			continue
		}
		if strings.HasPrefix(b, "_") {
			continue
		}

		if cmd.outputDir != "" {
			// blank template after execution
			if whitespaceRegex.MatchString(data) {
				continue
			}
			err = writeToFile(cmd.outputDir, m.Name, data)
			if err != nil {
				return err
			}
			continue
		}
		fmt.Printf("--- # Source: %s\n", m.Name)
		fmt.Println(data)
	}
	return nil
}

// write the <data> to <output-dir>/<name>
func writeToFile(outputDir string, name string, data string) error {
	outfileName := strings.Join([]string{outputDir, name}, string(filepath.Separator))

	err := ensureDirectoryForFile(outfileName)
	if err != nil {
		return err
	}

	f, err := os.Create(outfileName)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("--- # Source: %s\n%s", name, data))

	if err != nil {
		return err
	}

	fmt.Printf("wrote %s\n", outfileName)
	return nil
}

// check if the directory exists to create file. creates if don't exists
func ensureDirectoryForFile(file string) error {
	baseDir := path.Dir(file)
	_, err := os.Stat(baseDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return os.MkdirAll(baseDir, defaultDirectoryPermission)
}

// Merges source and destination map, preferring values from the source map
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}

// vals merges values from files specified via -f/--values and
// directly via --set or --set-string or --set-file, marshaling them to YAML
func vals(valueFiles valueFiles, values []string, stringValues []string, fileValues []string, CertFile, KeyFile, CAFile string) ([]byte, error) {
	base := map[string]interface{}{}

	// User specified a values files via -f/--values
	for _, filePath := range valueFiles {
		currentMap := map[string]interface{}{}

		var bytes []byte
		var err error
		if strings.TrimSpace(filePath) == "-" {
			bytes, err = ioutil.ReadAll(os.Stdin)
		} else {
			bytes, err = readFile(filePath, CertFile, KeyFile, CAFile)
		}

		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s", filePath, err)
		}
		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	// User specified a value via --set
	for _, value := range values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	// User specified a value via --set-string
	for _, value := range stringValues {
		if err := strvals.ParseIntoString(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set-string data: %s", err)
		}
	}

	// User specified a value via --set-file
	for _, value := range fileValues {
		reader := func(rs []rune) (interface{}, error) {
			bytes, err := readFile(string(rs), CertFile, KeyFile, CAFile)
			return string(bytes), err
		}
		if err := strvals.ParseIntoFile(value, base, reader); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set-file data: %s", err)
		}
	}

	return yaml.Marshal(base)
}

// printRelease prints info about a release if the Debug is true.
func (cmd *templateCmd) printRelease(rel *release.Release) {
	if rel == nil {
		return
	}
	// TODO: Switch to text/template like everything else.
	fmt.Printf("NAME:   %s\n", rel.Name)
	if settings.Debug {
		printRelease(rel)
	}
}

// locateChartPath looks for a chart directory in known places, and returns either the full path or an error.
//
// This does not ensure that the chart is well-formed; only that the requested filename exists.
//
// Order of resolution:
// - current working directory
// - if path is absolute or begins with '.', error out here
// - chart repos in $HELM_HOME
// - URL
//
// If 'verify' is true, this will attempt to also verify the chart.
func locateChartPath(repoURL, username, password, name, version string, verify bool, keyring,
	certFile, keyFile, caFile string) (string, error) {
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	if fi, err := os.Stat(name); err == nil {
		abs, err := filepath.Abs(name)
		if err != nil {
			return abs, err
		}
		if verify {
			if fi.IsDir() {
				return "", errors.New("cannot verify a directory")
			}
			if _, err := downloader.VerifyChart(abs, keyring); err != nil {
				return "", err
			}
		}
		return abs, nil
	}
	if filepath.IsAbs(name) || strings.HasPrefix(name, ".") {
		return name, fmt.Errorf("path %q not found", name)
	}

	crepo := filepath.Join(settings.Home.Repository(), name)
	if _, err := os.Stat(crepo); err == nil {
		return filepath.Abs(crepo)
	}

	dl := downloader.ChartDownloader{
		HelmHome: settings.Home,
		Out:      os.Stdout,
		Keyring:  keyring,
		Getters:  getter.All(settings),
		Username: username,
		Password: password,
	}
	if verify {
		dl.Verify = downloader.VerifyAlways
	}
	if repoURL != "" {
		chartURL, err := repo.FindChartInAuthRepoURL(repoURL, username, password, name, version,
			certFile, keyFile, caFile, getter.All(settings))
		if err != nil {
			return "", err
		}
		name = chartURL
	}

	if _, err := os.Stat(settings.Home.Archive()); os.IsNotExist(err) {
		os.MkdirAll(settings.Home.Archive(), 0744)
	}

	filename, _, err := dl.DownloadTo(name, version, settings.Home.Archive())
	if err == nil {
		lname, err := filepath.Abs(filename)
		if err != nil {
			return filename, err
		}
		debug("Fetched %s to %s\n", name, filename)
		return lname, nil
	} else if settings.Debug {
		return filename, err
	}

	return filename, fmt.Errorf("failed to download %q (hint: running `helm repo update` may help)", name)
}

func generateName(nameTemplate string) (string, error) {
	t, err := template.New("name-template").Funcs(sprig.TxtFuncMap()).Parse(nameTemplate)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = t.Execute(&b, nil)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func defaultNamespace() string {
	if ns, _, err := kube.GetConfig(settings.KubeContext, settings.KubeConfig).Namespace(); err == nil {
		return ns
	}
	return "default"
}

//readFile load a file from the local directory or a remote file with a url.
func readFile(filePath, CertFile, KeyFile, CAFile string) ([]byte, error) {
	u, _ := url.Parse(filePath)
	p := getter.All(settings)

	// FIXME: maybe someone handle other protocols like ftp.
	getterConstructor, err := p.ByScheme(u.Scheme)

	if err != nil {
		return ioutil.ReadFile(filePath)
	}

	getter, err := getterConstructor(filePath, CertFile, KeyFile, CAFile)
	if err != nil {
		return []byte{}, err
	}
	data, err := getter.Get(filePath)
	return data.Bytes(), err
}
