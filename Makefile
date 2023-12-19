# find indexs that are newer than cache.gob
# @qaqland 2023-12-17

MIRROR := /home/qaq/rsync/

INDEXS := $(shell find $(MIRROR) -name 'APKINDEX.tar.gz')
CACHES := $(INDEXS:%APKINDEX.tar.gz=%cache.gob)

all: $(CACHES)

# will create and delete cache.lock when processing
%cache.gob: %APKINDEX.tar.gz
	echo $< ; touch $@ 

.PHONY: clean
clean:
	find $(MIRROR) -name 'cache.*' -type f -delete
