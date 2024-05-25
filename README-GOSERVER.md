## Sample RestAPI

### Essential

- Go 1.20

### Steps For Basic Setup

1. check presence of development.env file

2. run `go mod tidy` to install dependencies

   For thet you need to make some changes in your .bashrc file

3. check go path

   ```bash
   go env GOPATH
   ```

---

### Starting the API's

1. Execute following to run go server:

   `go run cmd/app.go`

   `-e` flag can be use to give any configuration file other than `.development.env`, E.g. to use `.test.env` as active configuration the following command can be used:

   `go run cmd/app.go -e test`

2. Following can be used to check setup:

   `curl 'localhost:7000/dopamine/v1/healthcheck'`

---

### Extras

1. Install [air](https://github.com/cosmtrek/air), this will help in auto-reloading of your server whenever you save any change.
2. To start server with this you just need to enter `air` in commandline.
3. Install [golangci lint](https://golangci-lint.run/usage/install/).
4. This will help to follow better coding standards.

---

### How things are working

---
