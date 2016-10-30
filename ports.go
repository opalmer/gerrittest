package gerrittest

const (
	// InternalHTTPPort is the port Gerrit is listening for http requests
	// on inside the container.
	InternalHTTPPort = uint16(8080)

	// InternalSSHPort is the port Gerrit is listening for ssh connection on
	// inside the container.
	InternalSSHPort = uint16(29418)
)
