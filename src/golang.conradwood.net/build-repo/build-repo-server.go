package main

// this is a bit random, particularly the upload token we hand back to the client
// isn't really verified yet
// don't use it in an untrusted environment!

import (
	"fmt"
	"google.golang.org/grpc"
	//	"github.com/golang/protobuf/proto"
	"errors"
	"flag"
	pb "golang.conradwood.net/build-repo/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UploadMetaData struct {
	Token    string
	Filename string
	Storeid  string
	// the path under which we store the files
	Storepath string
	Created   time.Time
}

// static variables
var (
	port     = flag.Int("port", 5004, "The server port")
	httpport = flag.Int("http_port", 5005, "The http server port")
	base     = "/srv/build-repository/artefacts"
	src      = rand.NewSource(time.Now().UnixNano())
	uploads  = make(map[string]UploadMetaData)
	httpPort int
)

/**************************************************
* helpers
***************************************************/
//https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func RandString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits

	)

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func FindUploadMetaData(token string) UploadMetaData {
	umd := uploads[token]
	umd.Storepath = umd.Storeid
	return umd
}
func StoreUploadMetaData(vfilename string, vtoken string, vstoreid string) {
	umd := UploadMetaData{
		Token:    vtoken,
		Filename: vfilename,
		Storeid:  vstoreid,
		Created:  time.Now(),
	}
	uploads[vtoken] = umd

}

/**************************************************
* main entry
***************************************************/
func main() {
	flag.Parse() // parse stuff. see "var" section above
	listenAddr := fmt.Sprintf(":%d", *port)
	httpPort = *httpport
	lis, err := net.Listen("tcp4", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	http.HandleFunc("/", upload)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	s := new(BuildRepoServer)
	pb.RegisterBuildRepoManagerServer(grpcServer, s) // created by proto

	fmt.Printf("Starting  http service on port %d\n", httpPort)
	go http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil)

	fmt.Printf("Starting BuildRepo Manager service on %s\n", listenAddr)
	grpcServer.Serve(lis)

}

func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("HTTP Upload method: %s, content-type: %s\n", r.Method, r.Header.Get("Content-Type"))
	token := r.URL.Path
	if !strings.HasPrefix(token, "/upload/") {
		fmt.Println("Invalid token: ", token)
		return
	}
	token = strings.TrimPrefix(token, "/upload/")
	//fmt.Printf("Token: \"%s\"\n", token)
	umd := FindUploadMetaData(token)

	fname := fmt.Sprintf("%s/%s", umd.Storepath, umd.Filename)
	fmt.Printf("Receiving file %s => %s\n", umd.Filename, fname)
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("Failed to open file %s: %v\n", fname, err)
		return
	}
	defer f.Close()
	io.Copy(f, r.Body)

}

/**********************************
* implementing the functions here:
***********************************/
type BuildRepoServer struct {
	wtf int
}

// in C we put methods into structs and call them pointers to functions
// in java/python we also put pointers to functions into structs and but call them "objects" instead
// in Go we don't put functions pointers into structs, we "associate" a function with a struct.
// (I think that's more or less the same as what C does, just different Syntax)
func (s *BuildRepoServer) CreateBuild(ctx context.Context, cr *pb.CreateBuildRequest) (*pb.CreateBuildResponse, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		fmt.Println("Error getting peer ")
	}
	if cr.Repository == "" {
		return nil, errors.New("Missing repository name")
	}
	if cr.CommitID == "" {
		return nil, errors.New("Missing commit id")
	}
	if cr.CommitMSG == "" {
		return nil, errors.New("Missing commit message")
	}
	if cr.Branch == "" {
		return nil, errors.New("Missing branch name")
	}
	if cr.BuildID == 0 {
		return nil, errors.New("Missing build id")
	}

	resp := pb.CreateBuildResponse{}
	dir := fmt.Sprintf("%s/%s/%s/%d", base, cr.Repository, cr.Branch, cr.BuildID)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		fmt.Println("Failed to create directory ", dir, err)
		return &resp, err
	}
	fmt.Println("Created directory:", dir)
	fmt.Println(peer.Addr, "called createbuild")
	resp.BuildStoreid = dir

	// write env to file in directory
	metafile := fmt.Sprintf("%s/meta.txt", dir)
	f, err := os.OpenFile(metafile, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("Failed to open file %s: %v\n", metafile, err)
		return nil, err
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("COMMIT_ID=%s\n", cr.CommitID))
	f.WriteString(fmt.Sprintf("BUILD_ID=%s\n", cr.BuildID))

	linkdir := fmt.Sprintf("%s/%s/%s", base, cr.Repository, cr.Branch)
	err = UpdateSymLink(linkdir, int(cr.BuildID))
	if err != nil {
		fmt.Printf("Failed to create symlink in %s: %v\n", linkdir, err)
		return nil, err
	}
	return &resp, nil
}

func UpdateSymLink(dir string, latestBuild int) error {
	linkName := fmt.Sprintf("%s/latest", dir)
	fmt.Printf("linking \"latest\" in dir %s to %d\n", dir, latestBuild)
	err := os.Chdir(dir)
	if err != nil {
		fmt.Printf("Failed to chdir to %s: %v\n", dir, err)
		return err
	}

	err = os.Symlink(fmt.Sprintf("%d", latestBuild), "latest")
	if err == nil {
		return nil
	}
	if os.IsExist(err) {
		os.Remove(linkName)
		err = os.Symlink(fmt.Sprintf("%d", latestBuild), "latest")
		if err != nil {
			fmt.Printf("Tried to remove symlink but still failed to create it: %s: %s\n", linkName, err)
			return err
		}
	} else {
		fmt.Printf("Failed to create symlink in %s: %v\n", dir, err)
		return err
	}
	return nil
}

func (s *BuildRepoServer) GetUploadSlot(ctx context.Context, pr *pb.UploadSlotRequest) (*pb.UploadSlotResponse, error) {

	res := &pb.UploadSlotResponse{}
	storeid := pr.BuildStoreid
	fname := pr.Filename
	fname = filepath.Clean(fname)
	if filepath.IsAbs(fname) {
		return res, errors.New("file must be relative")
	}
	if !strings.HasPrefix(storeid, base) {
		fmt.Printf("Base=\"%s\", but token sent was: \"%s\"\n", base, storeid)
		return res, errors.New("storeid is invalid")
	}
	fbase := filepath.Dir(fname)
	sp := GetBuildStoreDir(pr.BuildStoreid)
	absDir := fmt.Sprintf("%s/%s", sp, fbase)
	fmt.Printf("Filebase: \"%s\" (%s)\n", fbase, absDir)
	err := os.MkdirAll(absDir, 0777)
	if err != nil {
		fmt.Println("Failed to create directory ", absDir, err)
		return res, err
	}
	fmt.Printf("Request to upload file \"%s\" to store \"%s\"\n", fname, storeid)
	token := RandString(256)
	res.Token = token
	res.Port = int32(httpPort)
	StoreUploadMetaData(fname, token, storeid)
	return res, nil
}
func (s *BuildRepoServer) Ping(ctx context.Context, pr *pb.PingRequest) (*pb.PingResponse, error) {
	fmt.Println("pong")
	return nil, nil
}

func GetBuildStoreDir(id string) string {
	return id
}
