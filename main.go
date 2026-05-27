package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mcp-claude-tools/internal/mcp"
	"mcp-claude-tools/internal/tools"
)

var (
	shellMgr *tools.ShellManager
	fileTool *tools.FileTool
)

func main() {
	// Inisialisasi direktori kerja aman (bisa disesuaikan lewat ENV)
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Gagal mengambil working directory: %v", err)
	}
	
	shellMgr = tools.NewShellManager()
	fileTool = tools.NewFileTool(baseDir)

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", handleMCP)

	// Proteksi Timeout HTTP Server Server untuk mencegah Slowloris
	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	// Mengatur saluran penangkapan sinyal penutupan (Graceful Shutdown)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server MCP Claude Tools berjalan pro di http://localhost:8080/mcp")
		log.Printf("Direktori File Aman diatur ke: %s", baseDir)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Menutup server secara bertahap (Graceful Shutdown)...")
	// Di sini Anda bisa menambahkan logic untuk mematikan sisa bash proses latar belakang jika diperlukan
	log.Println("Server berhasil dihentikan dengan aman.")
}

func handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Hanya menerima metode POST", http.StatusMethodNotAllowed)
		return
	}

	var req mcp.JSONRPCRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, -32700, "Parse error JSON tidak valid", req.ID)
		return
	}

	if req.Method != "tools/call" {
		sendError(w, -32601, "Metode tidak ditemukan", req.ID)
		return
	}

	var outputText string
	var toolErr error

	args := req.Params.Arguments

	switch req.Params.Name {
	case "bash":
		cmd, _ := args["command"].(string)
		to, _ := args["timeout"].(float64)
		bg, _ := args["run_in_bg"].(bool)
		bgID, _ := args["bg_id"].(string)
		outputText, toolErr = shellMgr.RunBash(cmd, int(to), bg, bgID)

	case "bash_output":
		bgID, _ := args["bg_id"].(string)
		out, serr, running := shellMgr.GetBashOutput(bgID)
		outputText = fmt.Sprintf("=== STDOUT ===\n%s\n=== STDERR ===\n%s\nStatus: Berjalan (%t)", out, serr, running)

	case "kill_shell":
		bgID, _ := args["bg_id"].(string)
		toolErr = shellMgr.KillShell(bgID)
		if toolErr == nil {
			outputText = fmt.Sprintf("Proses %s berhasil dimatikan.", bgID)
		}

	case "baca":
		path, _ := args["path"].(string)
		offset, _ := args["offset_lines"].(float64)
		limit, _ := args["limit_lines"].(float64)
		outputText, toolErr = fileTool.Baca(path, int(offset), int(limit))

	case "tulis":
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)
		outputText, toolErr = fileTool.Tulis(path, content)

	case "sunting":
		path, _ := args["path"].(string)
		oldStr, _ := args["old_string"].(string)
		newStr, _ := args["new_string"].(string)
		outputText, toolErr = fileTool.Sunting(path, oldStr, newStr)

	case "glob":
		pattern, _ := args["pattern"].(string)
		var matches []string
		matches, toolErr = fileTool.Glob(pattern)
		if toolErr == nil {
			data, _ := json.MarshalIndent(matches, "", "  ")
			outputText = string(data)
		}

	case "grep":
		pattern, _ := args["pattern"].(string)
		outputText, toolErr = fileTool.Grep(pattern)

	default:
		sendError(w, -32601, fmt.Sprintf("Alat '%s' tidak dikenali", req.Params.Name), req.ID)
		return
	}

	response := mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	if toolErr != nil {
		response.Result = mcp.ToolResult{
			Content: []mcp.TextContent{{Type: "text", Text: toolErr.Error()}},
			IsError: true,
		}
	} else {
		response.Result = mcp.ToolResult{
			Content: []mcp.TextContent{{Type: "text", Text: outputText}},
			IsError: false,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func sendError(w http.ResponseWriter, code int, message string, id interface{}) {
	w.Header().Set("Content-Type", "application/json")
	resp := mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		Error:   &mcp.RPCError{Code: code, Message: message},
		ID:      id,
	}
	_ = json.NewEncoder(w).Encode(resp)
}
