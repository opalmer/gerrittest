FROM java:openjdk-8-jre-alpine

MAINTAINER Oliver Palmer <oliverpalmer@opalmer.com>

ARG GERRIT_HOME=/var/gerrit
ARG GERRIT_SITE=${GERRIT_HOME}

# Initial setup. These will rarely change so we perform these steps first
# so the results are cached.
RUN set -x \
    && apk add --update --no-cache git openssh openssl bash curl \
    && adduser -D -h "${GERRIT_HOME}" -g "Gerrit User" "gerrit"

# The below are set the by the Makefile
ARG GERRIT_MAJOR_VERSION=""
ARG GERRIT_MINOR_VERSION=""
ARG GERRIT_MICRO_VERSION=""
ARG GERRIT_WAR_SHA1=""
ARG GERRIT_WAR_URL=https://gerrit-releases.storage.googleapis.com/gerrit-${GERRIT_MAJOR_VERSION}.${GERRIT_MINOR_VERSION}.${GERRIT_MICRO_VERSION}.war

# Drop in the Gerrit war file and setup the site.
RUN set -x \
    && mkdir -p ${GERRIT_HOME}/bin \
    && curl -Lso ${GERRIT_HOME}/bin/gerrit.war ${GERRIT_WAR_URL} \
    && echo "${GERRIT_WAR_SHA1}  ${GERRIT_HOME}/bin/gerrit.war" | sha1sum -c - \
    && java -jar ${GERRIT_HOME}/bin/gerrit.war init --batch --no-auto-start -d ${GERRIT_SITE}

# Add files
COPY /entrypoint.sh /
COPY config/gerrit.config ${GERRIT_SITE}/etc/gerrit.config
COPY plugins/${GERRIT_MAJOR_VERSION}.${GERRIT_MINOR_VERSION}/* ${GERRIT_SITE}/plugins/

ENV GERRIT_CANONICAL_URL "http://127.0.0.1:8080/"
ENV GERRIT_HOME ${GERRIT_HOME}
ENV GERRIT_SITE ${GERRIT_SITE}

WORKDIR ${GERRIT_HOME}
EXPOSE 8080 29418

CMD ["/entrypoint.sh"]
