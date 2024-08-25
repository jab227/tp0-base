#!/usr/bin/env sh

set -eu

echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"

if [ ! -f ./dockergen/dockergen ]; then
    set -x
    go build -C ./dockergen -o dockergen
    set +x
fi

./dockergen/dockergen -filename=$1 -clients=$2
