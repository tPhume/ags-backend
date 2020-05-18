ARG GO_VERSION=1.13.6

# First Stage
FROM golang:${GO_VERSION}-alpine AS dev

# Install Git
RUN apk add --update git

# Set Go build Env
ENV GO111MODULE="on" \
    CGO_ENABLED=0 \
    GOOS=linux

# Set Application level Env
ENV APP_PATH="/backend"

WORKDIR ${APP_PATH}

# Copy and cache Go Modules
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy rest of the code
COPY . .

# Build binaries
RUN go build -ldflags="-s -w" -o main ${APP_PATH}/cmd/main/main.go
RUN chmod +x main

# Second Stage
FROM alpine AS prod

# Set application level Env
ENV APP_PATH="/backend"

WORKDIR ${APP_PATH}

# Copy binaries from dev stage
COPY --from=dev ${APP_PATH}/main main

ENTRYPOINT ["/backend/main"]
