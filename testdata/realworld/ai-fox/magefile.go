//go:build mage
// +build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"os"
	"os/exec"
)

// PATH - defined from Makefile variable
const PATH = "/.cargo/bin:$(HOME)/.cargo/bin:$(PATH)"
// MODELS_DIR - defined from Makefile variable
const MODELS_DIR = "models"
// MODEL_REPO - defined from Makefile variable
const MODEL_REPO = "unsloth/Qwen3.5-0.8B-GGUF"
// MODEL_FILE - defined from Makefile variable
const MODEL_FILE = "Qwen3.5-0.8B-Q4_K_M.gguf"
// MODEL_PATH - defined from Makefile variable
const MODEL_PATH = "models/Qwen3.5-0.8B-Q4_K_M.gguf"
// DOCKER_IMAGE - defined from Makefile variable
const DOCKER_IMAGE = "python:3.11-slim"
// HOST - defined from Makefile variable
const HOST = "0.0.0.0"
// PORT - defined from Makefile variable
const PORT = "8080"
// MAX_CONTEXT_LEN - defined from Makefile variable
const MAX_CONTEXT_LEN = "4096"
// GPU_MEM_FRACTION - defined from Makefile variable
const GPU_MEM_FRACTION = "0.85"
// MAX_BATCH_SIZE - defined from Makefile variable
const MAX_BATCH_SIZE = "32"
// BLOCK_SIZE - defined from Makefile variable
const BLOCK_SIZE = "16"
// BENCH_CONCURRENCY - defined from Makefile variable
const BENCH_CONCURRENCY = "4"
// BENCH_REQUESTS - defined from Makefile variable
const BENCH_REQUESTS = "50"
// BENCH_PROMPT - defined from Makefile variable
const BENCH_PROMPT = "Write a short paragraph about the Rust programming language."
// BENCH_MAX_TOKENS - defined from Makefile variable
const BENCH_MAX_TOKENS = "128"

// Help runs help
func Help() error {

	if err := sh.Run("echo", "Targets:"); err != nil {
		return err
	}
	if err := sh.Run("echo", "make", "install-rust", "Install", "Rust", "toolchain", "(run", "once", "if", "not", "installed)"); err != nil {
		return err
	}
	if err := sh.Run("echo", "make", "download-model", "Download", "Qwen3.5-0.8B-Q4_K_M.gguf", "from", "HuggingFace", "to", "models/"); err != nil {
		return err
	}
	if err := sh.Run("echo", "make", "build", "Compile", "release", "binaries"); err != nil {
		return err
	}
	if err := sh.Run("echo", "make", "run", "Build", "and", "start", "the", "server"); err != nil {
		return err
	}
	if err := sh.Run("echo", "make", "dev", "Start", "with", "verbose", "logging", "(RUST_LOG=debug)"); err != nil {
		return err
	}
	if err := sh.Run("echo", "make", "test", "Run", "unit", "tests"); err != nil {
		return err
	}
	if err := sh.Run("echo", "make", "check", "Fast", "type-check", "without", "producing", "a", "binary"); err != nil {
		return err
	}
	if err := sh.Run("echo", "make", "ci", "Run", "the", "full", "CI", "suite", "locally", "(fmt", "+", "clippy", "+", "tests)"); err != nil {
		return err
	}
	if err := sh.Run("echo", "make", "setup", "Install", "git", "pre-push", "hook", "(run", "once", "after", "cloning)"); err != nil {
		return err
	}
	if err := sh.Run("echo", "make", "bench", "Run", "the", "integrated", "benchmark", "against", "a", "running", "server"); err != nil {
		return err
	}
	if err := sh.Run("echo", "make", "docker", "Build", "the", "Docker", "image"); err != nil {
		return err
	}
	if err := sh.Run("echo", "make", "docker-run", "Start", "the", "server", "via", "docker", "compose"); err != nil {
		return err
	}
	if err := sh.Run("echo"); err != nil {
		return err
	}
	if err := sh.Run("echo", "Variables", "(override", "with", "make", "run", "VAR=value):"); err != nil {
		return err
	}
	if err := sh.Run("echo", "MODEL_PATH=models/Qwen3.5-0.8B-Q4_K_M.gguf"); err != nil {
		return err
	}
	if err := sh.Run("echo", "HOST=0.0.0.0", "PORT=8080"); err != nil {
		return err
	}
	if err := sh.Run("echo", "MAX_CONTEXT_LEN=4096"); err != nil {
		return err
	}
	if err := sh.Run("echo", "GPU_MEM_FRACTION=0.85"); err != nil {
		return err
	}
	if err := sh.Run("echo", "MAX_BATCH_SIZE=32"); err != nil {
		return err
	}
	if err := sh.Run("echo", "BENCH_CONCURRENCY=4", "BENCH_REQUESTS=50"); err != nil {
		return err
	}
	return nil

}
// InstallRust runs install-rust
func InstallRust() error {

	var cmd *exec.Cmd
	// Complex command: command -v cargo >/dev/null 2>&1 && \
	cmd = exec.Command("command", "-v", "cargo", ">/dev/null", "2>&1", "&&")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: (echo "Rust already installed:"; cargo --version) || \
	cmd = exec.Command("(echo", "Rust already installed:;", "cargo", "--version)", "||")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: (curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y && \
	cmd = exec.Command("(curl", "--proto", "=https", "--tlsv1.2", "-sSf", "https://sh.rustup.rs", "|", "sh", "-s", "--", "-y", "&&")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: . /.cargo/env && cargo --version && \
	cmd = exec.Command(".", "/.cargo/env", "&&", "cargo", "--version", "&&")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: echo Rust installed. Run: source ~/.cargo/env && make build)
	cmd = exec.Command("echo", "Rust", "installed.", "Run:", "source", "~/.cargo/env", "&&", "make", "build)")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil

}
// DownloadModel runs download-model
func DownloadModel() error {

	var cmd *exec.Cmd
	if err := sh.Run("mkdir", "-p", "models"); err != nil {
		return err
	}
	if err := sh.Run("echo", "Downloading", "Qwen3.5-0.8B-Q4_K_M.gguf", "from", "unsloth/Qwen3.5-0.8B-GGUF..."); err != nil {
		return err
	}
	if err := sh.Run("docker", "run", "--rm"); err != nil {
		return err
	}
	// Complex command: e PIP_ROOT_USER_ACTION=ignore \
	cmd = exec.Command("e", "PIP_ROOT_USER_ACTION=ignore")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: v "/models:/data" \
	cmd = exec.Command("v", "/models:/data")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: w /data \
	cmd = exec.Command("w", "/data")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := sh.Run("python:3.11-slim"); err != nil {
		return err
	}
	// Complex command: sh -c "pip install --quiet huggingface_hub && \
	cmd = exec.Command("sh", "-c", "pip install --quiet huggingface_hub && ")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: python -c \"from huggingface_hub import hf_hub_download; \
	cmd = exec.Command("python", "-c", "\"from", "huggingface_hub", "import", "hf_hub_download;")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := sh.Run("hf_hub_download(repo_id=unsloth/Qwen3.5-0.8B-GGUF,", "filename=Qwen3.5-0.8B-Q4_K_M.gguf,", "local_dir=.)\""); err != nil {
		return err
	}
	if err := sh.Run("echo", "Model", "saved", "to", "models/Qwen3.5-0.8B-Q4_K_M.gguf"); err != nil {
		return err
	}
	return nil

}
// Check runs check
func Check() error {

	if err := sh.Run("cargo", "check"); err != nil {
		return err
	}
	return nil

}
// Ci runs ci
func Ci() error {

	if err := sh.Run("FOX_SKIP_LLAMA=1", "cargo", "fmt", "--all", "--", "--check"); err != nil {
		return err
	}
	if err := sh.Run("FOX_SKIP_LLAMA=1", "cargo", "clippy", "--all-targets", "--features", "test-helpers", "--", "-D", "warnings"); err != nil {
		return err
	}
	if err := sh.Run("FOX_SKIP_LLAMA=1", "cargo", "test", "--all", "--features", "test-helpers"); err != nil {
		return err
	}
	return nil

}
// Setup runs setup
func Setup() error {

	if err := sh.Run("bash", "scripts/install-hooks.sh"); err != nil {
		return err
	}
	return nil

}
// Build runs build
func Build() error {

	var cmd *exec.Cmd
	// Complex command: command -v cargo >/dev/null 2>&1 || \
	cmd = exec.Command("command", "-v", "cargo", ">/dev/null", "2>&1", "||")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: (echo "Rust not found. Run: make install-rust" && exit 1)
	cmd = exec.Command("(echo", "Rust not found. Run: make install-rust", "&&", "exit", "1)")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := sh.Run("cargo", "build", "--release"); err != nil {
		return err
	}
	return nil

}
// Run runs run (build)
func Run() error {

	mg.Deps(mg.F(Build))
	var cmd *exec.Cmd
	// Complex command: test -f "models/Qwen3.5-0.8B-Q4_K_M.gguf" || \
	cmd = exec.Command("test", "-f", "models/Qwen3.5-0.8B-Q4_K_M.gguf", "||")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: (echo "Model not found at models/Qwen3.5-0.8B-Q4_K_M.gguf. Run: make download-model" && exit 1)
	cmd = exec.Command("(echo", "Model not found at models/Qwen3.5-0.8B-Q4_K_M.gguf. Run: make download-model", "&&", "exit", "1)")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := sh.Run("./target/release/fox"); err != nil {
		return err
	}
	// Complex command: -model-path models/Qwen3.5-0.8B-Q4_K_M.gguf \
	cmd = exec.Command("-model-path", "models/Qwen3.5-0.8B-Q4_K_M.gguf")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -host 0.0.0.0 \
	cmd = exec.Command("-host", "0.0.0.0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -port 8080 \
	cmd = exec.Command("-port", "8080")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -max-context-len 4096 \
	cmd = exec.Command("-max-context-len", "4096")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -gpu-memory-fraction 0.85 \
	cmd = exec.Command("-gpu-memory-fraction", "0.85")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -max-batch-size 32
	cmd = exec.Command("-max-batch-size", "32")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil

}
// Dev runs dev (build)
func Dev() error {

	mg.Deps(mg.F(Build))
	var cmd *exec.Cmd
	// Complex command: test -f "models/Qwen3.5-0.8B-Q4_K_M.gguf" || \
	cmd = exec.Command("test", "-f", "models/Qwen3.5-0.8B-Q4_K_M.gguf", "||")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: (echo "Model not found at models/Qwen3.5-0.8B-Q4_K_M.gguf. Run: make download-model" && exit 1)
	cmd = exec.Command("(echo", "Model not found at models/Qwen3.5-0.8B-Q4_K_M.gguf. Run: make download-model", "&&", "exit", "1)")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := sh.Run("RUST_LOG=debug", "./target/release/fox"); err != nil {
		return err
	}
	// Complex command: -model-path models/Qwen3.5-0.8B-Q4_K_M.gguf \
	cmd = exec.Command("-model-path", "models/Qwen3.5-0.8B-Q4_K_M.gguf")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -host 0.0.0.0 \
	cmd = exec.Command("-host", "0.0.0.0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -port 8080 \
	cmd = exec.Command("-port", "8080")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -max-context-len 4096 \
	cmd = exec.Command("-max-context-len", "4096")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -gpu-memory-fraction 0.85 \
	cmd = exec.Command("-gpu-memory-fraction", "0.85")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -max-batch-size 32
	cmd = exec.Command("-max-batch-size", "32")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil

}
// Test runs test
func Test() error {

	if err := sh.Run("cargo", "test"); err != nil {
		return err
	}
	return nil

}
// Bench runs bench (build)
func Bench() error {

	mg.Deps(mg.F(Build))
	var cmd *exec.Cmd
	if err := sh.Run("echo", "Running", "benchmark", "against", "0.0.0.0:8080..."); err != nil {
		return err
	}
	if err := sh.Run("./target/release/fox-bench"); err != nil {
		return err
	}
	// Complex command: -url http://0.0.0.0:8080 \
	cmd = exec.Command("-url", "http://0.0.0.0:8080")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -model Qwen3.5-0.8B-Q4_K_M.gguf \
	cmd = exec.Command("-model", "Qwen3.5-0.8B-Q4_K_M.gguf")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -concurrency 4 \
	cmd = exec.Command("-concurrency", "4")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -requests 50 \
	cmd = exec.Command("-requests", "50")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -max-tokens 128 \
	cmd = exec.Command("-max-tokens", "128")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Complex command: -prompt "Write a short paragraph about the Rust programming language."
	cmd = exec.Command("-prompt", "Write a short paragraph about the Rust programming language.")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil

}
// Docker runs docker
func Docker() error {

	if err := sh.Run("docker", "build", "-t", "ferrumox:latest", "."); err != nil {
		return err
	}
	return nil

}
// DockerRun runs docker-run
func DockerRun() error {

	if err := sh.Run("docker", "compose", "up"); err != nil {
		return err
	}
	return nil

}
