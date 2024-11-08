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

    RUN apk add gcc musl-dev

    ENV CGO_ENABLED=1
    RUN go test ./...

build:
    FROM +godeps

    COPY . .

    RUN apk add gcc libc-dev

    ENV CGO_ENABLED=1
    ARG --required COMMIT_SHA
    ARG --required COMMIT_REF
    RUN go build -o calendar-sync -ldflags " \
        -X 'calendar-sync/cmd.CommitSHA=${COMMIT_SHA}' \
        -X 'calendar-sync/cmd.BuildDate=$(date)' \
        -X 'calendar-sync/cmd.CommitRef=${COMMIT_REF}' \
    " .

    SAVE ARTIFACT calendar-sync AS LOCAL calendar-sync

image:
    FROM alpine:${ALPINE_VERSION}

    ARG --required COMMIT_SHA
    ARG --required COMMIT_REF

    COPY (+build/calendar-sync --COMMIT_REF=${COMMIT_REF} --COMMIT_SHA=${COMMIT_SHA}) /bin/

    RUN /bin/calendar-sync --version

release:
    FROM +image

    ENTRYPOINT /bin/calendar-sync

    ARG --required image
    SAVE IMAGE --push ${image}
