FROM java:openjdk-8-jre-alpine

MAINTAINER Oliver Palmer <oliverpalmer@opalmer.com>

# Commonly overriden inputs.
ARG GERRIT_MAJOR_VERSION=2.13
ARG GERRIT_VERSION=2.13.5
ARG GERRIT_WAR_SHA1=4995181ada7c6dc0465c671167ac00f962d18ae6

# Less common overrides
ARG PLUGIN_VERSION=stable-${GERRIT_MAJOR_VERSION}
ARG BOUNCY_CASTLE_VERSION=1.54
ARG PLUGIN_VERSION=stable-${GERRIT_MAJOR_MINOR}

ARG GERRIT_HOME=/var/gerrit
ARG GERRIT_SITE=${GERRIT_HOME}/review_site
ARG GERRIT_USER=gerrit
ARG GERRIT_INIT_ARGS=""
ARG GERRIT_CANONICAL_URL="http://localhost:8080/"

# Establish environment variables (used by entrypoint.sh)
ENV GERRIT_HOME ${GERRIT_HOME}
ENV GERRIT_SITE ${GERRIT_SITE}
ENV GERRIT_CANONICAL_URL ${GERRIT_CANONICAL_URL}

# Create user and home of the installation
RUN set -x \
    && apk add --update --no-cache git openssh openssl bash curl

# Pull in all the remote files
RUN set -x \
    && adduser -D -h "${GERRIT_HOME}" -g "Gerrit User" "${GERRIT_USER}" && mkdir -p ${GERRIT_HOME} \
    && curl -so ${GERRIT_HOME}/events-log.jar https://gerrit-ci.gerritforge.com/job/plugin-events-log-${PLUGIN_VERSION}/lastSuccessfulBuild/artifact/buck-out/gen/plugins/events-log/events-log.jar \
    && curl -so ${GERRIT_HOME}/delete-project.jar https://gerrit-ci.gerritforge.com/job/plugin-delete-project-${PLUGIN_VERSION}/lastSuccessfulBuild/artifact/buck-out/gen/plugins/delete-project/delete-project.jar \
    && curl -so ${GERRIT_HOME}/bcprov-jdk15on-${BOUNCY_CASTLE_VERSION}.jar https://repo1.maven.org/maven2/org/bouncycastle/bcprov-jdk15on/${BOUNCY_CASTLE_VERSION}/bcprov-jdk15on-${BOUNCY_CASTLE_VERSION}.jar \
    && curl -so ${GERRIT_HOME}/bcpkix-jdk15on-${BOUNCY_CASTLE_VERSION}.jar https://repo1.maven.org/maven2/org/bouncycastle/bcpkix-jdk15on/${BOUNCY_CASTLE_VERSION}/bcpkix-jdk15on-${BOUNCY_CASTLE_VERSION}.jar \
    && curl -so ${GERRIT_HOME}/gerrit.war http://gerrit-releases.storage.googleapis.com/gerrit-${GERRIT_VERSION}.war  \
    && echo "${GERRIT_WAR_SHA1}  ${GERRIT_HOME}/gerrit.war" > ${GERRIT_HOME}/gerrit.war.sha1 \
    && sha1sum -c ${GERRIT_HOME}/gerrit.war.sha1 && rm ${GERRIT_HOME}/gerrit.war.sha1

# Add files
COPY /entrypoint.sh /
COPY config/gerrit.config ${GERRIT_SITE}/etc/gerrit.config

WORKDIR ${GERRIT_HOME}
EXPOSE 8080 29418

CMD ["/entrypoint.sh"]
