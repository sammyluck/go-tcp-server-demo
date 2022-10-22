all: server client

server: cmd/server/main.go
	go build github.com/sammyluck/tcp-server-demo4-with-syncpool/cmd/server
client: cmd/client/main.go
	go build github.com/sammyluck/tcp-server-demo4-with-syncpool/cmd/client

clean:
	rm -fr ./server
	rm -fr ./client