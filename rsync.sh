#! /bin/sh
# this script only sync APKINDEX files
# @qaqland 2023-12-16

src="rsync://mirrors.tuna.tsinghua.edu.cn/alpine/"
des="/home/qaq/rsync"

# TODO: flock

rsync \
    --archive \
    --prune-empty-dirs \
    --verbose \
    --include="*/" \
    --include="edge/**/APKINDEX.tar.gz" \
    --include="v3.18/**/APKINDEX.tar.gz" \
    --include="v3.19/**/APKINDEX.tar.gz" \
    --exclude="*" \
    $src $des

# --delete \

# example make running:
# MIRROR_DIR=$des \
# MASTER_KEY=1234567890 \
# AINDEX_BIN=/home/qaq/projects/apkindex/aindex \
# make -f /home/qaq/Makefile
