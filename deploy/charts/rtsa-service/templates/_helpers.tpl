{{/* CLASSIFICATION: UNCLASSIFIED */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "rtsa-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "rtsa-service.fullname" -}}
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

{{- define "rtsa-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels.
*/}}
{{- define "rtsa-service.labels" -}}
helm.sh/chart: {{ include "rtsa-service.chart" . }}
{{ include "rtsa-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: rtsa
classification: unclassified
{{- end -}}

{{/*
Selector labels.
*/}}
{{- define "rtsa-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "rtsa-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
ServiceAccount name.
*/}}
{{- define "rtsa-service.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "rtsa-service.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{/*
Container image reference. Fails the render if no repository is supplied.
*/}}
{{- define "rtsa-service.image" -}}
{{- if not .Values.image.repository -}}
{{- fail "image.repository is required (e.g. <acr-login-server>/<service>)" -}}
{{- end -}}
{{- $tag := default .Chart.AppVersion .Values.image.tag -}}
{{- printf "%s:%s" .Values.image.repository $tag -}}
{{- end -}}
