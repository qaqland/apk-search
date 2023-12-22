#!/usr/bin/python
# init meilisearch indexes(settings) and delete old
# @qaqland 2023-12-21

import argparse
import glob, os, sys


try:
    import meilisearch
except:
    print("Install meilisearch package of python")
    os.system("pip install meilisearch")
    import meilisearch
finally:
    from meilisearch.errors import MeilisearchError

parser = argparse.ArgumentParser(
    description="init meilisearch indexes(settings) and delete old",
    formatter_class=argparse.ArgumentDefaultsHelpFormatter,
)
parser.add_argument("path", metavar="RSYNC_PATH", help="rsync mirror files' directory")
parser.add_argument(
    "--url",
    default="http://localhost:7700",
    help="meilisearch address",
)
parser.add_argument(
    "--key", metavar="MASTER_KEY", help="meilisearch master key", required=True
)
args = parser.parse_args()

key = args.key
url = args.url
path = args.path

client = meilisearch.Client(url, key)

try:
    resp = client.get_keys()
except MeilisearchError as e:
    print(e)
    sys.exit(1)

has_key = False
for key in resp.results:
    if len(key.actions) != 1 or key.actions[0] != "search":
        continue
    if key.indexes[0] != "*":
        continue
    has_key = True
    print(f"Search Key: {key.key}")

if not has_key:
    resp = client.create_key(
        options={
            "description": "Search API Key",
            "actions": ["search"],
            "indexes": ["*"],
            "expiresAt": None,
        }
    )
    print(f"Search Key: {resp.key} (New)")

resp = client.get_indexes()
count = resp["total"]
print(f"Number of Existing Indexes: {count}")
has_indexes = []
if count != 0:
    for index in resp["results"]:
        has_indexes.append(index.uid)

indexes = glob.glob("*/community/*/APKINDEX.tar.gz", root_dir=path)
if len(indexes) == 0:
    print(
        f'Nothing has been found in "{path}", make sure the path is like "HERE"/v3.18/main/x86_64/APKINDEX..'
    )
    sys.exit(1)

settings = {
    "displayedAttributes": [
        "package",
        "version",
        "description",
        "origin",
        "build_time",
    ],
    "distinctAttribute": None,
    "faceting": {"maxValuesPerFacet": 100},
    "filterableAttributes": ["build_time", "maintainer", "repository", "id"],
    "pagination": {"maxTotalHits": 1000},
    "rankingRules": ["words", "typo", "attribute", "proximity", "sort", "exactness"],
    "searchableAttributes": ["package", "provides", "description"],
    "sortableAttributes": ["build_time"],
    "stopWords": [],
    "synonyms": {},
    "typoTolerance": {
        "disableOnAttributes": [],
        "disableOnWords": [],
        "enabled": True,
        "minWordSizeForTypos": {"oneTypo": 5, "twoTypos": 9},
    },
}

for ipath in indexes:
    parts = ipath.split("/")
    index = f'{parts[-4].replace(".", "_")}_{parts[-2]}'
    if index in has_indexes:
        has_indexes.remove(index)
    task = client.index(index).update_settings(settings)
    try:
        client.wait_for_task(task.task_uid, timeout_in_ms=50000, interval_in_ms=100)
    except MeilisearchError as e:
        print(e)
        sys.exit(1)
    else:
        print(f"{index:>16}: succeeded")

for index in has_indexes:
    print(f"Remove Unnecessary Index: {index}")
    client.delete_index(index)

print("Initialize the Search Indexes Done!")
