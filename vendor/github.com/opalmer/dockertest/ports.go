package dockertest

import (
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
)

// Protocol is a string representing a protocol such as TCP or UDP.
type Protocol string

const (
	// RandomPort may be passed to a function to indicate
	// that a random port should be chosen.
	RandomPort uint16 = 0

	// ProtocolTCP represents a tcp Protocol.
	ProtocolTCP Protocol = "tcp"

	// ProtocolUDP represents a udp Protocol.
	ProtocolUDP Protocol = "udp"
)

// Port represents a port spec that may be used by the *Ports struct
// to expose a port on a container.
type Port struct {
	// Private is the port exposed on the inside of the container
	// that you wish to map.
	Private uint16

	// Public is the publicly facing port that you wish to expose
	// the Private port on. Note, this may be `RandomPort` if you
	// wish to expose a random port instead of a specific port.
	Public uint16 `json:"port"`

	// Address is the IP address to expose the port mapping
	// on. By default, 0.0.0.0 will be used.
	Address string `json:"address"`

	// Protocol is the network protocol to expose.
	Protocol Protocol `json:"protocol"`
}

// Port converts the struct into a nat.Port
func (s *Port) Port() (nat.Port, error) {
	if s.Protocol == "" {
		return nat.Port(0), errors.New("Protocol not specified")
	}
	return nat.NewPort(
		string(s.Protocol), fmt.Sprintf("%d", s.Private))
}

// Binding converts the struct in a a nat.PortBinding. If no address
// has been given 0.0.0.0 will be used for the host ip.
func (s *Port) Binding() nat.PortBinding {
	address := s.Address
	if address == "" {
		address = "0.0.0.0"
	}
	return nat.PortBinding{
		HostIP:   address,
		HostPort: fmt.Sprintf("%d", s.Public),
	}
}

// Ports is when to convey port exposures to RunContainer()
type Ports struct {
	// Specs is a map of internal to external ports. The external
	// port may be the same as the internal port or it may be the
	// constant `RandomPort` if you wish for Docker to chose a port
	// for you.
	Specs []*Port
}

// Bindings will take the port specs and return port bindings that can
// be used in by the container.HostConfig.PortBindings field.
func (p *Ports) Bindings() (nat.PortMap, error) {
	ports := nat.PortMap{}

	for _, spec := range p.Specs {
		port, err := spec.Port()
		if err != nil {
			return nil, err
		}

		_, exists := ports[port]
		if !exists {
			ports[port] = []nat.PortBinding{}
		}

		ports[port] = append(ports[port], spec.Binding())
	}

	return ports, nil
}

// Add is a shortcut function to add a new port spec.
func (p *Ports) Add(port *Port) {
	p.Specs = append(p.Specs, port)
}

// NewPorts will produces a new *Ports struct that's ready to be
// modified.
func NewPorts() *Ports {
	return &Ports{Specs: []*Port{}}
}
