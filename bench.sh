#! /bin/bash

set -u

prev=$1
what=$2
curr=`git rev-parse --abbrev-ref HEAD`
rnd=$(head -c4 </dev/urandom|xxd -p)

bench() {
	local exp=".*"
    if [[ ! -z $2 ]]; then
    	$exp = $2
    fi
    filename=$(echo "$rnd-$1.bench" | tr "/" "_")
    if [[ -e "${filename}" ]]; then
        echo "Already exists ${filename}"
    else
        local backup=`git rev-parse --abbrev-ref HEAD`
        git checkout "$1"
        echo -n "Creating ${filename}... "
        go test ./... -run=none -benchmem -bench="$exp" > "${filename}"
        echo "OK"
        git checkout ${backup}
        sleep 5
    fi
}

bench ${prev} ${what}
bench ${curr} ${what}

benchstat "$rnd-${to}.bench" "$rnd-${current}.bench"
