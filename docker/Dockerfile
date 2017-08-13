FROM java:openjdk-8-jre-alpine

MAINTAINER Oliver Palmer <oliverpalmer@opalmer.com>

ARG GERRIT_HOME=/var/gerrit
ARG GERRIT_SITE=${GERRIT_HOME}
ARG GERRIT_USER=gerrit
ARG GERRIT_INIT_ARGS=""
ARG GERRIT_CANONICAL_URL="http://localhost:8080/"
ENV GERRIT_HOME ${GERRIT_HOME}
ENV GERRIT_SITE ${GERRIT_SITE}
ENV GERRIT_CANONICAL_URL ${GERRIT_CANONICAL_URL}

# Initial setup. These will rarely change so we perform these steps first
# so the results are cached.
RUN set -x \
    && apk add --update --no-cache git openssh openssl bash curl \
    && adduser -D -h "${GERRIT_HOME}" -g "Gerrit User" "${GERRIT_USER}"

ARG BOUNCY_CASTLE_VERSION=1.54
ARG GERRIT_WAR_URL=https://gerrit-releases.storage.googleapis.com/gerrit-2.14.2.war
ARG GERRIT_WAR_SHA1=e57c23f17457e14ecfd8da6b2750a94fe65f8970

# Drop in the Gerrit war file and setup the site.
RUN set -x \
    && mkdir -p ${GERRIT_HOME}/bin \
    && curl -Lso ${GERRIT_HOME}/bin/gerrit.war ${GERRIT_WAR_URL} \
    && echo "${GERRIT_WAR_SHA1}  ${GERRIT_HOME}/bin/gerrit.war" | sha1sum -c - \
    && java -jar ${GERRIT_HOME}/bin/gerrit.war init --batch --no-auto-start -d ${GERRIT_SITE}
#    && curl -so ${GERRIT_SITE}/lib/bcprov-jdk15on-${BOUNCY_CASTLE_VERSION}.jar https://repo1.maven.org/maven2/org/bouncycastle/bcprov-jdk15on/${BOUNCY_CASTLE_VERSION}/bcprov-jdk15on-${BOUNCY_CASTLE_VERSION}.jar \
#    && curl -so ${GERRIT_SITE}/lib/bcpkix-jdk15on-${BOUNCY_CASTLE_VERSION}.jar https://repo1.maven.org/maven2/org/bouncycastle/bcpkix-jdk15on/${BOUNCY_CASTLE_VERSION}/bcpkix-jdk15on-${BOUNCY_CASTLE_VERSION}.jar \

# Add files
COPY /entrypoint.sh /
COPY config/gerrit.config ${GERRIT_SITE}/etc/gerrit.config

WORKDIR ${GERRIT_HOME}
EXPOSE 8080 29418

CMD ["/entrypoint.sh"]
