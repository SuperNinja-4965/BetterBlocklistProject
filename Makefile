.PHONY: all
all: clean BlocklistManager run

run: BlocklistManager
	./BlocklistManager

.PHONY: BlocklistManager
BlocklistManager:
	go mod download
	go -C ./src/cmd/BlocklistManager build -ldflags="-extldflags=-static" -o ../../../BlocklistManager

.PHONY: clean
clean:
	if [ -f BlocklistManager ]; then rm BlocklistManager; fi
