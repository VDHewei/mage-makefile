//go:build mage
// +build mage

package main

import (
	"github.com/magefile/mage/sh"
	"os"
	"os/exec"
)

// GOHOSTOS - defined from Makefile variable
const GOHOSTOS = "windows"
// GOPATH - defined from Makefile variable
const GOPATH = "C:\\Users\\wei\\go"
// VERSION - defined from Makefile variable
const VERSION = "81258cd"
// _DEFAULT_GOAL - defined from Makefile variable
const _DEFAULT_GOAL = "help"
// Git_Bash - defined from Makefile variable
const Git_Bash = "C:/Program Files/Git/mingw64/bin/ C:/Program Files/Git/bin/bash.exe"
// INTERNAL_PROTO_FILES - defined from Makefile variable
const INTERNAL_PROTO_FILES = ""
// API_PROTO_FILES - defined from Makefile variable
const API_PROTO_FILES = ""

// Init runs init
func Init() error {

	if err := sh.Run("go", "install", "google.golang.org/protobuf/cmd/protoc-gen-go@latest"); err != nil {
		return err
	}
	if err := sh.Run("go", "install", "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"); err != nil {
		return err
	}
	if err := sh.Run("go", "install", "github.com/go-kratos/kratos/cmd/kratos/v2@latest"); err != nil {
		return err
	}
	if err := sh.Run("go", "install", "github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest"); err != nil {
		return err
	}
	if err := sh.Run("go", "install", "github.com/google/gnostic/cmd/protoc-gen-openapi@latest"); err != nil {
		return err
	}
	if err := sh.Run("go", "install", "github.com/google/wire/cmd/wire@latest"); err != nil {
		return err
	}
	return nil

}
// Config runs config
func Config() error {

	if err := sh.Run("protoc", "--proto_path=./internal", "--proto_path=./third_party", "--go_out=paths=source_relative:./internal"); err != nil {
		return err
	}
	return nil

}
// Api runs api
func Api() error {

	if err := sh.Run("protoc", "--proto_path=./api", "--proto_path=./third_party", "--go_out=paths=source_relative:./api", "--go-http_out=paths=source_relative:./api", "--go-grpc_out=paths=source_relative:./api", "--openapi_out=fq_schema_naming=true,default_response=false:."); err != nil {
		return err
	}
	return nil

}
// Build runs build
func Build() error {

	var cmd *exec.Cmd
	// Complex command: mkdir -p bin/ && go build -ldflags -X main.Version=81258cd -o ./bin/ ./...
	cmd = exec.Command("sh", "-c", "mkdir -p bin/ && go build -ldflags -X main.Version=81258cd -o ./bin/ ./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil

}
// Generate runs generate
func Generate() error {

	if err := sh.Run("go", "generate", "./..."); err != nil {
		return err
	}
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return err
	}
	return nil

}
// All runs all
func All() error {

	if err := sh.Run("make", "api"); err != nil {
		return err
	}
	if err := sh.Run("make", "config"); err != nil {
		return err
	}
	if err := sh.Run("make", "generate"); err != nil {
		return err
	}
	return nil

}
// Help runs help
func Help() error {

	var cmd *exec.Cmd
	if err := sh.Run("echo"); err != nil {
		return err
	}
	if err := sh.Run("echo", "Usage:"); err != nil {
		return err
	}
	if err := sh.Run("echo", "make", "[target]"); err != nil {
		return err
	}
	if err := sh.Run("echo"); err != nil {
		return err
	}
	if err := sh.Run("echo", "Targets:"); err != nil {
		return err
	}
	// Complex command:  /^[a-zA-Z\-\_0-9]+:/ {  helpMessage = match(lastLine, /^# (.*)/);  if (helpMessage) {  helpCommand = substr($$1, 0, index($$1, ":"));  helpMessage = substr(lastLine, RSTART + 2, RLENGTH);  printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage;  }  }  { lastLine = $$0 }
	cmd = exec.Command("sh", "-c", " /^[a-zA-Z\\-\\_0-9]+:/ {  helpMessage = match(lastLine, /^# (.*)/);  if (helpMessage) {  helpCommand = substr($$1, 0, index($$1, \":\"));  helpMessage = substr(lastLine, RSTART + 2, RLENGTH);  printf \"\\033[36m%-22s\\033[0m %s\\n\", helpCommand,helpMessage;  }  }  { lastLine = $$0 }")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil

}
