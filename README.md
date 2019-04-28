A simple utility for downloading Forge installers

## Usage

```
./forge-downloader MC_VERSION
```

where `MC_VERSION` is a standard Minecraft version such as 1.13.2

## Build

Clone this repo outside of `$GOPATH` and using Go 1.11 or newer:

```
go build
```

## Release snapshot

```
goreleaser --snapshot --rm-dist
```
