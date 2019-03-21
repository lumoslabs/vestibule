FROM ubuntu:bionic
COPY vest /bin/vest
RUN { \
  echo '#!/usr/bin/dumb-init /bin/sh'; \
  echo 'exec /bin/vest $@'; \
  } >/entrypoint.sh \
  && chmod 755 /entrypoint.sh \
  && apt-get update -qq \
  && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends dumb-init
ENTRYPOINT [ "/entrypoint.sh" ]
