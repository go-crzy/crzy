# crzy

Update your test environments in 3 seconds

## working principle

`crzy` is mainly 2 things: (1) a GIT server that triggers a `test`, `build` and
`run` on a receive; (2) an HTTP proxy that rollout the new version of your API.

## a simple example

The best way to understand how it works is to use it. You can check
[carnage-sh/color](https://github.com/carnage-sh/color) as an example. First
install and run `crzy`:

```shell
go get github.com/carnage-sh/crzy
crzy -server -repository color.git
```

> Note: we assume you are working on the `main` branch, if that is not the
> case, add the `-head` flag with the name of your branch to the `crzy`
> command.

This commands above starts the GIT server that can be used to push you
program to the `http://localhost:8080/color.git` URL. It also creates
a proxy server on `http://localhost:8081`. You can then push your project
the to `crzy`, with a simple `git push` command. Here is a complete example
that includes the setup:

```shell
git clone https://github.com/carnage-sh/color.git color
cd color
git remote add server http://localhost:8080/color.git
git push server
```

The API is now proxied and the next push will perform a blue/green update
of your test environment...

## the secret sauce

`crzy` is not magic and there is a few assumptions for your program to work
with it:

- We depend on `go` and `git` on the server
- We assume your program is using Go and can be build from a simple
  `go build` on the root of your repository. You should rely on `go mod`
- We assume the program relies on HTTP and we can change its listening port
  with the `PORT` environment variable
- We are running the `main` branch of your project

`crzy` will be improved to manage broader use cases. If you like the idea,
need support for another programming language or protocol or simply cannot
figure out how to make it work, do not hesitate to open an
[issue](https://github.com/carnage-sh/crzy/issues).

## known issues

`crzy` does not support Windows
