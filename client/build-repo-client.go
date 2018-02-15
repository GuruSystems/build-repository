package main

import (
	"os"
	"log"
	"net"
	"fmt"
	"flag"
	"time"
	"bytes"
	"errors"
	"strings"
	"net/http"
	"io/ioutil"
	//
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	//
	pb "proto"
)

// static variables for flag parser
var (
	serverAddr  = flag.String("server_addr", "localhost:5004", "The build-repo server address in the format of host:port")
	reponame    = flag.String("repository", "", "name of repository")
	branchname  = flag.String("branch", "", "branch of commit")
	commitid    = flag.String("commitid", "", "commit")
	commitmsg   = flag.String("commitmsg", "", "commit message")
	buildnumber = flag.Int("build", 0, "build number")
	distDir     = flag.String("distdir", "dist", "Default directory to upload")
	dryrun      = flag.Bool("n", false, "dry-run")
	versionfile = flag.String("versionfile", "", "filename of a versionfile to update with buildid")
	versiondir  = flag.String("versiondir", "", "directory to scan for buildversion.go files (update files with buildid)")
	info        = flag.Bool("info", false, "Get information about the repo")
	offset      = flag.Int("offset", -1, "if >0 read a block from the file in the repo beginning at the specified offset")
	blocksize   = flag.Int("blocksize", 512, "Default block size when reading block from file from repo")
	filename    = flag.String("filename", "", "Filename from which to retrieve a block")
)

func main() {
	flag.Parse()
	if *versionfile != "" {
		updateVersionFile(*versionfile)
		os.Exit(0)
	}
	if *versiondir != "" {
		updateVersionDir(*versiondir)
		os.Exit(0)
	}
	if *info {
		getInfo()
		os.Exit(0)
	}
	if *offset >= 0 {
		getBlock()
		os.Exit(0)
	}
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
	AddDirIfExists("deployment", &files)

	if *dryrun {
		for _, file := range files {
			fmt.Printf("Uploading file: %s\n", file)
		}
		return
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

// get an arbitrary block from a file from repo
// this is useful for OTA
func getBlock() {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	fmt.Println("Connecting to server...")
	conn, err := grpc.Dial(*serverAddr, opts...)
	bail(err, "failed to dial")

	defer conn.Close()
	ctx := context.Background()
	client := pb.NewBuildRepoManagerClient(conn)
	f := &pb.File{
		Repository: *reponame,
		Branch:     *branchname,
		BuildID:    uint64(*buildnumber),
		Filename:   *filename,
	}
	glv := &pb.GetBlockRequest{
		File:   f,
		Offset: uint64(*offset),
		Size:   uint32(*blocksize),
	}
	glr, err := client.GetBlock(ctx, glv)
	bail(err, "Failed to read block")
	fmt.Printf("Response:\n")
	fmt.Printf("Size=%d, Offset=%d\n", glr.Size, glr.Offset)
	fmt.Printf("Data: [%s]\n", glr.Data)
}

// connect to server, get latest version information of a given repo
func getInfo() {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	fmt.Println("Connecting to server...")
	conn, err := grpc.Dial(*serverAddr, opts...)
	bail(err, "failed to dial")

	defer conn.Close()
	ctx := context.Background()
	client := pb.NewBuildRepoManagerClient(conn)
	glv := &pb.GetLatestVersionRequest{
		Repository: *reponame,
		Branch:     *branchname,
	}
	glr, err := client.GetLatestVersion(ctx, glv)
	bail(err, "failed to get latest version")
	fmt.Printf("Latest Version: %d\n", glr.BuildID)
}

// end info stuff

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

func AddDirIfExists(dirname string, files *[]string) error {
	df, err := ioutil.ReadDir(dirname)
	if err != nil {
		fmt.Printf("Failed to read directory \"%s\": %s\n,", dirname, err)
		return err
	}
	for _, file := range df {
		fmt.Println(file.Name())
		*files = append(*files, fmt.Sprintf("%s/%s", dirname, file.Name()))
	}
	return nil
}
func bail(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Printf("%s: %s", msg, err)
	os.Exit(10)
}

// recursively go through directory and process all files called buildversion.go
func updateVersionDir(dname string) {
	fos, err := ioutil.ReadDir(dname)
	bail(err, "Unable to read dir")
	for _, file := range fos {
		if file.IsDir() {
			updateVersionDir(fmt.Sprintf("%s/%s", dname, file.Name()))
			continue
		}
		if file.Name() != "buildversion.go" {
			continue
		}
		fullname := fmt.Sprintf("%s/%s", dname, file.Name())
		fmt.Printf("File: %s\n", fullname)
		updateVersionFile(fullname)

	}
}

func updateVersionFile(fname string) {
	bs, err := ioutil.ReadFile(fname)
	bail(err, "Failed to readfile")
	lines := string(bs)
	var buffer bytes.Buffer
	changed := false
	for _, line := range strings.Split(lines, "\n") {
		if !strings.Contains(line, "// AUTOMATIC VERSION UPDATE: OK") {
			buffer.WriteString(line)
			buffer.WriteString("\n")
			continue
		}
		if strings.Contains(line, "Buildnumber") {
			changed = true
			line = strings.Replace(line, "0", fmt.Sprintf("%d", *buildnumber), 1)
		} else if strings.Contains(line, "Build_date_string") {
			changed = true
			line = strings.Replace(line, "today", time.Now().UTC().Format("2006-01-02T15:04:05-0700"), 1)
		} else if strings.Contains(line, "Build_date") {
			changed = true
			line = strings.Replace(line, "0", fmt.Sprintf("%d", time.Now().Unix()), 1)
		}
		buffer.WriteString(line)
		buffer.WriteString("\n")

	}
	if !changed {
		fmt.Printf("File %s was not changed\n", fname)
		return
	}
	s := buffer.String()
	if *buildnumber != 0 {
		err := ioutil.WriteFile(fname, []byte(s), 0777)
		bail(err, "Failed to write versionfile")
		fmt.Printf("File %s updated\n", fname)
	} else {
		fmt.Printf("File %s would have been updated to:\n%s\n", fname, s)
	}

}
