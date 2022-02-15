package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	canpackets "github.com/Liquid-Propulsion/canpackets/go"
	"github.com/angelodlfrtr/go-can"
)

var port = 8881
var CurrentCANServer *CANServer

func initCAN() {
	log.Printf("Starting TCP Server on :%d", port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	server := new(CANServer)
	server.listener = listener
	server.clientMutex = sync.Mutex{}
	go server.run()
	CurrentCANServer = server
}

type CANServer struct {
	listener    net.Listener
	clientMutex sync.Mutex
	client      net.Conn
}

func (server *CANServer) run() {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			continue
		}
		server.clientMutex.Lock()
		server.client = conn
		server.clientMutex.Unlock()
		go func() {
			jsonDec := json.NewDecoder(server.client)

			for {
				frm := &can.Frame{}

				if err := jsonDec.Decode(frm); err != nil {
					if err == net.ErrClosed || err == io.EOF {
						break
					}

					panic(err)
				}

				server.handleFrame(frm)
			}
		}()
	}
}

func (server *CANServer) handleFrame(frame *can.Frame) {
	switch frame.ArbitrationID {
	case 0x00:
		power_state := canpackets.PowerPacket{}
		power_state.Decode(frame.Data[:])
		fmt.Println(frame.Data)
		CurrentFakeEngine.handlePower(power_state.SystemPowered, power_state.Siren)
	case 0x01:
		stage := canpackets.StagePacket{}
		stage.Decode(frame.Data[:])
		CurrentFakeEngine.handleStage(stage.SolenoidState)
	case 0x04:
		CurrentFakeEngine.ping()
	case 0x06:
		blink := canpackets.BlinkPacket{}
		blink.Decode(frame.Data[:])
		CurrentFakeEngine.handleBlink(blink.NodeId)
	}
}

func (server *CANServer) sendFrame(frame *can.Frame) error {
	if server.client != nil {
		frmBytes, err := json.Marshal(frame)
		if err != nil {
			return err
		}

		frmBytes = append(frmBytes, []byte("\r\n")...)

		if _, err := server.client.Write(frmBytes); err != nil {
			log.Println("error while write to client conn:", err.Error())
			server.client.Close()
			server.clientMutex.Lock()
			server.client = nil
			server.clientMutex.Unlock()
		}
	}
	return nil
}

func (server *CANServer) sendSensorValue(id uint8, value uint32) error {
	packet := canpackets.SensorDataPacket{
		SensorId:   id,
		SensorData: value,
	}
	frame, err := createFrame(0x03, packet.Encode())
	if err != nil {
		return err
	}
	return server.sendFrame(frame)
}

func (server *CANServer) sendPong(id uint8, nodeType canpackets.NodeType) error {
	packet := canpackets.PongPacket{
		NodeId:   id,
		NodeType: nodeType,
	}
	frame, err := createFrame(0x05, packet.Encode())
	if err != nil {
		return err
	}
	return server.sendFrame(frame)
}

func createFrame(id uint32, src []byte) (*can.Frame, error) {
	length := len(src)
	if length > 8 {
		return nil, errors.New("packet too long")
	}
	var dst [8]uint8
	copy(dst[:], src[:length])
	frame := can.Frame{
		ArbitrationID: id,
		DLC:           uint8(length),
		Data:          dst,
	}
	return &frame, nil
}
