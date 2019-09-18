FROM golang:1.12

RUN apt-get update && \
    wget -q https://github.com/gohugoio/hugo/releases/download/v0.57.2/hugo_0.57.2_Linux-64bit.tar.gz && \
    tar xzf hugo_0.57.2_Linux-64bit.tar.gz && \
    mkdir -p /builder/bin && \
    mv hugo /builder/bin && \
    rm hugo_0.57.2_Linux-64bit.tar.gz

RUN apt-get -y update && \
    apt-get -y install \
        apt-transport-https \
        ca-certificates \
        curl \
        jq \
        make \
        zip \
        unzip \
        software-properties-common && \
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - && \
    apt-key fingerprint 0EBFCD88 && \
    add-apt-repository \
       "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
       xenial \
       edge" && \
    apt-get -y update && \
    apt-get -y install docker-ce=5:18.09.0~3-0~ubuntu-xenial && \
    curl -sL https://deb.nodesource.com/setup_9.x | bash - && \
    apt-get install -y nodejs npm gcc python2.7 python-dev python-setuptools python-pip && \
    wget -qO- https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.tar.gz | tar zxv -C /builder && \
    CLOUDSDK_PYTHON="python2.7" /builder/google-cloud-sdk/install.sh --usage-reporting=false \
        --bash-completion=false \
        --disable-installation-options && \
    npm install -g firebase-tools

ENV PATH=/builder/google-cloud-sdk/bin/:/builder/bin:$PATH

RUN curl https://raw.githubusercontent.com/helm/helm/master/scripts/get | bash && \
    helm init --client-only

COPY valet-linux-amd64 /usr/local/bin/valet
ENTRYPOINT [ "/usr/local/bin/valet" ]

