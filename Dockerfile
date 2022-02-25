FROM golang:1.16-alpine AS build
WORKDIR /app
ADD . /app
RUN apk add build-base
RUN echo "Starting Build" && \
    CC=$(which musl-gcc) go build -buildmode=pie -trimpath --ldflags '-w -linkmode external -extldflags "-static"' && \
    echo "Completed Build" 

FROM alpine:latest

WORKDIR /app
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=build /app/webex-teams-cli /app/webex-teams-cli
COPY --from=build /app/run.sh /app/run.sh
ENV PATH="/app:${PATH}"
CMD ["/app/webex-teams-cli", "-v"] 