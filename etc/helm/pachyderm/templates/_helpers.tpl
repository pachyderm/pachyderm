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

{{- define "pachyderm.ingressproto" -}}
{{- if or .Values.ingress.tls.enabled .Values.ingress.uriHttpsProtoOverride -}}
https
{{- else -}}
http
{{- end -}}
{{- end }}

{{- define "pachyderm.issuerURI" -}}
{{- if .Values.oidc.issuerURI -}}
{{ .Values.oidc.issuerURI }}
{{- else if .Values.ingress.host -}}
{{- printf "%s://%s/dex" (include "pachyderm.ingressproto" .) .Values.ingress.host -}}
{{- else if not .Values.ingress.enabled -}}
{{- if eq .Values.pachd.service.type "NodePort" -}}
http://pachd:1658
{{- else -}}
http://pachd:30658
{{- end -}}
{{- else -}}
{{ fail "For Authentication, an OIDC Issuer for this pachd must be set." }}
{{- end -}}
{{- end }}

{{- /*
reactAppRuntimeIssuerURI: The URI without the path of the user accessible issuerURI. 
ie. In local deployments, this is http://localhost:30658, while the issuer URI is http://pachd:30658
In deployments where the issuerURI is user accessible (ie. Via ingress) this would be the issuerURI without the path
*/ -}}
{{- define "pachyderm.reactAppRuntimeIssuerURI" -}}
{{- if .Values.console.config.reactAppRuntimeIssuerURI -}}
{{ .Values.console.config.reactAppRuntimeIssuerURI }}
{{- else if .Values.ingress.host -}}
{{- printf "%s://%s" (include "pachyderm.ingressproto" .) .Values.ingress.host -}}
{{- else if not .Values.ingress.enabled -}}
http://localhost:30658
{{- end }}
{{- end }}

{{- define "pachyderm.consoleRedirectURI" -}}
{{- if .Values.console.config.oauthRedirectURI -}}
{{ .Values.console.config.oauthRedirectURI }}
{{- else if .Values.ingress.host -}}
{{- printf "%s://%s/oauth/callback/?inline=true" (include "pachyderm.ingressproto" .) .Values.ingress.host -}}
{{- else if not .Values.ingress.enabled -}}
http://localhost:4000/oauth/callback/?inline=true
{{- else -}}
{{ fail "To connect Console to Pachyderm, Console's Redirect URI must be set." }}
{{- end }}
{{- end }}

{{- define "pachyderm.pachdRedirectURI" -}}
{{- if .Values.pachd.oauthRedirectURI -}}
{{ .Values.pachd.oauthRedirectURI -}}
{{- else if .Values.ingress.host -}}
{{- printf "%s://%s/authorization-code/callback" (include "pachyderm.ingressproto" .) .Values.ingress.host -}}
{{- else if not .Values.ingress.enabled -}}
http://localhost:30657/authorization-code/callback
{{- else -}}
{{ fail "For Authentication, an OIDC Redirect URI for this pachd must be set." }}
{{- end -}}
{{- end }}

{{- define "pachyderm.pachdPeerAddress" -}}
pachd-peer.{{ .Release.Namespace }}.svc.cluster.local:30653
{{- end }}

{{- define "pachyderm.localhostIssuer" -}}
{{- if .Values.pachd.localhostIssuer -}}
  {{- if eq .Values.pachd.localhostIssuer "true" -}}
    true
  {{- else if eq .Values.pachd.localhostIssuer "false" -}}
    false
  {{- else -}}
    {{- fail "pachd.localhostIssuer must either be set to the string value of \"true\" or \"false\"" }}
  {{- end -}}
{{- else if .Values.pachd.activateEnterpriseMember -}}
false
{{- else if not .Values.ingress.enabled -}}
true
{{- else if .Values.ingress.host -}}
false
{{- end -}}
{{- end }}

{{- define "pachyderm.userAccessibleOauthIssuerHost" -}}
{{- if .Values.oidc.userAccessibleOauthIssuerHost -}}
{{ .Values.oidc.userAccessibleOauthIssuerHost }}
{{- else if not .Values.ingress.enabled -}}
localhost:30658
{{- end -}}
{{- end }}

{{- define "pachyderm.idps" -}}
{{- if .Values.oidc.upstreamIDPs }}
{{ toYaml .Values.oidc.upstreamIDPs | indent 4 }}
{{- else if .Values.oidc.mockIDP }}
    - id: test
      name: test
      type: mockPassword
      jsonConfig: '{"username": "admin", "password": "password"}'
{{- else }}
    {{- fail "either oidc.upstreamIDPs or oidc.mockIDP must be set in non-LOCAL deployments" }}
{{- end }}
{{- end }}
