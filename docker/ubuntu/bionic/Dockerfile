FROM ubuntu:bionic
COPY ["bule", "vest", "/bin/"]
RUN { \
  echo '#!/usr/bin/dumb-init /bin/sh'; \
  echo "bule /var/run/secrets.json || true"; \
  echo 'exec vest $@'; \
  } >/entrypoint.sh \
  && chmod 755 /entrypoint.sh \
  && apt-get update -qq \
  && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends dumb-init jq ca-certificates
ENTRYPOINT [ "/entrypoint.sh" ]
