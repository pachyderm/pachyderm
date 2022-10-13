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

{{- define "pachyderm.hostproto" -}}
{{- if or .Values.ingress.tls.enabled .Values.proxy.tls.enabled .Values.ingress.uriHttpsProtoOverride -}}
https
{{- else -}}
http
{{- end -}}
{{- end }}

{{- define "pachyderm.host" -}}
{{- if .Values.ingress.enabled -}}
{{ required "if ingress is enabled, an ingress.host is required" .Values.ingress.host }}
{{- else if .Values.proxy.enabled -}}
{{ .Values.proxy.host }}
{{- end -}}
{{- end }}



{{- define "pachyderm.issuerURI" -}}
{{- if .Values.oidc.issuerURI -}}
{{- if and .Values.proxy.enabled (not (hasSuffix "/dex" .Values.oidc.issuerURI)) -}}
{{ fail (printf "With the proxy enabled, oidc.issuerURI must end with /dex, not %s" .Values.oidc.issuerURI) }}
{{- end -}}
{{ .Values.oidc.issuerURI }}
{{- else if and .Values.ingress.host .Values.ingress.enabled -}}
{{- printf "%s://%s/dex" (include "pachyderm.hostproto" .) .Values.ingress.host -}}
{{- else if and .Values.proxy.host .Values.proxy.enabled -}}
{{- printf "%s://%s/dex" (include "pachyderm.hostproto" .) .Values.proxy.host -}}
{{- else if .Values.proxy.enabled -}}
http://pachd:30658/dex
{{- else -}}
{{- if eq .Values.pachd.service.type "NodePort" -}}
http://pachd:1658
{{- else -}}
http://pachd:30658
{{- end -}}
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
{{- else if (include "pachyderm.host" .) -}}
{{- printf "%s://%s" (include "pachyderm.hostproto" .) (include "pachyderm.host" .) -}}
{{- else if not .Values.ingress.enabled -}}
http://localhost:30658
{{- end }}
{{- end }}

{{- define "pachyderm.consoleRedirectURI" -}}
{{- if .Values.console.config.oauthRedirectURI -}}
{{ .Values.console.config.oauthRedirectURI }}
{{- else if (include "pachyderm.host" .) -}}
{{- printf "%s://%s/oauth/callback/?inline=true" (include "pachyderm.hostproto" .) (include "pachyderm.host" .) -}}
{{- else if not .Values.ingress.enabled -}}
http://localhost:4000/oauth/callback/?inline=true
{{- else -}}
{{ fail "To connect Console to Pachyderm, Console's Redirect URI must be set." }}
{{- end }}
{{- end }}

{{- define "pachyderm.pachdRedirectURI" -}}
{{- if .Values.pachd.oauthRedirectURI -}}
{{ .Values.pachd.oauthRedirectURI -}}
{{- else if (include "pachyderm.host" .) -}}
{{- printf "%s://%s/authorization-code/callback" (include "pachyderm.hostproto" .) (include "pachyderm.host" .) -}}
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
{{- else -}}
true
{{- end -}}
{{- end }}

{{- define "pachyderm.userAccessibleOauthIssuerHost" -}}
{{- if .Values.oidc.userAccessibleOauthIssuerHost -}}
{{ .Values.oidc.userAccessibleOauthIssuerHost }}
{{- else if (include "pachyderm.host" .) -}}
{{- (include "pachyderm.host" .) -}}
{{- else  -}}
localhost:30658
{{- end -}}
{{- end }}

{{- define "pachyderm.idps" -}}
{{- if .Values.oidc.upstreamIDPs -}}
{{ toYaml .Values.oidc.upstreamIDPs }}
{{- else if .Values.oidc.mockIDP -}}
- id: test
  name: test
  type: mockPassword
  config:
    username: admin
    password: password
{{- else }}
    {{- fail "either oidc.upstreamIDPs or oidc.mockIDP must be set in non-LOCAL deployments" }}
{{- end }}
{{- end }}

{{- define "pachyderm.withEnterprise" }}
{{- if or (include "pachyderm.enterpriseLicenseKeySecretName" . ) .Values.pachd.activateEnterpriseMember }}
true
{{- end }}
{{- end }}

{{- define "pachyderm.enterpriseLicenseKeySecretName" -}}
{{- if .Values.pachd.enterpriseLicenseKeySecretName }}
{{ .Values.pachd.enterpriseLicenseKeySecretName }}
{{- else if .Values.pachd.enterpriseLicenseKey }}
pachyderm-license
{{- end }}
{{- end }}

{{- define "pachyderm.enterpriseSecretSecretName" -}}
{{- if .Values.pachd.enterpriseSecretSecretName }}
{{ .Values.pachd.enterpriseSecretSecretName }}
{{- else if or .Values.pachd.enterpriseSecret (include "pachyderm.enterpriseLicenseKeySecretName" . ) }}
pachyderm-enterprise
{{- end }}
{{- end }}

{{- define "pachyderm.upstreamIDPsSecretName" -}}
{{- if .Values.oidc.upstreamIDPsSecretName }}
{{ .Values.oidc.upstreamIDPsSecretName }}
{{- else if or .Values.oidc.upstreamIDPs .Values.oidc.mockIDP }}
pachyderm-identity
{{- end }}
{{- end }}

{{- define "pachyderm.mockIDPRoleBindings" -}}
{{- if and .Values.oidc.mockIDP (not .Values.oidc.upstreamIDPsSecretName) (not .Values.oidc.upstreamIDPs) }}
user:kilgore@kilgore.trout:
- clusterAdmin
{{- end }}
{{- end }}
