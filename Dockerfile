FROM golang:1.24  AS build
ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /app

RUN apt update && apt install -y make ca-certificates tzdata

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN go mod download

COPY . .

RUN go build -o app

FROM scratch
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /app/roles /roles
COPY --from=build /app/static /static
COPY --from=build /app/app /app

ENTRYPOINT ["/app"]