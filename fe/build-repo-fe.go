package main

// simple http server serving http
// and providing download options for
// things in the build repo

import (
	"os"
	"fmt"
	"flag"
	"errors"
	"strings"
	"net/http"
	"path/filepath"
	"html/template"
	//
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	//
	pb "proto"
)

// static variables
var (
	httpPort  = flag.Int("http_port", 7005, "The http server port")
	buildRepo = flag.String("server_addr", "localhost:5004", "The build-repo server address in the format of host:port")
	buildconn *grpc.ClientConn
	basePath  = flag.String("http_path", "/Buildrepo/", "The build-repo server path of the url")
)

/**************************************************
* helpers
***************************************************/
const (
	listRepos    = 1
	listBranches = 2
	listVersions = 3
	listFiles    = 4
	getFile      = 5
)

type PathInfo struct {
	Repository string
	Branch     string
	Version    string
	Filename   string
	// reqTarget indicates what we're trying to reach
	ReqTarget int
	// output extension?
	output string
}

type RenderInfo struct {
	// current path to stuff (prefix to links)
	Path    string
	Top     string
	Up      string
	Pi      *PathInfo
	Entries []*pb.RepoEntry
}

/**************************************************
* main entry
***************************************************/
func parsePath(path string) *PathInfo {
	if strings.HasPrefix(path, *basePath) {
		path = strings.TrimLeft(path, *basePath)
	}
	if path == "/favicon.ico" { // annoying
		return nil
	}
	fmt.Printf("Request for \"%s\"\n", path)
	res := PathInfo{output: "html"} // default serve result as html

	if path == "/" || path == "" {
		res.ReqTarget = listRepos
		return &res
	}
	/*	path, err := filepath.Rel("/", path)
		if err != nil {
			fmt.Printf("Failed to parse url path \"%s\": %s\n", path, err)
			return nil
		}
	*/
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
		fmt.Printf("#%d: \"%s\"\n", i, s)
		if i == 0 {
			res.Repository = s
			res.ReqTarget = listBranches
		} else if i == 1 {
			res.Branch = s
			res.ReqTarget = listVersions
		} else if i == 2 {
			res.Version = s
			res.ReqTarget = listFiles
		} else if i == 3 {
			res.Filename = s
			res.ReqTarget = getFile
		}
	}
	return &res
}

// we serve a directory listen or send a file, depending
// on what we find on disk
func serveContent(pi *PathInfo, w http.ResponseWriter, r *http.Request) error {

	var err error
	var entries []*pb.RepoEntry

	ctx := context.Background()
	client := pb.NewBuildRepoManagerClient(buildconn)
	if buildconn == nil {
		return errors.New("Unable to create RPC client")
	}
	ri := RenderInfo{
		Pi: pi,
		Top: *basePath,
	}
	// default template (only different for files)
	tmplname := fmt.Sprintf("templates/directory_listing.%s", pi.output)

	if pi.ReqTarget == listRepos {
		// list repos
		sr := &pb.ListReposRequest{}
		sa, e := client.ListRepos(ctx, sr)
		if e != nil {
			return e
		}
		entries = sa.Entries
		ri.Path = ""
	} else if pi.ReqTarget == listBranches {
		// list branches
		sr := pb.ListBranchesRequest{Repository: pi.Repository}
		sa, e := client.ListBranches(ctx, &sr)
		if e != nil {
			return e
		}
		entries = sa.Entries
		ri.Path = fmt.Sprintf("%s", pi.Repository)
	} else if pi.ReqTarget == listVersions {
		// list versions
		sr := pb.ListVersionsRequest{Repository: pi.Repository,
			Branch: pi.Branch}
		sa, e := client.ListVersions(ctx, &sr)
		if e != nil {
			return e
		}
		ri.Path = fmt.Sprintf("%s/%s", pi.Repository, pi.Branch)
		entries = sa.Entries
	} else if pi.ReqTarget == listFiles {
		// list files
		sr := pb.ListFilesRequest{Repository: pi.Repository,
			Branch: pi.Branch, Version: pi.Version}
		sa, e := client.ListFiles(ctx, &sr)
		if e != nil {
			return e
		}
		ri.Path = fmt.Sprintf("%s/%s/%s", pi.Repository, pi.Branch, pi.Version)
		entries = sa.Entries

	} else {

		return errors.New(fmt.Sprintf("Type %d not implemented", pi.ReqTarget))
	}
	if err != nil {
		fmt.Println("Failed to retrieve %v: %s\n", pi, err)
		return err
	}

	// move the path in to our subpath
	ri.Path = filepath.Clean(fmt.Sprintf("%s/%s", *basePath, ri.Path))

	tmpl, err := template.ParseFiles(tmplname)
	if err != nil {
		fmt.Printf("Failed to parse template %s: %s\n", tmplname, err)
		return err
	}
	ri.Entries = entries
	err = tmpl.Execute(w, ri)
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

	flag.Parse() // parse stuff. see "var" section above

	http.HandleFunc("/", handleRequest)

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	fmt.Println("Connecting to server...")
	if buildconn, err := grpc.Dial(*buildRepo, opts...); err != nil {
		fmt.Printf("fail to dial: %v", err)
		os.Exit(10)
	}

	//defer buildconn.Close()
	ctx := context.Background()
	client := pb.NewBuildRepoManagerClient(buildconn)
	fmt.Printf("Created connection %v and %v\n", ctx, client)
	fmt.Printf("Listening on port %d...\n", *httpPort)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), nil); err != nil {
		fmt.Printf("Failed to listen on port %d: %s\n", *httpPort, err)
	}
}
