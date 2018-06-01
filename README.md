# echosounder

`echosounderd` is a simple REPL TCP service with a HTTP statistics service embedded.

## Usage

You may simply run it by cloning it to `$GOPATH/src/echosounder` and run

```sh
go run cmd/echosounderd/echosounderd.go --listen ':23' --stat-listen ':80'
```

**Note:** Without specifying `--listen` and `--stat-listen`, the default REPL TCP service and HTTP statistics service listen on `:10023` and `10080`.

Or run it as a docker container

```sh
docker-compose up -d
```

### REPL TCP service

```sh
$ telnet localhost 23
Trying ::1...
telnet: connect to address ::1: Connection refused
Trying fe80::1...
telnet: connect to address fe80::1: Connection refused
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
>>> dummy
        Title: sunt aut facere repellat provident occaecati excepturi optio reprehenderit
        Body: quia et suscipit
suscipit recusandae consequuntur expedita et cum
reprehenderit molestiae ut ut quas totam
nostrum rerum est autem sunt rem eveniet architecto
>>> stat

        Current connection: 1
        Total request: 1
        Processed request: 1
        Waiting request: 0
>>> quit
bye~
Connection closed by foreign host.
```

3 commands are supported: `dummy`, `stat`, `quit`.

- `dummy` sends a request to an external API service jsonplaceholder.typicode.com, gets the first post and returns it back to client.
- `stat` shows current statistics just like what you request from HTTP statistics server (but in different format).
- `quit` exits REPL and closes the connection.

### REPL HTTP service

```sh
$ curl http://localhost/stat
{"current_connection":0,"total_request":1,"processed_request":1,"waiting_request":0}
```
