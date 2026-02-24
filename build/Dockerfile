FROM alpine

RUN apk add --no-cache wine nodejs npm bash \
    && npm i -g pkg@5.8.1

ENV WINEDEBUG=-all
ENV XDG_RUNTIME_DIR=/tmp
RUN wineboot --init

WORKDIR /build
COPY assets ./assets
RUN test -d assets/cm/books && \
    test -d assets/cm/personalities && \
    test -f assets/cm/TheKing350.exe

COPY runbook ./runbook
COPY scripts ./scripts
RUN chmod +x ./scripts/*.sh
RUN mkdir -p /build/dist
RUN mkdir -p /build/logs

CMD ["bash", "./scripts/build-all.sh"]
