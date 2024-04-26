#!/bin/bash
echo "Installing mssql-tools - ODBC and Go"
curl -sSL https://packages.microsoft.com/keys/microsoft.asc | (OUT=$(apt-key add - 2>&1) || echo "$OUT")
DISTRO=$(lsb_release -is | tr '[:upper:]' '[:lower:]')
CODENAME=$(lsb_release -cs)
RELEASE=$(lsb_release -rs)

echo "Installing mssql-tools - ODBC"
echo "deb [arch=amd64,arm64] https://packages.microsoft.com/repos/microsoft-${DISTRO}-${CODENAME}-prod ${CODENAME} main" >/etc/apt/sources.list.d/microsoft.list
apt-get update
ACCEPT_EULA=Y apt-get -y install unixodbc-dev msodbcsql18 libunwind8 mssql-tools18
mv /opt/mssql-tools18 /opt/mssql-tools
#Note: creating symbolic link conflicts with Go version
#ln -s /opt/mssql-tools/bin/* /usr/local/bin/

echo "Installing mssql-tools - Go"
echo "deb [arch=amd64,arm64,armhf] https://packages.microsoft.com/${DISTRO}/${RELEASE}/prod ${CODENAME} main" >/etc/apt/sources.list.d/microsoft-prod.list
apt-get update

ACCEPT_EULA=Y apt-get -y install sqlcmd

# Note: not yet available in the Debian 12 bookworm repo. Installing manually.
# wget https://packages.microsoft.com/debian/11/prod/pool/main/s/sqlcmd/sqlcmd_1.2.1-1_bullseye_all.deb
# dpkg --install sqlcmd_1.2.1-1_bullseye_all.deb
# rm -f sqlcmd_1.2.1-1_bullseye_all.deb

echo "Installing sqlpackage"
curl -sSL -o sqlpackage.zip "https://aka.ms/sqlpackage-linux"
mkdir /opt/sqlpackage
unzip sqlpackage.zip -d /opt/sqlpackage
rm sqlpackage.zip
chmod a+x /opt/sqlpackage/sqlpackage
ln -s /opt/sqlpackage/sqlpackage /usr/local/bin/
