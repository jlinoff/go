# Test the utilities.
all: bin/tester
	bin/tester

clean:
	find . -type f -name '*~' -delete
	rm -rf src bin pkg

bin/tester:	src/tester/tester.go \
		src/github.com/jlinoff/go/msg/msg.go \
		src/github.com/jlinoff/go/run/run.go
	GOPATH=$$(pwd) ; \
	cd src/tester ; \
	go install

src/github.com/jlinoff/go/msg/msg.go:
	GOPATH=$$(pwd) go get github.com/jlinoff/go/msg

src/github.com/jlinoff/go/run/run.go:
	GOPATH=$$(pwd) go get github.com/jlinoff/go/run

src/tester/tester.go: tester.go
	@[ ! -d src/tester ] && mkdir -p src/tester || true
	@rm -f $@
	cp $< $@
