FROM alpine

# ENV GIN_MODE=release
# ENV PORT=8080

COPY dist/kingapi /opt
COPY dist/personalities.json /opt
WORKDIR /opt

CMD ["./kingapi"]