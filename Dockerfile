FROM alpine:3.7

COPY bin/app /usr/bin/weave-policy-validator
COPY entrypoint.sh /usr/bin/weave-validator

RUN chmod +x /usr/bin/weave-policy-validator
RUN chmod +x /usr/bin/weave-validator
