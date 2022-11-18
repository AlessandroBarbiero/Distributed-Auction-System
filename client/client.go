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

		switch input[0] {
		case 's': // Show command
			c.Show()
		case 'b': // Bid command
			amountStr := input[2:]
			amount, amErr := strconv.Atoi(amountStr)
			if amErr != nil {
				log.Printf("Client inserted wrong input\n")
				fmt.Println("Please insert a valid command like [b number]")
			} else {
				c.Bid(int64(amount))
			}

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
	msgBack := false
	for i, serverConnection := range c.serverConnections {
		showReply, err := serverConnection.Show(context.Background(), &auction.ShowRequest{})
		if err != nil {
			log.Printf("Server with port %d is down, go for the next one\n", c.serverPorts[i])
		} else {
			if msgBack == false {
				log.Printf("Client %d asked for the status of auction %s\n", c.id, showReply.ObjectName)
				fmt.Printf("Current bid: %d, Current winner: %d, Item: %s, Seconds left: %d\n", showReply.CurrentBid, showReply.WinningClientId, showReply.ObjectName, showReply.SecondsTillEnd)
				msgBack = true
			}
		}
	}
}

func (c *Client) Bid(amount int64) {
	msgBack := false
	var brokenServersIdx []int
	for i, serverConnection := range c.serverConnections {
		bidReply, err := serverConnection.Bid(context.Background(), &auction.BidRequest{
			ClientId: c.id,
			Amount:   amount,
		})

		if err != nil {
			log.Printf("Server with port %d is down, go for the next one\n", c.serverPorts[i])
			brokenServersIdx = append(brokenServersIdx, i)
		} else {
			if msgBack == false {
				if c.id == -1 {
					c.id = bidReply.ClientId
				}
				if bidReply.Success == true {
					log.Printf("Client %d bidded succesfully, new best bid %d\n", c.id, bidReply.BestBid)
					fmt.Printf("Bid successful, new best bid %d\n", bidReply.BestBid)
				} else {
					log.Printf("Client %d tried to bid unsuccesfully\n", c.id)
					fmt.Printf("Bid unsuccessful, current best bid: %d\n", bidReply.BestBid)
				}

				msgBack = true
			}
		}
	}
	for _, idx := range brokenServersIdx {
		c.serverConnections = removeFromSlice(c.serverConnections, idx)
	}

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
