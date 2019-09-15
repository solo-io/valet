FROM google/cloud-sdk:alpine

RUN apk add --update ca-certificates
RUN apk add --update make cmake gcc g++
RUN apk add --update git

RUN gcloud components install kubectl

EXPOSE 8080

COPY valet-linux-amd64 /usr/local/bin/valet

ENTRYPOINT [ "/usr/local/bin/valet" ]