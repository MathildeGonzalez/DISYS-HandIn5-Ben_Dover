// Client and frontend for the auction application
package main

import (
	proto "Auction/grpc"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

//A Client has a slice of all replication managers that it connects to and sends requests to
//When the frontend registers that a replication manager has failed, it removes it from the slice

type Client struct{
	id 					int32
	replicationManagers []*proto.AuctionServer
}

func main(){
	//Create a client struct
	//add all replication managers to the slice

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		scan := scanner.Text()
		if strings.HasPrefix(scan, "bid") {
			bid, err := strconv.Atoi(strings.Split(scan, " ")[1])
			if err != nil {
				fmt.Println("Bid is not a number")
				return
			}
			//send bid to all replication managers
			
		} else if scan == "result" {
			//
		}
	}
}


