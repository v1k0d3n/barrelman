{{/*
Copyright 2018 Kubernetes/Flagship and it's Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/}}

{{- define "kubernetes-common.snippets.kubernetes_resources" -}}
{{- $envAll := index . 0 -}}
{{- $component := index . 1 -}}
{{- if $envAll.Values.pod.resources.enabled -}}
resources:
  limits:
    cpu: {{ $component.limits.cpu | quote }}
    memory: {{ $component.limits.memory | quote }}
  requests:
    cpu: {{ $component.requests.cpu | quote }}
    memory: {{ $component.requests.memory | quote }}
{{- end -}}
{{- end -}}