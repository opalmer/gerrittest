#!/bin/bash -e

set_gerrit_config() {
  git config -f "${GERRIT_SITE}/etc/gerrit.config" "$@"
}

set_gerrit_config gerrit.canonicalWebUrl ${GERRIT_CANONICAL_URL}

java -jar ${GERRIT_WAR} init --batch --no-auto-start -d ${GERRIT_SITE}
exec java -jar ${GERRIT_WAR} daemon --console-log -d ${GERRIT_SITE}

