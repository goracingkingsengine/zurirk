#!/bin/bash

branch=`git rev-parse --abbrev-ref HEAD`
release=${branch#release.}
when=`date +"%F %T"`

# make goos goarch extension
function make {
        GOOS=$1 GOARCH=$2 go build \
                -a \
                -ldflags "-X 'main.buildVersion=$release' -X 'main.buildTime=$when'" \
                -o zurichess-$release-$1-$2$3 \
                bitbucket.org/zurichess/zurichess/zurichess
}

make   linux amd64 ""
make   linux   386 ""
make windows amd64 ".exe"
make windows   386 ".exe"
