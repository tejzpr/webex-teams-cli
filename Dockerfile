FROM golang:1.26.0-alpine3.23 AS build
WORKDIR /app
ADD . /app
RUN apk add git
RUN echo "Starting Build" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -trimpath -ldflags='-s -w' && \
    ./webex-teams-cli -v && \
    mkdir -p /dist/app && mkdir -p /dist/etc/ssl/certs/ && \
    mv /etc/ssl/certs/ca-certificates.crt /dist/etc/ssl/certs/ && \
    mv /app/webex-teams-cli /dist/app/webex-teams-cli && \
    mv /app/run.sh /dist/app/run.sh && \
    echo "Completed Build" 

FROM scratch
COPY --from=build /dist/ /
ENV PATH="/app:${PATH}"
ENTRYPOINT ["/app/webex-teams-cli"]
CMD ["-v"]