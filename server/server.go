package main

import (
	"auctionSystem/grpc"
	"context"
	"log"
	"math"
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
	port int64
	// Store a progressive number for the Ids of the clients in order to give a univoque id to each client
	idCounter  int64
	clients    []int64
	currentBid HighestBid
	mutex      sync.RWMutex
}

type HighestBid struct {
	clientId  int64
	item      AuctionItem
	startTime time.Time
}

type AuctionItem struct {
	name          string
	bid           int64
	auctionLength int64
}

var auctionItems = [...]AuctionItem{
	{name: "Item1", bid: 50, auctionLength: 20},
	{name: "Item2", bid: 20, auctionLength: 45},
	{name: "Item3", bid: 20, auctionLength: 45},
	{name: "Item4", bid: 20, auctionLength: 45},
	{name: "Item5", bid: 20, auctionLength: 45},
}

// Add this part if we want to use parametric port on call of the method
// var port = flag.Int("port", 0, "server port number")

func main() {
	// Get the port from the command line when the server is run
	// flag.Parse()
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	// Hardcoded port
	port := &arg1

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
	list, err := net.Listen("tcp", ":"+strconv.Itoa(int(server.port)))

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

	go server.updateBids()

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
	for {
		s.mutex.Lock()
		s.currentBid.item = auctionItems[counter%numOfItems]
		s.currentBid.startTime = time.Now()
		log.Printf("Auction for item %v started at %v and lasts %v seconds\n", s.currentBid.item.name, s.currentBid.startTime, s.currentBid.item.auctionLength)
		s.mutex.Unlock()

		time.Sleep(time.Duration(s.currentBid.item.auctionLength) * time.Second)
		counter++
		log.Printf("Auction for item %v ended at %v\n", s.currentBid.item.name, time.Now())
	}
}

func (s *Server) Bid(ctx context.Context, request *auctionSystem.BidRequest) (*auctionSystem.BidReply, error) {
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
		log.Printf("Bid %v from client %v accepted, current bid is %v\n", request.Amount, request.ClientId, bestBid)
	} else {
		success = false
		s.mutex.RLock()
		defer s.mutex.RUnlock()

		bestBid = s.currentBid.item.bid
		log.Printf("Bid %v from client %v declined, current bid is %v\n", request.Amount, request.ClientId, bestBid)
	}

	return &auctionSystem.BidReply{ClientId: id, Success: success, BestBid: bestBid}, nil
}

func (s *Server) Show(ctx context.Context, request *auctionSystem.ShowRequest) (*auctionSystem.ShowReply, error) {
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

	log.Printf("Show request received, current bid %v", currentBid)

	return &auctionSystem.ShowReply{CurrentBid: currentBid, WinningClientId: winningClient, ObjectName: name, SecondsTillEnd: secondsLeft}, nil
}

func (s *Server) getSecondsTillEnd() int64 {
	diff := time.Now().Sub(s.currentBid.startTime)
	return s.currentBid.item.auctionLength - int64(math.Round(diff.Seconds()))
}
