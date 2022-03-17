FROM golang:1.16-alpine AS build
WORKDIR /app
ADD . /app
RUN apk add build-base
RUN echo "Starting Build" && \
    CC=$(which musl-gcc) go build -buildmode=pie -trimpath --ldflags '-s -w -linkmode external -extldflags "-static"' && \
    echo "Completed Build" 

FROM scratch

WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/webex-teams-cli /app/webex-teams-cli
COPY --from=build /app/run.sh /app/run.sh
ENV PATH="/app:${PATH}"
CMD ["/app/webex-teams-cli", "-v"] 