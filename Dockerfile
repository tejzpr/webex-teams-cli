FROM golang:1.25-alpine AS build
WORKDIR /app
ADD . /app
RUN apk add gcc musl-dev git
RUN echo "Starting Build" && \
    CC=$(which musl-gcc) CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -buildmode=pie -trimpath --ldflags '-s -w -linkmode external -extldflags "-static"' && \
    ./webex-teams-cli -v && \
    mkdir -p /dist/app && mkdir -p /dist/etc/ssl/certs/ && \
    mv /etc/ssl/certs/ca-certificates.crt /dist/etc/ssl/certs/ && \
    mv /app/webex-teams-cli /dist/app/webex-teams-cli && \
    mv /app/run.sh /dist/app/run.sh && \
    echo "Completed Build" 

FROM scratch
COPY --from=build /dist/ /
ENV PATH="/app:${PATH}"
CMD ["/app/webex-teams-cli", "-v"]