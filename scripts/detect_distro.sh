#!/bin/bash

detect_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        echo "$ID"
        return
    fi

    if command -v lsb_release >/dev/null 2>&1; then
        lsb_release -si
        return
    fi

    if [ -f /etc/debian_version ]; then
        echo "debian"
    elif [ -f /etc/redhat-release ]; then
        echo "redhat"
    elif [ -f /etc/arch-release ]; then
        echo "arch"
    elif [ -f /etc/alpine-release ]; then
        echo "alpine"
    elif [ -f /etc/gentoo-release ]; then
        echo "gentoo"
    elif [ -f /etc/slackware-version ]; then
        echo "slackware"
    elif [ -f /etc/SuSE-release ]; then
        echo "suse"
    elif [ -f /etc/mandriva-release ]; then
        echo "mandriva"
    elif [ -f /etc/fedora-release ]; then
        echo "fedora"
    elif [ -f /etc/centos-release ]; then
        echo "centos"
    elif [ -f /etc/void-release ]; then
        echo "void"
    else
        echo "unknown"
    fi
}

detect_distro
