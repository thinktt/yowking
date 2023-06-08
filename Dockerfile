FROM benjymous/docker-wine-headless

RUN wine cmd /c echo Wine is setup || true
COPY dist/* /opt/yowking/
COPY assets/* /opt/yowking/
WORKDIR /opt/yowking


# ADD dist.tar /opt/yeoldwiz
# WORKDIR /opt/yeoldwiz
# COPY --from=node:16-buster-slim /usr/local/bin/node .

# ENV ENG_CMD="/usr/bin/wine enginewrap.exe"
CMD ["./kingapi"]