# syntax=docker/dockerfile:1

# The web and Go stages run on the build machine and cross-compile, so a multi-arch build emulates only apt.
FROM --platform=$BUILDPLATFORM node:26-slim AS web
WORKDIR /src/web
COPY web/package.json web/pnpm-lock.yaml ./
# Node 25 dropped corepack, so pnpm is installed at the version package.json pins.
RUN npm install --global "$(node -p "require('./package.json').packageManager")" \
    && pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

FROM --platform=$BUILDPLATFORM golang:1.27 AS server
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/dist ./web/dist
ARG VERSION=docker
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/udemy-studio ./cmd/udemy-studio

FROM debian:trixie-slim
RUN apt-get update \
    && apt-get install --yes --no-install-recommends ca-certificates ffmpeg aria2 yt-dlp \
    && rm -rf /var/lib/apt/lists/*
RUN useradd --uid 1000 --create-home studio \
    && mkdir /library /data \
    && chown studio:studio /library /data
COPY --from=server /out/udemy-studio /usr/local/bin/udemy-studio
USER studio
VOLUME ["/data"]
EXPOSE 8765
ENTRYPOINT ["udemy-studio"]
CMD ["serve", "--addr", "0.0.0.0:8765", "--library", "/library", "--data-dir", "/data"]
