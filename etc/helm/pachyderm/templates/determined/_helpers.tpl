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

{{- define "determined.dbHost" -}}
    {{- if .Values.determined.db.hostAddress }}
        {{- .Values.determined.db.hostAddress }}
    {{- else }}
        {{- "determined-db-service-" }}{{ .Release.Name }}
    {{- end -}}
{{- end -}}

{{- define "determined.genai.PVCName" -}}
    {{- if .Values.determined.genai.sharedPVCName }}
        {{- .Values.determined.genai.sharedPVCName }}
    {{- else }}
        {{- "genai-pvc-" }}{{ .Release.Name }}
    {{- end -}}
{{- end -}}

{{- define "determined.genai.allResourcePoolNames" -}}
    {{- $orig_resource_pool_data := (required "A valid .Values.determined.resourcePools entry required!" .Values.determined.resourcePools) }}
    {{- $resource_pools := list -}}
    {{- range $v := $orig_resource_pool_data }}
        {{- $resource_pools = mustAppend $resource_pools $v.pool_name }}
    {{- end }}
    {{ toJson $resource_pools }}
{{- end }}

