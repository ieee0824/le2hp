package main

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/ieee0824/getenv"
)

type option struct {
	cmd  string
	args []string
}

func getOpt() (*option, error) {
	cmd := getenv.String("SUB_CMD")
	argsString := getenv.String("SUB_CMD_ARGS")
	if cmd == "" {
		return nil, errors.New("command is empty")
	}
	return &option{
		cmd:  cmd,
		args: strings.Split(argsString, ","),
	}, nil
}

var backend *exec.Cmd

func init() {
	log.SetFlags(log.Llongfile)
	opt, err := getOpt()
	if err != nil {
		log.Fatalln(err)
	}

	backend = exec.Command(opt.cmd, opt.args...)
	if err := backend.Start(); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sig
		backend.Process.Kill()
	}()

	director := func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = ":8080"
	}

	rp := &httputil.ReverseProxy{Director: director}

	lambda.Start(httpadapter.NewV2(rp).ProxyWithContext)
}
