FROM golang:1.12

ENV PATH=$PATH:$PWD/go/bin

RUN apt-get update && apt-get install -y --no-install-recommends rsync

ADD "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=6.38.0&source=github-rel" cf.tgz
RUN tar xzf cf.tgz cf
RUN chmod +x cf

ADD "https://github.com/cloudfoundry/bosh-bootloader/releases/download/v6.9.0/bbl-v6.9.0_linux_x86-64" bbl
RUN chmod +x bbl

ADD "https://github.com/cloudfoundry-incubator/credhub-cli/releases/download/2.0.0/credhub-linux-2.0.0.tgz" credhub.tgz
RUN tar xzf credhub.tgz ./credhub
RUN chmod +x credhub

ADD "https://github.com/cloudfoundry/bosh-cli/releases/download/v5.2.2/bosh-cli-5.2.2-linux-amd64" bosh
RUN chmod +x bosh

ADD "https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64" jq
RUN chmod +x jq

RUN cp cf /bin/cf
RUN cp bbl /bin/bbl
RUN cp credhub /bin/credhub
RUN cp bosh /bin/bosh
RUN cp jq /bin/jq
