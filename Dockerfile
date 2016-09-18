FROM ubuntu:latest
MAINTAINER Sean Lewis
ENV SYSTEM gitlab
ENV TYPE repo
ENV TERM xterm

RUN apt-get -y update
RUN apt-get -y install wget
# install postgres 9.5.2
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ `cat /etc/lsb-release | grep CODENAME | awk -F'=' '{print $2}'`-pgdg main" > /etc/apt/sources.list.d/pgdg.list
RUN wget -q https://www.postgresql.org/media/keys/ACCC4CF8.asc -O - | apt-key add -

RUN apt-get -y update && \
    apt-get -y install \
    postgresql-9.5

ADD meme_coin /
#ADD files files/

#RUN mkdir -p /MFer && \
    #mkdir -p /root/.aws && \
    #mv /files/pgpass /root/.pgpass && \
    #chmod 0600 /root/.pgpass && \
    #mv /files/backup-cron /etc/cron.d/backup-cron && \
    #chmod 0644 /etc/cron.d/backup-cron && \
    #mv /files/backup.sh /backup.sh && \
    #mv /files/backup_pass /MFer/backup_pass && \
    #mv /files/aws_config /root/.aws/config && \
    #mv /files/extra.sh /

RUN ls -al /

#ENTRYPOINT [ "/stats_server" ]

#ENTRYPOINT [ "/bin/bash", "-c" ]
#CMD [ "bash /files/extra.sh" ]
# set up main.conf
#RUN sed -i "s/^System=.*$/System=$SYSTEM/g" /mf/main.conf
#RUN sed -i "s/^Type=.*$/Type=$TYPE/g" /mf/main.conf

## Copy Dockerfile over to image
ADD Dockerfile /Dockerfile.$SYSTEM
ADD docker_start.sh /docker_start.sh

## Replaces the ami_machine_id bin file that use to exist on instances ##
#RUN echo "START" >> /MFer/machine_id \
    #echo "$SYSTEM" >> /MFer/machine_id \
    #echo "Build Key: " `uuidgen` >> /MFer/machine_id \
    #echo "Build Date: "`date +%s` >> /MFer/machine_id
#RUN cat /Dockerfile.$SYSTEM >> /MFer/machine_id
#RUN echo "END" >> /MFer/machine_id
#-----------------------------------------------------------------------------------

ENTRYPOINT [ "/bin/bash", "-c" ]
CMD [ "bash /docker_start.sh" ]
