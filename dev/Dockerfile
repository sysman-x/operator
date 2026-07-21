FROM golang:1.24

ARG OPERATOR_SDK_VERSION=v1.41.1
ARG KUBECTL_VERSION=v1.34.1
ARG DOCKER_VERSION=29.6.2

RUN apt-get update && \
    apt-get install -y \
        git \
        make \
        curl \
        ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# kubectl
RUN curl -fsSL -o /usr/local/bin/kubectl \
    https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl

# operator-sdk
RUN curl -fsSL -o /usr/local/bin/operator-sdk \
    https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk_linux_amd64 && \
    chmod +x /usr/local/bin/operator-sdk

# docker CLI only (no daemon)
RUN curl -fsSL -o /tmp/docker.tgz \
    https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz && \
    tar xzvf /tmp/docker.tgz -C /tmp && \
    mv /tmp/docker/docker /usr/local/bin/docker && \
    rm -rf /tmp/docker /tmp/docker.tgz

WORKDIR /workspace

CMD ["bash"]

