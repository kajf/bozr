#!/bin/sh

# Heavily modified from https://gist.github.com/sjparkinson/327dc78c60ab81a06c946630b4288910

help() {
    cat <<'EOF'
Install a binary release of Bozr

Usage:
    install.sh [options]

Options:
    -h, --help      Display this message
    --tag TAG       Tag (version) of the binName to install (default <latest release>)
    --to LOCATION   Where to install the binary (default /usr/local/bin)
EOF
}

say() {
    echo "install.sh: $1"
}

say_err() {
    say "$1" >&2
}

err() {
    if [ -n "$td" ]; then
        rm -rf "$td"
    fi

    say_err "ERROR $1"
    exit 1
}

need() {
    if ! command -v "$1" > /dev/null 2>&1; then
        err "need $1 (command not found)"
    fi
}

binName="bozr"
git="kajf/bozr"
version="0.9.2"
tag="v$version"

while test $# -gt 0; do
    case $1 in
        --help | -h)
            help
            exit 0
            ;;
        --tag)
            tag=$2
            shift
            ;;
        --to)
            dest=$2
            shift
            ;;
        *)
            ;;
    esac
    shift
done

# Dependencies
need basename
need curl
need install
need mkdir
need mktemp
need tar

if [ -z "$git" ]; then
    # shellcheck disable=SC2016
    err 'must specify a git repository using `--git`. Example: `install.sh --git bozr/cross`'
fi

url="https://github.com/$git"

if [ "$(curl --head --write-out "%{http_code}\n" --silent --output /dev/null "$url")" -eq "404" ]; then
  err "GitHub repository $git does not exist"
fi

say_err "GitHub repository: $url"

if [ -z "$binName" ]; then
    binName=$(echo "$git" | cut -d'/' -f2)
fi

say_err "Binary: $binName"

if [ -z "$dest" ]; then
    dest="/usr/local/bin"
fi

if [ -e "$dest/$binName" ] && [ $force = false ]; then
    err "$binName already exists in $dest, use --force to overwrite the existing binary"
fi

url="$url/releases"

platform="$(uname -s)"

case "$(uname -s)" in
"Darwin")
  platform="darwin"
  ;;
"Linux")
  platform="linux"
  ;;
esac

arch="$(uname -m)"

url="$url/download/$tag/${binName}_${version}_${platform}_${arch}.tar.gz"

say_err "Downloading: $url"

if [ "$(curl --head --write-out "%{http_code}\n" --silent --output /dev/null "$url")" -eq "404" ]; then
  err "$url does not exist, you will need to build $binName from source"
fi

td=$(mktemp -d || mktemp -d -t tmp)
curl -sL "$url" | tar -C "$td" -xz

say_err "Installing to: $dest"

for f in "$td"/*; do
    [ -e "$f" ] || break # handle the case of no *.wav files

    test -x "$f" || continue

    if [ -e "$dest/$binName" ] && [ $force = false ]; then
        err "$binName already exists in $dest"
    else
        mkdir -p "$dest"
        cp "$f" "$dest"
        chmod 0755 "$dest/$binName"
    fi
done

rm -rf "$td"