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

{{- define "pachyderm.issuerURI" -}}
{{- if .Values.console.config.issuerURI }}
{{ .Values.console.config.issuerURI }}
{{- else if .Values.ingress.host -}}
https://{{ .Values.ingress.host }}/dex
{{- else if eq .Values.deployTarget "LOCAL" -}}
http://pachd:1658
{{- end }}
{{- end }}

{{- define "pachyderm.reactAppRuntimeIssuerURI" -}}
{{- if .Values.console.config.reactAppRuntimeIssuerURI }}
{{ .Values.console.config.reactAppRuntimeIssuerURI }}
{{- else if .Values.ingress.host -}}
https://{{ .Values.ingress.host }}/dex
{{- else if eq .Values.deployTarget "LOCAL" -}}
http://localhost:30658/
{{- end }}
{{- end }}

{{- define "pachyderm.consoleRedirectURI" -}}
{{- if .Values.console.config.oauthRedirectURI }}
{{ .Values.console.config.oauthRedirectURI }}
{{- else if .Values.ingress.host -}}
https://{{ .Values.ingress.host }}/oauth/callback/?inline=true
{{- else if eq .Values.deployTarget "LOCAL" -}}
http://localhost:4000/oauth/callback/?inline=true
{{- end }}
{{- end }}

{{- define "pachyderm.pachdRedirectURI" -}}
{{- if .Values.pachd.oauthRedirectURI }}
{{ .Values.console.config.oauthRedirectURI }}
{{- else if .Values.ingress.host -}}
https://{{ .Values.ingress.host }}/authorization-code/callback
{{- else if eq .Values.deployTarget "LOCAL" -}}
http://localhost:30657/authorization-code/callback
{{- end }}
{{- end }}

{{- define "pachyderm.pachdPeerAddress" -}}
pachd-peer.{{ .Release.Namespace }}.svc.cluster.local:30653
{{- end }}
