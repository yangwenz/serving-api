# serving-api

The backend for ML model APIs.

## Get-Started
Install dependencies:
```shell
go mod tidy
```

Install mockgen:
```shell
go install go.uber.org/mock/mockgen@v0.2.0
go get go.uber.org/mock/mockgen/model@v0.2.0
```

Run tests:
```shell
make test
```

Run on local machine:
```shell
make server
```
