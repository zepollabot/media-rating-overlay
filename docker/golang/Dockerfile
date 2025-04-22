FROM golang:1.24-alpine AS base
LABEL authors="pietro.pollarolo@gmail.com"

FROM base AS dev

WORKDIR /go/src/app
COPY . .

RUN go mod download

CMD ["go", "run", "main.go"]

FROM dev AS test

COPY --from=dev . .

RUN go install github.com/vektra/mockery/v2@v2.53.3

CMD ["go", "run", "main.go"]

FROM base AS builder

WORKDIR /

# Copy only necessary files for production build
COPY go.mod go.sum ./
COPY main.go ./
COPY configs/ ./configs/
# Copy internal directory excluding test files
COPY --from=dev /go/src/app/internal/ ./internal/
RUN find ./internal -type f -name "*_test.go" -delete && \
    find ./internal -type d -name "mocks" -exec rm -rf {} +

RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -o /go/bin/app

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian11 AS prod

ENV ENV=PROD

#USER nonroot:nonroot

COPY --from=builder /go/bin/app /
COPY internal/processor/text/fonts /internal/processor/text/fonts
COPY internal/processor/image/data /internal/processor/image/data

CMD ["/app"]