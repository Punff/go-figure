#! /bin/bash

detect_pkg_manager() {
    if command -v apt >/dev/null 2>&1; then
        echo "apt"
    elif command -v dnf >/dev/null 2>&1; then
        echo "dnf"
    elif command -v yum >/dev/null 2>&1; then
        echo "yum"
    elif command -v pacman >/dev/null 2>&1; then
        echo "pacman"
    elif command -v apk >/dev/null 2>&1; then
        echo "apk"
    elif command -v zypper >/dev/null 2>&1; then
        echo "zypper"
    elif command -v emerge >/dev/null 2>&1; then
        echo "portage"
    elif command -v xbps-install >/dev/null 2>&1; then
        echo "xbps"
    elif command -v slackpkg >/dev/null 2>&1; then
        echo "slackpkg"
    else
        echo "unknown"
    fi
}

detect_pkg_manager
