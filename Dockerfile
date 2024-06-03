## Build the caddy webserver with the latest adobe_usage_tracker
##
## for a multi-platform build, do:
##
## docker build --platform linux/amd64,linux/arm64 -t clickonetwo/adobe_usage_tracker:latest .

FROM caddy:2.8.1-builder AS builder

RUN xcaddy build \
    --with github.com/clickonetwo/tracker@latest

FROM caddy:2.8.1

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
