#! /bin/bash

bench() {
    if [[ -z $2 ]]; then
    	$2 = ".*"
    fi
    filename="/tmp/$1.bench"
    if test -e "${filename}";
    then
        echo "Already exists ${filename}"
    else
        backup=`git rev-parse --abbrev-ref HEAD`
        git checkout $1
        echo -n "Creating ${filename}... "
        go test ./... -run=NONE -bench=$2 > "${filename}" -benchmem
        echo "OK"
        git checkout ${backup}
        sleep 5
    fi
}


to=$1
current=`git rev-parse --abbrev-ref HEAD`

bench ${to} $2
bench ${current} $2

benchcmp $3 "/tmp/${to}.bench" "/tmp/${current}.bench"
