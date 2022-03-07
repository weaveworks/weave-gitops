# shellcheck shell=bash
if ! command -v mkcert &> /dev/null
then
    echo "mkcert is not installed, consider following this instructions: https://github.com/FiloSottile/mkcert#installation "
    exit
else
    mkcert -install
    mkcert localhost
fi