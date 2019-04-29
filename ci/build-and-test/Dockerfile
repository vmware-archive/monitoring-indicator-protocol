FROM golang:1.12

ENV PATH=$PATH:$PWD/go/bin

RUN apt-get update
RUN apt-get install -y --no-install-recommends rsync jq lsb-core

RUN echo "deb http://packages.cloud.google.com/apt cloud-sdk-$(lsb_release -c -s) main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list

RUN curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN apt-get update
RUN apt-get install --yes google-cloud-sdk

RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
RUN chmod +x kubectl

ADD "https://github.com/kubernetes-sigs/kustomize/releases/download/v2.0.3/kustomize_2.0.3_linux_amd64" kustomize
RUN chmod +x kustomize

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

RUN mv kubectl /bin/kubectl
RUN mv kustomize /bin/kustomize
RUN mv cf /bin/cf
RUN mv bbl /bin/bbl
RUN mv credhub /bin/credhub
RUN mv bosh /bin/bosh
