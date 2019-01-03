#!/bin/bash -e
set -Eeo pipefail; [ -n "$DEBUG" ] && set -x; set -u

working_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

dupes="$(
    grep -irE "^  name:" "$working_dir"/*/*.yml | \
    awk '{print $NF}' | \
    sort | \
    uniq -c | \
    sort -r
)"
if [ "$(echo "$dupes" | awk '{print $1}' | sort -u | wc -l | awk '{print $NF}')" -ne 1 ]; then
    echo "All test objects do not have unique names:"
    echo "$dupes"
    exit 1
fi

function cleanup {
    for f in "$working_dir/valid"/*; do
        kubectl delete -f "$f" > /dev/null 2>&1
    done
}
trap cleanup EXIT

kubectl apply -f "$working_dir/../config/300-indicator-document.yaml" > /dev/null

failed=false

for f in $(ls "$working_dir/valid"/* | grep -v 'pending'); do
    if out="$(kubectl apply -f "$f" 2>&1)"; then
        echo "PASSED: valid/$(basename "$f")"
    else
        echo "FAILED: valid/$(basename "$f")"
        echo "$out"
        failed=true
    fi
done

for f in $(ls "$working_dir/invalid"/* | grep -v 'pending'); do
    if out="$(kubectl apply -f "$f" 2>&1)"; then
        echo "FAILED: invalid/$(basename "$f")"
        echo "$out"
        failed=true
    else
        echo "PASSED: invalid/$(basename "$f")"
    fi
done

if [ "$failed" != false ]; then
    exit 1
fi

go test "$working_dir/../..." -race -count 1
