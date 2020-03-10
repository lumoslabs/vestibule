FROM alpine:3.8

COPY ["bule", "vest", "/bin/"]

RUN { \
  echo '#!/usr/bin/dumb-init /bin/sh'; \
  echo 'mkdir -p /var/run/vestibule && bule /var/run/vestibule/secrets || true'; \
  echo 'exec vest $@'; \
  } >/entrypoint.sh \
  && chmod 755 /entrypoint.sh \
  && apk add --update dumb-init jq ca-certificates

ENTRYPOINT [ "/entrypoint.sh" ]
