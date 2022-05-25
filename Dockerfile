FROM alpine:3.7

COPY bin/app /usr/bin/weave-iac-validator
COPY entrypoint.sh /usr/bin/weave-validator

RUN chmod +x /usr/bin/weave-iac-validator
RUN chmod +x /usr/bin/weave-validator
