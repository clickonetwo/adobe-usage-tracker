## Build the caddy webserver with the latest adobe_usage_tracker
##
## For a multi-platform build, do:
##
## docker build --platform linux/amd64,linux/arm64 -t clickonetwo/adobe_usage_tracker .

FROM caddy:2.8.1-builder AS builder

RUN xcaddy build \
    --with github.com/clickonetwo/tracker@v1.0.0

FROM caddy:2.8.1

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
