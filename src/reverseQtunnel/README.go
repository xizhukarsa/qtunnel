# reverse based on qtunnel

## purpose
use to replace simple jump server

## how work be done
1s. start proxy server on host with public ip, open port for proxy client;
2. proxy client run on host with real server, connect to local service and proxy server;
3. when recive connection from proxy client, open an adtional port for user;
4. when receive user data , send to related proxy client , then to local service; 
5. local service data flow reverse link in step 4;

## how to used
* public host with ip1;
* local service with port1;
* run server like `rqtunnel -listen=:1234 -webServicePort=:1111 -crypto=	rc4 -secret=1234`;
* run client like `rqtunnel -clientmode=true -remoteAddr=ip1:1234 -localAddr=:port1 -name=localservice1 -crypto=	rc4 -secret=1234`
* list tunnel with `ip1:1111/listItem`, return `{"list":[{"name":"localservice1", "port":port2}]}`;
* connect local service with `ip1:port2`