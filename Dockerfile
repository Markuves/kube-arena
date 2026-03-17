FROM golang:1.25-bookworm AS runtime-builder

WORKDIR /perf-src

RUN git clone https://github.com/kubernetes/perf-tests.git . && \
    cd clusterloader2 && \
    GOOS=linux GOARCH=amd64 go build -o /out/clusterloader2 ./cmd/clusterloader.go

FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    curl \
    docker.io \
    git \
    iptables \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/local/bin

# Install kubectl
RUN curl -sSL -o kubectl https://dl.k8s.io/release/v1.30.0/bin/linux/amd64/kubectl && \
    chmod +x kubectl

# Install KIND
RUN curl -sSL -o kind https://kind.sigs.k8s.io/dl/v0.23.0/kind-linux-amd64 && \
    chmod +x kind

# Install Terraform
RUN curl -sSL -o terraform.zip https://releases.hashicorp.com/terraform/1.7.5/terraform_1.7.5_linux_amd64.zip && \
    apt-get update && apt-get install -y --no-install-recommends unzip && \
    unzip terraform.zip -d /usr/local/bin && \
    rm terraform.zip && \
    apt-get purge -y unzip && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

COPY --from=runtime-builder /out/clusterloader2 /usr/local/bin/clusterloader2
COPY runner/ks-arena-runner /usr/local/bin/ks-arena-runner
RUN chmod +x /usr/local/bin/ks-arena-runner

ENTRYPOINT ["ks-arena-runner"]

