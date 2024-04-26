#!/bin/bash
dacpac="false"
sqlfiles="false"
SApassword=$1
dacpath=$2
sqlpath=$3
host=${4:-mssql-fixture}
port=${5:-1433}

echo "SELECT * FROM SYS.DATABASES" | dd of=testsqlconnection.sql
# shellcheck disable=SC2034
for i in {1..60}; do
  /opt/mssql-tools/bin/sqlcmd -S "$host","$port" -U sa -P "$SApassword" -d master -C -i testsqlconnection.sql >/dev/null
  # shellcheck disable=SC2181
  if [ $? -eq 0 ]; then
    echo "SQL server ready"
    break
  else
    echo "Not ready yet..."
    sleep 1
  fi
done
rm testsqlconnection.sql

for f in "$dacpath"/*; do
  if [[ $f == "$dacpath"/*".dacpac" ]]; then
    dacpac="true"
    echo "Found dacpac $f"
  fi
done

for f in "$sqlpath"/*; do
  if [[ $f == "$sqlpath"/*".sql" ]]; then
    sqlfiles="true"
    echo "Found SQL file $f"
  fi
done

if [ $sqlfiles == "true" ]; then
  for f in "$sqlpath"/*; do
    if [[ $f == "$sqlpath"/*".sql" ]]; then
      echo "Executing $f"
      /opt/mssql-tools/bin/sqlcmd -S "$host","$port" -U sa -P "$SApassword" -d master -C -i "$f"
    fi
  done
fi

if [ "$dacpac" == "true" ]; then
  for f in "$dacpath"/*; do
    if [[ $f == "$dacpath"/*".dacpac" ]]; then
      dbname=$(basename "$f" ".dacpac")
      echo "Deploying dacpac $f"
      /opt/sqlpackage/sqlpackage /Action:Publish /SourceFile:"$f" /TargetServerName:"$host","$port" /TargetDatabaseName:"$dbname" /TargetUser:sa /TargetPassword:"$SApassword"
    fi
  done
fi
