package gerrittest

// Config is used to tell the *runner struct what setup steps
// to perform, where to listen for services, etc.
type Config struct {
	// Image is the name of docker image to run.
	Image string

	// PortSSH is the port to expose the SSH service on.
	PortSSH uint16

	// PortHTTP is the port to expose the HTTP service on.
	PortHTTP uint16

	// CleanupOnFailure indicates if the container should be kept around
	// after we're done and/or after failure.
	CleanupOnFailure bool
}
