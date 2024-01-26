FROM golang:alpine AS build
ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /go/src/app
RUN apk add make ca-certificates tzdata git

ENV GOPRIVATE="gitlab.com/healthcare-integration"
ARG gitlab_user
ENV gitlab_user=$gitlab_user
ARG gitlab_personal_token
ENV gitlab_personal_token=$gitlab_personal_token
RUN git config --global url."https://${gitlab_user}:${gitlab_personal_token}@gitlab.com:".insteadOf "https://gitlab.com"



COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make  build



FROM scratch
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /go/src/app/app /app


ENTRYPOINT ["/app"]

