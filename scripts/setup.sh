#!/bin/bash

ADDRESS=localhost
_=$(docker-machine active 2> /dev/null)
if [ "$?" -eq "0" ]; then
    ADDRESS=$(docker-machine ip $active)
    eval $(docker-machine env $active)
fi

_=$(docker rm -f gerrittest 2> /dev/null)

# Start the container, remove the container if it fails
container=$(docker run -d -p 8080:8080 -p 29418:29418 --name gerrittest opalmer/gerrittest)
if [ "$?" -ne "0" ]; then
  echo "WARNING: Removing container $container"
  echo "WARNING: command failed!"
  docker rm $container
  exit 1
fi

# Wait on Gerrit to come up
COUNTER=0
echo "Waiting up to 120 seconds for Gerrit to come up"
while [  $COUNTER -lt 120 ]; do
   curl -s http://$ADDRESS:8080/ -o /dev/null
   if [ "$?" -eq "0" ]; then
       break
   fi
   let COUNTER=COUNTER+1
   sleep 1
done

# Create the admin user and login once
set -ex
curl -o /dev/null -s -L http://$ADDRESS:8080/login/%23%2F?account_id=1000000  --cookie-jar GerritAccount
curl -o /dev/null -s -L http://$ADDRESS:8080/a/accounts/self --digest -u admin:secret
set +x
rm GerritAccount

# Generate, add and test an ssh public key
tmpdir=$(mktemp -d)
set -x
ssh-keygen -b 2048 -t rsa -f $tmpdir/id_rsa -q -N ""
curl -o /dev/null -s -L -X POST http://$ADDRESS:8080/a/accounts/self/sshkeys --digest -u admin:secret -H "Content-Type: plain/text" --data-binary @$tmpdir/id_rsa.pub
ssh -o LogLevel=quiet -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i $tmpdir/id_rsa.pub -p 29418 admin@$ADDRESS # test public key
