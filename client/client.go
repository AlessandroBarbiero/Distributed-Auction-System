package auctionSystem

import (
	auction "auctionSystem/grpc"
	"bufio"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strconv"
)

type Client struct {
	id               int64
	serverConnection auction.AuctionClient
}

var (
	//	clientPort = flag.Int("cPort", 0, "client port number")
	serverPort = flag.Int("sPort", 0, "server port number (should match the port used for the server)")
)

func main() {
	// Parse the flags to get the port for the server
	flag.Parse()

	// Create a client
	client := &Client{
		id: -1, // id -1 means the client doesn't have yet an id
	}

	// Connect to the server
	client.connectToServer()

	// Process client requests
	client.handleRequests()

}

func (c *Client) handleRequests() {

	fmt.Println("Hi there! Welcome to this auction!")
	fmt.Print("Choose an action:\n- Show (s)\n- Bid (b [value])")
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
		}
	}
}

func (c *Client) connectToServer() {
	// Dial the server at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", *serverPort)
	} else {
		log.Printf("Connected to the server at port %d\n", *serverPort)
	}
	c.serverConnection = auction.NewAuctionClient(conn)
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
