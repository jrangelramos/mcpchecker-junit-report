.PHONY: build clean install

build:
	go build -o mcpchecker-junit-report

clean:
	rm -f mcpchecker-junit-report junit-report*.xml

install: build
	go install
