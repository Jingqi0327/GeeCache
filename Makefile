server1:
	./server -port=8001 -api=1 -geeRegistryServer=1

server2:
	./server -port=8002 

server3:
	./server -port=8003

build:
	rm server 
	go build -o server

.PHONY: server1 server2 server3 build