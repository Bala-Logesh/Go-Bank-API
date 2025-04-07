# A Banking API created using Go

A RESTful API build using Golang without frameworks for a mock Banking service

## Packages used
- gorilla/mux
- godotenv
- pq
- x/crypto
- golang-jwt

## Developer mode

Enable hot reloading with [Air](https://github.com/air-verse/air)

```
export GOBIN=$(go env GOPATH)/bin
$GOBIN/air
```

## Production mode

```
make build
```

```
make run
```

# Instructions to setup Postgres on Docker

```
docker pull postgres
```

```
docker run -d --name postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=postgres -p 5432:5432 postgres
```