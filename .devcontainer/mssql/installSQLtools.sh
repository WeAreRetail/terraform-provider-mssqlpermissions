#!/bin/bash
echo "Installing mssql-tools - ODBC and Go"
curl -sSL https://packages.microsoft.com/keys/microsoft.asc | (OUT=$(apt-key add - 2>&1) || echo "$OUT")
DISTRO=$(lsb_release -is | tr '[:upper:]' '[:lower:]')
RELEASE=$(lsb_release -rs)

# Install the Linux Software Repository for Microsoft Products
curl -sSL -O https://packages.microsoft.com/config/"$DISTRO"/"$RELEASE"/packages-microsoft-prod.deb
dpkg -i packages-microsoft-prod.deb
rm packages-microsoft-prod.deb
sudo apt-get update

# Install the mssql-tools
ACCEPT_EULA=Y apt-get -y install unixodbc-dev msodbcsql18 libunwind8 mssql-tools18
ln -s /opt/mssql-tools18 /opt/mssql-tools
#Note: creating symbolic link conflicts with Go version
#ln -s /opt/mssql-tools/bin/* /usr/local/bin/

# Install the mssql-tools - Go
# Not available in multiple distributions repo. Install manually.
echo "Installing mssql-tools - Go"
echo "Looking for the latest GitHub release on https://github.com/microsoft/go-sqlcmd"
LATEST_RELEASE=$(curl -s https://api.github.com/repos/microsoft/go-sqlcmd/releases/latest | jq -r .tag_name)
echo "Found latest release: $LATEST_RELEASE"
# Download and install the latest release
curl -sSL -o sqlcmd-linux-amd64.tar.bz2 "https://github.com/microsoft/go-sqlcmd/releases/download/$LATEST_RELEASE/sqlcmd-linux-amd64.tar.bz2"
mkdir -p /opt/go-sqlcmd
tar -xjf sqlcmd-linux-amd64.tar.bz2 -C /opt/go-sqlcmd
rm sqlcmd-linux-amd64.tar.bz2
chmod a+x /opt/go-sqlcmd/sqlcmd
ln -s /opt/go-sqlcmd/sqlcmd /usr/local/bin/

# Install sqlpackage
echo "Installing sqlpackage"
curl -sSL -o sqlpackage.zip "https://aka.ms/sqlpackage-linux"
mkdir /opt/sqlpackage
unzip sqlpackage.zip -d /opt/sqlpackage
rm sqlpackage.zip
chmod a+x /opt/sqlpackage/sqlpackage
ln -s /opt/sqlpackage/sqlpackage /usr/local/bin/
