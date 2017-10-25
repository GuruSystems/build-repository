package main

// see: https://grpc.io/docs/tutorials/basic/go.html

import (
	"fmt"
	"google.golang.org/grpc"
	//	"github.com/golang/protobuf/proto"
	"flag"
	"golang.org/x/net/context"
	//	"net"
	pb "golang.conradwood.net/build-repo/proto"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

// static variables for flag parser
var (
	serverAddr = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
	port       = flag.Int("port", 10000, "The server port")
)

func main() {
	flag.Parse()
	files := flag.Args()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	fmt.Println("Connecting to server...")
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	ctx := context.Background()

	client := pb.NewBuildRepoManagerClient(conn)
	req := pb.CreateBuildRequest{
		Repository: "testrepo",
		CommitID:   "testaccess",
		Branch:     "master",
		BuildID:    1,
		CommitMSG:  "none",
	}
	resp, err := client.CreateBuild(ctx, &req)
	if err != nil {
		log.Fatalf("fail to create build: %v", err)
	}
	fmt.Printf("Response to createbuild was: %v\n", resp)
	storeid := resp.BuildStoreid
	serverHost, _, err := net.SplitHostPort(*serverAddr)
	// why "dog"? I learned of the "range" operator from an example
	// which uses "dogs" - cnw
	for _, dog := range files {
		fmt.Printf("Uploading \"%s\"\n", dog)
		file, err := os.Open(dog)
		if err != nil {
			fmt.Printf("Unable to open \"%s\": %s\n", dog, err)
			continue
		}
		defer file.Close()

		ureq := &pb.UploadSlotRequest{
			BuildStoreid: storeid,
			Filename:     dog,
		}
		resp, err := client.GetUploadSlot(ctx, ureq)
		if err != nil {
			fmt.Printf("Failed to upload %s: %v\n", dog, err)
			continue
		}
		//fmt.Println("Upload Token:", resp)
		url := fmt.Sprintf("http://%s:%d/upload/%s", serverHost, resp.Port, resp.Token)
		//fmt.Println("URL: ", url)
		res, err := http.Post(url, "binary/octet-stream", file)
		if err != nil {
			fmt.Println("Failed to upload: ", err)
			continue
		}

		defer res.Body.Close()
		message, _ := ioutil.ReadAll(res.Body)
		fmt.Printf(string(message))

	}
}
