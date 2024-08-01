package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	quic "github.com/Abdoueck632/mp-quic"
	"github.com/Abdoueck632/quic-third-transfer/config"
	"github.com/Abdoueck632/quic-third-transfer/utils"
)

// Tuple pour représenter les plages d'octets
type Range struct {
	start int
	end   int
}
type Streams struct {
	id     int
	stream quic.Stream
}
type fenetre struct {
	id_relais int
	donnees   int
	is_ack    bool
}
type relais struct {
	id         int
	part       int
	nombre_ack int
	stream     quic.Stream
	pathId     int
}
type Monitor struct {
	curseur        int
	relais         []relais
	fenetre        []fenetre
	nombre_relais  int
	ackPaquets     []config.Ack
	basePort       int
	taille_donnees int
	donnees        []int
	taile_fenetre  int
	sync           bool
}

var taille_fenetre = 200
var ackPaquets = []config.Ack{}

func main() {

	// Vérifier qu'au moins un argument (adresse de relais) a été passé
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <AddrRelay1> <AddrRelay2> ...")
		return
	}

	// Récupérer les adresses des relais passées en argument
	AddrRelays := os.Args[1:]
	dataMigration := config.DataMigration{}
	filename := make([]byte, 64)

	listener, err := quic.ListenAddr(config.Addr, utils.GenerateTLSConfig(), config.QuicConfig)
	utils.HandleError(err)
	f, err := os.Create("serveur_SSLKEYLOGFILE.bin")
	if err != nil {
		utils.HandleError(err)
	} else {
		defer f.Close()
	}

	fmt.Println("Server started! Waiting for streams from client ...")

	sess, err := listener.Accept()
	utils.HandleError(err)

	stream, err := sess.AcceptStream()
	utils.HandleError(err)

	fmt.Println("session created : ", sess.RemoteAddr())

	//read filename of the client
	stream.Read(filename)
	//sess.ClosePath(0)

	dataMigration.FileName = strings.Trim(string(filename), ":")

	if err != nil {
		log.Fatalf("loadDerivedKeys: %s", err)
	}
	//time.Sleep(10 * time.Second)
	//send to the first server relay
	for {
		if sess.GetLenPaths() == 2 {
			break
		}
	}
	lines, err := loadDerivedKeys("/derivateK.in.json")
	dataMigration.CrytoKey = lines
	name := "./storage-server/" + dataMigration.FileName
	file, err := os.Open(name)

	fileInfo, err := file.Stat()

	fileSize := utils.FillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := utils.FillString(fileInfo.Name(), 64)
	stream.Write([]byte(fileSize))
	stream.Write([]byte(fileName))
	/*seuil := 100000
	result := splitWindow(0, seuil, config.BUFFERSIZE, 2)
	fmt.Println(result)*/
	//dataMigration.TabBuffer = result[0]
	dataMigration.WritteOffset = 74

	monitor := Monitor{
		nombre_relais:  len(AddrRelays),
		basePort:       4242,
		taile_fenetre:  taille_fenetre,
		donnees:        tabIntGenerate(int(fileInfo.Size()), 0, config.BUFFERSIZE),
		taille_donnees: int(fileInfo.Size()),
	}

	monitor.fenetre = generateFenetreList(monitor.donnees[:taille_fenetre])
	monitor.curseur = taille_fenetre

	for i, addr := range AddrRelays {
		relais := relais{
			pathId: 2 * (i + 1),
			id:     i + 1,
			part:   100,
		}
		monitor.relais = append(monitor.relais, relais)
		fmt.Printf("Relais %d: pathId=%d, id=%d, addr=%s\n", i, relais.pathId, relais.id, addr)
	}

	result := attribuerDonneesAuxRelais(&monitor)
	dataMigration.FileSize = monitor.taille_donnees
	for i, addr := range AddrRelays {
		dataMigration.IdRelay++
		fmt.Printf("Relay %d: %s\n", i+1, addr)
		dataMigration.TabBuffer = result[i]
		SendRelayData(addr, dataMigration, sess, stream, monitor.relais[i].pathId)
		//dataMigration.TabBuffer = result[1]
		dataMigration.WritteOffset += config.BUFFERSIZE
		dataMigration.StartAt = config.BUFFERSIZE

	}

	//SendRelayData(AddrRelay2, dataMigration, sess, stream, 4)
	for i := 0; i < len(AddrRelays); i++ {
		port := monitor.basePort + i + 1
		addr := fmt.Sprintf("0.0.0.0:%d", port)
		_, stream := acceptConnectionServer(addr)
		fmt.Printf("Accepted connection on %s for relay %d\n", addr, i+1)
		monitor.relais[i].stream = stream

	}

	/*_, monitor.relais[0].stream = acceptConnectionServer("0.0.0.0:4243")
	_, monitor.relais[1].stream = acceptConnectionServer("0.0.0.0:4244")
	//stream.CancelRead()
	/*streams := []Streams{
		{id: 1, stream: stream1},
		{id: 2, stream: stream2},
	}*/
	//sess.ClosePath(0)
	//sess.ClosePath(1)
	//s1.Cancel(err)
	//s2.Cancel(err)
	//ranges := calculateRanges(int(fileInfo.Size()), 100000, 80*config.BUFFERSIZE)
	fmt.Printf("%+v", monitor)

	var wg sync.WaitGroup
	wg.Add(len(monitor.relais))
	for i, r := range monitor.relais {

		fmt.Printf("Lancement de la goroutine pour le relais %d\n", i)

		go readFromStream(r, &monitor, &wg)
	}
	wg.Wait()
	fmt.Println("All elements processed")

	//var data = make([]byte, 30)

	//size1, _ := stream1.Read(data)
	//if size1 > 0 {
	//seuil += 40000
	//result = splitWindow(seuil-40000, seuil, config.BUFFERSIZE, 2)
	//dataMigration.TabBuffer = result[0]
	//fmt.Println(result)
	/*
		size2, _ := stream2.Read(data)
		if size2 > 0 {
			seuil += 40000
			result = splitWindow(seuil-40000, seuil, config.BUFFERSIZE, 2)
			dataMigration.TabBuffer = result[0]
			fmt.Println(result)
			ack, err := json.Marshal(config.PlageBuffer{TabBuffer: result[0]})
			if err != nil {
				log.Fatal(err)
			}
			stream1.Write([]byte(utils.FillString(string(ack), 30)))
			ack, err = json.Marshal(config.PlageBuffer{TabBuffer: result[0]})

			stream2.Write([]byte(utils.FillString(string(ack), 30)))
		}*/

}
func sendMessage(monitor *Monitor, result [][]int) {
	for i, r := range monitor.relais {
		var ack []byte
		if monitor.curseur >= len(monitor.donnees) {
			for _, r := range monitor.relais {
				data := make([]byte, 30)
				data = []byte(utils.FillString("Finish", 30))
				r.stream.Write(data)
			}
			return
		}
		//if monitor.curseur+r.part > len(monitor.donnees) {
		ack, err := json.Marshal(config.PlageBuffer{TabBuffer: result[i]})
		//} else {
		//ack, err = json.Marshal(config.PlageBuffer{TabBuffer: result[i]})
		//}

		//fmt.Printf("%d R: [%d,%d] ", r.id, result[i][0], result[i][1])
		r.stream.Write([]byte(utils.FillString(string(ack), 30)))
		if err != nil {
			log.Fatal(err)
		}
		/*
			if i+1 < len(ranges) {

				ack1, _ := json.Marshal(config.PlageBuffer{TabBuffer: []int{ranges[i+1].start, ranges[i+1].end}})

				streams[1].stream.Write([]byte(utils.FillString(string(ack1), 100)))
			}*/
	}
}
func verifierEtMettreAJourAck(m *Monitor, ackPaquet config.Ack) bool {
	for i := 0; i < len(m.fenetre); i++ {
		if m.fenetre[i].donnees == int(ackPaquet.Offset-74) {
			m.fenetre[i].is_ack = true
			m.relais[ackPaquet.IdRelay-1].nombre_ack += 1
			fmt.Println("id relais ", m.relais[ackPaquet.IdRelay-1])
			return true
		}
	}
	return false
}

// Fonction pour avancer la fenêtre
func avancerFenetre(m *Monitor) int {
	cpt := 0
	// Si le premier élément de la fenêtre a été accusé de réception
	for m.fenetre[0].is_ack {

		// Déplacer tous les éléments de la fenêtre d'une position vers la gauche
		for j := 0; j < m.taile_fenetre-1; j++ {
			m.fenetre[j] = m.fenetre[j+1]
		}
		// Ajouter un nouveau paquet de données à la fin de la fenêtre
		if m.curseur < len(m.donnees) {
			m.fenetre[m.taile_fenetre-1] = fenetre{
				id_relais: -1,
				donnees:   m.donnees[m.curseur],
				is_ack:    false,
			}
			m.curseur++
		}
	}
	for i := 0; i < len(m.fenetre); i++ {
		if m.fenetre[i].id_relais == -1 {
			cpt++
		}
	}
	return cpt
}
func generateFenetreList(donnees []int) []fenetre {
	fenetreList := make([]fenetre, len(donnees))

	for i, val := range donnees {
		fenetreList[i] = fenetre{
			id_relais: -1,
			donnees:   val,
			is_ack:    false,
		}
	}

	return fenetreList
}
func attribuerDonneesAuxRelais(m *Monitor) [][]int {
	plages := make([][]int, len(m.relais))

	for i := 0; i < len(m.relais); i++ {
		// Initialiser la plage pour chaque relais
		plages[i] = []int{-1, 0} // [-1, 0] signifie qu'aucune plage n'a été attribuée

		// Initialiser un compteur pour la capacité restante de ce relais
		capaciteRestante := m.relais[i].part
		for j := 0; j < m.taile_fenetre; j++ {
			if m.fenetre[j].id_relais == -1 && capaciteRestante > 0 {
				// Attribuer le relais à la fenêtre
				m.fenetre[j].id_relais = m.relais[i].id

				// Définir le début de la plage si ce n'est pas encore fait
				if plages[i][0] == -1 {
					plages[i][0] = m.fenetre[j].donnees
				}
				// Toujours mettre à jour la fin de la plage
				plages[i][1]++

				// Décrémenter la capacité restante
				capaciteRestante--
				fmt.Printf("%d \n", capaciteRestante)
			}
		}
	}
	fmt.Printf("Plages des attributions: %+v\n", plages)
	return plages
}
func tabIntGenerate(sizeData int, def_offset int, buffer int) []int {
	if sizeData <= 0 {
		return []int{}
	}

	// Calculez la taille nécessaire pour le tableau
	var size int
	if sizeData > def_offset {
		size = (sizeData - def_offset + buffer - 1) / buffer
	} else {
		size = 0
	}

	// Créez le tableau avec la taille calculée
	nombres := make([]int, size)

	for i, j := def_offset, 0; i < sizeData && j < size; i, j = i+buffer, j+1 {
		nombres[j] = i
	}

	return nombres
}

// calculPart calcule le ratio des offsets reçus par chaque relais
func calculPart(m *Monitor, capacite_reste int) {
	// Utiliser une map pour compter le nombre d'offsets reçus par chaque relais

	totalAcks := 0

	// Parcourir les ackPaquets et compter les offsets par relais
	for i, r := range m.relais {
		totalAcks += r.nombre_ack
		fmt.Printf("nombre ack r %d  %d \n", i+1, r.nombre_ack)
	}
	fmt.Printf("Capacité restante %d \n", capacite_reste)
	fmt.Printf("total ack %d \n", totalAcks)

	// Remplir la slice avec les ratios
	for j := 0; j < len(m.relais); j++ {
		m.relais[j].part = int(float64(capacite_reste) * float64(m.relais[j].nombre_ack) / float64(totalAcks))
		fmt.Printf("part %d \n", m.relais[j].part)

	}

}

func toutesPlagesDefinies(plages [][]int) bool {
	for _, plage := range plages {
		if plage[1] == 0 {
			return false
		}
	}
	return true
}

// Fonction pour lire les données d'un flux et vérifier le message "Sync"
func readFromStream(r relais, m *Monitor, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Goroutine pour le relais %d démarrée\n", r.id)
	for {

		buffer := make([]byte, 30)
		for {
			// Lire les données dans le buffer

			n, err := r.stream.Read(buffer)
			if err != nil {
				if err != io.EOF {
					fmt.Println("Error reading from stream:", err)
				}

			}

			// Afficher les données lues
			if n > 0 {
				//fmt.Printf("Read %d bytes: %s\n", n, string(buffer[:n]))
				config.WriteFile("exemple.txt", string(buffer[:n]))
				traitementMessage(buffer, m, r)

			}
		}

	}
}
func traitementMessage(buffer []byte, m *Monitor, r relais) {
	message := strings.Trim(string(buffer), ":")
	config.WriteFile("serverLog.txt", message)
	fmt.Printf("Received from relay %d: %s Thioune", r.id, message)
	if message == "Sync" {
		config.WriteFile("exemple1.txt", fmt.Sprintf(" \n Avant avacement de la fenetre %v", m.fenetre))
		cpt_new_element := avancerFenetre(m)
		fmt.Printf("Hello, %+v", m.fenetre)
		config.WriteFile("exemple.txt", fmt.Sprintf(" \n Apres avacement de la fenetre %v", m.fenetre))
		//fmt.Printf("Hello, %+v", cpt_new_element)

		calculPart(m, cpt_new_element)

		result := attribuerDonneesAuxRelais(m)

		if toutesPlagesDefinies(result) == true {
			fmt.Println("Toutes les plages ne sont pas définies, réattribution...")

		} else {
			fmt.Println("Toutes les plages sont définies.")
			for i := 0; i < len(m.relais); i++ {
				m.relais[i].nombre_ack = 0
			}
			sendMessage(m, result)

		}
		// implement finish to quit program
		fmt.Println(m.ackPaquets)

	} else {
		//js := strings.Trim(string(message), ":")
		ackPaquet := config.Ack{}
		err := json.Unmarshal([]byte(message), &ackPaquet)
		if err != nil {
			log.Printf("Error unmarshalling message from relay %d: %v\n", r.id, err)

		}
		verifierEtMettreAJourAck(m, ackPaquet)
		m.ackPaquets = append(m.ackPaquets, ackPaquet)
		fmt.Println(m.ackPaquets)

	}
}

// Fonction pour afficher les répartitions des lots entre les serveurs
func displayDistribution(ranges []Range) {
	for i := 0; i < len(ranges); i += 2 {
		fmt.Printf("%d R1: [%d,%d] ", (i/2)+1, ranges[i].start, ranges[i].end)
		if i+1 < len(ranges) {
			fmt.Printf("R2: [%d,%d]\n", ranges[i+1].start, ranges[i+1].end)
		} else {
			fmt.Println()
		}
	}
}
func splitWindow(debut, windowSize, bufferSize, numRelays int) [][]int {
	halfWindowSize := (debut + windowSize) / 2 // Calculer la moitié de la fenêtre
	relays := make([][]int, numRelays)

	// Pour le premier relais
	relays[0] = []int{debut, halfWindowSize - bufferSize}

	// Pour le deuxième relais
	relays[1] = []int{halfWindowSize, windowSize - bufferSize}

	return relays
}
func acceptConnectionServer(addr string) (quic.Session, quic.Stream) {

	// Écoute des connexions entrantes
	listener, err := quic.ListenAddr(addr, utils.GenerateTLSConfig(), config.QuicConfig)
	utils.HandleError(err)
	fmt.Println("Server started! Waiting for streams from client...")

	// Acceptation d'une nouvelle session
	sess, err := listener.Accept()
	utils.HandleError(err)
	fmt.Println("Session created: ", sess.RemoteAddr())

	// Acceptation d'un nouveau flux dans la session
	stream, err := sess.AcceptStream()
	utils.HandleError(err)
	fmt.Println("Stream created: ", stream.StreamID())
	fmt.Println("Connected to server, start receiving the file name and file size")
	var data = make([]byte, 10)
	stream.Read(data)
	return sess, stream
}
func SendRelayData(relayaddr string, dataMigration config.DataMigration, sess quic.Session, stream quic.Stream, idpath int) quic.Stream {

	dataMigration.IpAddr = fmt.Sprintf("%v", sess.RemoteAddrById(1))

	dataMigration.Once, dataMigration.Obit, dataMigration.Id = sess.GetCryptoSetup().GetOncesObitID()
	dataMigration.RelayNumber = 2
	sessServer, err := quic.DialAddr(relayaddr, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("session created with secondary server: ", sessServer.RemoteAddr())

	streamServer, err := sessServer.OpenStream()
	utils.HandleError(err)

	dataMigration.IpAddr = utils.FillString(dataMigration.IpAddr, 20)

	dataMigration.FileName = utils.FillString(dataMigration.FileName, 64) // par defaut fileInfo.Name()import socket
	dataMigration.IdPathToCreate = idpath
	//fmt.Println("session created: ", sess.RemoteAddr())

	fmt.Println("stream created...")
	fmt.Println("Client connected")

	if verifyOrder(sess, dataMigration.CrytoKey[2]) != true {
		fmt.Println("False in verification")
		dataMigration.CrytoKey[0], dataMigration.CrytoKey[1] = inverseByte(dataMigration.CrytoKey[0], dataMigration.CrytoKey[1])
		dataMigration.CrytoKey[2], dataMigration.CrytoKey[3] = inverseByte(dataMigration.CrytoKey[2], dataMigration.CrytoKey[3])
	}

	dataByte, err := json.Marshal(dataMigration)
	if err != nil {
		log.Fatal(err)
	}

	streamServer.Write([]byte(utils.FillString(string(dataByte), 1000)))
	fmt.Println("%+v", dataMigration)
	sess.SetIPAddress("172.10.15.56:4242", 0)
	return streamServer

}
func verifyOrder(sess quic.Session, otherIV []byte) bool {
	forw, _, _ := sess.GetCryptoSetup().GetAEADs()
	if bytes.Equal(forw.GetOtherIV(), otherIV) == true {
		return true
	}
	return false

}
func inverseByte(first, second []byte) ([]byte, []byte) {
	return second, first
}

// writeLines writes the lines to the given file.
func saveDerivedKeys(data [][]byte, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Parcourt le tableau et écrit chaque élément dans le fichier
	dataByte, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write(dataByte)

	return err
}

func loadDerivedKeys(path string) ([][]byte, error) {
	datas, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	//defer datas.Close()

	// Sépare le fichier en lignes
	var derivedKeys [][]byte
	json.Unmarshal(datas, &derivedKeys)

	fmt.Printf("%v\n", derivedKeys)
	return derivedKeys, nil
}
func stringTobytes(line string) []byte {
	return []byte(line)
}
func stringTobytes2(tab []string) [][]byte {
	var s [][]byte
	for _, mybte := range tab {
		s = append(s, stringTobytes(mybte))
	}
	fmt.Println(s)
	return s
}
func convertStringSliceToByteSliceSlice(s []string) [][]byte {
	var result [][]byte
	for _, str := range s {
		var bytes []byte
		for _, r := range []rune(str) {
			buf := make([]byte, utf8.RuneLen(r))
			utf8.EncodeRune(buf, r)
			bytes = append(bytes, buf...)
		}
		result = append(result, bytes)
	}
	return result
}

// Fonction pour calculer la plage d'octets pour chaque serveur
func calculateRanges(fileSize, start, lotSize int) []Range {
	var ranges []Range
	for i := start; i < fileSize; i += lotSize {
		end := i + lotSize
		if end > fileSize {
			end = fileSize
		}
		ranges = append(ranges, Range{i, end})
	}
	return ranges
}
