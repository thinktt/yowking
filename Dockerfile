FROM i386/alpine

RUN apk update
RUN apk add wine
RUN apk add nodejs
RUN apk add vim
RUN winecfg
RUN apk add xvfb

ENV WINEDEBUG=-all
ENV DISPLAY=:0.0

COPY dist /opt/yowking/
WORKDIR /opt/yowking

ENV ENG_CMD="/usr/bin/wine enginewrap.exe"
CMD ["./kingapi"]