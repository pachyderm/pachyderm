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

{{- define "pachyderm.consoleSecret" -}}
{{ include "defaultOrStableHash" (dict "defaultValue" .Values.console.config.oauthClientSecret "hashSalt" "consoleSecret" "onlyDefault" .Release.IsUpgrade) }}
{{- end -}}

{{- define "pachyderm.clusterDeploymentId" -}}
{{ include "defaultOrStableHash" (dict "defaultValue" .Values.pachd.clusterDeploymentID "hashSalt" "deploymentID" "onlyDefault" .Release.IsUpgrade) }}
{{- end -}}

{{- define "pachyderm.enterpriseSecret" -}}
{{ include "defaultOrStableHash" (dict "defaultValue" .Values.pachd.enterpriseSecret "hashSalt" "enterpriseSecret" "onlyDefault" .Release.IsUpgrade) }}
{{- end -}}


## Input:
## This function expects a context containing 'defaultValue', 'hashSalt' and 'onlyDefault' keys
##
## Behavior: 
## Use 'defaultValue' if it's provided, otherwise use the current date/time to create a stable hash
## incorporating the 'hashSalt' value to produce a different hash per client template's usage. 
## This way different caller templates can produce different random values.
##
## Discussions: 
## 1.) The reason we use a a truncated time string to create the hash, 
## is because we want to this template to produce the same hash each 
## time it's executed in a release. In this case, the current timestamp 
## gives us a unique enough value to use as a hash seed. We truncate it 
## to the minute to reduce the odds that it will produce different timestamp 
## values between renderings of this template.
##
## 2.) 'onlyDefault' is used to force this template to fail if a default value is not provided
{{- define "defaultOrStableHash" -}}
{{- if .defaultValue }}
    {{- .defaultValue }}
{{- else }}
    {{- if .onlyDefault }}
        {{- fail "must provide a default value during release upgrade." }}
    {{- else }}
        {{- derivePassword 1 "long" (now | toString | trunc 16) .hashSalt "pachyderm" -}}
    {{- end }}
{{- end }}
{{- end -}}