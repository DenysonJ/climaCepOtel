FROM golang:1.26 AS build
WORKDIR /app
COPY . .
# SERVICE define qual main compilar: cmd/cep ou cmd/clima
ARG SERVICE=cep
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/climaCEP ./cmd/${SERVICE}

FROM scratch
WORKDIR /app
COPY --from=build /app/climaCEP .
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["./climaCEP"]

