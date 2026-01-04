ALL_TARGETS := linux-amd64 \
	linux-386 \
	linux-arm64 \
	darwin-amd64 \
	darwin-arm64 \
	windows-amd64 \
	windows-386 \
	windows-arm64

# excluded darwin/386 - not supported

LDFLAGS := -ldflags "-w -s"

default : filesurf.go 
	go build $(LDFLAGS) -o bin/filesurf $<

$(ALL_TARGETS) : filesurf.go
	GOOS=$(word 1,$(subst -, ,$@)) \
	GOARCH=$(word 2,$(subst -, ,$@)) \
	go build $(LDFLAGS) -o bin/filesurf-$@ $<

all : $(ALL_TARGETS)

.PHONY : clean
clean :
	-rm -f bin/filesurf*
