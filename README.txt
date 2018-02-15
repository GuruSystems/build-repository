It's a generic build-artefact-repository.
One pushes "builds" (as in binaries) to this repository.
Intent is to add hooks to the server to trigger other systems on any new build.


this is an entire GO directory.
set GOPATH to this level to compile

start server with --port [number]

submit files with
client --serveradddr=[ip:port] --branch=xxx --repository=xxx --commitid=xxxx --commitmsg=xxx --buildid=xxx [filenames...]



