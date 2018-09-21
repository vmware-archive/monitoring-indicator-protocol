FROM golang:1.11

ENV PATH=$PATH:$PWD/go/bin

RUN wget -O cf.tgz "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=6.38.0&source=github-rel"
RUN tar xzf cf.tgz cf
RUN chmod +x cf

RUN wget -O bbl "https://github.com/cloudfoundry/bosh-bootloader/releases/download/v6.9.0/bbl-v6.9.0_linux_x86-64"
RUN chmod +x bbl

RUN wget -O credhub.tgz "https://github.com/cloudfoundry-incubator/credhub-cli/releases/download/2.0.0/credhub-linux-2.0.0.tgz"
RUN tar xzf credhub.tgz ./credhub
RUN chmod +x credhub

RUN wget -O bosh "https://github.com/cloudfoundry/bosh-cli/releases/download/v5.2.2/bosh-cli-5.2.2-linux-amd64"
RUN chmod +x bosh

RUN wget -O jq "https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64"
RUN chmod +x jq

RUN cp cf /bin/cf
RUN cp bbl /bin/bbl
RUN cp credhub /bin/credhub
RUN cp bosh /bin/bosh
RUN cp jq /bin/jq
