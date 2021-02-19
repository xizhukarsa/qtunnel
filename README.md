# reverse-qTunnel

Reverse tunnel based on qTunnel;

## how to use

1. start listen on 9092 for tunnel client, listen on 8081 for user to connect local service:
	`server-cli -ternelAddr=":9092" -addr=":8081" -secret="secret"` 
2. connect server with `serverip:9092` and connect local service with `:8080`
	`client-cli -ternelAddr="serverip:9092" -addr=":8080" -secret="secret" -clientmode=true` 
3. when user request on `serverip:8081`, server proxy user request to local service on 8080;

## how to build
after get all dependencies, run :
`go build src/reverseQtunnel/main.go`

## use case 
* remote work debug;
* open sub network services;

## more secure
* with `-crypto=aes256cfb` use aes for cryption;
