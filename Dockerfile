# Build stage
FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /zuginbox ./cmd/zuginbox

# Runtime: Temurin JRE optional for Mustang PDF validation
FROM eclipse-temurin:21-jre-jammy
COPY --from=build /zuginbox /usr/local/bin/zuginbox
# Download Mustang separately, e.g. scripts/fetch-mustang.sh
ENTRYPOINT ["zuginbox"]
