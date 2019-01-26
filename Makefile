.PHONY: default build test benchmark update_source clean protos

default: test

enums:
	@go get -u -a golang.org/x/tools/cmd/stringer
	@stringer -type ObjectType nodes/object_type.go
	@stringer -type SortByDir nodes/sort_by_dir.go
	@stringer -type StmtType nodes/stmt_type.go

build:
	go get github.com/juju/errors
	go get github.com/kataras/go-errors
	go get github.com/kataras/golog
	go get github.com/readystock/golog
	go get github.com/stretchr/testify/assert
	go get github.com/kr/pretty
	go build

test: protos enums build
	go test -v ./ ./nodes

protos:
	protoc -I=$(PROTOS_DIRECTORY) --go_out=./nodes $(PROTOS_DIRECTORY)/context.proto

benchmark:
	go build -a
	go test -test.bench=. -test.run=XXX -test.benchtime 10s -test.benchmem -test.cpu=4
	#go test -c -o benchmark
	#GODEBUG=schedtrace=100 ./benchmark -test.bench=BenchmarkRawParseCreateTableParallel -test.run=XXX -test.benchtime 20s -test.benchmem -test.cpu=16

# --- Below only needed for releasing new versions

LIB_PG_QUERY_TAG = 10-1.0.2

root_dir := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
LIB_TMPDIR = $(root_dir)/tmp
LIBDIR = $(LIB_TMPDIR)/libpg_query
LIBDIRGZ = $(TMPDIR)/libpg_query-$(LIB_PG_QUERY_TAG).tar.gz
PROTOS_DIRECTORY = ./protos

$(LIBDIR): $(LIBDIRGZ)
	mkdir -p $(LIBDIR)
	cd $(LIB_TMPDIR); tar -xzf $(LIBDIRGZ) -C $(LIBDIR) --strip-components=1

$(LIBDIRGZ):
	mkdir -p $(LIB_TMPDIR)
	curl -o $(LIBDIRGZ) https://codeload.github.com/lfittl/libpg_query/tar.gz/$(LIB_PG_QUERY_TAG)

update_source: $(LIBDIR)
	rm -f parser/*.{c,h}
	rm -fr parser/include
	# Reduce everything down to one directory
	cp -a $(LIBDIR)/src/* parser/
	mv parser/postgres/* parser/
	rmdir parser/postgres
	cp -a $(LIBDIR)/pg_query.h parser/include
	# Make sure every .c file in the top-level directory is its own translation unit
	mv parser/*{_conds,_defs,_helper}.c parser/include
	# Other support files
	rm -fr testdata
	cp -a $(LIBDIR)/testdata testdata
	# Update nodes directory
	ruby scripts/generate_nodes.rb

clean:
	-@ $(RM) -r $(LIB_TMPDIR)
