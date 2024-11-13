FROM alpine

# ENV GIN_MODE=release
# ENV PORT=8433

COPY dist/kingapi /opt
COPY dist/personalities.json /opt
COPY dist/certs /opt/certs
WORKDIR /opt

CMD ["./kingapi"]