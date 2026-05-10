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
const VERSION = ""
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

	var cmd *exec.Cmd
	if err := sh.Run("protoc", "--proto_path=./internal"); err != nil {
		return err
	}
	// Complex command: -proto_path=./third_party \
	cmd = exec.Command("-proto_path=./third_party")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -go_out=paths=source_relative:./internal \
	cmd = exec.Command("-go_out=paths=source_relative:./internal")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil

}
// Api runs api
func Api() error {

	var cmd *exec.Cmd
	if err := sh.Run("protoc", "--proto_path=./api"); err != nil {
		return err
	}
	// Complex command: -proto_path=./third_party \
	cmd = exec.Command("-proto_path=./third_party")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -go_out=paths=source_relative:./api \
	cmd = exec.Command("-go_out=paths=source_relative:./api")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -go-http_out=paths=source_relative:./api \
	cmd = exec.Command("-go-http_out=paths=source_relative:./api")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -go-grpc_out=paths=source_relative:./api \
	cmd = exec.Command("-go-grpc_out=paths=source_relative:./api")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -openapi_out=fq_schema_naming=true,default_response=false:. \
	cmd = exec.Command("-openapi_out=fq_schema_naming=true,default_response=false:.")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil

}
// Build runs build
func Build() error {

	var cmd *exec.Cmd
	// Complex command: mkdir -p bin/ && go build -ldflags -X main.Version= -o ./bin/ ./...
	cmd = exec.Command("mkdir", "-p", "bin/", "&&", "go", "build", "-ldflags", "-X", "main.Version=", "-o", "./bin/", "./...")
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
	if err := sh.Run("/^[a-zA-Z-_0-9]+:/", "{"); err != nil {
		return err
	}
	// Complex command: helpMessage = match(lastLine, /^# (.*)/); \
	cmd = exec.Command("helpMessage", "=", "match(lastLine,", "/^#", "(.*)/);")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := sh.Run("if", "(helpMessage)", "{"); err != nil {
		return err
	}
	// Complex command: helpCommand = substr($$1, 0, index($$1, ":")); \
	cmd = exec.Command("helpCommand", "=", "substr($$1,", "0,", "index($$1,", ":));")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
	cmd = exec.Command("helpMessage", "=", "substr(lastLine,", "RSTART", "+", "2,", "RLENGTH);")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
	cmd = exec.Command("printf", "033[36m%-22s033[0m %sn,", "helpCommand,helpMessage;")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := sh.Run("}"); err != nil {
		return err
	}
	if err := sh.Run("}"); err != nil {
		return err
	}
	if err := sh.Run("{", "lastLine", "=", "$$0", "}"); err != nil {
		return err
	}
	return nil

}
