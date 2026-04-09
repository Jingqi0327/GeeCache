lsof -ti:8003 | xargs -r kill -9
lsof -ti:8002 | xargs -r kill -9
lsof -ti:8001 | xargs -r kill -9
lsof -ti:9999 | xargs -r kill -9
lsof -ti:8000 | xargs -r kill -9