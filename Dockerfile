FROM golang:1.18-alpine AS build
WORKDIR /app
ADD . /app
RUN apk add gcc musl-dev upx git
RUN echo "Starting Build" && \
    CC=$(which musl-gcc) GOOS=linux GOARCH=amd64 go build -a -tags -buildmode=pie -trimpath --ldflags '-s -w -linkmode external -extldflags "-static"' && \
    upx --best --lzma ./webex-teams-cli && \
    ./webex-teams-cli -v && \
    echo "Completed Build" 

FROM scratch

WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/webex-teams-cli /app/webex-teams-cli
COPY --from=build /app/run.sh /app/run.sh
ENV PATH="/app:${PATH}"
CMD ["/app/webex-teams-cli", "-v"] 