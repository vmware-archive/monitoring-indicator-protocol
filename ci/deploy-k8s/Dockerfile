FROM google/cloud-sdk:240.0.0-slim

RUN gcloud components install kubectl --quiet
RUN curl -o /usr/local/bin/kustomize -L https://github.com/kubernetes-sigs/kustomize/releases/download/v2.0.3/kustomize_2.0.3_linux_amd64
RUN chmod u+x /usr/local/bin/kustomize
