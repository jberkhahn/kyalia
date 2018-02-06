package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/icza/gowut/gwu"
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

	connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", creds.Username, creds.Password, creds.Hostname, creds.Port, creds.Name)

	s.Db, err = sql.Open("mysql", connString)
	server.FreakOut(err)
	defer s.Db.Close()
	err = s.Db.Ping()
	server.FreakOut(err)

	win := gwu.NewWindow("main", "psifos")
	win.Style().SetFullWidth()
	win.SetHAlign(gwu.HACenter)
	win.SetCellPadding(2)

	win.Add(gwu.NewLabel("Vote Below by Clicking on Your Favorite"))
	btnsPanel := gwu.NewNaturalPanel()
	btn1 := gwu.NewButton("Dogs")
	btn2 := gwu.NewButton("Cats")

	btn1.AddEHandler(server.NewButtonHandler(s, "categoryOne"), gwu.ETypeClick)
	btn2.AddEHandler(server.NewButtonHandler(s, "categoryTwo"), gwu.ETypeClick)

	btnsPanel.Add(btn1)
	btnsPanel.Add(btn2)

	win.Add(btnsPanel)

	setupBtn := gwu.NewButton("Setup Database")
	setupBtn.AddEHandler(server.NewButtonHandler(s, "setup"), gwu.ETypeClick)
	win.Add(setupBtn)

	serv := gwu.NewServer("psifos", fmt.Sprintf("0.0.0.0:%d", portNumber))
	serv.AddWin(win)
	serv.SetText("Psifos")
	serv.Start()
}

// func main() {
// 	s := server.NewPsifosServer()
// 	port := os.Getenv("PORT")
// 	portNumber, err := strconv.Atoi(port)
// 	server.FreakOut(err)

// 	connBytes := os.Getenv("VCAP_SERVICES")

// 	myServices := &server.VcapServices{}
// 	server.FreakOut(err)
// 	err = json.Unmarshal([]byte(connBytes), myServices)
// 	server.FreakOut(err)
// 	creds := myServices.Pmysql[0].Credentials

// 	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", creds.Username, creds.Password, creds.Hostname, creds.Port, creds.Name)

// 	s.Db, err = sql.Open("mysql", connString)
// 	server.FreakOut(err)
// 	defer s.Db.Close()
// 	err = s.Db.Ping()
// 	server.FreakOut(err)

// 	s.Start(portNumber)
// 	defer s.Stop()
// }
