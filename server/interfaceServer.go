package main

import (
    "os"
    "fmt"
    "net/http"
    //
//    "github.com/golangdaddy/tarantula/router/common"
    standard "github.com/golangdaddy/tarantula/router/standard"
    "github.com/golangdaddy/tarantula/markup"
    "github.com/golangdaddy/tarantula/web/validation"
)

const (
    CONST_SERVER_PORT = 7005
    CONST_ARTEFACTS_PATH = "/srv/build-repository/artefacts"
    CONST_PAGE_TITLE = "BUILD-REPO SERVER"
)

func (app *App) ServeInterface() {

    app.page_head = g.HEAD().Add(
        g.STYLESHEET("https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css"),
    )

    if app.Error(
        os.MkdirAll(CONST_ARTEFACTS_PATH, 0600),
    ) {
        panic("SERVER FAILED...")
    }

    root, router := standard.NewRouter(app, "buildrepo")

    root.Add("/login")

        repos := root.Add("/repos")

            repos.GET(app.apiRepos)

        repo := repos.Param(validation.String(1, 64), "repo")

            repo.GET(app.apiRepo)

        branch := repo.Param(validation.String(1, 64), "branch")

            branch.GET(app.apiRepoBranch)

        build := branch.Param(validation.Int(), "build")

            build.GET(app.apiRepoBranchBuild)

        binary := build.Param(validation.String(1, 64), "binary")

            binary.GET(app.apiRepoBranchBuildBinary)


    root.Add("/files").Folder(CONST_ARTEFACTS_PATH)

    app.Debugf("SERVING ON PORT: %d", CONST_SERVER_PORT)

    panic(http.ListenAndServe(fmt.Sprintf(":%d", CONST_SERVER_PORT), router))
}
