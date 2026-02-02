package api

import (
	"encoding/json"
	"net/http"

	"github.com/rohansx/vgx/pkg/scanner"
	"github.com/rohansx/vgx/pkg/types"
)

type ScanRequest struct {
    Files []string `json:"files"`
    Repo  string   `json:"repo"`
}

type ScanResponse struct {
    Status         string                `json:"status"`
    Vulnerabilities []types.Vulnerability `json:"vulnerabilities"`
    Message        string                `json:"message,omitempty"`
}

// HandleScan processes scan requests from clients
func HandleScan(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var req ScanRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendErrorResponse(w, "Invalid request format")
        return
    }
    
    // Process file scan
    vulnerabilities, err := scanner.ScanFiles(req.Files)
    if err != nil {
        sendErrorResponse(w, "Scan failed: "+err.Error())
        return
    }
    
    // Send response
    response := ScanResponse{
        Status: "success",
        Vulnerabilities: vulnerabilities,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// HandleHealth returns service health status
func HandleHealth(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
    })
}

func sendErrorResponse(w http.ResponseWriter, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    
    response := ScanResponse{
        Status:  "error",
        Message: message,
    }
    
    json.NewEncoder(w).Encode(response)
}
