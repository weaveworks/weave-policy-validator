FROM ubuntu:22.04

RUN apt-get update \
    && apt-get install -y ca-certificates \
    && update-ca-certificates

COPY bin/app /usr/bin/weave-iac-validator
COPY entrypoint_azure.sh /usr/bin/weave-validator

RUN chmod +x /usr/bin/weave-iac-validator
RUN chmod +x /usr/bin/weave-validator
