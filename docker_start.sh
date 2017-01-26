#!/bin/bash

export PATH=$PATH:/usr/local/go/bin
export GOPATH=/builds/go
export PATH=$GOPATH/bin:$PATH
#service postgresql start
mkdir -p /mnt/containers/meme_coin/pg/
chown postgres -R /mnt/containers/meme_coin/pg
su -c "/usr/lib/postgresql/9.5/bin/initdb /mnt/containers/meme_coin/pg/" -m postgres
su -c "/usr/lib/postgresql/9.5/bin/pg_ctl -D /mnt/containers/meme_coin/pg/ -l /mnt/containers/meme_coin/pg/pg.log start" -m postgres
sleep 10
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
