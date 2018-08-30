FROM golang:1.11

ENV GOPATH=$PWD/go
ENV PATH=$PATH:$GOPATH/bin

RUN wget -O cf.tgz "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=6.38.0&source=github-rel"
RUN tar xzf cf.tgz cf
RUN chmod +x cf

RUN wget -O bbl "https://github.com/cloudfoundry/bosh-bootloader/releases/download/v6.9.0/bbl-v6.9.0_linux_x86-64"
RUN chmod +x bbl

RUN wget -O credhub.tgz "https://github.com/cloudfoundry-incubator/credhub-cli/releases/download/2.0.0/credhub-linux-2.0.0.tgz"
RUN tar xzf credhub.tgz ./credhub
RUN chmod +x credhub


RUN cp cf /bin/cf
RUN cp bbl /bin/bbl
RUN cp credhub /bin/credhub
