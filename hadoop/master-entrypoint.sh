#!/bin/bash
for (( i=0; i<=$SLAVES_COUNT; i++ ))
do
  echo "$SLAVES_SS_NAME-$i.$SLAVES_SVC_NAME.$NAMESPACE.svc.cluster.local" >>  $HADOOP_CONF_DIR/workers
done

for (( i=0; i<=$SLAVES_COUNT; i++ ))
do
	while :
	do
		ip=$(dig +short "$SLAVES_SS_NAME-$i.$SLAVES_SVC_NAME.$NAMESPACE.svc.cluster.local")
		if [ "$ip" != "" ]; then
			echo "$ip $SLAVES_SS_NAME-$i.$SLAVES_SVC_NAME.$NAMESPACE.svc.cluster.local" >> /etc/hosts
			echo "$SLAVES_SS_NAME-$i.$SLAVES_SVC_NAME.$NAMESPACE.svc.cluster.local is running"
			break
		else
			echo "$SLAVES_SS_NAME-$i.$SLAVES_SVC_NAME.$NAMESPACE.svc.cluster.local is not running"
			sleep 1
		fi
	done
done

sed -i -e "s/<MASTER_ENDPOINT>/$MASTER_ENDPOINT/g" $HADOOP_CONF_DIR/yarn-site.xml
sed -i -e "s/master/$MASTER_ENDPOINT/g" $HADOOP_CONF_DIR/core-site.xml
sed -i -e "s/master/$MASTER_ENDPOINT/g" $HADOOP_CONF_DIR/mapred-site.xml
sed -i -e "s/hadoop_distribution_directory/$HADOOP_HOME/g" $HADOOP_CONF_DIR/mapred-site.xml

sed -i -e "s/without-password/yes/g" /etc/ssh/sshd_config
password=$(cat /etc/rootpwd/password)
echo "root:$password" | chpasswd
echo "hadoop:$password" | chpasswd

/etc/init.d/ssh start
mkdir -p /hadoop/ && chown -R hadoop.hadoop /hadoop/
gosu hadoop hadoop namenode -format
gosu hadoop start-all.sh

tail -f /dev/null