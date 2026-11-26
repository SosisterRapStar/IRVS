#!/bin/sh
set -eu

if [ "$#" -ne 1 ] || [ -z "$1" ]; then
    printf 'Использование: %s <kubeconfig>\n' "${0##*/}" >&2
    exit 2
fi
if [ ! -r "$1" ]; then
    printf 'Не удалось прочитать kubeconfig: %s\n' "$1" >&2
    exit 2
fi

: "${DB_USER:?Укажите DB_USER в env-файле}"
: "${DB_NAME:?Укажите DB_NAME в env-файле}"
: "${DB_PASSWORD:?Укажите DB_PASSWORD в env-файле}"

umask 077
secret_dir=$(mktemp -d)
trap 'rm -r -- "$secret_dir"' EXIT
trap 'exit 1' HUP INT TERM

printf '%s' "$DB_USER" > "$secret_dir/POSTGRES_USER"
printf '%s' "$DB_NAME" > "$secret_dir/POSTGRES_DB"
printf '%s' "$DB_PASSWORD" > "$secret_dir/POSTGRES_PASSWORD"

set -- "--kubeconfig=$1" --namespace=sositer-rap-star
secret_manifest=$(kubectl "$@" create secret generic postgres-auth \
    --from-file="$secret_dir" --dry-run=client -o yaml)

printf '%s\n' "$secret_manifest" |
    kubectl "$@" apply --server-side --field-manager=k8s-secrets --force-conflicts -f -
