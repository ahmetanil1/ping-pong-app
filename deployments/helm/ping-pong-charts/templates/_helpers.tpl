{{- define "hepapi-case.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{- define "hepapi-case.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}


{{- define "hepapi-case.labels" -}}
helm.sh/chart: {{ include "hepapi-case.name" . }}-{{ .Chart.Version | replace "+" "_" }}
{{ include "hepapi-case.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}


{{- define "hepapi-case.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hepapi-case.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}