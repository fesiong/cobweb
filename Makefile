build: clean
	mkdir -p -v ./bin
	GOOS=linux GOARCH=amd64 go build -o ./bin/spider cobweb/app

clean:
	rm -rf ./bin