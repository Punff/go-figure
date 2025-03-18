#!/bin/bash

list_system_packages() {
    if command -v dpkg >/dev/null 2>&1; then
        dpkg --get-selections | awk '{print $1}'
    elif command -v dnf >/dev/null 2>&1; then
        dnf list installed | awk '{print $1}' | tail -n +2
    elif command -v yum >/dev/null 2>&1; then
        yum list installed | awk '{print $1}' | tail -n +2
    elif command -v pacman >/dev/null 2>&1; then
        pacman -Qq
    elif command -v apk >/dev/null 2>&1; then
        apk list --installed | awk '{print $1}'
    elif command -v zypper >/dev/null 2>&1; then
        zypper se --installed-only | awk '{print $2}' | tail -n +5
    elif command -v qlist >/dev/null 2>&1; then
        qlist -I
    elif command -v xbps-query >/dev/null 2>&1; then
        xbps-query -l | awk '{print $2}'
    elif [ -d /var/log/packages ]; then
        ls /var/log/packages/
    else
        echo "System package manager not detected."
    fi
}

list_snap_packages() {
    if command -v snap >/dev/null 2>&1; then
        snap list | awk 'NR>1 {print $1}'
    else
        echo "Snap not installed."
    fi
}

list_flatpak_packages() {
    if command -v flatpak >/dev/null 2>&1; then
        flatpak list --app | awk '{print $1}'
    else
        echo "Flatpak not installed."
    fi
}

case "$1" in
    system) list_system_packages ;;
    snap) list_snap_packages ;;
    flatpak) list_flatpak_packages ;;
    all)
        list_system_packages
        echo
        list_snap_packages
        echo
        list_flatpak_packages
        ;;
    *)
        echo "Usage: system | snap | flatpak | all "
        exit 1
        ;;
esac
echo "Usage: system | snap | flatpak | all "
exit 1
