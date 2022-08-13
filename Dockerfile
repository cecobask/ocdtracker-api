FROM golang:1.18-alpine as builder
WORKDIR /go/src/github.com/cecobask/ocdtracker-api
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/ocdtracker-api

FROM gcr.io/distroless/static-debian11
COPY --from=builder /go/bin/ocdtracker-api .
COPY --from=builder /go/src/github.com/cecobask/ocdtracker-api/internal/db/postgres/migration migration
EXPOSE 8080 8080
CMD ["/ocdtracker-api"]