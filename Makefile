gen: clean
	go generate

go clean:
	rm -f *.gen.go