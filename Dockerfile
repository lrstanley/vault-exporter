FROM centos:7
ADD _output/bin/vault_exporter /usr/bin
ENTRYPOINT ["/usr/bin/vault_exporter"]
