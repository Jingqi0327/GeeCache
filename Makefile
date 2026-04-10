server1:
	./server -port=8001 

server2:
	./server -port=8002 

server3:
	./server -port=8003 -api=1 -geeRegistryServer=1

build:
	rm server 
	go build -o server

stop:
	sh stop.sh

tom:
	curl "http://localhost:9999/api?key=Tom"

tom3:
	sh tom3.sh

jack:
	curl "http://localhost:9999/api?key=Jack"

sam:
	curl "http://localhost:9999/api?key=Sam"

.PHONY: server1 server2 server3 build stop tom tom3 jack sam