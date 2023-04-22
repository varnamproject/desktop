#!/bin/sh

export varnamDir=$(cd "$(dirname "$0")" && pwd)
export LD_LIBRARY_PATH="$varnamDir:$LD_LIBRARY_PATH"

cd $varnamDir
./varnam --config config.toml
