package main

import (
    "os"
    "strings"
    "strconv"
    "io/ioutil"
    "path/filepath"
    //
    "github.com/golangdaddy/tarantula/markup"
    "github.com/golangdaddy/tarantula/web"
)

func breadcrumbs(a []string) *g.ELEMENT {

    c := g.DIV().Class("row")

    for i, x := range a {

        c.Add(
            g.DIV().Class("col-sm").Add(
                g.H(3, "/"),
            ),
            g.DIV().Class("col-sm").Add(
                g.A().Href(
                    "/" + strings.Join(a[:i+1], "/"),
                ).Add(
                    g.H(3, x),
                ),
            ),
        )

    }

    return g.DIV().Class("container").Add(
        c,
    )
}

func (app *App) apiRepos(req web.RequestInterface) *web.ResponseStatus {

    files, err := ioutil.ReadDir(CONST_ARTEFACTS_PATH)
    if app.Error(err) {
        return req.Fail()
    }

    page_body := g.BODY()

    page := g.HTML().Add(
        app.page_head.New().Add(
            g.TITLE(CONST_PAGE_TITLE + " - " + "repositories"),
        ),
        page_body,
    )

    crumbs := []string{"repos"}

    page_body.Add(
        breadcrumbs(crumbs),
    )

    content := g.DIV().Class("column")

    for _, f := range files {

        name := f.Name()

        content.Add(
            g.DIV().Add(
                g.A().Href(
                    "/" + strings.Join(crumbs, "/") + "/" + name,
                ).Add(
                    g.BUTTON().Inner(name),
                ),
            ),
        )

    }

    page.Add(content)

    return req.Respond(page)
}

func (app *App) apiRepo(req web.RequestInterface) *web.ResponseStatus {

    repo := req.Param("repo").(string)

    files, err := ioutil.ReadDir(
        strings.Join(
            []string{CONST_ARTEFACTS_PATH, repo},
            "/",
        ),
    )
    if app.Error(err) {
        return req.Fail()
    }

    page_body := g.BODY()

    page := g.HTML().Add(
        app.page_head.New().Add(
            g.TITLE(CONST_PAGE_TITLE + " - " + "builds"),
        ),
        page_body,
    )

    crumbs := []string{"repos", repo}

    page_body.Add(
        breadcrumbs(crumbs),
    )

    content := g.DIV().Class("column")

    for _, f := range files {

        name := f.Name()

        content.Add(
            g.DIV().Add(
                g.A().Href(
                    "/" + strings.Join(crumbs, "/") + "/" + name,
                ).Add(
                    g.BUTTON().Inner(name),
                ),
            ),
        )

    }

    page.Add(content)

    return req.Respond(page)
}

func (app *App) apiRepoBranch(req web.RequestInterface) *web.ResponseStatus {

    repo := req.Param("repo").(string)
    branch := req.Param("branch").(string)

    files, err := ioutil.ReadDir(
        strings.Join(
            []string{CONST_ARTEFACTS_PATH, repo, branch},
            "/",
        ),
    )
    if app.Error(err) {
        return req.Fail()
    }

    page_body := g.BODY()

    page := g.HTML().Add(
        app.page_head.New().Add(
            g.TITLE(CONST_PAGE_TITLE + " - " + "builds"),
        ),
        page_body,
    )

    crumbs := []string{"repos", repo, branch}

    page_body.Add(
        breadcrumbs(crumbs),
    )

    content := g.DIV().Class("column")

    for _, f := range files {

        name := f.Name()

        content.Add(
            g.DIV().Add(
                g.A().Href(
                    "/" + strings.Join(crumbs, "/") + "/" + name,
                ).Add(
                    g.BUTTON().Inner(name),
                ),
            ),
        )

    }

    page.Add(content)

    return req.Respond(page)
}

func (app *App) apiRepoBranchBuild(req web.RequestInterface) *web.ResponseStatus {

    repo := req.Param("repo").(string)
    branch := req.Param("branch").(string)
    build := req.Param("build").(int)

    files, err := ioutil.ReadDir(
        strings.Join(
            []string{CONST_ARTEFACTS_PATH, repo, branch, strconv.Itoa(build)},
            "/",
        ),
    )
    if app.Error(err) {
        return req.Fail()
    }

    page_body := g.BODY()

    page := g.HTML().Add(
        app.page_head.New().Add(
            g.TITLE(CONST_PAGE_TITLE + " - " + "builds"),
        ),
        page_body,
    )

    crumbs := []string{"repos", repo, branch, strconv.Itoa(build)}

    page_body.Add(
        breadcrumbs(crumbs),
    )

    content := g.DIV().Class("column")

    for _, f := range files {

        name := f.Name()

        content.Add(
            g.A().Href(
                "/" + strings.Join(crumbs, "/") + "/" + name,
            ).Add(
                g.BUTTON().Inner(name),
            ),
        )

    }

    page.Add(content)

    return req.Respond(page)
}

func (app *App) apiRepoBranchBuildBinary(req web.RequestInterface) *web.ResponseStatus {

    repo := req.Param("repo").(string)
    branch := req.Param("branch").(string)
    build := req.Param("build").(int)

    page_body := g.BODY()

    page := g.HTML().Add(
        app.page_head.New().Add(
            g.TITLE(CONST_PAGE_TITLE + " - " + "builds"),
        ),
        page_body,
    )

    crumbs := []string{"repos", repo, strconv.Itoa(build)}

    page_body.Add(
        breadcrumbs(crumbs),
    )

    content := g.DIV().Class("column")

    err := filepath.Walk(
        strings.Join(
            []string{CONST_ARTEFACTS_PATH, repo, branch, strconv.Itoa(build)},
            "/",
        ),
        func(path string, info os.FileInfo, err error) error {

            if err != nil {
                return err
            }

            name := info.Name()

            page_body.Add(
                g.DIV().Add(
                    g.A().Href(
                        "/files/" + strings.Join(crumbs[1:], "/") + "/" + name,
                    ).Add(
                        g.BUTTON().Inner(name),
                    ),
                ),
            )

            return nil
        },
    )
    if app.Error(err) {
        return req.Fail()
    }

    page_body.Add(
        content,
    )

    return req.Respond(page)
}
