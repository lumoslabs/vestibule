FROM alpine:3.8
COPY vest /bin/vest
ENV USER=root
RUN { \
  echo '#!/usr/bin/dumb-init /bin/sh'; \
  echo 'exec /bin/vest $@'; \
  } >/entrypoint.sh \
  && chmod 755 /entrypoint.sh \ 
  && apk add --update dumb-init
ENTRYPOINT [ "/entrypoint.sh" ]