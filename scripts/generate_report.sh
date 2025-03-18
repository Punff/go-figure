#!/bin/bash

detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS_NAME=$NAME
    elif command -v lsb_release >/dev/null 2>&1; then
        OS_NAME=$(lsb_release -si)
    else
        OS_NAME="Unknown"
    fi
    echo "$OS_NAME"
}

detect_package_manager() {
    if command -v apt >/dev/null 2>&1; then
        echo "apt"
    elif command -v dnf >/dev/null 2>&1; then
        echo "dnf"
    elif command -v pacman >/dev/null 2>&1; then
        echo "pacman"
    elif command -v yum >/dev/null 2>&1; then
        echo "yum"
    elif command -v zypper >/dev/null 2>&1; then
        echo "zypper"
    elif command -v apk >/dev/null 2>&1; then
        echo "apk"
    elif command -v snap >/dev/null 2>&1; then
        echo "snap"
    elif command -v flatpak >/dev/null 2>&1; then
        echo "flatpak"
    else
        echo "Unknown"
    fi
}

generate_json() {
    OS=$(detect_os)
    PACKAGE_MANAGER=$(detect_package_manager)

    echo "{" > packages.json
    echo "\"os\": \"$OS\"," >> packages.json
    echo "\"package_manager\": \"$PACKAGE_MANAGER\"," >> packages.json

    echo '"system": [' >> packages.json
    if command -v dpkg >/dev/null 2>&1; then
        dpkg --get-selections | awk '{print "\"" $1 "\""}' | paste -sd, >> packages.json
    elif command -v dnf >/dev/null 2>&1; then
        dnf list installed | awk '{print "\"" $1 "\""}' | tail -n +2 | paste -sd, >> packages.json
    elif command -v yum >/dev/null 2>&1; then
        yum list installed | awk '{print "\"" $1 "\""}' | tail -n +2 | paste -sd, >> packages.json
    elif command -v pacman >/dev/null 2>&1; then
        pacman -Qq | awk '{print "\"" $1 "\""}' | paste -sd, >> packages.json
    elif command -v apk >/dev/null 2>&1; then
        apk list --installed | awk '{print "\"" $1 "\""}' | paste -sd, >> packages.json
    elif command -v zypper >/dev/null 2>&1; then
        zypper se --installed-only | awk '{print "\"" $2 "\""}' | tail -n +5 | paste -sd, >> packages.json
    elif command -v qlist >/dev/null 2>&1; then
        qlist -I | awk '{print "\"" $1 "\""}' | paste -sd, >> packages.json
    elif command -v xbps-query >/dev/null 2>&1; then
        xbps-query -l | awk '{print "\"" $2 "\""}' | paste -sd, >> packages.json
    elif [ -d /var/log/packages ]; then
        ls /var/log/packages/ | awk '{print "\"" $1 "\""}' | paste -sd, >> packages.json
    fi
    echo "]," >> packages.json

    # List Snap Packages
    echo '"snap": [' >> packages.json
    if command -v snap >/dev/null 2>&1; then
        snap list | awk 'NR>1 {print "\"" $1 "\""}' | paste -sd, >> packages.json
    fi
    echo "]," >> packages.json

    # List Flatpak Packages
    echo '"flatpak": [' >> packages.json
    if command -v flatpak >/dev/null 2>&1; then
        flatpak list --app | awk '{print "\"" $1 "\""}' | paste -sd, >> packages.json
    fi
    echo "]" >> packages.json

    # Close JSON structure
    echo "}" >> packages.json
}

generate_json
