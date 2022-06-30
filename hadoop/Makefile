HADOOP_VERSION=3.3.1

all: hadoop

hadoop:
	docker buildx build --load  -f Dockerfile -t harbor.boncloud.com/hadoop/hadoop:${HADOOP_VERSION} .
	docker push harbor.boncloud.com/hadoop/hadoop:${HADOOP_VERSION}