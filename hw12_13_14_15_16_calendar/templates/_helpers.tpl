{{/*
Базовое имя чарта, усечённое до 63 символов (ограничение DNS-имён Kubernetes).
*/}}
{{- define "calendar.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Полное имя релиза — префикс для всех ресурсов чарта.
*/}}
{{- define "calendar.fullname" -}}
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

{{- define "calendar.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Общие метки. Ожидает dict со ссылкой на корневой контекст (top) и именем компонента (component).
*/}}
{{- define "calendar.labels" -}}
helm.sh/chart: {{ include "calendar.chart" .top }}
{{ include "calendar.selectorLabels" . }}
app.kubernetes.io/version: {{ .top.Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .top.Release.Service }}
app.kubernetes.io/part-of: {{ include "calendar.name" .top }}
{{- end -}}

{{/*
Метки, по которым Deployment находит свои поды, а Service — эндпоинты.
*/}}
{{- define "calendar.selectorLabels" -}}
app.kubernetes.io/name: {{ include "calendar.name" .top }}
app.kubernetes.io/instance: {{ .top.Release.Name }}
app.kubernetes.io/component: {{ .component }}
{{- end -}}

{{/*
Полное имя образа: [registry/]repository:tag.
Ожидает dict с корневым контекстом (top) и секцией образа компонента (image).
*/}}
{{- define "calendar.image" -}}
{{- $registry := .image.registry | default .top.Values.image.registry -}}
{{- $tag := .image.tag | default .top.Values.image.tag | default .top.Chart.AppVersion -}}
{{- if $registry -}}
{{- printf "%s/%s:%s" $registry .image.repository $tag -}}
{{- else -}}
{{- printf "%s:%s" .image.repository $tag -}}
{{- end -}}
{{- end -}}

{{/*
Имя Secret с паролями (созданного чартом или переданного через existingSecret).
*/}}
{{- define "calendar.secretName" -}}
{{- if .Values.existingSecret -}}
{{- .Values.existingSecret -}}
{{- else -}}
{{- printf "%s-secrets" (include "calendar.fullname" .) -}}
{{- end -}}
{{- end -}}

{{/*
Хост и порт PostgreSQL: сервис чарта либо внешняя БД.
*/}}
{{- define "calendar.postgres.host" -}}
{{- if .Values.postgresql.enabled -}}
{{- printf "%s-postgres" (include "calendar.fullname" .) -}}
{{- else -}}
{{- required "postgresql.external.host обязателен при postgresql.enabled=false" .Values.postgresql.external.host -}}
{{- end -}}
{{- end -}}

{{- define "calendar.postgres.port" -}}
{{- if .Values.postgresql.enabled -}}
{{- .Values.postgresql.service.port -}}
{{- else -}}
{{- .Values.postgresql.external.port -}}
{{- end -}}
{{- end -}}

{{/*
Хост и порт RabbitMQ: сервис чарта либо внешний брокер.
*/}}
{{- define "calendar.rabbitmq.host" -}}
{{- if .Values.rabbitmq.enabled -}}
{{- printf "%s-rabbitmq" (include "calendar.fullname" .) -}}
{{- else -}}
{{- required "rabbitmq.external.host обязателен при rabbitmq.enabled=false" .Values.rabbitmq.external.host -}}
{{- end -}}
{{- end -}}

{{- define "calendar.rabbitmq.port" -}}
{{- if .Values.rabbitmq.enabled -}}
{{- .Values.rabbitmq.service.amqpPort -}}
{{- else -}}
{{- .Values.rabbitmq.external.port -}}
{{- end -}}
{{- end -}}

{{/*
DSN PostgreSQL с произвольной подстановкой пароля.
Ожидает dict с корневым контекстом (top) и строкой пароля (password).
*/}}
{{- define "calendar.postgres.dsnWith" -}}
host={{ include "calendar.postgres.host" .top }} port={{ include "calendar.postgres.port" .top }} user={{ .top.Values.postgresql.auth.username }} password={{ .password }} dbname={{ .top.Values.postgresql.auth.database }} sslmode=disable
{{- end -}}

{{/*
DSN для конфиг-файлов приложения: пароль подставляет сам процесс через
os.ExpandEnv, поэтому используется shell-синтаксис ${POSTGRES_PASSWORD}.
*/}}
{{- define "calendar.postgres.dsn" -}}
{{- include "calendar.postgres.dsnWith" (dict "top" . "password" "${POSTGRES_PASSWORD}") -}}
{{- end -}}

{{/*
DSN для переменной окружения контейнера: kubelet раскрывает ссылки на другие
переменные в синтаксисе $(POSTGRES_PASSWORD).
*/}}
{{- define "calendar.postgres.dsnEnv" -}}
{{- include "calendar.postgres.dsnWith" (dict "top" . "password" "$(POSTGRES_PASSWORD)") -}}
{{- end -}}

{{- define "calendar.rabbitmq.dsn" -}}
amqp://{{ .Values.rabbitmq.auth.username }}:${RABBIT_PASSWORD}@{{ include "calendar.rabbitmq.host" . }}:{{ include "calendar.rabbitmq.port" . }}/
{{- end -}}

{{/*
Переменные окружения с паролями, которые подставляются в конфиг-файлы.
*/}}
{{- define "calendar.postgresPasswordEnv" -}}
- name: POSTGRES_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "calendar.secretName" . }}
      key: postgres-password
{{- end -}}

{{- define "calendar.rabbitPasswordEnv" -}}
- name: RABBIT_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "calendar.secretName" . }}
      key: rabbit-password
{{- end -}}

{{/*
Init-контейнеры ожидания зависимостей.
*/}}
{{- define "calendar.waitForPostgres" -}}
- name: wait-for-postgres
  image: {{ .Values.waitImage }}
  imagePullPolicy: {{ .Values.image.pullPolicy }}
  command:
    - sh
    - -c
    - |
      until nc -z {{ include "calendar.postgres.host" . }} {{ include "calendar.postgres.port" . }}; do
        echo "waiting for postgres..."
        sleep 2
      done
{{- end -}}

{{- define "calendar.waitForRabbitmq" -}}
- name: wait-for-rabbitmq
  image: {{ .Values.waitImage }}
  imagePullPolicy: {{ .Values.image.pullPolicy }}
  command:
    - sh
    - -c
    - |
      until nc -z {{ include "calendar.rabbitmq.host" . }} {{ include "calendar.rabbitmq.port" . }}; do
        echo "waiting for rabbitmq..."
        sleep 2
      done
{{- end -}}

{{/*
Секция imagePullSecrets, если они заданы.
*/}}
{{- define "calendar.imagePullSecrets" -}}
{{- with .Values.image.pullSecrets }}
imagePullSecrets:
{{- toYaml . | nindent 2 }}
{{- end }}
{{- end -}}
