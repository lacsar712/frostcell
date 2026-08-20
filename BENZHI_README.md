# Frostcell

冷链温区越限监测

## Build

```bash
export GOTOOLCHAIN=local
go build ./...
```

## Test

```bash
export GOTOOLCHAIN=local
go test ./... -count=1
```

## Docker (benzhi)

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh frostcell linux/amd64
./build_benzhi_docker.sh frostcell linux/arm64
```