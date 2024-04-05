package farmer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
)

// SensorConnection represents a device connection
// that connects to the Farm Server.
type SensorConnection struct {
	message string
	conn    *net.Conn
}

// Farmer represents a server receinving and sending
// messages to/from various sensors.
type Farmer struct {
	listener   net.Listener           // main server listener
	sensors    map[*net.Conn]string   // holds connections from all registered sensors
	register   chan *SensorConnection // register new sensor connection
	unregister chan *net.Conn         // unregistering sensors' connections
	broadcast  chan string            // chan used for sending msg to all connected sensors
}

// NewFarmer takes a string representing network address and creates a new Farm Server.
func NewFarmer(addr string) (*Farmer, error) {
	f := Farmer{
		register:   make(chan *SensorConnection, 1),
		unregister: make(chan *net.Conn, 1),
		sensors:    make(map[*net.Conn]string),
		broadcast:  make(chan string, 20),
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	f.listener = l
	return &f, nil
}

func (f *Farmer) Start() {
	go f.listen()

	for {
		select {
		case s := <-f.register:
			f.sensors[s.conn] = s.message
			f.broadcast <- fmt.Sprintf("Sensor: %s registered\n", s.message)
			go f.handle(s)
		case msg := <-f.broadcast:
			// send message for all connected sensors
			for conn := range f.sensors {
				_, err := fmt.Fprint(*conn, msg)
				if err != nil {
					fmt.Printf("sending message to sensors failed: %v", err)
				}
			}
		case conn := <-f.unregister:
			delete(f.sensors, conn)
			fmt.Printf("sensor unregistered")
		}
	}
}

func (f *Farmer) listen() {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			fmt.Printf("farmer: failure to create connection: %v\n", err)
			continue
		}

		r := bufio.NewReader(conn)
		msg, err := r.ReadString('\n')
		if err != nil {
			fmt.Printf("farmer: failure to read from client: %v\n", err)
		}

		s := &SensorConnection{
			message: msg,
			conn:    &conn,
		}
		f.register <- s
	}
}

func (f *Farmer) handle(s *SensorConnection) {
	r := bufio.NewReader(*s.conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Printf("sensor connection closed")
			}
			f.unregister <- s.conn
			break
		}
		f.broadcast <- fmt.Sprintf("sensor message: %s", msg)
	}
}

func ListenAndServe(addr string) (err error) {
	farmer, err := NewFarmer(addr)
	if err != nil {
		return err
	}
	farmer.Start()
	return nil
}

// Sensor represents a client that connects to
// the Farmer server.
type Sensor struct {
	serialNo string
	conn     net.Conn
}

// Close closes connection between Farmer (Server) and Sensor (client).
func (s *Sensor) Close() {
	s.conn.Close()
}

// Read returns a message the Sensor received from the Farmer Server.
func (s *Sensor) Read() (string, error) {
	r := bufio.NewReader(s.conn)
	msg, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return msg, nil
}

// Send takes a string and sends it to the Farmer (connected server).
func (s *Sensor) Send(msg string) error {
	_, err := fmt.Fprintf(s.conn, "SensorID: %s, Message: %s", s.serialNo, msg)
	if err != nil {
		return err
	}
	return nil
}

// ConnectSensor creates new sensor connection to the Farmer server.
func ConnectSensor(serialNo, addr string) (*Sensor, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	sensor := Sensor{
		serialNo: serialNo,
		conn:     conn,
	}
	return &sensor, nil
}
