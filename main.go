package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func invokeWasmModule(modname string, wasmPath string, env map[string]string) (string, error) {
	ctx := context.Background()

	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	_, err := r.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(func(v uint32) {
			log.Printf("[%v]: %v", modname, v)
		}).
		Export("log_i32").
		NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, ptr uint32, len uint32) {
			if bytes, ok := mod.Memory().Read(ptr, len); ok {
				log.Printf("[%v]: %v", modname, string(bytes))
			} else {
				log.Printf("[%v]: log_string: unable to read wasm memory", modname)
			}
		}).
		Export("log_string").
		Instantiate(ctx)

	if err != nil {
		return "", err
	}

	wasmObj, err := os.ReadFile(wasmPath)
	if err != nil {
		return "", err
	}

	var stdoutBuf bytes.Buffer
	config := wazero.NewModuleConfig().WithStdout(&stdoutBuf)

	for k, v := range env {
		config = config.WithEnv(k, v)
	}

	_, err = r.InstantiateWithConfig(ctx, wasmObj, config)
	if err != nil {
		return "", err
	}

	return stdoutBuf.String(), nil
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 1 {
		http.Error(w, "want /{modulename} prefix", http.StatusBadRequest)
	}

	mod := parts[0]
	log.Printf("module %v requested with query %v", mod, r.URL.Query())

	env := map[string]string{
		"http_path":   r.URL.Path,
		"http_method": r.Method,
		"http_host":   r.Host,
		"http_query":  r.URL.Query().Encode(),
		"remote_addr": r.RemoteAddr,
	}

	modpath := fmt.Sprintf("target/%v.wasm", mod)
	log.Printf("loading module %v", modpath)
	out, err := invokeWasmModule(mod, modpath, env)
	if err != nil {
		log.Printf("error loading module %v", modpath)
		http.Error(w, "unable to find module "+modpath, http.StatusNotFound)
		return
	}

	fmt.Fprint(w, out)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", httpHandler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
