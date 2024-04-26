#! /bin/bash
VERSION="v1.57.2"
echo "Installing golangci-lint"
curl -sS https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh -o install.sh
chmod +x ./install.sh
./install.sh -b /usr/local/bin $VERSION
rm -f ./install.sh
golangci-lint --version
echo "Installed golangci-lint"
