import json
import urllib.request
import tarfile

packages = []
branch = ["edge"]
repository = ["main", "community", "testing"]
# architecture default x86_64

# end without "/"
mirror = "https://mirrors.tuna.tsinghua.edu.cn/alpine"

# https://mirrors.tuna.tsinghua.edu.cn/alpine/edge/main/x86_64/APKINDEX.tar.gz

def parser_apkindex(data, bra, repo):
    package = {}
    for line in data.split('\n'):
        if line == '':
            if package:
                package["id"] = f"{bra}-{repo}-" + package["P"]
                packages.append(package)
                package = {}
        else:
            key, value = line.split(':', 1)
            package[key] = value.strip()

for bra in branch:
    for repo in repository:
        file = f"{bra}_{repo}_APKINDEX.tar.gz"
        urllib.request.urlretrieve(
            f"{mirror}/{bra}/{repo}/x86_64/APKINDEX.tar.gz", file
        )
        with tarfile.open(file, "r:gz") as tar:
            tar.extractall()
        with open('APKINDEX', 'r', encoding="utf-8") as input_file:
            data = input_file.read()
            parser_apkindex(data, bra, repo)

print("all:", len(packages))

with open('data.json', 'w') as jsonfile:
    json.dump(packages, jsonfile)
