package main

// don't use it in an untrusted environment!

// also we WILL leak memory, because we don't clean up the maps if
// client doesn't go through the completeUpload RPC

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
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type StoreMetaData struct {
	StoreID    string
	StorePath  string
	BuildID    int
	CommitID   string
	Commitmsg  string
	Branch     string
	Repository string
}

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
	hooksdir = flag.String("hooks", "/srv/build-repository/hooks", "Directory to search for hooks")
	src      = rand.NewSource(time.Now().UnixNano())
	uploads  = make(map[string]UploadMetaData)
	storeids = make(map[string]StoreMetaData)
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
	umd.Storepath = StoreIDToDir(umd.Storeid)
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

func StoreIDToDir(storeid string) string {
	return storeids[storeid].StorePath
}
func DirToStoreID(dir string, buildid int, commitid string, msg string, branch string, repo string) string {
	id := RandString(128)
	smd := StoreMetaData{
		BuildID:    buildid,
		CommitID:   commitid,
		Commitmsg:  msg,
		Branch:     branch,
		Repository: repo,
	}
	smd.StorePath = dir
	smd.StoreID = id
	storeids[id] = smd
	return id
}
func getStoreByID(id string) StoreMetaData {
	return storeids[id]
}
func removeStoreID(id string) {
	delete(storeids, id)
}

// check if it's a valid name for a repo or branch,
// basically no / or .. or . or so allowed
func isValidName(path string) bool {
	if path == "" {
		return false
	}
	if strings.Contains(path, "/") {
		return false
	}
	if strings.Contains(path, ".") {
		return false
	}
	return true
}

/**************************************************
* main entry
***************************************************/
func main() {
	flag.Parse() // parse stuff. see "var" section above
	listenAddr := fmt.Sprintf(":%d", *port)
	*httpport = *httpport
	lis, err := net.Listen("tcp4", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	http.HandleFunc("/", upload)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	s := new(BuildRepoServer)
	pb.RegisterBuildRepoManagerServer(grpcServer, s) // created by proto

	fmt.Printf("Starting  http service on port %d\n", *httpport)
	go http.ListenAndServe(fmt.Sprintf(":%d", *httpport), nil)

	fmt.Printf("Starting BuildRepo Manager service on %s\n", listenAddr)
	grpcServer.Serve(lis)

}

func upload(w http.ResponseWriter, r *http.Request) {
	//fmt.Printf("HTTP Upload method: %s, content-type: %s\n", r.Method, r.Header.Get("Content-Type"))
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
	return

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
	resp.BuildStoreid = DirToStoreID(dir, int(cr.BuildID), cr.CommitID, cr.CommitMSG, cr.Branch, cr.Repository)

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
	sp := StoreIDToDir(pr.BuildStoreid)
	if !strings.HasPrefix(sp, base) {
		fmt.Printf("Base=\"%s\", but token sent was: \"%s\"\n", base, sp)
		return res, errors.New("storeid is invalid")
	}
	fbase := filepath.Dir(fname)

	absDir := fmt.Sprintf("%s/%s", sp, fbase)
	//fmt.Printf("Filebase: \"%s\" (%s)\n", fbase, absDir)
	err := os.MkdirAll(absDir, 0777)
	if err != nil {
		fmt.Println("Failed to create directory ", absDir, err)
		return res, err
	}
	//fmt.Printf("Request to upload file \"%s\" to store \"%s\"\n", fname, storeid)
	token := RandString(256)
	res.Token = token
	res.Port = int32(*httpport)
	StoreUploadMetaData(fname, token, storeid)
	return res, nil
}

// all done, now call any hooks we might find
func (s *BuildRepoServer) UploadsComplete(ctx context.Context, udr *pb.UploadDoneRequest) (*pb.UploadDoneResponse, error) {
	resp := &pb.UploadDoneResponse{}
	if udr.BuildStoreid == "" {
		fmt.Println("No builds store id (to complete)")
		return nil, errors.New("missing build store id")

	}
	store := getStoreByID(udr.BuildStoreid)
	path := store.StorePath
	fmt.Println("Completed uploads for:", path)
	if !strings.HasPrefix(path, base) {
		fmt.Printf("Invalid path \"%s\", must start with \"%s\"\n", path, base)
		return nil, errors.New("Invalid path")
	}
	relpath := strings.SplitAfter(path, base)
	if len(relpath) != 2 {
		fmt.Printf("Invalid len (%d)\n", len(relpath))
		return nil, errors.New("Invalid path")
	}
	hookdir := fmt.Sprintf("%s/%s", *hooksdir, relpath[1])
	hookdir = filepath.Clean(hookdir)
	hookdir = filepath.Dir(hookdir)
	fmt.Printf("Looking for hooks in %s\n", hookdir)
	execute(store, hookdir, "post-upload")
	removeStoreID(store.StoreID)
	return resp, nil
}

// returns true if it executed it, false if file not found
// and error if file exists but something went wrong
func execute(store StoreMetaData, dir string, scriptname string) (bool, error) {
	fname := fmt.Sprintf("%s/%s", dir, scriptname)
	st, err := os.Stat(fname)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%s does not exist\n", fname)
			return false, nil
		}
		return false, err
	}
	if !st.Mode().IsRegular() {
		fmt.Printf("%s is not a regular file\n", fname)
		return false, errors.New("hook is not a regular file")
	}
	fmt.Println("Executing ", fname)
	cmd := exec.Command(fname)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("COMMIT_ID=%s", store.CommitID),
		fmt.Sprintf("COMMIT_MSG=%s", store.Commitmsg),
		fmt.Sprintf("GIT_BRANCH=%s", store.Branch),
		fmt.Sprintf("BUILD_NUMBER=%d", store.BuildID),
		fmt.Sprintf("PROJECT_NAME=%s", store.Repository),
		fmt.Sprintf("REPOSITORY=%s", store.Repository),
		fmt.Sprintf("DIST=%s", store.StorePath),
		fmt.Sprintf("BUILDDIR=%s", store.StorePath),
		fmt.Sprintf("ARTEFACT=%s", store.StorePath),
	)
	out, err := cmd.CombinedOutput()
	fmt.Printf("Output: %s\n", out)
	if err != nil {
		fmt.Printf("failed to execute %s: %s\n", fname, err)
		return true, err
	}

	return true, nil
}
func (s *BuildRepoServer) ListRepos(ctx context.Context, req *pb.ListReposRequest) (*pb.ListReposResponse, error) {
	res := pb.ListReposResponse{}
	e, err := ReadEntries(base)
	res.Entries = e
	if err != nil {
		return nil, err
	}
	return &res, nil
}
func (s *BuildRepoServer) ListBranches(ctx context.Context, req *pb.ListBranchesRequest) (*pb.ListBranchesResponse, error) {
	repo := req.Repository
	fmt.Printf("Listing branches of repository %s\n", repo)
	if !isValidName(repo) {
		return nil, errors.New(fmt.Sprintf("Invalid name \"%s\"", repo))
	}
	repodir := fmt.Sprintf("%s/%s", base, repo)
	res := pb.ListBranchesResponse{}
	e, err := ReadEntries(repodir)
	res.Entries = e
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *BuildRepoServer) ListVersions(ctx context.Context, req *pb.ListVersionsRequest) (*pb.ListVersionsResponse, error) {
	repo := req.Repository
	if !isValidName(repo) {
		return nil, errors.New(fmt.Sprintf("Invalid repo name \"%s\"", repo))
	}
	branch := req.Branch
	if !isValidName(branch) {
		return nil, errors.New(fmt.Sprintf("Invalid branch name \"%s\"", branch))
	}
	fmt.Printf("Listing versions for repo %s and branch %s\n", repo, branch)
	repodir := fmt.Sprintf("%s/%s/%s", base, repo, branch)
	res := pb.ListVersionsResponse{}
	e, err := ReadEntries(repodir)
	res.Entries = e
	if err != nil {
		return nil, err
	}
	return &res, nil
}
func (s *BuildRepoServer) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	repo := req.Repository
	if !isValidName(repo) {
		return nil, errors.New(fmt.Sprintf("Invalid repo name \"%s\"", repo))
	}
	branch := req.Branch
	if !isValidName(branch) {
		return nil, errors.New(fmt.Sprintf("Invalid branch name \"%s\"", branch))
	}
	build := req.Version
	if !isValidName(build) {
		return nil, errors.New(fmt.Sprintf("Invalid build name \"%s\"", build))
	}
	fmt.Printf("Listing versions for repo %s and branch %s and build %s\n", repo, branch, build)
	repodir := fmt.Sprintf("%s/%s/%s/%s", base, repo, branch, build)
	res := pb.ListFilesResponse{}
	x, err := ReadEntries(repodir)
	if err != nil {
		return nil, err
	}
	res.Entries = x
	return &res, nil
}

func ReadEntries(dir string) ([]*pb.RepoEntry, error) {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var res []*pb.RepoEntry
	for idx, fi := range fis {
		fmt.Printf("%d. Repo: %s\n", idx, fi.Name())
		re := pb.RepoEntry{}
		re.Name = fi.Name()
		re.Type = 1
		if fi.IsDir() {
			re.Type = 2
		}
		res = append(res, &re)
	}
	return res, nil
}
