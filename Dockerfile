FROM alpine:3.7

COPY bin/app /usr/bin/weave-validator
COPY entrypoint.sh entrypoint.sh

RUN chmod +x /usr/bin/weave-validator
RUN chmod +x entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
