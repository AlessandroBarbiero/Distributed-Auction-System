package main

import (
	auction "auctionSystem/grpc"
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strconv"
)

type Client struct {
	id                int64
	serverPorts       []int
	serverConnections []auction.AuctionClient
}

func main() {

	f, err := os.OpenFile("client.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	log.SetOutput(f)

	// Create an empty client
	client := &Client{
		id: -1, // id -1 means the client doesn't have yet an id
	}

	// Store the ports from the DNS_cache.info file
	client.readDnsCache()

	// Connect to the servers
	client.connectToServers()

	// Process client requests
	client.handleRequests()

}

func (c *Client) handleRequests() {

	fmt.Println("Hi there! Welcome to this auction!")
	fmt.Print("Choose an action:\n- Show (s)\n- Bid (b [value])\n")
	// Wait for input in the client terminal
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		log.Printf("Client wrote: %s\n", input)

		switch input {
		case "s": // Show command
			c.Show()
		case "b": // Bid command
			amount := 0 //TODO: Add reading amount from command line
			c.Bid(amount)
		default:
			fmt.Println("Please insert a valid command")
			log.Printf("Client inserted wrong input\n")
		}
		fmt.Print("Choose an action:\n- Show (s)\n- Bid (b [value])\n")
	}
}

func (c *Client) connectToServers() {
	for _, serverPort := range c.serverPorts {
		// Dial the server at the specified port.
		conn, err := grpc.Dial("localhost:"+strconv.Itoa(serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Could not connect to the server with port: %d\n", serverPort)
		} else {
			log.Printf("Connected to the server at port %d\n", serverPort)
			c.serverConnections = append(c.serverConnections, auction.NewAuctionClient(conn))
		}
	}

	return
}

func (c *Client) Show() {
	c.connectToServer()

	showReply, err := c.serverConnection.Show(context.Background(), &auction.ShowRequest{})
	if err != nil {
		log.Printf(err.Error())
	} else {
		log.Printf("Current bid: %s, Current winner: %s, Item: %s, Seconds left: %s\n", showReply.CurrentBid, showReply.WinningClientId, showReply.ObjectName, showReply.SecondsTillEnd)
	}
}

func (c *Client) Bid(amount int64) {
	timeReturnMessage, err := serverConnection.AskForTime(context.Background(), &auction.AskForTimeMessage{
		ClientId: int64(client.id),
	})
}

// Read DNS Cache file to save ports of the servers
func (c *Client) readDnsCache() {
	name := "client/DNS_Cache.info"
	file, err := os.Open(name)
	if err != nil {
		log.Fatalln("Couldn't read file with server addresses")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Read DNS Cache file and save the ports of the servers
	for scanner.Scan() {
		port, e := strconv.Atoi(scanner.Text())
		if e != nil {
			log.Fatalln("Invalid value in DNS cache file")
		}
		c.serverPorts = append(c.serverPorts, port)
	}

}

// Remove the element at index i from s
func removeFromSlice(s []auction.AuctionClient, i int) []auction.AuctionClient {
	if i == len(s) {
		return s[:i]
	}
	return append(s[:i], s[i+1:]...)
}
