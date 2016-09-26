#!/bin/bash

#service postgresql start
mkdir -p /mnt/containers/meme_coin/pg/
su -c "/usr/lib/postgresql/9.5/bin/initdb /mnt/containers/meme_coin/pg/" -m postgres
su -c "/usr/lib/postgresql/9.5/bin/pg_ctl -D /mnt/containers/meme_coin/pg/ -l /mnt/containers/meme_coin/pg/pg.log start" -m postgres
sleep 10
su -c "createdb money" -m postgres
su -c "createuser -d -s memebot" -m postgres
su -c "psql -c 'alter user memebot password \$\$password\$\$;'" -m postgres
#su -c "psql -c 'drop table money'" -m postgres
su -c "psql -c 'create table money(id SERIAL PRIMARY KEY, discord_id VARCHAR(100), name VARCHAR(100), current_money numeric DEFAULT(1000), total_money numeric DEFAULT(0), won_money numeric DEFAULT(0), lost_money numeric DEFAULT(0), given_money numeric DEFAULT(0), received_money numeric DEFAULT(0), earned_money numeric DEFAULT(1000), spent_money numeric DEFAULT(0), mine_time timestamptz NOT NULL DEFAULT(now()));' money" -m postgres
su -c "psql -c 'create table units(discord_id VARCHAR(100) PRIMARY KEY, miner numeric DEFAULT(0), robot numeric DEFAULT(0), swarm numeric DEFAULT(0), fracker numeric DEFAULT(0);' money" -m postgres

/meme_coin
