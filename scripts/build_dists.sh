#!/bin/bash

bin_dir=python/src/numerous/cli/bin

function python_platforms() {
    case "$1" in
        'linux_amd64') echo "manylinux_2_17_x86_64";;
        'linux_arm64') echo "manylinux_2_17_aarch64";;
        'windows_amd64') echo "win_amd64";;
        'windows_arm64') echo "win_arm64";;
        'darwin_amd64') echo "macosx_10_0_x86_64 macosx_11_0_x86_64 macosx_12_0_x86_64 macosx_13_0_x86_64 macosx_14_0_x86_64";;
        'darwin_arm64') echo "macosx_10_0_arm64 macosx_11_0_arm64 macosx_12_0_arm64 macosx_13_0_arm64 macosx_14_0_arm64";;
        *) >&2 echo "Unexpected platform $1"; exit 1;;
    esac
}

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

function build_platform_dist() {
    plat_names="$(python_platforms $1)"
    copy_bin $1
    for plat_name in $plat_names; do
        echo "Building wheel for $1 => $plat_name"
        python -m build --wheel "-C=--build-option=--plat-name=$plat_name"
    done
}

function build_multi_dist {
    copy_bins
    python -m build
}

function build_dists() {
    build_platform_dist "linux_amd64"
    build_platform_dist "linux_arm64"
    build_platform_dist "windows_amd64"
    build_platform_dist "windows_arm64"
    build_platform_dist "darwin_amd64"
    build_platform_dist "darwin_arm64"
    build_multi_dist
}

mkdir $bin_dir 2>/dev/null
if [[ -z "$1" ]]; then
    build_dists
elif [[ "$1" = "any" ]]; then
    build_multi_dist
else
    build_platform_dist $1
fi

