// Server/Replication manager
package main

import (
	proto "Auction/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

type ReplicationManager struct {
	proto.UnimplementedAuctionServer
	biddingMap    map[string]int32
	port          int32
	isBiddingOver bool
}

func main() {
	//Parse the port number from the command line
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000

	// Create a RM struct with the port and an empty map for the bids
	replicationManager := &ReplicationManager{
		port:       ownPort,
		biddingMap: make(map[string]int32),
	}

	// Start the server
	startServer(replicationManager)
}

func startServer(replicationManager *ReplicationManager) {

	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", replicationManager.port))

	if err != nil {
		log.Fatalf("Could not create the Replication Manager %v", err)
	}
	log.Printf("Started Replication Manager at port: %d\n", replicationManager.port)

	// Register the grpc server and serve its listener
	proto.RegisterAuctionServer(grpcServer, replicationManager)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

func (replicationManager *ReplicationManager) Bid(ctx context.Context, bidMessage *proto.BidMessage) (*proto.Acknowledgement, error) {
	//If map is empty, start the bidding phase
	if len(replicationManager.biddingMap) == 0 {
		go replicationManager.startBidding()
	}

	//Return error-status if bidding is over
	if replicationManager.isBiddingOver {
		return &proto.Acknowledgement{Status: "fail - bidding is over"}, nil
	}

	//Get the current highest bid
	_, currentHighestBid := replicationManager.getHighestBid()
	//Check if the received bid is higher than the current highest bid
	if bidMessage.Amount < currentHighestBid {
		//Return error
		return &proto.Acknowledgement{Status: "fail - bid too low"}, nil
	}

	//Add the new Bid to the map for the Client
	replicationManager.biddingMap[bidMessage.Id] = bidMessage.Amount

	//Return succesful
	return &proto.Acknowledgement{Status: "success"}, nil
}

func (replicationManager *ReplicationManager) GetResult(ctx context.Context, empty *proto.Empty) (*proto.Outcome, error) {
	//Get the current highest bid and bidder
	currentHighestBidder, currentHighestBid := replicationManager.getHighestBid()
	//If bidding is over, we return both the winner and the winning bid
	if replicationManager.isBiddingOver {
		winnerString := currentHighestBidder
		return &proto.Outcome{Winner: winnerString, HighestBid: currentHighestBid}, nil
	}
	//If bidding is not over, we only return the current highest bid
	return &proto.Outcome{Winner: "", HighestBid: currentHighestBid}, nil
}

//Helper method to get the highest bid and bidder from the map of bids
func (replicationManager *ReplicationManager) getHighestBid() (string, int32) {
	var currentHighestBidder string
	currentHighestBid := int32(0)

	//Run through the map to find highest bid and bidder
	for key, value := range replicationManager.biddingMap {
		if value > currentHighestBid {
			currentHighestBid = value
			currentHighestBidder = key
		}
	}
	return currentHighestBidder, currentHighestBid
}

//Method to start the bidding phase which is now hardcoded to last 60 seconds
func (replicationManager *ReplicationManager) startBidding() {
	time.Sleep(60 * time.Second)
	replicationManager.isBiddingOver = true
}
