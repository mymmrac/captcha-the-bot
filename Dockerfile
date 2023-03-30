FROM golang:1.20-alpine AS source

WORKDIR /captcha-the-bot

RUN go env -w CGO_ENABLED="0"

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

FROM source AS dev

RUN go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install github.com/cespare/reflex@latest

CMD reflex --decoration="none" -R "bin/" -s -- sh -c \
    "dlv debug --output ./bin/captcha-the-bot --headless --continue --accept-multiclient --listen :40000 --api-version=2 --log ./"

FROM source AS build

RUN apk --update add ca-certificates upx && update-ca-certificates

RUN go build -ldflags="-s -w" -o /bin/captcha-the-bot . && upx --best --lzma /bin/captcha-the-bot

FROM scratch AS release

COPY --from=mymmrac/mini-health:latest /mini-health /mini-health
HEALTHCHECK CMD ["/mini-health", "-e", "CAPTCHA_THE_BOT_LISTEN_URL", "/health"]

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/captcha-the-bot /captcha-the-bot

ENTRYPOINT ["/captcha-the-bot"]
