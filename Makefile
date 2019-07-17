# perf-analyzer Makefile

PACKAGE=github.com/openshift-scale/perf-analyzer
DIR_SCRAPER=cmd/scraper
PACKAGE_SCRAPER=$(PACKAGE)/$(DIR_SCRAPER)
PACKAGE_SCRAPER_BIN=$(lastword $(subst /, ,$(PACKAGE_SCRAPER)))
PATH_SCRAPER=$(OUT_DIR)/$(PACKAGE_SCRAPER_BIN)
DIR_COMPARE=cmd/compare
PACKAGE_COMPARE=$(PACKAGE)/$(DIR_COMPARE)
PACKAGE_COMPARE_BIN=$(lastword $(subst /, ,$(PACKAGE_COMPARE)))
PATH_COMPARE=$(OUT_DIR)/$(PACKAGE_COMPARE_BIN)
ENVVAR=GOOS=linux CGO_ENABLED=0 GOARCH=amd64 
OUT_DIR=_output
ASSETS_COMPARE=$(shell find pkg/result -name \*.go)
ASSETS_SCRAPER=$(shell find pkg/config -name \*.go)

.PHONY: all
all: $(PATH_SCRAPER) $(PATH_COMPARE)

$(PATH_SCRAPER): $(ASSETS_SCRAPER) ./$(DIR_SCRAPER)/main.go
	$(ENVVAR) go build -o $(PATH_SCRAPER) $(PACKAGE_SCRAPER)

$(PATH_COMPARE): $(ASSETS_COMPARE) ./$(DIR_COMPARE)/main.go
	$(ENVVAR) go build -o $(PATH_COMPARE) $(PACKAGE_COMPARE)

.PHONY: clean
clean:
	rm -rf $(OUT_DIR)
