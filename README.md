# Apk-Search

alpine linux package search online

## Setup Alpine Linux Mirror

cron & rsync

## Build Parser AINDEX

```
$ go build -o aindex main.go
```

## Init Settings in Meilisearch

pip install `meilisearch` first

```
$ python ./init-search-index.py --help
```

It delete old indexs and create new based on rsync file tree.

Move `indexes.json` to html's public dir and change KEY in `html/src/Key.jsx`

## Update Indexs in MeiliSearch

Look into `Makefile` first.

```
$ make -j$(nproc)
```
