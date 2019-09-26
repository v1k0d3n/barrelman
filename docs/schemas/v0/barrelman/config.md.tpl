## Overview

{{ .Description }}

## Specification

{{ range .Properties -}}
### {{ .Name }}

{{ if .Description }}{{ .Description }}{{ end }}

{{ range .Attributes }}
* {{ . }}
{{- end }}

{{ end }}
