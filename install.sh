#!/bin/sh

LOCK_ARCH=${LOCK_ARCH:-amd64}

LATEST_RELEASE=$(curl --silent https://api.github.com/repos/cosee-gitlab/lock/releases/latest \
        | grep '"tag_name"' \
        | sed -E 's/.*"v([^"]+)".*/\1/' )   # use sed instead of grep -P, since that might be unavailable
RELEASE=${RELEASE:-${LATEST_RELEASE}}

echo "Installing gitlab-lock to ~/.local/bin in version ${RELEASE} and LOCK_ARCH=${LOCK_ARCH}"

curl --silent -L "https://github.com/cosee-gitlab/lock/releases/download/v${RELEASE}/lock_${RELEASE}_linux_${LOCK_ARCH}.tar.gz" \
        | tar -xz -C ~/.local/bin/ lock
