UNIX_ENVS := linux-amd64 linux-386 linux-arm64 darwin-amd64 darwin-arm64
WIN_ENVS  := windows-amd64 windows-386 windows-arm64

OUTDIR = bin

UNIX_TARGETS := $(UNIX_ENVS:%=filesurf-%)
WIN_TARGETS  := $(WIN_ENVS:%=filesurf-%.exe)
UNIX_TARGETS := $(UNIX_TARGETS:%=$(OUTDIR)/%)
WIN_TARGETS  := $(WIN_TARGETS:%=$(OUTDIR)/%)

LDFLAGS := -ldflags "-w -s"

TARGET_NAME := $(OUTDIR)/filesurf-$(shell go env GOOS)-$(shell go env GOARCH)

ifeq ($(OS),Windows_NT)
TARGET_NAME := $(TARGET_NAME).exe
endif

default : $(TARGET_NAME)

$(UNIX_TARGETS) $(WIN_TARGETS) : TARGET_OS = $(word 2,$(subst -, ,$@))
$(UNIX_TARGETS) $(WIN_TARGETS) : TARGET_ARCH = $(word 3,$(subst -, ,$(@:%.exe=%)))

$(UNIX_TARGETS) $(WIN_TARGETS) : filesurf.go
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build $(LDFLAGS) -o $@ $<

all : $(UNIX_TARGETS) $(WIN_TARGETS)

.PHONY : clean cleaner run
clean :
	-rm -f bin/filesurf-*

cleaner : clean
	go clean --cache

run : $(TARGET_NAME)
	./$<
