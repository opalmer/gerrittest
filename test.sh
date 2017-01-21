#!/bin/bash -ex

container_id=$(gerrittest --log-level debug run)
gerrittest --log-level debug wait $container_id
gerrittest --log-level debug get-port ssh $container_id
gerrittest --log-level debug get-port http $container_id
gerrittest --log-level debug create-admin $container_id
