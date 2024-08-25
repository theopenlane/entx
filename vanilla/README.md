# Setting up a vanilla ent example

To setup a vanilla instance of `ent` to use as an example (or for a new repository), copy the entirety of the files / directory in `_example` to wherever it's destination is going to be. Once copied, update the import paths inside of any of the go files, most notably `gen_schema.go`.

Now, you'll want to init `go` by running a command similar to this _within the _example_ directory.

```bash
go mod init github.com/theopenlane/[YOUR_REPO_NAME]/_example
```

Be sure to update this to whatever path / directory the example will live in. This will create a `go.mod` and `go.sum` file that are needed to import the necessary packages used in the setup. Run `go mod tidy` after you init. You may need to install additional packages by using `go install` or `go get` for the respective packages.

Now, run `go generate ./...` in the `_example` directory - this should spit out a ton of new directories and files within `_example/ent`. The remainder of the tasks can be run using the Taskfile included in this example (e.g. task gqlgen).

Update your example to be specific to your repo / use case!