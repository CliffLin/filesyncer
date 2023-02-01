build:
	go build -o bin/filesyncer cmd/filesyncer/main.go
install:
	sudo cp bin/filesyncer /usr/local/bin/
