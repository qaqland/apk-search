# find indexs that are newer than cache.gob
# @qaqland 2023-12-17

MIRROR_DIR ?= "/home/qaq/rsync/"
MASTER_KEY ?= "1234567890"
SEARCH_URL ?= "http://127.0.0.1:7700"
AINDEX_BIN ?= "/home/qaq/projects/apkindex/aindex"

INDEXS := $(shell find $(MIRROR_DIR) -name 'APKINDEX.tar.gz')
CACHES := $(INDEXS:%APKINDEX.tar.gz=%cache.gob)

all: $(CACHES)

# sleep is needed on low performance server
%cache.gob: %APKINDEX.tar.gz
	@$(AINDEX_BIN) -path $< -key $(MASTER_KEY) -url $(SEARCH_URL); sleep 60

.PHONY: clean
clean:
	find $(MIRROR_DIR) -name 'cache.*' -type f -delete
