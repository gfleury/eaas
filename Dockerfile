FROM tsuru/go:latest

# This is needed for flow, and the weirdos that built it in ocaml:
RUN sudo apt-get update && sudo apt-get install -y build-essential mongodb
RUN sudo /var/lib/tsuru/go/install

# grab etcd
WORKDIR /tmp
RUN curl --silent -L -o etcd-v3.3.5-linux-amd64.tar.gz https://github.com/coreos/etcd/releases/download/v3.3.5/etcd-v3.3.5-linux-amd64.tar.gz

# install and clean up
RUN tar -zxf etcd-v3.3.5-linux-amd64.tar.gz
RUN rm etcd-v3.3.5-linux-amd64.tar.gz
RUN sudo mv etcd-v3.3.5-linux-amd64/etcd /usr/local/bin
RUN rm -rf etcd-v3.3.5-linux-amd64
WORKDIR /
