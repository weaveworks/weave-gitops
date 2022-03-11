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
    # shellcheck disable=SC2155,SC2086
    local titled="$(tr '[:lower:]' '[:upper:]' <<<${param:0:1})${param:1}"
    echo "$titled"
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
    local checksum_path=""

    dldir="$(mktempdir)"
    mkdir "${dldir}/${cmd}"
    do_curl "${dldir}/${cmd}.tar.gz" "${url}"
    ## if checksum_path is set validate tarball
    if [ -n "${checksum_path}" ]
    then
        ## need to validate file here because unpacking the tar changes the hash
        validate_file "${checksum_path}" "${dldir}/${cmd}.tar.gz" "${cmd}"
    fi
    tar -C "${dldir}/${cmd}" -xvf "${dldir}/${cmd}.tar.gz"
    mv "${dldir}/${cmd}/${cmd}" "${path}"
    rm -rf "${dldir}"
}

do_curl_tarball_with_path() {
    local cmd="${1}"
    # shellcheck disable=SC2206
    local url_and_path=(${2//;/ })
    local default_path="${HOME}/.wego/bin"
    local path="${3:-${default_path}}"

    dldir="$(mktempdir)"
    mkdir -p "${dldir}/${cmd}/${url_and_path[1]}"
    do_curl "${dldir}/${cmd}.tar.gz" "${url_and_path[0]}"
    tar -C "${dldir}/${cmd}" -xvf "${dldir}/${cmd}.tar.gz"
    mv "${dldir}/${cmd}/${url_and_path[1]}" "${path}/${cmd}"
}

validate_file() {
    local checksums_path="${1}"
    local file_path="${2}"
    local cmd="${3}"
    # shellcheck disable=SC2155
    local digest=$(openssl dgst -sha256 "${file_path}" | cut -d ' ' -f 2)
    if ! grep "${digest}" "${checksums_path}"; then
        echo "${cmd} is not a valid file"
        rm -rf "${file_path}"
        exit 1
    fi
}
