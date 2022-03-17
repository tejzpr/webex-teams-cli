FROM golang:1.18-alpine AS build
WORKDIR /app
ADD . /app
RUN apk add gcc musl-dev upx git
RUN echo "Starting Build" && \
    CC=$(which musl-gcc) GOOS=linux GOARCH=amd64 go build -a -tags -buildmode=pie -trimpath --ldflags '-s -w -linkmode external -extldflags "-static"' && \
    upx --best --lzma ./webex-teams-cli && \
    ./webex-teams-cli -v && \
    echo "Completed Build" 

RUN mkdir -p /dist/app && mkdir -p /dist/etc/ssl/certs/ && \
    mv /etc/ssl/certs/ca-certificates.crt /dist/etc/ssl/certs/ && \
    mv /app/webex-teams-cli /dist/app/webex-teams-cli && \
    mv /app/run.sh /dist/app/run.sh

FROM scratch
COPY --from=build /dist/ /
WORKDIR /app
ENV PATH="/app:${PATH}"
CMD ["/app/webex-teams-cli", "-v"] 