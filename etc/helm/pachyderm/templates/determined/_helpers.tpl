{{- define "determined.secretPath" -}}
/mount/determined/secrets/
{{- end -}}

{{- define "determined.masterPort" -}}
8081
{{- end -}}

{{- define "pachyderm.pachdAddress" -}}
{{- if .Values.determined.integrations.pachyderm.address -}}
{{ .Values.determined.integrations.pachyderm.address -}}
{{- else -}}
grpc://pachd.{{ .Release.Namespace }}.svc.cluster.local:30650
{{- end -}}
{{- end -}}
