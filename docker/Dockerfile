FROM java:openjdk-8-jre-alpine

MAINTAINER Oliver Palmer <oliverpalmer@opalmer.com>

# Commonly overriden inputs
ARG GERRIT_MAJOR_VERSION=2.13
ARG GERRIT_VERSION=2.13.2
ARG GERRIT_WAR_SHA1=37db9f32f91363100e963e78846dd1f52fbaab50

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

# Grab gosu for easy step-down from root
RUN set -x \
    && apk add --update --no-cache git openssh openssl bash

# Create user, pull in gerrit.war
RUN set -x \
    && adduser -D -h "${GERRIT_HOME}" -g "Gerrit User" "${GERRIT_USER}" \
    && mkdir -p ${GERRIT_HOME} \
    && wget http://gerrit-releases.storage.googleapis.com/gerrit-${GERRIT_VERSION}.war -O ${GERRIT_HOME}/gerrit.war \
    && echo "${GERRIT_WAR_SHA1}  ${GERRIT_HOME}/gerrit.war" > ${GERRIT_HOME}/gerrit.war.sha1 \
    && sha1sum -c ${GERRIT_HOME}/gerrit.war.sha1 \
    && rm ${GERRIT_HOME}/gerrit.war.sha1

# Add plugins
ADD https://gerrit-ci.gerritforge.com/job/plugin-delete-project-${PLUGIN_VERSION}/lastSuccessfulBuild/artifact/buck-out/gen/plugins/delete-project/delete-project.jar ${GERRIT_HOME}/delete-project.jar
ADD https://gerrit-ci.gerritforge.com/job/plugin-events-log-${PLUGIN_VERSION}/lastSuccessfulBuild/artifact/buck-out/gen/plugins/events-log/events-log.jar ${GERRIT_HOME}/events-log.jar

# Add bouncy castle
ADD https://repo1.maven.org/maven2/org/bouncycastle/bcprov-jdk15on/${BOUNCY_CASTLE_VERSION}/bcprov-jdk15on-${BOUNCY_CASTLE_VERSION}.jar ${GERRIT_HOME}/bcprov-jdk15on-${BOUNCY_CASTLE_VERSION}.jar
ADD https://repo1.maven.org/maven2/org/bouncycastle/bcpkix-jdk15on/${BOUNCY_CASTLE_VERSION}/bcpkix-jdk15on-${BOUNCY_CASTLE_VERSION}.jar ${GERRIT_HOME}/bcpkix-jdk15on-${BOUNCY_CASTLE_VERSION}.jar

# Add files
ADD /entrypoint.sh /
ADD config/gerrit.config ${GERRIT_SITE}/etc/gerrit.config

WORKDIR ${GERRIT_HOME}
EXPOSE 8080 29418

CMD ["/entrypoint.sh"]
