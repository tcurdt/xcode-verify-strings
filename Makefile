NAME=xcode-verify-strings
VERSION=v0.0.1

BUILDS=\
  darwin-arm64  \
  # darwin-amd64  \
  # linux-arm64   \
  # windows-amd64 \
  # linux-386     \
  # linux-amd64   \
  # linux-arm     \
  # windows-386   \

MAKEFLAGS += --silent

DISTS_PATTERN=dist/$(NAME)-$(VERSION)-%.tgz
DISTS=$(BUILDS:%=$(DISTS_PATTERN))

BINS_PATTERN=bin/%/$(NAME)
BINS=$(BUILDS:%=$(BINS_PATTERN))

.PHONY:
all: $(BINS)

$(BINS): OS = $(word 1,$(subst -, ,$*))
$(BINS): ARCH = $(word 2,$(subst -, ,$*))
$(BINS): $(BINS_PATTERN):
	echo "building bin: [$(OS) $(ARCH)]"
	mkdir -p `dirname $@`
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $@
	touch $@

$(DISTS): OS = $(word 1,$(subst -, ,$*))
$(DISTS): ARCH = $(word 2,$(subst -, ,$*))
$(DISTS): $(DISTS_PATTERN): $(BINS)
	echo "building dist: [$(OS) $(ARCH)]"
	mkdir -p dist
	cd bin/$(OS)-$(ARCH) && tar czf ../../$@ .
	touch $@

.PHONY: clean
clean:
	rm -rf dist bin

.PHONY: test
test: all
	go test

.PHONY: run
run:
	go run $(shell find . -name "*.go" -and -not -name "*_test.go")

.PHONY: install
install: $(BINS)
	cp "bin/darwin-amd64/$(NAME)" "$(HOME)/.bin/$(NAME)"

.PHONY: release
release: $(DISTS)
	git tag -a $(VERSION) -m "releasing $(VERSION)"
	git push --tags origin master
