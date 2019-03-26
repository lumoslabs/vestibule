FROM alpine:3.8
COPY ["bule", "vest", "/bin/"]
RUN { \
  echo '#!/usr/bin/dumb-init /bin/sh'; \
  echo 'exec $@'; \
  } >/entrypoint.sh \
  && chmod 755 /entrypoint.sh \ 
  && apk add --update dumb-init ca-certificates
ENTRYPOINT [ "/entrypoint.sh" ]
