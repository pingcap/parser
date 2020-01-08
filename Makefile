.PHONY: all parser clean

all: fmt parser

test: fmt parser
	sh test.sh

parser: parser.go hintparser.go

%arser.go: prefix = $(@:parser.go=)
%arser.go: %arser.y bin/goyacc
	@echo "bin/goyacc -o $@ -p yy$(prefix) -t $(prefix)Parser $<"
	@bin/goyacc -o $@ -p yy$(prefix) -t $(prefix)Parser $< || ( rm -f $@ && echo 'Please check y.output for more information' && exit 1 )
	@rm -f y.output

bin/goyacc: goyacc/main.go
	GO111MODULE=on go build -o bin/goyacc goyacc/main.go

fmt: bin/goyacc
	@echo "gofmt (simplify)"
	@gofmt -s -l -w . 2>&1 | awk '{print} END{if(NR>0) {exit 1}}'

clean:
	go clean -i ./...
	rm -rf *.out
	rm -f parser.go hintparser.go
