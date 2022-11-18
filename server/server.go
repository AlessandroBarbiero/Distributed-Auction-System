package main

import (
	"auctionSystem/grpc"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type Server struct {
	auctionSystem.UnimplementedAuctionServer
	name string
	port int
	// Store a progressive number for the Ids of the clients in order to give a univoque id to each client
	idCounter  int64
	clients    []int64
	currentBid HighestBid
	mutex      sync.RWMutex
}

type HighestBid struct {
	clientId      int64
	item          AuctionItem
	auctionLength int64
	startTime     time.Time
}

type AuctionItem struct {
	name string
	bid  int64
}

var auctionItems = [...]AuctionItem{
	{name: "Item1", bid: 50},
	{name: "Item2", bid: 20},
}

// Add this part if we want to use parametric port on call of the method
// var port = flag.Int("port", 0, "server port number")

func main() {
	// Get the port from the command line when the server is run
	// flag.Parse()

	// Hardcoded port
	port_value := 8080
	port := &port_value

	// Create a server struct
	server := &Server{
		name:      "serverName",
		port:      *port,
		idCounter: 0,
		clients:   make([]int64, 4),
		mutex:     sync.RWMutex{},
	}

	// Start the server
	startServer(server)
}

func startServer(server *Server) {
	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	list, err := net.Listen("tcp", ":"+strconv.Itoa(server.port))

	if err != nil {
		log.Fatalf("Could not create the server %v", err)
	}

	f, err := os.OpenFile("server.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	log.SetOutput(f)

	log.Printf("Started server at port: %d\n", server.port)

	// Register the grpc server and serve its listener
	auctionSystem.RegisterAuctionServer(grpcServer, server)
	serveError := grpcServer.Serve(list)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

func (s *Server) updateBids() {
	numOfItems := len(auctionItems)
	counter := 0

	minAuctionLenght := 20
	maxAuctionLength := 60
	for {
		s.mutex.Lock()
		s.currentBid.item = auctionItems[counter%numOfItems]
		s.currentBid.startTime = time.Now()
		s.currentBid.auctionLength = int64(rand.Intn(maxAuctionLength-minAuctionLenght) + minAuctionLenght)
		s.mutex.Unlock()

		time.Sleep(time.Duration(s.currentBid.auctionLength) * time.Second)
		counter++
	}
}

func (s *Server) Bid(request auctionSystem.BidRequest) auctionSystem.BidReply {
	var id int64
	var success bool
	var bestBid int64
	log.Printf("Received bid request\n")
	if request.ClientId == -1 {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		s.idCounter++
		id = s.idCounter
		s.clients = append(s.clients, id)
	} else {
		id = request.ClientId
	}

	if request.Amount > s.currentBid.item.bid {
		success = true
		s.mutex.Lock()
		defer s.mutex.Unlock()

		s.currentBid.item.bid = request.Amount
		s.currentBid.clientId = request.ClientId
		bestBid = request.Amount
	} else {
		success = false
		s.mutex.RLock()
		defer s.mutex.RUnlock()

		bestBid = s.currentBid.item.bid
	}

	return auctionSystem.BidReply{ClientId: id, Success: success, BestBid: bestBid}
}

func (s *Server) Show(request auctionSystem.ShowRequest) auctionSystem.ShowReply {
	var secondsLeft int64
	var currentBid int64
	var winningClient int64
	var name string

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	name = s.currentBid.item.name
	winningClient = s.currentBid.clientId
	currentBid = s.currentBid.item.bid
	secondsLeft = s.getSecondsTillEnd()

	return auctionSystem.ShowReply{CurrentBid: currentBid, WinningClientId: winningClient, ObjectName: name, SecondsTillEnd: secondsLeft}
}

func (s *Server) getSecondsTillEnd() int64 {
	diff := time.Now().Sub(s.currentBid.startTime)
	return s.currentBid.auctionLength - int64(math.Round(diff.Seconds()))
}
