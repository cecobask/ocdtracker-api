FROM golang:1.18-alpine as builder
WORKDIR /go/src/github.com/cecobask/ocd-tracker-api
COPY . .
RUN go mod download && CGO_ENABLED=0 go build -o /go/bin/ocd-tracker-api

FROM gcr.io/distroless/static-debian11
COPY --from=builder /go/bin/ocd-tracker-api .
EXPOSE 8080 8080
CMD ["/ocd-tracker-api"]