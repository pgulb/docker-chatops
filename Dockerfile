FROM golang:1.22.3-alpine3.20 AS buildstage

COPY . .
RUN go mod download && cd bot/exec && go build -o /app/chatops -ldflags="-extldflags=-static"

FROM scratch
COPY --from=buildstage /app/chatops /app/chatops
WORKDIR /app
CMD [ "/app/chatops" ]
