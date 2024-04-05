package farmer_test

import (
	"net"
	"strconv"
	"testing"

	"github.com/qba73/farmer"
)

func TestFarmerConnection(t *testing.T) {
	t.Parallel()

	errChan := make(chan error)
	p := randomFreePort()
	addr := net.JoinHostPort("localhost", strconv.Itoa(p))

	go func() {
		errChan <- farmer.ListenAndServe(addr)
	}()

	_, err := farmer.ConnectSensor("Temp-02", addr)
	if err != nil {
		t.Fatalf("Sensor can't establish connection: %v", err)
	}
}

func TestSensorSendsMessageOnSuccesfullConnectionToFarmServer(t *testing.T) {
	t.Parallel()

	port := randomFreePort()
	addr := net.JoinHostPort("localhost", strconv.Itoa(port))

	errChan := make(chan error)
	go func() {
		errChan <- farmer.ListenAndServe(addr)
	}()

	sensor, err := farmer.ConnectSensor("Temp-01", addr)
	if err != nil {
		t.Fatal(err)
	}
	err = sensor.Send("10C" + "\n")
	if err != nil {
		t.Fatal(err)
	}

	got, err := sensor.Read()
	if err != nil {
		t.Fatal(err)
	}
	want := "Sensor: SensorID: Temp-01, Message: 10C\n"
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func randomFreePort() int {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	tcpAddr, err := net.ResolveTCPAddr("tcp", l.Addr().String())
	if err != nil {
		panic(err)
	}
	return tcpAddr.Port
}
