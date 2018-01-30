package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/swetharepakula/psifos/server"
)

func main() {
	s := server.NewPsifosServer()
	port := os.Getenv("PORT")
	portNumber, err := strconv.Atoi(port)
	server.FreakOut(err)

	connBytes := os.Getenv("VCAP_SERVICES")

	myServices := &server.VcapServices{}
	server.FreakOut(err)
	err = json.Unmarshal([]byte(connBytes), myServices)
	server.FreakOut(err)
	creds := myServices.Pmysql[0].Credentials

	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", creds.Username, creds.Password, creds.Hostname, creds.Port, creds.Name)

	s.Db, err = sql.Open("mysql", connString)
	server.FreakOut(err)
	defer s.Db.Close()
	err = s.Db.Ping()
	server.FreakOut(err)

	s.Start(portNumber)
	defer s.Stop()
}
