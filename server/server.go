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
	biddingMap    map[int32]int32
	port          int32
	isBiddingOver bool
}

func main() {
	//Parse the port number from the command line
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000

	// Create a server struct with the port and the above slice of connections
	replicationManager := &ReplicationManager{
		port:       ownPort,
		biddingMap: make(map[int32]int32),
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

	if replicationManager.isBiddingOver {
		return &proto.Acknowledgement{Status: "fail - bidding is over"}, nil
	}

	//We asssume that the bid is higher than the last bid, and can overwrite the last bid in the map.
	_, currentHighestBid := replicationManager.getHighestBid()
	//Check if the bid was higher then the current highest bid
	if bidMessage.Amount < currentHighestBid {
		//Return error
		return &proto.Acknowledgement{Status: "fail"}, nil
	}

	//Add the new Bid to the map for the Client.
	replicationManager.biddingMap[bidMessage.Id] = bidMessage.Amount

	//Return succesful
	return &proto.Acknowledgement{Status: "success"}, nil
}

func (replicationManager *ReplicationManager) GetResult(ctx context.Context, empty *proto.Empty) (*proto.Outcome, error) {
	currentHighestBidder, currentHighestBid := replicationManager.getHighestBid()
	if replicationManager.isBiddingOver {
		winnerString := "Client " + strconv.Itoa(int(currentHighestBidder))
		return &proto.Outcome{Winner: winnerString, HighestBid: currentHighestBid}, nil
	}
	return &proto.Outcome{Winner: "", HighestBid: currentHighestBid}, nil
}

func (replicationManager *ReplicationManager) getHighestBid() (int32, int32) {
	currentHighestBidder := int32(0)
	currentHighestBid := int32(0)

	//Run through the hashmap, to find highest bidder
	for key, value := range replicationManager.biddingMap {
		if value > currentHighestBid {
			currentHighestBid = value
			currentHighestBidder = key
		}
	}
	return currentHighestBidder, currentHighestBid
}

func (replicationManager *ReplicationManager) startBidding() {
	//Here the time you can bid is 60 seconds.
	time.Sleep(60 * time.Second)
	replicationManager.isBiddingOver = true
}
