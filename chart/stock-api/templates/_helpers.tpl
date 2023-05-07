{{/* vim: set filetype=mustache: */}}

{{/*
stock-api home
*/}}
{{- define "stock-api.home" -}}
{{- print "/stock-api" -}}
{{- end -}}

{{/*
Expand the name of the chart.
*/}}
{{- define "stock-api.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Expand the namespace of the chart.
*/}}
{{- define "stock-api.namespace" -}}
{{- default .Release.Namespace .Values.namespace  -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "stock-api.fullname" -}}
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

{{/*
Define cluster's name
*/}}
{{- define "stock-api.cluster.name" -}}
{{- if .Values.clusterName }}
{{- .Values.clusterName }}
{{- else -}}
{{- template "stock-api.fullname" .}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "stock-api.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the common labels.
*/}}
{{- define "stock-api.standardLabels" -}}
app: {{ template "stock-api.name" . }}
chart: {{ template "stock-api.chart" . }}
release: {{ .Release.Name }}
heritage: {{ .Release.Service }}
cluster: {{ template "stock-api.cluster.name" . }}
{{- end }}

{{/*
Create the template labels.
*/}}
{{- define "stock-api.template.labels" -}}
app: {{ template "stock-api.name" . }}
release: {{ .Release.Name }}
cluster: {{ template "stock-api.cluster.name" . }}
{{- end }}

{{/*
Create the match labels.
*/}}
{{- define "stock-api.matchLabels" -}}
app: {{ template "stock-api.name" . }}
release: {{ .Release.Name }}
{{- end }}
