all:
	go fmt .
	go vet .
	golint .
	staticcheck .
