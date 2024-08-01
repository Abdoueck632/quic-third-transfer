
package main

import (
	"fmt"
	"log"
	"sort"
)

type Relay struct {
	Fenetre     []int
	BdwR        int
	NumberAckR  int
	PartR       int
	Name        string
}

func (r *Relay) ReceivedAck(fenetre *[]int, number int) {
	r.NumberAckR++
	for i, v := range *fenetre {
		if v == number {
			*fenetre = append((*fenetre)[:i], (*fenetre)[i+1:]...)
			break
		}
	}
}

func (r *Relay) CalculPart(sizeFrame, sumAck int) int {
	r.PartR = int(r.NumberAckR * sizeFrame / sumAck)
	r.NumberAckR = 0
	return r.PartR
}

func (r *Relay) SetPart(newPart int) {
	r.PartR = newPart
}

func (r *Relay) SetNumberAck(newNumber int) {
	r.NumberAckR = newNumber
}

type Motor struct {
	Fenetre      []int
	Size         int
	SizeData     int
	DataToSend   []int
	Cluters      int
	NewFenetre   []int
	Relays       []*Relay
	NotFull      int
	PertPuissance int
	IsSync       bool
}

func NewMotor(size, sizeData int, relays []*Relay, isSync bool) *Motor {
	return &Motor{
		Size:       size,
		SizeData:   sizeData,
		DataToSend: tabIntGenerate(sizeData),
		Cluters:    0,
		NewFenetre: []int{},
		Relays:     relays,
		NotFull:    0,
		PertPuissance: 0,
		IsSync:     isSync,
	}
}



func (m *Motor) SumAck() int {
	sum := 0
	for _, relay := range m.Relays {
		sum += relay.NumberAckR
	}
	return sum
}

func (m *Motor) CalculParts() {
	index := 0
	sumPart := 0
	somme := m.SumAck()
	for index < len(m.Relays)-1 {
		sumPart += m.Relays[index].CalculPart(m.Size, somme)
		index++
	}
	m.Relays[index].SetPart(m.Size - sumPart)
	for _, relay := range m.Relays {
		relay.SetNumberAck(0)
		m.PrintAcksRelays()
	}
}

func (m *Motor) SupplyData() {
	if len(m.Fenetre) < m.Size {
		if len(m.Fenetre) == 0 {
			m.Fenetre = m.DataToSend[m.Cluters:m.Cluters+m.Size]
			m.Cluters += m.Size
			m.NewFenetre = append(m.NewFenetre, m.Fenetre...)
			return
		}
		for m.Cluters < len(m.DataToSend) && len(m.Fenetre) < m.Size && (m.DataToSend[m.Cluters]-m.Fenetre[0] < m.Size) {
			m.Fenetre = append(m.Fenetre, m.DataToSend[m.Cluters])
			m.NewFenetre = append(m.NewFenetre, m.DataToSend[m.Cluters])
			m.Cluters++
		}
	}
}

func (m *Motor) SendDataRelay(partR int) []int {
	iteration := 0
	bufferR := []int{}
	for len(m.NewFenetre) > 0 && iteration < partR {
		bufferR = append(bufferR, m.NewFenetre[0])
		iteration++
		m.NewFenetre = m.NewFenetre[1:]
	}
	return bufferR
}

func (m *Motor) EtatDesRelais() bool {
	for _, relay := range m.Relays {
		if len(relay.Fenetre) != 0 {
			return false
		}
	}
	return true
}

func (m *Motor) SortRelay() []*Relay {
	if len(m.Relays) == 1 {
		return m.Relays
	}
	sort.Slice(m.Relays, func(i, j int) bool {
		return m.Relays[i].PartR > m.Relays[j].PartR
	})
	m.NotFull++
	return m.Relays
}

func (m *Motor) GetMaxFenetreRSize() int {
	maxRelay := m.Relays[0]
	for _, relay := range m.Relays {
		if relay.BdwR > maxRelay.BdwR {
			maxRelay = relay
		}
	}
	return maxRelay.BdwR
}

func (m *Motor) SendDataRelays() {
	for _, relay := range m.Relays {
		relay.Fenetre = append(relay.Fenetre, m.SendDataRelay(relay.PartR)...)
	}
}

func (m *Motor) DistributeDataToRelays() {
	m.SupplyData()
	m.Relays = m.SortRelay()
	m.SendDataRelays()
}

func (m *Motor) ReceivedAck(relay *Relay, number int) {
	relay.ReceivedAck(&m.Fenetre, number)
}

func (m *Motor) PrintBWRelays() {
	for _, relay := range m.Relays {
		log.Printf("BW %s: %d", relay.Name, relay.BdwR)
	}
}

func (m *Motor) PrintFenetreRelays() {
	for _, relay := range m.Relays {
		log.Printf("Fenetre d'envoi %s: %v", relay.Name, relay.Fenetre)
	}
}

func (m *Motor) PrintPartsRelays() {
	for _, relay := range m.Relays {
		log.Printf("Part de %s: %d", relay.Name, relay.PartR)
	}
}

func (m *Motor) PrintAcksRelays() {
	for _, relay := range m.Relays {
		log.Printf("%s Nombre de données envoyées: %d", relay.Name, relay.NumberAckR)
	}
}

type FrameQueue struct {
	Tab  []int
	Size int
}

func NewFrameQueue(size int) *FrameQueue {
	return &FrameQueue{
		Tab:  []int{},
		Size: size,
	}
}

func (fq *FrameQueue) CheckSize() bool {
	return len(fq.Tab) < fq.Size
}

func (fq *FrameQueue) AddInOrder(fileClient []int) []int {
	if len(fq.Tab) == 0 {
		return fileClient
	}
	for {
		lastElement := 0
		if len(fileClient) > 0 {
			lastElement = fileClient[len(fileClient)-1]
		}
		nextIndex := -1
		for i, v := range fq.Tab {
			if v == lastElement+1 {
				nextIndex = i
				break
			}
		}
		if nextIndex == -1 {
			break
		}
		fileClient = append(fileClient, fq.Tab[nextIndex])
		fq.Tab = append(fq.Tab[:nextIndex], fq.Tab[nextIndex+1:]...)
	}
	return fileClient
}

func (fq *FrameQueue) AddElements(motor *Motor, relay *Relay) int {
	numberAdd := 0
	if len(relay.Fenetre) == 0 && motor.Cluters < len(motor.DataToSend) {
		log.Println("--> Synchronisation")
		motor.PrintAcksRelays()
		if motor.IsSync {
			motor.CalculParts()
		}
		log.Println("--> Calcul des parts")
		motor.PrintPartsRelays()
		motor.DistributeDataToRelays()
		log.Println("--> Distribuer les données aux relais")
		log.Printf("Fenetre d'envoi du moteur %v", motor.Fenetre)
		motor.PrintFenetreRelays()
	}
	if fq.CheckSize() && len(relay.Fenetre) > 0 {
		fq.AddInFrameQueue(relay.Fenetre[0])
		numberAdd = 1
		relay.ReceivedAck(&motor.Fenetre, relay.Fenetre[0])
		relay.Fenetre = relay.Fenetre[1:]
	}
	motor.PertPuissance += relay.BdwR - numberAdd
	return numberAdd
}

func (fq *FrameQueue) Recept(motor *Motor) {
	count := 0
	maxSize := motor.GetMaxFenetreRSize()
	for count < maxSize {
		for i := 0; i < len(motor.Relays); i++ {
			if motor.Relays[i].BdwR > count {
				fq.AddElements(motor, motor.Relays[i])
			}
		}
		count++
	}
}

func (fq *FrameQueue) AddInFrameQueue(element int) {
	fq.Tab = append(fq.Tab, element)
}

func SendSync(frameQueue *FrameQueue, motor *Motor, dataReceived *[]int) map[string]int {
	motor.DistributeDataToRelays()
	log.Println("--------------------------------------")
	log.Println("Résumé")
	log.Println("--------")
	log.Printf("Données à envoyer %v", motor.DataToSend)
	log.Printf("Fenetre du moteur