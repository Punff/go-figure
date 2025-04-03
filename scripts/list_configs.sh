#!/bin/bash

output_file="configs.txt"
> "$output_file"

cnt=0
pkg_cnt=$(wc -l < packages.txt)

while read -r pkg; do
    cnt=$((cnt+1))

    found_path=$(find /etc /var ~/.config -type d -iname "*$pkg*" -o -type f -iname "*$pkg*" 2>/dev/null | head -n 1)

    if [[ -z "$found_path" ]]; then
        for dir in /etc /etc/default /etc/sysconfig; do
            if [[ -f "$dir/$pkg" || -d "$dir/$pkg" ]]; then
                found_path="$dir/$pkg"
                break
            fi
        done
    fi

    if [[ -n "$found_path" ]]; then
        echo "$pkg: $found_path" >> "$output_file"
    else
        echo "$pkg: Not found" >> "$output_file"
    fi

    printf "\rProgress: %d/%d" "$cnt" "$pkg_cnt"
done < packages.txt
