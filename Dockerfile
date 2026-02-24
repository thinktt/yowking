FROM alpine

RUN apk add --no-cache wine 

ENV WINEDEBUG=-all
ENV XDG_RUNTIME_DIR=/tmp
RUN wineboot --init

COPY dist /opt/yowking/
WORKDIR /opt/yowking

ENV ENG_CMD="/usr/bin/wine enginewrap.exe"
ENV GIN_MODE=release
CMD ["./kingworker"]