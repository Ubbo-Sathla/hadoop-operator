FROM ubuntu:focal-20220531 AS base

ENV HADOOP_CONF_DIR /etc/hadoop

ENV JAVA_HOME /opt/jdk1.8.0_311
ENV HADOOP_HOME /opt/hadoop-3.3.1

ENV PATH $HADOOP_HOME/sbin:$HADOOP_HOME/bin:$JAVA_HOME/bin:$PATH

RUN apt update || true && apt install -y  openssh-server openssh-client net-tools dnsutils wget curl jq || true && rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true && rm -rf /var/lib/{apt,dpkg,cache,log}
RUN wget -O /usr/local/bin/gosu https://github.com/tianon/gosu/releases/download/1.14/gosu-amd64  && chmod +x /usr/local/bin/gosu

ADD hdfs /

WORKDIR /opt
RUN wget http://10.127.252.101/dav/apache/hadoop/hadoop-3.3.1.tar.gz && tar -zxvf hadoop-3.3.1.tar.gz && wget http://10.127.252.101/dav/java/jdk-8u311-linux-x64.tar.gz && tar -zxvf jdk-8u311-linux-x64.tar.gz && rm -rf hadoop-3.3.1.tar.gz jdk-8u311-linux-x64.tar.gz

WORKDIR /
RUN groupadd -r hadoop && useradd -m -r -g hadoop hadoop && chown -R hadoop.hadoop $JAVA_HOME $HADOOP_HOME $HADOOP_CONF_DIR

USER hadoop
RUN ssh-keygen -f ~/.ssh/id_rsa -t rsa -N '' && /bin/cp /home/hadoop/.ssh/id_rsa.pub /home/hadoop/.ssh/authorized_keys
USER root

FROM base

ADD master-entrypoint.sh /master-entrypoint.sh
ADD slave-entrypoint.sh /slave-entrypoint.sh

RUN chmod +x /master-entrypoint.sh /slave-entrypoint.sh