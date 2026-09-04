FROM golang:1.25.3 AS build

WORKDIR /src
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
ARG UBOM_BUILD_ID=dev
RUN CGO_ENABLED=0 go build -ldflags "-X ubom-v4/api.BuildID=${UBOM_BUILD_ID}" -o /out/ubom ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache wget && adduser -D -u 65532 ubom
COPY --from=build /out/ubom /ubom
USER ubom
EXPOSE 8080
ENTRYPOINT ["/ubom"]
