package pkg

import "testing"

func Test_createPortSequenceAndFailure(t *testing.T) {
	_, err := createPortSequence(0, 10)
	if err != errInvalidPortRange {
		t.Error("port rangs should be invalid")
	}
}

func Test_createPortSequenceAndSuccess(t *testing.T) {
	port, err := createPortSequence(8090, 8100)
	if err != nil {
		t.Error("port sequence should be created")
		t.FailNow()
	}
	if len(port.list) != 11 {
		t.Error("there should be 10 ports")
	}
}

func Test_getPortAndSuccess(t *testing.T) {
	port, err := createPortSequence(8090, 8090)
	if err != nil {
		t.Error("port sequence should be created")
		t.FailNow()
	}
	p, _ := port.getPort()
	if p != "8090" {
		t.Error("port sequence should be 8090")
	}
}

func Test_getPortAndFail(t *testing.T) {
	port := &port{
		list: []string{},
	}
	_, err := port.getPort()

	if err != errNoPortAvailable {
		t.Error("getPort should fail with errNoPortAvailable")
	}
}

func Test_releasePort(t *testing.T) {
	port := &port{
		list: []string{},
	}
	port.releasePort("8090")
	p, _ := port.getPort()
	if p != "8090" {
		t.Error("port sequence should be 8090")
	}
}
