// Client and frontend for the auction application
package main

import (
	proto "Auction/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//A Client has a slice of all replication managers that it connects to and sends requests to
//When the frontend registers that a replication manager has failed, it removes it from the slice

type Client struct {
	id                  int32
	replicationManagers []int32
	auctionClients      []proto.AuctionClient
}

//make frontent struct

func main() {
	clientId, _ := strconv.ParseInt(os.Args[1], 10, 32)

	//Create a client struct
	//add all replication managers to the slice

	client := &Client{
		id:                  int32(clientId),
		replicationManagers: []int32{5000, 5001, 5002},
		auctionClients:      []proto.AuctionClient{},
	}

	//Connect to all replication managers
	client.connectToServers()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		scan := scanner.Text()
		if strings.HasPrefix(scan, "bid") {
			bidAmount, err := strconv.Atoi(strings.Split(scan, " ")[1])
			if err != nil {
				fmt.Println("Bid is not a number")
				return
			}
			//send bid to all replication managers
			client.sendBid(int32(bidAmount))

		} else if scan == "result" {
			client.getResult()
		}
	}
}

// Function to send bid to all replication managers by looping through auctionClients
// We assume that there always be a minimum of one functioning server.
func (client *Client) sendBid(bidAmount int32) {
	//The function can maximally remove one server each call of sendBid()
	var toDelete proto.AuctionClient

	for _, auctionClient := range client.auctionClients {
		ack, err := auctionClient.Bid(context.Background(), &proto.BidMessage{Id: client.id, Amount: bidAmount})
		if err != nil {
			log.Printf("Could not send bid to server: Connection lost!") //remove the Server from slice.
			toDelete = auctionClient
		} else {
			log.Printf("Received response from server: %v", ack)
		}
	}
	if toDelete != nil {
		client.auctionClients = removeElement(client.auctionClients, toDelete)
	}
}

// Function to request result of auction
func (client *Client) getResult() {
	//ask the first replication manager for the result, since it is the first to be updated

	outcome, err := client.auctionClients[0].GetResult(context.Background(), &proto.Empty{})
	if err != nil {
		log.Fatalf("Could not receive result from server: %v", err)
	}
	if len(outcome.Winner) == 0 {
		log.Printf("The current highest bid is %d", outcome.HighestBid)
	} else {
		log.Printf("The auction is over! The winner is %s, with the bid of: %d", outcome.Winner, outcome.HighestBid)
	}
}

func (client *Client) connectToServers() {
	// Dial the server at the specified port.
	for _, port := range client.replicationManagers {
		conn, err := grpc.Dial("localhost:"+strconv.Itoa(int(port)), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Could not connect to port %d", port)
		} else {
			log.Printf("Connected to the server at port %d\n", port)
		}
		client.auctionClients = append(client.auctionClients, proto.NewAuctionClient(conn))
	}
}

// Helper method used to remove a specific element from a Slice
func removeElement(slice []proto.AuctionClient, element proto.AuctionClient) []proto.AuctionClient {
	for i, v := range slice {
		if v == element {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice // Element not found
}
