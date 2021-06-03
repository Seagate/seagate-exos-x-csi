{{- define "exosx.labels" -}}
app.kubernetes.io/name: {{ .Chart.Name | kebabcase }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "exosx.extraArgs" -}}
{{- range .extraArgs }}
  - {{ toYaml . }}
{{- end }}
{{- end -}}