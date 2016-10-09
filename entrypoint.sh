#!/bin/bash -e


java -jar ${GERRIT_WAR} init --batch --no-auto-start -d ${GERRIT_SITE}
exec java -jar ${GERRIT_WAR} daemon --console-log -d ${GERRIT_SITE}

