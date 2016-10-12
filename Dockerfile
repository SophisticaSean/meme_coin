FROM ubuntu:latest
MAINTAINER Sean Lewis
ENV SYSTEM gitlab
ENV TYPE repo
ENV TERM xterm

RUN apt-get -y update
RUN apt-get -y install wget vim
# install postgres 9.5.2
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ `cat /etc/lsb-release | grep CODENAME | awk -F'=' '{print $2}'`-pgdg main" > /etc/apt/sources.list.d/pgdg.list
RUN wget -q https://www.postgresql.org/media/keys/ACCC4CF8.asc -O - | apt-key add -

RUN apt-get -y update && \
    apt-get -y install \
    postgresql-9.5

ADD meme_coin /

## Copy Dockerfile over to image
ADD Dockerfile /Dockerfile.$SYSTEM
ADD docker_start.sh /docker_start.sh

#-----------------------------------------------------------------------------------

ENTRYPOINT [ "/bin/bash", "-c" ]
CMD [ "bash /docker_start.sh" ]
