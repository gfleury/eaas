FROM tsuru/go:latest

# This is needed for flow, and the weirdos that built it in ocaml:
RUN apt-get update && apt-get install -y build-essential mongodb

# grab etcd
ENV ETCD_VERSION v3.3.5
ENV ETCD_FILE etcd-$ETCD_VERSION-linux-amd64
RUN curl --silent -L -o $ETCD_FILE.tar.gz https://github.com/coreos/etcd/releases/download/$ETCD_VERSION/$ETCD_FILE.tar.gz

# install and clean up
RUN tar -zxf $ETCD_FILE.tar.gz
RUN rm $ETCD_FILE.tar.gz
RUN mv $ETCD_FILE/etcd /usr/local/bin
RUN rm -rf $ETCD_FILE
