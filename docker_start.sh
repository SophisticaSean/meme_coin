#!/bin/bash

export PATH=$PATH:/usr/local/go/bin
export GOPATH=/builds/go
export PATH=$GOPATH/bin:$PATH
mkdir -p /mnt/containers/meme_coin/pg/
chown postgres -R /mnt/containers/meme_coin/pg
su -c "/usr/lib/postgresql/9.5/bin/initdb /mnt/containers/meme_coin/pg/" -m postgres
su -c "/usr/lib/postgresql/9.5/bin/pg_ctl -D /mnt/containers/meme_coin/pg/ -l /mnt/containers/meme_coin/pg/pg.log start" -m postgres

RETRIES=10

until su -c "psql -c 'select 1'" -m postgres > /dev/null 2>&1 || [ $RETRIES -eq 0 ]; do
  echo "Waiting for postgres server, $((RETRIES--)) remaining attempts..."
  sleep 1
done

if [ $RETRIES -eq 0 ]; then
    exit 1
fi

su -c "createdb money" -m postgres
su -c "createdb test" -m postgres
su -c "createuser -d -s memebot" -m postgres
su -c "psql -c 'alter user memebot password \$\$password\$\$;'" -m postgres

if [[ $TEST != "" ]]; then
  cd /builds/go/src/github.com/SophisticaSean/meme_coin/.
  go get -v
  go test ./...
else
  /meme_coin
fi
