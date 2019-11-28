jangled: $(wildcard *.go) $(wildcard util/*.go)
	CGO_ENABLED=0 go build -ldflags "-w -s"

server-test: jangled
	./jangled -staging -nopanic=false $(JANGLED_OPTS)

MONGO_PORT := 3600
MONGO_DBPATH := $(HOME)/jangle-mongodb
MONGO := mongod

launch-mongo-test:
	mkdir -p "$(MONGO_DBPATH)"
	$(MONGO) --dbpath "$(MONGO_DBPATH)" --port $(MONGO_PORT) --replSet jangle-mongodb

NODE := node
BOT_TOKEN := 42

launch-simple-bot:
	env TOKEN=$(BOT_TOKEN) $(NODE) jstests/example.js
