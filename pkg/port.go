package pkg

import (
	"errors"
	"fmt"
	"sync"
)

type port struct {
	sync.Mutex
	list []string
}

var errInvalidPortRange error = errors.New("invalidportrange")
var errNoPortAvailable error = errors.New("noport")

func createPortSequence(min, max int) (*port, error) {
	if min < 1024 || min > max || max > 65534 {
		return nil, errInvalidPortRange
	}
	list := []string{}
	for i := max; i >= min; i-- {
		list = append(list, fmt.Sprintf("%d", i))
	}
	return &port{
		list: list,
	}, nil
}

func (p *port) getPort() (string, error) {
	p.Lock()
	defer p.Unlock()
	if len(p.list) == 0 {
		return "", errNoPortAvailable
	}
	output := p.list[len(p.list)-1]
	p.list = p.list[:len(p.list)-1]
	return output, nil
}

func (p *port) releasePort(port string) {
	p.Lock()
	defer p.Unlock()
	p.list = append(p.list, port)
}
