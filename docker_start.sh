#!/bin/bash

service postgresql start
mkdir -p /mnt/containers/meme_coin/pg/
su -c "/usr/lib/postgresql/9.5/bin/initdb /mnt/containers/meme_coin/pg/" -m postgres
su -c "/usr/lib/postgresql/9.5/bin/pg_ctl -D /mnt/containers/meme_coin/pg/ -l /mnt/containers/meme_coin/pg/pg.log start" -m postgres
su -c "createdb money" -m postgres
su -c "createuser -d -s memebot" -m postgres
su -c "psql -c 'alter user memebot password \$\$password\$\$;'" -m postgres
su -c "psql -c 'create table money(id SERIAL, discord_id VARCHAR(100) not null, name VARCHAR(100) not null, current_money numeric DEFAULT 1000, total_money numeric DEFAULT 0, won_money numeric DEFAULT 0, lost_money numeric DEFAULT 0, given_money numeric DEFAULT 0, received_money numeric DEFAULT 0);' money" -m postgres


/meme_coin
