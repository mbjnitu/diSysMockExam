package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	gRPC "github.com/mbjnitu/diSysMockExam/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")

var serverPort1 = flag.String("server1", "5401", "Tcp server1")
var serverPort2 = flag.String("server2", "5402", "Tcp server2")
var serverPort3 = flag.String("server3", "5403", "Tcp server3")

var server1 gRPC.TemplateClient //the server
var server2 gRPC.TemplateClient //the server
var server3 gRPC.TemplateClient //the server
var ServerConn1 grpc.ClientConn //the server connection
var ServerConn2 grpc.ClientConn //the server connection
var ServerConn3 grpc.ClientConn //the server connection

func main() {
	//parse flag/arguments
	flag.Parse()

	fmt.Println("--- CLIENT APP ---")

	//log to file instead of console
	//f := setLog()
	//defer f.Close()

	//connect to server and close the connection when program closes
	fmt.Println("--- join Server ---")
	ConnectToServer(&server1, &ServerConn1, serverPort1)
	ConnectToServer(&server2, &ServerConn2, serverPort2)
	ConnectToServer(&server3, &ServerConn3, serverPort3)
	defer ServerConn1.Close()
	defer ServerConn1.Close()
	defer ServerConn3.Close()

	//start the listening to user calls
	parseInput()
}

func ConnectToServer(_server *gRPC.TemplateClient, _conn *grpc.ClientConn, _port *string) {

	//dial options
	//the server is not using TLS, so we use insecure credentials
	//(should be fine for local testing but not in the real world)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	//dial the server, with the flag "server", to get a connection to it
	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *_port)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", *_port), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	// makes a client from the server connection and saves the connection
	// and prints rather or not the connection was is READY
	*_server = gRPC.NewTemplateClient(conn)
	_conn = conn
	log.Println("the connection is: ", conn.GetState().String())
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Ready for command:")
	fmt.Println("--------------------")

	//Infinite loop to listen for clients input.
	for {
		fmt.Print("-> ")

		//Read input into var input and any errors into err
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input) //Trim input

		//Convert string to int64, return error if the int is larger than 32bit or not a number
		val, err := strconv.ParseInt(input, 10, 64)
		if err == nil {
			incrementVal(val)
			continue
		} else {
			fmt.Print("Increment only acceps a number")
		}
	}
}

// Error: couldn't get this to work with a reference.
func incrementVal(val int64) {
	servers := []gRPC.TemplateClient{server1, server2, server3}
	values := []int64{}
	//create amount type
	amount := &gRPC.Amount{
		ClientName: *clientsName,
		Value:      val, //cast from int to int32
	}

	for i := 0; i < 3; i++ {
		//Make gRPC call to server with amount, and recieve acknowlegdement back.
		ack, err := servers[i].Increment(context.Background(), amount)
		if err != nil || ack == nil {
			log.Printf("The server 500%d has crashed \n", i)
			continue
		}
		values = append(values, ack.NewValue)
	}
	fmt.Printf("Returned increment value is %d\n", values[0])
}

// Function which returns a true boolean if the connection to the server is ready, and false if it's not.
func conReady(conn *grpc.ClientConn) bool {
	return conn.GetState().String() == "READY"
}
