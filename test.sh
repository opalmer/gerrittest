#!/bin/bash

container_id=$(gerrittest --log-level debug run)
ssh_port=$(gerrittest --log-level debug get-port ssh $container_id)
http_port=$(gerrittest --log-level debug get-port http $container_id)
sleep 5
rsa_key=$(gerrittest --log-level debug create-admin $container_id)
echo $container_id $ssh_port $http_port $rsa_key