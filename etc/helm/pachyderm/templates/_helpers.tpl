{{- /*
SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
SPDX-License-Identifier: Apache-2.0
*/ -}}
{{- /* vim: set filetype=mustache: */ -}}

{{- define "pachyderm.storageBackend" -}}
{{- if eq .Values.deployTarget "" }}
{{ fail "deployTarget must be set" }}
{{- end }}
{{- if .Values.pachd.storage.backend -}}
{{ .Values.pachd.storage.backend }}
{{- else if eq .Values.deployTarget "AMAZON" -}}
AMAZON
{{- else if eq .Values.deployTarget "GOOGLE" -}}
GOOGLE
{{- else if eq .Values.deployTarget "MICROSOFT" -}}
MICROSOFT
{{- else if eq .Values.deployTarget "LOCAL" -}}
LOCAL
{{- else -}}
{{ fail "pachd.storage.backend required when no matching deploy target found" }}
{{- end -}}
{{- end -}}

{{- define "pachyderm.clusterDeploymentId" -}}
{{ default (randAlphaNum 32) .Values.pachd.clusterDeploymentID }}
{{- end -}}

{{- define "pachyderm.imagePullSecrets" -}}
{{- if .Values.global.imagePullSecrets }}
imagePullSecrets:
  {{- range .Values.global.imagePullSecrets }}
  - name: {{ . }}
  {{- end }}
{{- end }}
{{- end -}}
