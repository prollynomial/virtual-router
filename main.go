package main

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"os"
)

var ipt, _ = iptables.New()

/**
 * Responsibilities include:
 *  - port forwarding
 *  - static NAT (which is just port forwarding but a bunch of times)
 *  - subnet access control
 *  - front everything with APIs
 */
func main() {
	if os.Geteuid() != 0 {
		panic("Not running as root!")
	}

	router := mux.NewRouter()

	// Port forwarding rules
	router.HandleFunc("/pfrs", createPortForwardingRule).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}

type Pfr struct {
	FromIp        net.IP `json:"from"`
	ToIp          net.IP `json:"to"`
	FromPortStart int    `json:"fromPortStart"`
	FromPortEnd   int    `json:"fromPortEnd"`
	ToPortStart   int    `json:"toPortStart"`
	ToPortEnd     int    `json:"toPortEnd"`
}

type ApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func createPortForwardingRule(res http.ResponseWriter, req *http.Request) {
	var pfr Pfr
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&pfr)
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()

	destination := fmt.Sprintf("%s:%d-%d", pfr.ToIp.String(), pfr.ToPortStart, pfr.ToPortEnd)
	sourcePorts := fmt.Sprintf("%d:%d", pfr.FromPortStart, pfr.FromPortEnd)
	err = ipt.AppendUnique("nat", "PREROUTING", "--protocol", "tcp", "--dst", pfr.FromIp.String(), "--dport", sourcePorts, "-j", "DNAT", "--to-destination", destination)
	if err != nil {
		failWithMessage(res, "PREROUTING rule creation failed")
		return
	}

	// TODO this only has to be in nat/POSTROUTING once - why do it here?
	err = ipt.AppendUnique("nat", "POSTROUTING", "-j", "MASQUERADE")
	if err != nil {
		failWithMessage(res, "POSTROUTING rule creation failed")
		return
	}

	// TODO include a reference so that consumers can delete PFRs as well
	err = writeJsonResponse(res, http.StatusOK, ApiResponse{Success: true})
	if err != nil {
		panic("Couldn't write successful response")
	}
}

func failWithMessage(w http.ResponseWriter, message string) error {
	return writeJsonResponse(w, http.StatusInternalServerError, ApiResponse{
		Success: false,
		Message: message,
	})
}

func writeJsonResponse(w http.ResponseWriter, status int, res ApiResponse) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	return encoder.Encode(res)
}
