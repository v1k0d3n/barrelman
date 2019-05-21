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
	"github.com/cirrocloud/yamlpack"
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

	bfest "github.com/charter-oss/barrelman/pkg/manifest"
	"github.com/charter-oss/barrelman/pkg/version"
	"github.com/charter-oss/structured"
	"github.com/charter-oss/structured/errors"
	"github.com/charter-oss/structured/log"
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

//TemplateCmd is the configuration and state for the template command
type TemplateCmd struct {
	Namespace        string
	ValueFiles       valueFiles
	ChartPath        string
	Out              io.Writer
	Values           []string
	StringValues     []string
	FileValues       []string
	NameTemplate     string
	ShowNotes        bool
	ReleaseName      string
	ReleaseIsUpgrade bool
	RenderFiles      []string
	KubeVersion      string
	OutputDir        string
	Options          *CmdOptions
	Config           *Config
	Log              structured.Logger
	LogOptions       *[]string
}

//NewTemplateCmd returns an initialized *TemplateCmd
func NewTemplateCmd() *TemplateCmd {

	logOptions := &[]string{}
	return &TemplateCmd{
		Options:    &CmdOptions{},
		Config:     &Config{},
		LogOptions: logOptions,
	}
}

func (cmd *TemplateCmd) Run() error {
	var err error
	ver := version.Get()
	log.WithFields(log.Fields{
		"Version": ver.Version,
		"Commit":  ver.Commit,
		"Branch":  ver.Branch,
	}).Info("Barrelman")

	if cmd.Options.ConfigFile != "" {
		cmd.Config, err = GetConfigFromFile(cmd.Options.ConfigFile)
		if err != nil {
			return errors.Wrap(err, "got error while loading config")
		}
	} else {
		cmd.Config = GetEmptyConfig()
	}

	archives, manifestName, err := processManifest(&bfest.Config{
		DataDir:      cmd.Options.DataDir,
		ManifestFile: cmd.Options.ManifestFile,
		AccountTable: cmd.Config.Account,
	}, cmd.Options.NoSync)
	if err != nil {
		return errors.Wrap(err, "template failed")
	}

	for _, v := range archives.List {
		log.WithFields(log.Fields{
			"File":         v.Path,
			"MetaName":     v.MetaName,
			"Namespace":    v.Namespace,
			"ChartName":    v.ChartName,
			"ManifestName": manifestName,
		}).Debug("Template")
		if err := cmd.Export(v); err != nil {
			return errors.WithFields(errors.Fields{
				"file": v.Path,
				"name": v.MetaName,
			}).Wrap(err, "Export failed")
		}
	}
	return nil
}

func (cmd *TemplateCmd) RunWithSections(ys []*yamlpack.YamlSection) error {
	var err error
	ver := version.Get()
	log.WithFields(log.Fields{
		"Version": ver.Version,
		"Commit":  ver.Commit,
		"Branch":  ver.Branch,
	}).Info("Barrelman")

	if cmd.Options.ConfigFile != "" {
		cmd.Config, err = GetConfigFromFile(cmd.Options.ConfigFile)
		if err != nil {
			return errors.Wrap(err, "got error while loading config")
		}
	} else {
		cmd.Config = GetEmptyConfig()
	}

	archives, err := processManifestSections(&bfest.Config{
		DataDir:      cmd.Options.DataDir,
		ManifestFile: cmd.Options.ManifestFile,
		AccountTable: cmd.Config.Account,
	}, ys, cmd.Options.NoSync)
	if err != nil {
		return errors.Wrap(err, "template failed")
	}

	for _, v := range archives.List {
		log.WithFields(log.Fields{
			"File":      v.Path,
			"MetaName":  v.MetaName,
			"Namespace": v.Namespace,
			"ChartName": v.ChartName,
		}).Debug("Template")
		if err := cmd.Export(v); err != nil {
			return errors.WithFields(errors.Fields{
				"file": v.Path,
				"name": v.MetaName,
			}).Wrap(err, "Export failed")
		}
	}
	return nil
}

//func (cmd *TemplateCmd) Export(inChart io.Reader) error {
func (cmd *TemplateCmd) Export(as *bfest.ArchiveSpec) error {

	// verify that output-dir exists if provided
	if cmd.OutputDir != "" {
		_, err := os.Stat(cmd.OutputDir)
		if os.IsNotExist(err) {
			return fmt.Errorf("output-dir '%s' does not exist", cmd.OutputDir)
		}
	}
	// get combined values and create config
	rawVals, err := vals(as.Overrides, cmd.ValueFiles, cmd.Values, cmd.StringValues, cmd.FileValues, "", "", "")
	if err != nil {
		return err
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	c, err := chartutil.LoadArchive(as.Reader)
	if err != nil {
		return errors.Wrap(err, "chart load failed")
	}

	config := &chart.Config{Raw: string(rawVals), Values: map[string]*chart.Value{}}
	if msgs := validation.IsDNS1123Subdomain(as.ReleaseName); as.ReleaseName != "" && len(msgs) > 0 {
		return fmt.Errorf("release name %s is invalid: %s", as.ReleaseName, strings.Join(msgs, ";"))
	}

	// If template is specified, try to run the template.
	if cmd.NameTemplate != "" {
		cmd.ReleaseName, err = generateName(cmd.NameTemplate)
		if err != nil {
			return err
		}
	}

	renderOpts := renderutil.Options{
		ReleaseOptions: chartutil.ReleaseOptions{
			Name:      as.ReleaseName,
			IsInstall: !cmd.ReleaseIsUpgrade,
			IsUpgrade: cmd.ReleaseIsUpgrade,
			Time:      timeconv.Now(),
			Namespace: as.Namespace,
		},
		KubeVersion: cmd.KubeVersion,
	}

	renderedTemplates, err := renderutil.Render(c, config, renderOpts)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Name":      as.ReleaseName,
		"Config":    config,
		"Namespace": as.Namespace,
		"Info":      &release.Info{LastDeployed: timeconv.Timestamp(time.Now())},
	}).Debug("chart")

	listManifests := manifest.SplitManifests(renderedTemplates)
	var manifestsToRender []manifest.Manifest

	// if we have a list of files to render, then check that each of the
	// provided files exists in the chart.
	if len(cmd.RenderFiles) > 0 {
		for _, f := range cmd.RenderFiles {
			missing := true
			if !filepath.IsAbs(f) {
				newF, err := filepath.Abs(filepath.Join(cmd.ChartPath, f))
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
				toJoin := append([]string{cmd.ChartPath}, manifestPathSplit...)
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
		if !cmd.ShowNotes && b == "NOTES.txt" {
			continue
		}
		if strings.HasPrefix(b, "_") {
			continue
		}

		if cmd.OutputDir != "" {
			// blank template after execution
			if whitespaceRegex.MatchString(data) {
				continue
			}
			err = writeToFile(cmd.OutputDir, b, data)
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

	_, err = f.WriteString(fmt.Sprintf("--- # Source: %s\n%s\n", name, data))

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
func vals(overrideBytes []byte, valueFiles valueFiles, values []string, stringValues []string, fileValues []string, CertFile, KeyFile, CAFile string) ([]byte, error) {
	base := map[string]interface{}{}
	overrides := map[string]interface{}{}
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
			return []byte{}, errors.WithFields(errors.Fields{
				"File": filePath,
			}).Wrap(err, "failed to parse file")
		}
		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	// User specified a value via --set
	for _, value := range values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, errors.Wrap(err, "failed parsing --set")
		}
	}

	// User specified a value via --set-string
	for _, value := range stringValues {
		if err := strvals.ParseIntoString(value, base); err != nil {
			return []byte{}, errors.Wrap(err, "failed parsing --set-string")
		}
	}

	// User specified a value via --set-file
	for _, value := range fileValues {
		reader := func(rs []rune) (interface{}, error) {
			bytes, err := readFile(string(rs), CertFile, KeyFile, CAFile)
			return string(bytes), err
		}
		if err := strvals.ParseIntoFile(value, base, reader); err != nil {
			return []byte{}, errors.Wrap(err, "failed parsing --set-file")
		}
	}

	if err := yaml.Unmarshal(overrideBytes, &overrides); err != nil {
		return []byte{}, errors.Wrap(err, "failed to parse overrides")
	}
	base = mergeValues(base, overrides)
	return yaml.Marshal(base)
}

// printRelease prints info about a release if the Debug is true.
func (cmd *TemplateCmd) printRelease(rel *release.Release) {
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
