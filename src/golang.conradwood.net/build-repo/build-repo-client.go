package main

// see: https://grpc.io/docs/tutorials/basic/go.html

import (
	"fmt"
	"google.golang.org/grpc"
	//	"github.com/golang/protobuf/proto"
	"errors"
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
	serverAddr  = flag.String("server_addr", "localhost:5004", "The server address in the format of host:port")
	reponame    = flag.String("repository", "", "name of repository")
	branchname  = flag.String("branch", "", "branch of commit")
	commitid    = flag.String("commitid", "", "commit")
	commitmsg   = flag.String("commitmsg", "", "commit message")
	buildnumber = flag.Int("build", 0, "build number")
	distDir     = flag.String("distdir", "dist", "Default directory to upload")
)

func main() {
	flag.Parse()
	files := flag.Args()
	if len(files) == 0 {
		fmt.Printf("No files specified on commandline, using \"%s\" as default\n", *distDir)
		df, err := ioutil.ReadDir(*distDir)
		if err != nil {
			fmt.Printf("Failed to read directory \"%s\": %s\n,", *distDir, err)
			os.Exit(5)
		}
		for _, file := range df {
			fmt.Println(file.Name())
			files = append(files, fmt.Sprintf("%s/%s", *distDir, file.Name()))
		}
	}
	opts := []grpc.DialOption{grpc.WithInsecure()}
	fmt.Println("Connecting to server...")
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	ctx := context.Background()

	client := pb.NewBuildRepoManagerClient(conn)
	fmt.Printf("New build %d in repo %s\n", *buildnumber, *reponame)
	req := pb.CreateBuildRequest{
		Repository: *reponame,
		CommitID:   *commitid,
		Branch:     *branchname,
		BuildID:    uint64(*buildnumber),
		CommitMSG:  *commitmsg,
	}
	resp, err := client.CreateBuild(ctx, &req)
	if err != nil {
		log.Fatalf("fail to create build: %v", err)
	}
	fmt.Printf("Response to createbuild was: %v\n", resp)
	storeid := resp.BuildStoreid
	serverHost, _, err := net.SplitHostPort(*serverAddr)

	uploadFiles(ctx, client, storeid, serverHost, files)
	udr := &pb.UploadDoneRequest{BuildStoreid: storeid}
	_, err = client.UploadsComplete(ctx, udr)
	if err != nil {
		fmt.Printf("Failed to complete uploads: %v\n", err)
		os.Exit(5)
	}
}

func uploadFiles(ctx context.Context, client pb.BuildRepoManagerClient, storeid string, serverHost string, files []string) error {
	// why "dog"? I learned of the "range" operator from an example
	// which uses "dogs" - cnw
	for _, dog := range files {
		st, err := os.Stat(dog)
		if err != nil {
			fmt.Printf("Cannot stat %s: %s, skipping...\n", dog, err)
			continue
		}
		if st.Mode().IsDir() {
			var nfiles []string
			df, err := ioutil.ReadDir(dog)
			if err != nil {
				fmt.Printf("Failed to read directory \"%s\": %s\n,", dog, err)
				return errors.New("Failed to read directory")
			}
			for _, file := range df {
				nf := fmt.Sprintf("%s/%s", dog, file.Name())
				nfiles = append(nfiles, nf)
			}
			uploadFiles(ctx, client, storeid, serverHost, nfiles)
			continue

		}
		if !st.Mode().IsRegular() {
			fmt.Printf("Skipping %s - it's not a file\n", dog)
			continue
		}
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
	return nil
}
