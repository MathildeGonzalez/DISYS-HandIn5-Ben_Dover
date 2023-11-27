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
	id string
}

type Frontend struct {
	id                  string
	replicationManagers []int32
	auctionClients      []proto.AuctionClient
}

func main() {
	clientId := os.Args[1]

	client := &Client{
		id: string(clientId),
	}

	//Create a frontend with the id of the client and the ports of the replication managers
	//and a slice of auctionClients
	frontend := &Frontend{
		id:                  string(clientId),
		replicationManagers: []int32{5000, 5001, 5002},
		auctionClients:      []proto.AuctionClient{},
	}

	//Connect to all replication managers
	frontend.connectToServers()

	go listenToClient(client, frontend)

	for {
		//infinite for loop to keep the client running
	}
}

// Function to listen to client input
func listenToClient(client *Client, frontend *Frontend) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		scan := scanner.Text()
		if strings.HasPrefix(scan, "bid") {
			bidAmount, err := strconv.Atoi(strings.Split(scan, " ")[1])
			if err != nil {
				log.Println("Bid is not a number")
				return
			}
			//Send bid to frontend, who will then pass the bid on to all replication managers
			client.sendBid(int32(bidAmount), frontend)

		} else if scan == "result" {
			//Request result from frontend, who will then pass the request on to the first replication manager
			//and the frontend will return the response to the client
			client.getResult(frontend)
		}
	}
}

func (client *Client) sendBid(bidAmount int32, frontend *Frontend) {
	frontendResponse := frontend.sendBid(bidAmount)
	log.Printf("Client received response from frontend: %s", frontendResponse)
}

// Function to send bid to all replication managers by looping through auctionClients
// We assume that there is always a minimum of one functioning server
func (frontend *Frontend) sendBid(bidAmount int32) string {
	//The function can maximally remove one server each call of sendBid()
	var toDelete proto.AuctionClient
	var serverResponse string
	for _, auctionClient := range frontend.auctionClients {
		ack, err := auctionClient.Bid(context.Background(), &proto.BidMessage{Id: frontend.id, Amount: bidAmount})
		if err != nil {
			log.Printf("Frontend: Could not send bid to server: Connection lost!")
			//Mark this server for deletion
			toDelete = auctionClient
		} else {
			log.Printf("Frontend received: Received response from server: %v", ack)
			serverResponse = ack.Status
		}
	}

	//Remove the server marked for deletion if there is one
	if toDelete != nil {
		frontend.auctionClients = removeElement(frontend.auctionClients, toDelete)
	}

	//Return a response to the client
	return serverResponse
}

func (client *Client) getResult(frontend *Frontend) {
	frontendResponse := frontend.getResult()
	log.Printf("Client received from frontend: %s", frontendResponse)
}

// Function to request result of auction
func (frontend *Frontend) getResult() string {
	//Ask the first replication manager for the result
	//Since the first RM in the slice is always the first to be updated, it will 
	//always be the one with the most up-to-date result
	var serverResponse string
	outcome, err := frontend.auctionClients[0].GetResult(context.Background(), &proto.Empty{})
	if err != nil {
		//This error will happen, if the first RM in the slice is down
		log.Printf("Could not receive result from server: %v", err)

		//Remove the first RM from the slice
		frontend.auctionClients = removeElement(frontend.auctionClients, frontend.auctionClients[0])
		
		//Call the function again, to try the next RM in the slice
		return frontend.getResult()
	}

	//If there is no winner yet, we only return the highest bid
	if len(outcome.Winner) == 0 {
		serverResponse = fmt.Sprintf("The current highest bid is %d", outcome.HighestBid)
		log.Printf("Frontend received from server: The current highest bid is %d", outcome.HighestBid)
	} else {
		serverResponse = fmt.Sprintf("The auction is over! The winner is %s, with the bid of: %d", outcome.Winner, outcome.HighestBid)
		log.Printf("Frontend received from server: The auction is over! The winner is %s, with the bid of: %d", outcome.Winner, outcome.HighestBid)
	}

	return serverResponse
}

func (frontend *Frontend) connectToServers() {
	// Dial the servers at the specified port.
	for _, port := range frontend.replicationManagers {
		conn, err := grpc.Dial("localhost:"+strconv.Itoa(int(port)), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Could not connect to port %d", port)
		} else {
			log.Printf("Connected to the server at port %d\n", port)
		}
		//Add the connection to the slice of auctionClients
		frontend.auctionClients = append(frontend.auctionClients, proto.NewAuctionClient(conn))
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
