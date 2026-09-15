//go:build r2locals3

package r2locals3

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	wranglerReadyWait    = 60 * time.Second
	wranglerPollInterval = 200 * time.Millisecond
)

// start は wrangler experimental local S3 を起動し Peer を返す。
// 失敗は error で見える化する（黙って skip しない）。
func start(ctx context.Context) (*Peer, func(), error) {
	if ctx == nil {
		ctx = context.Background()
	}
	pkgDir, err := packageDir()
	if err != nil {
		return nil, nil, err
	}
	wranglerBin, err := resolveWrangler(pkgDir)
	if err != nil {
		return nil, nil, err
	}
	configPath := filepath.Join(pkgDir, "wrangler.local-s3.jsonc")
	if _, err := os.Stat(configPath); err != nil {
		return nil, nil, fmt.Errorf("r2locals3: wrangler config missing")
	}

	port, err := freeLocalPort()
	if err != nil {
		return nil, nil, fmt.Errorf("r2locals3: allocate port: %w", err)
	}

	persistDir, err := os.MkdirTemp("", "r2locals3-persist-*")
	if err != nil {
		return nil, nil, fmt.Errorf("r2locals3: persist dir: %w", err)
	}

	cmd := exec.CommandContext(
		ctx,
		wranglerBin,
		"dev",
		"--config", configPath,
		"--ip", "127.0.0.1",
		"--port", strconv.Itoa(port),
		"--local",
		"--persist-to", persistDir,
	)
	cmd.Dir = pkgDir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var logBuf strings.Builder
	cmd.Stdout = &logBuf
	cmd.Stderr = &logBuf
	if err := cmd.Start(); err != nil {
		_ = os.RemoveAll(persistDir)
		return nil, nil, fmt.Errorf("r2locals3: start wrangler: %w", err)
	}

	cleanup := func() {
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			_, _ = cmd.Process.Wait()
		}
		_ = os.RemoveAll(persistDir)
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d%s", port, s3APIPathPrefix)
	if err := waitReady(ctx, port, &logBuf); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("r2locals3: wrangler not ready: %w; log=%s", err, sanitizePeerLog(logBuf.String()))
	}

	peer := &Peer{
		BaseURL:         baseURL,
		AccessKeyID:     localAccessKeyID,
		SecretAccessKey: localSecretAccessKey,
		AccountID:       localAccountID,
		Bucket:          localBucket,
	}
	return peer, cleanup, nil
}

func packageDir() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("r2locals3: resolve package dir")
	}
	return filepath.Dir(file), nil
}

func resolveWrangler(pkgDir string) (string, error) {
	root, err := findRepoRoot(pkgDir)
	if err != nil {
		return "", err
	}
	bin := filepath.Join(root, "apps", "playback", "node_modules", ".bin", "wrangler")
	if runtime.GOOS == "windows" {
		bin += ".cmd"
	}
	if st, err := os.Stat(bin); err != nil || st.IsDir() {
		return "", errors.New("r2locals3: wrangler binary missing (apps/playback npm install)")
	}
	return bin, nil
}

func findRepoRoot(start string) (string, error) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "apps", "playback", "package.json")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("r2locals3: repo root not found")
		}
		dir = parent
	}
}

func freeLocalPort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer func() { _ = ln.Close() }()
	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("not tcp")
	}
	return addr.Port, nil
}

func waitReady(ctx context.Context, port int, logBuf *strings.Builder) error {
	deadline := time.Now().Add(wranglerReadyWait)
	client := &http.Client{Timeout: 2 * time.Second}
	probe := fmt.Sprintf("http://127.0.0.1:%d/", port)
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("%w; log=%s", err, sanitizePeerLog(logBuf.String()))
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout; log=%s", sanitizePeerLog(logBuf.String()))
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, probe, nil)
		if err != nil {
			return err
		}
		res, err := client.Do(req)
		if err == nil {
			_ = res.Body.Close()
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w; log=%s", ctx.Err(), sanitizePeerLog(logBuf.String()))
		case <-time.After(wranglerPollInterval):
		}
	}
}
