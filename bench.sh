#! /bin/bash

rnd=$(head -c4 </dev/urandom|xxd -p)

bench() {
	local exp=".*"
    if [[ ! -z $2 ]]; then
    	$exp = $2
    fi
    filename=$(echo "$rnd-$1.bench" | tr "/" "_")
    if test -e "${filename}";
    then
        echo "Already exists ${filename}"
    else
        backup=`git rev-parse --abbrev-ref HEAD`
        git checkout "$1"
        echo -n "Creating ${filename}... "
        go test ./... -run=NONE -bench="$exp" > "${filename}" -benchmem
        echo "OK"
        git checkout ${backup}
        sleep 5
    fi
}


to=$1
current=`git rev-parse --abbrev-ref HEAD`

bench ${to} $2
bench ${current} $2

benchcmp $3 "$rnd-${to}.bench" "$rnd-${current}.bench"
