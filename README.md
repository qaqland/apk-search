# apk search json

Example package: `aaudit`

Offical package details: <https://pkgs.alpinelinux.org/package/edge/main/x86_64/aaudit>

After run `py apk.py` we can get one json as below

```json
[
  {
    "C": "Q1BawafBgSS1e6GBSZxoKehEolr1A=",
    "P": "aaudit",
    "V": "0.7.2-r3",
    "A": "x86_64",
    "S": "3394",
    "I": "49152",
    "T": "Alpine Auditor",
    "U": "https://alpinelinux.org",
    "L": "Unknown",
    "o": "aaudit",
    "m": "Timo Ter\u00e4s <timo.teras@iki.fi>",
    "t": "1659792088",
    "c": "0714a84b7f79009ae8b96aef50216ed72f54b885",
    "D": "lua5.2 lua5.2-posix lua5.2-cjson lua5.2-pc lua5.2-socket",
    "p": "cmd:aaudit=0.7.2-r3",
    "id": "edge-main-aaudit"
  }
]
```

APKINDEX Format see offical wiki: <https://wiki.alpinelinux.org/wiki/Apk_spec>

Some fields are useless in our search, what we need are:

- P: Package
- V: Version
- A: Architecture
- S: Size
- I: Installed size
- T: Description
- U: Project
- L: License
- o: Origin
- m: Maintainer
- t: Build time
- c: Commit
- D: Depends

The last `id` is made for Meilisearch UUID
