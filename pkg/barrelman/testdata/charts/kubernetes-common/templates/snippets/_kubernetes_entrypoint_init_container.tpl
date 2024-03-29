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

{{- define "kubernetes-common.snippets.kubernetes_entrypoint_init_container" -}}
{{- $envAll := index . 0 -}}
{{- $deps := index . 1 -}}
{{- $mounts := index . 2 -}}
- name: init
  image: {{ $envAll.Values.images.tags.dep_check }}
  imagePullPolicy: {{ $envAll.Values.images.pull_policy }}
  env:
    - name: POD_NAME
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.name
    - name: NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    - name: INTERFACE_NAME
      value: eth0
    - name: PATH
      value: /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/
    - name: DEPENDENCY_SERVICE
      value: "{{ tuple $deps.services $envAll | include "kubernetes-common.utils.comma_joined_service_list" }}"
    - name: DEPENDENCY_JOBS
      value: "{{  include "kubernetes-common.utils.joinListWithComma" $deps.jobs }}"
    - name: DEPENDENCY_DAEMONSET
      value: "{{  include "kubernetes-common.utils.joinListWithComma" $deps.daemonset }}"
    - name: DEPENDENCY_CONTAINER
      value: "{{  include "kubernetes-common.utils.joinListWithComma" $deps.container }}"
    - name: DEPENDENCY_POD
      value: {{ if $deps.pod }}{{ toJson $deps.pod | quote }}{{ else }}""{{ end }}
    - name: COMMAND
      value: "echo done"
  command:
    - kubernetes-entrypoint
  volumeMounts:
{{ toYaml $mounts | indent 4 }}
{{- end -}}
