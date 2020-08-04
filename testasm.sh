#!/bin/sh
set -ex
if [ ! -f testdata/laplace.s ]; then
  curl -fsSLo testdata/laplace.s https://user.eng.umd.edu/~blj/RiSC/laplace.s
  git apply --no-index testdata/laplace.diff
fi
if [ ! -x testdata/a ]; then
  curl -fsSLo testdata/a.c https://user.eng.umd.edu/~blj/RiSC/a.c
  gcc -std=c89 -o testdata/a testdata/a.c
fi
./testdata/a ./testdata/laplace.s ./testdata/theirs.bin
go build -v ./cmd/asm
./asm -f ./testdata/laplace.s > ./testdata/ours.bin
git diff --no-index ./testdata/theirs.bin ./testdata/ours.bin
