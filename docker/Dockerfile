from golang:1.16-alpine

RUN apk add --no-cache wget git

ENV crzy_version=0.2.0

RUN wget -O crzy_${crzy_version}_linux_x86_64.tar.gz \
    https://github.com/go-crzy/crzy/releases/download/v${crzy_version}/crzy_${crzy_version}_linux_x86_64.tar.gz && \
    tar -zxvf crzy_${crzy_version}_linux_x86_64.tar.gz -C /bin crzy && \
    rm -f crzy_${crzy_version}_linux_x86_64.tar.gz

EXPOSE 8080
EXPOSE 8081

ENTRYPOINT ["/bin/crzy"]
CMD ["-server"]
