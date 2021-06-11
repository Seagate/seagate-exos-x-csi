{{- define "csidriver.labels" -}}
app.kubernetes.io/name: {{ .Chart.Name | kebabcase }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "csidriver.extraArgs" -}}
{{- range .extraArgs }}
  - {{ toYaml . }}
{{- end }}
{{- end -}}