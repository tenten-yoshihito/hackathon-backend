package controller

import (
	"encoding/json"
	"log"
	"net/http"
)

// ---------------------------------------------------------
//
//	共通ヘルパー関数
//
// ---------------------------------------------------------

// respondJSON : JSONレスポンスを返す共通関数(ここでは任意のinterface{}を容認)
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("fail: response encoding, %v\n", err)
	}
}

// respondError : エラーレスポンスを返す共通関数
func respondError(w http.ResponseWriter, status int, message string, err error) {
	log.Printf("error: %s: %v", message, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
