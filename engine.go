package main

import (
	"math"
	"time"

	canpackets "github.com/Liquid-Propulsion/canpackets/go"
	"github.com/spf13/cast"
)

var CurrentFakeEngine *FakeEngine

type FakeEngine struct {
	Sensors       map[uint8]Sensor
	Nodes         map[uint8]Node
	SolenoidState [64]bool
	lastPowerTime time.Time
	lastStageTime time.Time
	Power         bool
	Siren         bool
}

type Sensor struct {
	ID        uint8
	BaseValue uint
	BaseRange uint
}

type Node struct {
	ID        uint8
	NodeType  canpackets.NodeType
	BlinkTime time.Time
}

func initEngine() {
	engine := new(FakeEngine)
	engine.Sensors = make(map[uint8]Sensor)
	engine.Nodes = make(map[uint8]Node)
	go engine.run()
	CurrentFakeEngine = engine
}

func (engine *FakeEngine) run() {
	for {
		for id, sensor := range engine.Sensors {
			value, err := cast.ToUint32E(math.Sin(float64(time.Now().Nanosecond()*int(sensor.BaseRange)) + float64(sensor.BaseValue)))
			if err != nil {
				continue
			}
			CurrentCANServer.sendSensorValue(id, value)
		}
		if time.Since(engine.lastPowerTime) > time.Millisecond*42 {
			engine.Power = false
			engine.Siren = false
		}
		if time.Since(engine.lastStageTime) > time.Millisecond*42 {
			engine.SolenoidState = [64]bool{}
		}
		time.Sleep(time.Millisecond * 20)
	}
}

func (engine *FakeEngine) ping() {
	for id, node := range engine.Nodes {
		CurrentCANServer.sendPong(id, node.NodeType)
	}
}

func (engine *FakeEngine) addSensor(id int, baseValue uint, baseRange uint) error {
	safeID, err := cast.ToUint8E(id)
	if err != nil {
		return err
	}
	engine.Sensors[safeID] = Sensor{
		ID:        safeID,
		BaseValue: baseValue,
		BaseRange: baseRange,
	}
	return nil
}

func (engine *FakeEngine) removeSensor(id int) error {
	safeID, err := cast.ToUint8E(id)
	if err != nil {
		return err
	}
	delete(engine.Sensors, safeID)
	return nil
}

func (engine *FakeEngine) addNode(id int, nodeType canpackets.NodeType) error {
	safeID, err := cast.ToUint8E(id)
	if err != nil {
		return err
	}
	engine.Nodes[safeID] = Node{
		ID:       safeID,
		NodeType: nodeType,
	}
	return nil
}

func (engine *FakeEngine) removeNode(id int) error {
	safeID, err := cast.ToUint8E(id)
	if err != nil {
		return err
	}
	delete(engine.Nodes, safeID)
	return nil
}

func (engine *FakeEngine) handleBlink(id uint8) {
	if node, ok := engine.Nodes[id]; ok {
		node.BlinkTime = time.Now()
		engine.Nodes[id] = node
	}
}

func (engine *FakeEngine) handlePower(power bool, siren bool) {
	engine.lastPowerTime = time.Now()
	engine.Power = power
	engine.Siren = siren
}

func (engine *FakeEngine) handleStage(solenoidState [64]bool) {
	engine.lastStageTime = time.Now()
	engine.SolenoidState = solenoidState
}
