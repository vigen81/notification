FROM golang:alpine AS builder
ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /go/src/notification-service

RUN apk --update --no-cache add ca-certificates make git
ENV GOPRIVATE="gitlab.com/healthcare-integration/golang"
ARG gitlab_user
ENV gitlab_user=$gitlab_user
ARG gitlab_personal_token
ENV gitlab_personal_token=$gitlab_personal_token
RUN git config --global url."https://${gitlab_user}:${gitlab_personal_token}@gitlab.com:".insteadOf "https://gitlab.com"


COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN make build

FROM scratch
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /go/src/notification-service/notification-service /notification-service
ENTRYPOINT ["/notification-service"]
CMD []
