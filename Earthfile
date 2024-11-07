VERSION 0.8

ARG --global ALPINE_VERSION="3.19"

ci:
    BUILD +lint
    BUILD +test
    BUILD +image


godeps:
    ARG --required GOLANG_VERSION

    FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION}

    WORKDIR /src
    COPY go.mod go.sum ./
    RUN go mod download

lint:
    FROM +godeps

    ARG --required GOLANGCILINT_VERSION

    RUN echo "installing golangci-lint" \
        && wget https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCILINT_VERSION}/golangci-lint-${GOLANGCILINT_VERSION}-linux-amd64.tar.gz \
            --output-document golangcilint.tgz \
        && tar xvf golangcilint.tgz

    COPY . .

    RUN ./golangci-lint-${GOLANGCILINT_VERSION}-linux-amd64/golangci-lint run --timeout 10m

test:
    FROM +godeps

    COPY . .

    RUN go test ./...

build:
    FROM +godeps

    COPY . .

    RUN apk add gcc libc-dev

    ENV CGO_ENABLED=1
    ARG --required COMMIT_SHA
    RUN go build -o calendar-sync -ldflags "-X 'calendar-sync/cmd.CommitSHA=${COMMIT_SHA}' -X 'calendar-sync/cmd.BuildDate=$(date)'" .

    SAVE ARTIFACT calendar-sync AS LOCAL calendar-sync

image:
    FROM alpine:${ALPINE_VERSION}

    COPY +build/calendar-sync /bin/

    RUN /bin/calendar-sync --help

release:
    FROM +image

    ENTRYPOINT /bin/calendar-sync

    ARG --required image
    SAVE IMAGE --push ${image}
