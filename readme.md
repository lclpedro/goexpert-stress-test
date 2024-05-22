## Application Stress Test

#### Como executar..
1. Execução via aplicação
```shell
$ go run main.go --url=http://google.com --requests=1000 --concurrency=10
```

2Execução via docker

```shell
$ docker run lclpedro/stress-test-attack "./server" --url=http://google.com --requests=1000 --concurrency=10
```