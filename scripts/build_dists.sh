#!/bin/bash

bin_dir=python/src/numerous/cli/bin

function clean_bins() {
    rm -f $bin_dir/*
}

function copy_bin() {
    clean_bins
    cp build/$1 $bin_dir/cli
}

function copy_bins() {
    clean_bins
    cp build/darwin_amd64 build/darwin_arm64 build/linux_amd64 build/linux_arm64 build/windows_amd64 build/windows_arm64 $bin_dir
}

function build_wheels() {
    mkdir $bin_dir || echo "Already exists"

    copy_bin "linux_amd64"
    python -m build --wheel "-C=--build-option=--plat-name=manylinux_2_17_x86_64"

    copy_bin "linux_arm64"
    python -m build --wheel "-C=--build-option=--plat-name=manylinux_2_17_aarch64"

    copy_bin "windows_amd64"
    python -m build --wheel "-C=--build-option=--plat-name=win_amd64"

    copy_bin "windows_arm64"
    python -m build --wheel "-C=--build-option=--plat-name=win_arm64"

    copy_bin "darwin_amd64"
    python -m build --wheel "-C=--build-option=--plat-name=macosx_10_9_x86_64"

    copy_bin "darwin_arm64"
    python -m build --wheel "-C=--build-option=--plat-name=macosx_10_9_arm64"

    copy_bins
    python -m build
}

build_wheels
