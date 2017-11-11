package main

// simple http server serving http
// and providing download options for
// things in the build repo

import (
	"fmt"
	pb "golang.conradwood.net/build-repo/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"html/template"
	"strings"
	/*
		"google.golang.org/grpc"
		//	"github.com/golang/protobuf/proto"
	*/
	"errors"
	"flag"
	//	"html"
	/*
		pb "golang.conradwood.net/build-repo/proto"
		"golang.org/x/net/context"
		"google.golang.org/grpc/peer"
		"io"
		"log"
		"math/rand"
		"net"
	*/
	"net/http"
	"os"
	"path/filepath"
	/*
		"os/exec"
		"strings"
		"time"
	*///"io/ioutil"
)

// static variables
var (
	httpPort  = flag.Int("http_port", 7005, "The http server port")
	buildRepo = flag.String("server_addr", "localhost:5004", "The build-repo server address in the format of host:port")
	buildconn *grpc.ClientConn
)

/**************************************************
* helpers
***************************************************/
const (
	listRepos    = iota
	listBranches = iota
	listVersions = iota
	listFiles    = iota
	getFile      = iota
)

type PathInfo struct {
	Repository string
	Branch     string
	Version    string
	Filename   string
	// reqTarget indicates what we're trying to reach
	reqTarget int
	// output extension?
	output string
}

/**************************************************
* main entry
***************************************************/
func parsePath(path string) *PathInfo {
	fmt.Printf("Request for \"%s\"\n", path)
	res := PathInfo{}
	if path == "/" {
		res.reqTarget = listRepos
		return &res
	}
	path, err := filepath.Rel("/", path)
	if err != nil {
		fmt.Printf("Failed to parse url path: \"%s\"n", path)
		return nil
	}
	res.output = "html" // default serve result as html

	types := []string{"info", "txt", "xml"}
	for _, t := range types {
		tp := fmt.Sprintf(".%s", t)
		if strings.HasSuffix(path, tp) {
			res.output = t
			path = strings.TrimRight(path, tp)
		}
	}
	parts := strings.Split(path, "/")
	for i, s := range parts {
		if s == "" {
			fmt.Printf("Failed to parse url path: \"%s\" at pos #%d: \"%s\"!\n", path, i, s)
			return nil
		}
		if i == 0 {
			res.Repository = s
			res.reqTarget = listBranches
		} else if i == 1 {
			res.Branch = s
			res.reqTarget = listVersions
		} else if i == 2 {
			res.Version = s
			res.reqTarget = listFiles
		} else if i == 3 {
			res.Filename = s
			res.reqTarget = getFile
		}
	}
	return &res
}

// we serve a directory listen or send a file, depending
// on what we find on disk
func serveContent(pi *PathInfo, w http.ResponseWriter, r *http.Request) error {
	var entries []*pb.RepoEntry
	var err error

	ctx := context.Background()
	client := pb.NewBuildRepoManagerClient(buildconn)
	if buildconn == nil {
		return errors.New("Unable to create RPC client")
	}

	// default template (only different for files)
	tmplname := fmt.Sprintf("templates/directory_listing.%s", pi.output)

	if pi.reqTarget == listRepos {
		// list repos
		sr := pb.ListReposRequest{}
		sa, e := client.ListRepos(ctx, &sr)
		if e != nil {
			return e
		}
		entries = sa.Repositories
	} else if pi.reqTarget == listBranches {
		// list branches
		sr := pb.ListBranchesRequest{Repository: pi.Repository}
		sa, e := client.ListBranches(ctx, &sr)
		if e != nil {
			return e
		}
		entries = sa.Branches
	} else if pi.reqTarget == listVersions {
		// list versions
		sr := pb.ListVersionsRequest{Repository: pi.Repository,
			Branch: pi.Branch}
		sa, e := client.ListVersions(ctx, &sr)
		if e != nil {
			return e
		}
		entries = sa.Versions
	} else if pi.reqTarget == listFiles {
		// list versions
		sr := pb.ListFilesRequest{Repository: pi.Repository,
			Branch: pi.Branch, Version: pi.Version}
		sa, e := client.ListFiles(ctx, &sr)
		if e != nil {
			return e
		}
		entries = sa.Files
	} else {
		return errors.New(fmt.Sprintf("Type %d not implemented", pi.reqTarget))
	}
	if err != nil {
		fmt.Println("Failed to retrieve %v: %s\n", pi, err)
		return err
	}
	tmpl, err := template.ParseFiles(tmplname)
	if err != nil {
		fmt.Printf("Failed to parse template %s: %s\n", tmplname, err)
		return err
	}
	err = tmpl.Execute(w, entries)
	if err != nil {
		fmt.Printf("Failed to execute template %s\n", err)
		return err
	}
	return nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	pi := parsePath(r.URL.Path)
	if pi == nil {
		fmt.Fprintf(w, "Invalid url")
		return
	}
	err := serveContent(pi, w, r)
	if err != nil {
		fmt.Fprintf(w, "Failed to process request: %s\n", err)
		return
	}
	fmt.Printf("PathInfo: %v\n", pi)
	//fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func main() {
	var err error
	flag.Parse() // parse stuff. see "var" section above

	http.HandleFunc("/", handleRequest)

	opts := []grpc.DialOption{grpc.WithInsecure()}
	fmt.Println("Connecting to server...")
	buildconn, err = grpc.Dial(*buildRepo, opts...)
	if err != nil {
		fmt.Printf("fail to dial: %v", err)
		os.Exit(10)
	}
	//defer buildconn.Close()
	ctx := context.Background()
	client := pb.NewBuildRepoManagerClient(buildconn)
	fmt.Printf("Created connection %v and %v\n", ctx, client)
	fmt.Printf("Listening on port %d...\n", *httpPort)
	err = http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), nil)
	if err != nil {
		fmt.Printf("Failed to listen on port %d: %s\n", *httpPort, err)
	}
}
