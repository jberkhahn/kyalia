package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/icza/gowut/gwu"
	"github.com/swetharepakula/psifos/button"
	"github.com/swetharepakula/psifos/server"
)

func main() {
	s := server.NewPsifosServer()

	var portNumber int
	var err error
	port := os.Getenv("PORT")
	if port != "" {
		portNumber, err = strconv.Atoi(port)
		server.FreakOut(err)
	} else {
		portNumber = 8080
	}

	s.Db, err = sql.Open("mysql", server.GetVcapServicesCreds())
	server.FreakOut(err)
	defer s.Db.Close()
	err = s.Db.Ping()
	server.FreakOut(err)

	win := gwu.NewWindow("main", "psifos")
	win.Style().SetFullWidth()
	win.SetHAlign(gwu.HACenter)
	win.SetCellPadding(2)

	label := gwu.NewLabel("Vote Below By Clicking on Your Favorite")
	label.Style().SetFontSize("50pt")
	win.Add(label)

	btn1 := button.CreateButton("Dogs", "categoryOne", s)
	btn2 := button.CreateButton("Cats", "categoryTwo", s)
	setupBtn := button.CreateButton("Setup Database", "setup", s)

	win.Add(btn1)
	win.Add(btn2)

	win.Add(setupBtn)

	serv := gwu.NewServer("", fmt.Sprintf("0.0.0.0:%d", portNumber))
	serv.AddWin(win)
	serv.SetText("Psifos")
	serv.Start()
}
