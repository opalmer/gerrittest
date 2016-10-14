# This file contains functions for interacting with or setting
# up Gerrit in docker. This is intended for testing purposes and
# is meant to be sourced by other scripts.

# The image to use to run Gerrit in docker
DOCKER_IMAGE="${DOCKER_IMAGE:=opalmer/gerrittest:2.13.1}"

# The local port that docker should forward ssh to from the container.
PORT_SSH="${PORT_SSH:=29418}"

# The local port that docker should forward http to from the container.
PORT_HTTP="${PORT_HTTP:=8080}"

# The name to use when creating or removing the container.
CONTAINER_NAME="${CONTAINER_NAME:=gerrittest}"

# How long WaitForGerrit should wait for each service to come up.
WAIT_DURATION="${WAIT_DURATION:=120}"

# This function will run the requested command for docker-machine
# until the return code is 0. It is assumed that docker-machine itself
# exists on the local host.
# NOTE: This function exists because docker-machine can sometimes
# fail with a traceback. This appears to only occur on certain
# platforms but it's better safe than sorry since we rely on the
# output of docker-machine in a few places.
function dockermachine {
    local command=$1
    local retries=0

    while [ $retries -lt 50 ]; do
        output=$(docker-machine $command 2>&1)
        if [ $? -eq 0 ]; then
            echo $output
            return
        fi
        let retries=retries+1
    done

    echo "ERROR: docker-machine $command failed after $retries retries"
    exit 1
}

# Echos the proper address to use when connecting to Gerrit
function GerritAddress {
    # Explicit override
    if [ "$GERRIT_ADDRESS" != "" ]; then
        echo $GERRIT_ADDRESS
        return
    fi

    # If the docker-machine command does not exist then use localhost
    docker-machine version 2>1 /dev/null
    if [ $? -ne 0 ]; then
        echo "localhost"
        return
    fi

    local active=$(dockermachine active)
    echo $(dockermachine ip $active)
}


# Locates and echos the id of the container running gerrit.
function ContainerID {
    echo $(docker ps --quiet --filter name=$CONTAINER_NAME --filter status=running 2> /dev/null)
}

# Waits for Gerrit to start listening for http/ssh connections
function WaitForGerrit {
    local ssh_up=false
    local http_up=false
    local counter=0
    local address=$(GerritAddress)

    echo "Waiting for Gerrit to be expose http/ssh at $address"
    while [ $counter -lt $WAIT_DURATION ]; do
        if [ "$http_up" = false ]; then
            curl -o /dev/null -s http://$address:$PORT_HTTP
            if [ "$?" -eq "0" ]; then
                http_up=true
            fi
        fi

        if [ "$ssh_up" = false ]; then
            local out=$(ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -p $PORT_SSH $address 2>&1)
            if [[ $out == *"Permission denied (publickey)."* ]]; then
                ssh_up=true
            fi
        fi

        if [[ "$http_up" = true && "$ssh_up" = true ]]; then
            break
        fi

        let counter=counter+1
        sleep 1
    done

    if [ $counter -ge $WAIT_DURATION ]; then
        echo "ERROR: Timed out waiting for http/ssh to come up"
        exit 1
    fi
}

# Runs Gerrit in a container and waits for it to come up before returning.
function RunContainer {
    local address=$(GerritAddress)

    if [ "$(ContainerID)" = "" ]; then
        RunOrFail "docker run -d --name $CONTAINER_NAME -p $PORT_HTTP:8080 -p $PORT_SSH:29418 -e GERRIT_CANONICAL_URL=http://$address:$PORT_HTTP $DOCKER_IMAGE" \
            "ERROR: Failed to run container"
    else
        echo "Gerrit already running in container $(ContainerID)"
    fi

    WaitForGerrit
}

# Kills the running container then removes it.
function StopContainer {
    local container=$(ContainerID)
    if [ "$container" = "" ]; then
        return
    fi

    echo "Stopping container $container"
    docker kill $container > /dev/null || true  # might already be stopped
    docker rm -f $container > /dev/null
}

# Runs a command or fails with an error message
#  RunOrFail "ping -c 0 nosuchhost" "ERROR: Ping failed"
function RunOrFail {
    local command=$1
    local errormessage=$2

    $command
    code=$?

    if [ $code -ne 0 ]; then
        echo $errormessage
        echo "  command: $command"
        echo "  code: $code"
        exit $code
    else
        >&2 echo "$command (exit: $code)"
    fi
}

DIGEST_CURL="curl -sLo /dev/null -u admin:secret --digest --fail"

# Creates the administrator account
function CreateAdminAccount {
    local address=$(GerritAddress)
    local cookiefile=$(mktemp)

    echo "Creating admin account and testing login"

    RunOrFail \
        "curl -sLo /dev/null --fail  --cookie-jar $cookiefile http://$address:$PORT_HTTP/login/%23%2F?account_id=1000000" \
        "ERROR: Failed to create account"

    RunOrFail \
        "$DIGEST_CURL http://$address:$PORT_HTTP/a/accounts/self" \
        "ERROR: Failed to login to new account"
}

# Generates a new ssh private key and echos the path
function GenerateSSHPrivateKey {
    local tmpdir=$(mktemp -d)
    ssh-keygen -b 2048 -t rsa -f $tmpdir/id_rsa -q -N ""
    echo $tmpdir/id_rsa
}

# Adds the provided ssh key. Keys can be generated using GenerateSSHPrivateKey
function AddAdminSSHKey {
    local path=$1
    local address=$(GerritAddress)
    local url="http://$address:$PORT_HTTP/a/accounts/self/sshkeys"

    if [[ ! -f $path || $path = "" ]]; then
        echo "ERROR: Provided path does not exist or it was not provided"
        exit 1
    fi

    echo "Adding ssh key $path.pub"

    # Add the key
    RunOrFail \
        "$DIGEST_CURL -u admin:secret --digest -H \"Content-Type: plain/text\" -o /dev/null --fail --data-binary @$path.pub http://$address:$PORT_HTTP/a/accounts/self/sshkeys" \
        "ERROR: Failed to add ssh key"

    # Test the key
    RunOrFail \
        "ssh -o LogLevel=quiet -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i $path -p $PORT_SSH admin@$address gerrit version" \
        "ERROR: key test failed"
}
