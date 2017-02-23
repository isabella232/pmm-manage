package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var c confConfig

func main() {
	parseFlag()

	router := mux.NewRouter().PathPrefix(c.PathPrefix).Subrouter()
	router.HandleFunc("/v1/sshkey", getSSHKeyHandler).Methods("GET")
	router.HandleFunc("/v1/sshkey", setSSHKeyHandler).Methods("POST")

	router.HandleFunc("/v1/check-update", runCheckUpdateHandler).Methods("GET")
	router.HandleFunc("/v1/updates", getUpdateListHandler).Methods("GET")
	router.HandleFunc("/v1/updates", runUpdateHandler).Methods("POST")
	router.HandleFunc("/v1/updates/{timestamp}", getUpdateHandler).Methods("GET")
	router.HandleFunc("/v1/updates/{timestamp}", deleteUpdateHandler).Methods("DELETE")

	router.HandleFunc("/v1/users", getUserListHandler).Methods("GET")
	router.HandleFunc("/v1/users", createUserHandler).Methods("POST")
	router.HandleFunc("/v1/users/{username}", getUserHandler).Methods("GET")
	router.HandleFunc("/v1/users/{username}", deleteUserHandler).Methods("DELETE")

	// TODO: create separate handler with old password verification
	router.HandleFunc("/v1/users/{username}", createUserHandler).Methods("PATCH")

	log.Printf("PMM Configurator is started on %s address", c.ListenAddress)
	log.Fatal(http.ListenAndServe(c.ListenAddress, router))
}

func parseFlag() {
	flag.StringVar(
		&c.HtpasswdPath,
		"htpasswd-path",
		"/srv/nginx/.htpasswd",
		"htpasswd file location",
	)
	flag.StringVar(
		&c.ListenAddress,
		"listen-address",
		"127.0.0.1:7777",
		"Address and port to listen on: [ip_address]:port",
	)
	flag.StringVar(
		&c.PathPrefix,
		"url-prefix",
		"/configurator",
		"Prefix for the internal routes of web endpoints",
	)
	flag.StringVar(
		&c.SSHKeyPath,
		"ssh-key-path",
		"",
		"Path for SSH key",
	)
	flag.StringVar(
		&c.SSHKeyOwner,
		"ssh-key-owner",
		"admin",
		"Owner of SSH key",
	)
	flag.StringVar(
		&c.GrafanaDBPath,
		"grafana-db-path",
		"/srv/grafana/grafana.db",
		"grafana database location",
	)
	flag.StringVar(
		&c.PrometheusConfPath,
		"prometheus-conf-path",
		"/etc/prometheus.yml",
		"prometheus configuration file location",
	)
	flag.StringVar(
		&c.UpdateDirPath,
		"update-dir-path",
		"/srv/update",
		"update directory location",
	)
	flag.Parse()

	runSSHKeyChecks()
}

func returnSuccess(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(jsonResponce{
		Code:   http.StatusOK,
		Status: http.StatusText(http.StatusOK),
	})
}

func returnError(w http.ResponseWriter, req *http.Request, httpStatus int, title string, err error) {
	responce := jsonResponce{
		Code:   httpStatus,
		Status: http.StatusText(httpStatus),
		Title:  title,
	}
	if err != nil {
		responce.Detail = err.Error()
	}

	responceJSON, _ := json.Marshal(responce)
	log.Printf("%s %s: %s", req.Method, req.URL.String(), responceJSON)

	http.Error(w, string(responceJSON)+"\n", httpStatus)
}
