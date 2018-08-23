#!/bin/sh

LOCK_ARCH=${LOCK_ARCH:-amd64}

RELEASE=${RELEASE:-$(curl --silent https://api.github.com/repos/cosee-gitlab/lock/releases/latest | grep -Po '"tag_name": "v\K.*?(?=")')}

echo "Installing gitlab-lock to ~/.local/bin in version ${RELEASE} and LOCK_ARCH=${LOCK_ARCH}"

cd $(mktemp -d)
curl --silent -L "https://github.com/cosee-gitlab/lock/releases/download/v${RELEASE}/lock_${RELEASE}_linux_${LOCK_ARCH}.tar.gz" | tar -xz lock

mv lock ~/.local/bin