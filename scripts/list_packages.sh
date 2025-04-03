#!/bin/bash

> packages.txt
> snaps.txt
> flatpaks.txt

list_system_packages() {
    if command -v dpkg >/dev/null 2>&1; then
        apt-mark showmanual | sort >> packages.txt
    elif command -v dnf >/dev/null 2>&1; then
        dnf repoquery --userinstalled --qf "%{NAME}" | sort >> packages.txt
    elif command -v yum >/dev/null 2>&1; then
        yum list installed | awk '{print $1}' | tail -n +2 | sort >> packages.txt
    elif command -v pacman >/dev/null 2>&1; then
        pacman -Qe | awk '{print $1}' | sort >> packages.txt
    elif command -v apk >/dev/null 2>&1; then
        apk info --installed | sort >> packages.txt
    elif command -v zypper >/dev/null 2>&1; then
        zypper se --installed-only | awk '{print $2}' | tail -n +5 | sort >> packages.txt
    elif command -v qlist >/dev/null 2>&1; then
        qlist -Ive | sort >> packages.txt
    elif command -v xbps-query >/dev/null 2>&1; then
        xbps-query -m | sort >> packages.txt
    elif [ -d /var/log/packages ]; then
        ls /var/log/packages/ | cut -d'-' -f1 | sort >> packages.txt
    else
        echo "System package manager not detected."
    fi
}

list_snap_packages() {
    if command -v snap >/dev/null 2>&1; then
        snap list | awk 'NR>1 {print $1}' | sort >> snaps.txt
    else
        echo "Snap not installed."
    fi
}

list_flatpak_packages() {
    if command -v flatpak >/dev/null 2>&1; then
        flatpak list --app | awk '{print $1}' | sort >> flatpaks.txt
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
    help)
        echo "Usage: $0 {system | snap | flatpak | all}"
        exit 1
        ;;
esac

exit 0
