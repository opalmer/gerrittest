#!/bin/bash -ex

set_gerrit_config() {
  git config -f "${GERRIT_SITE}/etc/gerrit.config" "$@"
}

set_gerrit_config gerrit.canonicalWebUrl ${GERRIT_CANONICAL_URL}
exec java -jar ${GERRIT_SITE}/bin/gerrit.war daemon --console-log -d ${GERRIT_SITE}
