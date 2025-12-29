{{/*
Expand the name of the chart.
*/}}
{{- define "ndots-webhook.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ndots-webhook.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ndots-webhook.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ndots-webhook.labels" -}}
helm.sh/chart: {{ include "ndots-webhook.chart" . }}
{{ include "ndots-webhook.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ndots-webhook.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ndots-webhook.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Common annotations for all resources
*/}}
{{- define "ndots-webhook.annotations" -}}
{{- with .Values.commonAnnotations }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ndots-webhook.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "ndots-webhook.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the TLS secret
*/}}
{{- define "ndots-webhook.tlsSecretName" -}}
{{- printf "%s-tls" (include "ndots-webhook.fullname" .) }}
{{- end }}

{{/*
Create the webhook service DNS name
*/}}
{{- define "ndots-webhook.serviceDnsName" -}}
{{- printf "%s.%s.svc" (include "ndots-webhook.fullname" .) .Release.Namespace }}
{{- end }}

{{/*
Create the issuer name
*/}}
{{- define "ndots-webhook.issuerName" -}}
{{- if .Values.tls.certManager.issuer.name }}
{{- .Values.tls.certManager.issuer.name }}
{{- else }}
{{- printf "%s-issuer" (include "ndots-webhook.fullname" .) }}
{{- end }}
{{- end }}

