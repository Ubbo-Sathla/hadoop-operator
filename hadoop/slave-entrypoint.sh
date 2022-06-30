#!/bin/bash
sed -i -e "s/<MASTER_ENDPOINT>/$MASTER_ENDPOINT/g" $HADOOP_CONF_DIR/yarn-site.xml
sed -i -e "s/master/$MASTER_ENDPOINT/g" $HADOOP_CONF_DIR/core-site.xml
sed -i -e "s/master/$MASTER_ENDPOINT/g" $HADOOP_CONF_DIR/mapred-site.xml
sed -i -e "s/hadoop_distribution_directory/$HADOOP_HOME/g" $HADOOP_CONF_DIR/mapred-site.xml

mkdir -p /hadoop/ && chown -R hadoop.hadoop /hadoop/
/etc/init.d/ssh start
tail -f /dev/null