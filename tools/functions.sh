#!/usr/bin/env bash
# shellcheck shell=bash

log() {
    echo "•" "$@"
}

error() {
    log "error:" "$@"
    exit 1
}

title_case() {
    local param="${1}"
    local titled="$(tr '[:lower:]' '[:upper:]' <<<${param:0:1})${param:1}"
    echo $titled
}

os() {
    uname -s
}

goos() {
    local os
    os="$(uname -s)"
    case "${os}" in
    Linux*)
        echo linux
        ;;
    Darwin*)
        echo darwin
        ;;
    *)
        error "unknown OS: ${os}"
        ;;
    esac
}

arch() {
    uname -m
}

goarch() {
    local arch
    arch="$(uname -m)"
    case "${arch}" in
    armv5*)
        echo "armv5"
        ;;
    armv6*)
        echo "armv6"
        ;;
    armv7*)
        echo "armv7"
        ;;
    aarch64)
        echo "arm64"
        ;;
    arm64)
        echo "amd64"
        ;;
    x86)
        echo "386"
        ;;
    x86_64)
        echo "amd64"
        ;;
    i686)
        echo "386"
        ;;
    i386)
        echo "386"
        ;;
    *)
        error "uknown arch: ${arch}"
        ;;
    esac
}

mktempdir() {
    mktemp -d 2>/dev/null || mktemp -d -t 'wego'
}

do_curl() {
    local path="${1}"
    local url="${2}"

    log "Downloading ${url}"
    curl --progress-bar -fLo "${path}" "${url}"
}

do_curl_binary() {
    local cmd="${1}"
    local url="${2}"
    local default_path="${HOME}/.wego/bin"
    local path="${3:-${default_path}}"/"${cmd}"

    do_curl "${path}" "${url}"
    chmod +x "${path}"
}

do_curl_tarball() {
    local cmd="${1}"
    local url="${2}"
    local default_path="${HOME}/.wego/bin"
    local path="${3:-${default_path}}"/"${cmd}"

    dldir="$(mktempdir)"
    mkdir "${dldir}/${cmd}"
    do_curl "${dldir}/${cmd}.tar.gz" "${url}"
    mv "${dldir}/${cmd}.tar.gz" "${path}"
    rm -rf "${dldir}"
}

do_curl_tarball_with_path() {
    local cmd="${1}"
    local url_and_path=(${2//;/ })
    local default_path="${HOME}/.wego/bin"
    local path="${3:-${default_path}}"

    dldir="$(mktempdir)"
    mkdir -p "${dldir}/${cmd}/${url_and_path[1]}"
    do_curl "${dldir}/${cmd}.tar.gz" "${url_and_path[0]}"
    tar -C "${dldir}/${cmd}" -xvf "${dldir}/${cmd}.tar.gz"
    mv "${dldir}/${cmd}/${url_and_path[1]}" "${path}/${cmd}"
}

do_curl_txt() {
    local cmd="${1}"
    local url="${2}"
    local default_path="${HOME}/.wego/bin"
    local path="${3:-${default_path}}"/"${cmd}"

    dldir="$(mktempdir)"
    mkdir "${dldir}/${cmd}"
    do_curl "${dldir}/${cmd}.txt" "${url}"
    mv "${dldir}/${cmd}.txt" "${path}.txt"
    rm -rf "${dldir}"
}

