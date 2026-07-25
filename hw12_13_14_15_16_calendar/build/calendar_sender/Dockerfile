FROM golang:1.25-alpine AS build

ENV BIN_FILE=/opt/calendar/calendar-sender
ENV CODE_DIR=/go/src/

WORKDIR ${CODE_DIR}

COPY go.mod go.sum ./
RUN go mod download

COPY . ${CODE_DIR}

ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} ./cmd/calendar_sender

FROM alpine:3.20

LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="calendar_sender"
LABEL MAINTAINERS="student@otus.ru"

ENV BIN_FILE=/opt/calendar/calendar-sender
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE=/etc/calendar/sender_config.yaml
COPY ./configs/docker/sender_config.yaml ${CONFIG_FILE}

CMD ${BIN_FILE} -config ${CONFIG_FILE}
