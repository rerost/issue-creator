schedule: {{.Schedule}}
command: [{{- range $index, $var := .Commands}}{{- if ne $index 0}}, {{- end}}"{{$var}}"{{- end}}]
name: {{.Name}}
