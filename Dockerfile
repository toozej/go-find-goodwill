# setup project and deps
FROM golang:1.26-bookworm AS init

WORKDIR /go/go-find-goodwill/

COPY go.mod* go.sum* ./
RUN go mod download

COPY . ./

FROM init AS vet
RUN go vet ./...

# run tests
FROM init AS test
RUN go test -coverprofile c.out -v ./...

# build binary
FROM init AS build
ARG LDFLAGS

RUN CGO_ENABLED=0 go build -ldflags="${LDFLAGS}"
# Create a dummy data directory to copy over
RUN mkdir /data && chown 65532:65532 /data

# runtime image including CA certs and tzdata
FROM gcr.io/distroless/static-debian12:latest

# Copy data directory with correct ownership (65532 is nonroot)
COPY --from=build --chown=nonroot:nonroot /data /data

# Copy our static executable.
COPY --from=build /go/go-find-goodwill/go-find-goodwill /go/bin/go-find-goodwill

# Expose port for publishing as web service
EXPOSE 8081
# Run as nonroot user (distroless provides 'nonroot' user at UID 65532)
USER nonroot
# Run the binary.
ENTRYPOINT ["/go/bin/go-find-goodwill"]
