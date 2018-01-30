package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	chart "github.com/wcharczuk/go-chart"
)

type kyaliaServer struct {
	listener net.Listener
	db       *sql.DB
}

type VcapServices struct {
	Pmysql []ServiceInstances `json:"p-mysql"`
}

type ServiceInstances struct {
	Credentials Credentials `json:"credentials"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
}

type Row struct {
	Animal string
	Votes  int
}

func main() {
	server := NewKyaliaServer()
	port := os.Getenv("PORT")
	portNumber, err := strconv.Atoi(port)
	FreakOut(err)

	connBytes := os.Getenv("VCAP_SERVICES")

	myServices := &VcapServices{}
	err = json.Unmarshal([]byte(connBytes), myServices)
	FreakOut(err)
	creds := myServices.Pmysql[0].Credentials

	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", creds.Username, creds.Password, creds.Hostname, creds.Port, creds.Name)

	server.db, err = sql.Open("mysql", connString)
	FreakOut(err)
	defer server.db.Close()
	err = server.db.Ping()
	FreakOut(err)

	server.Start(portNumber)
	defer server.Stop()
}

func NewKyaliaServer() *kyaliaServer {
	return &kyaliaServer{}
}

func (s *kyaliaServer) Start(port int) {
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if e != nil {
		log.Fatal("listen error:", e)
	}

	http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.Contains(path, "/get/results") {
			w.Header().Set("refresh", "1")
			w.WriteHeader(200)
			rows, err := s.getAllRows()
			FreakOut(err)

			values := []chart.Value{}
			for _, v := range rows {
				values = append(values, chart.Value{Value: float64(v.Votes), Label: v.Animal})
			}
			pie := chart.PieChart{
				Width:  1024,
				Height: 1024,
				Values: values,
			}

			w.Header().Set("Content-Type", "image/png")
			err = pie.Render(chart.PNG, w)
			if strings.Contains(err.Error(), "must contain at least") {
				w.Header().Set("Content-Type", "text")
				w.WriteHeader(503)
				w.Write([]byte("Graph unavailble. Try writing data to the database."))
				return
			}
			FreakOut(err)
		}
	}))
}

func (s *kyaliaServer) Stop() {
	s.listener.Close()
}

func FreakOut(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *kyaliaServer) getAllRows() ([]Row, error) {
	rows, err := s.db.Query("select * from pets")
	if err != nil {
		return []Row{}, err
	}

	defer rows.Close()

	animals := []Row{}

	for rows.Next() {

		row := Row{}
		err = rows.Scan(&row.Animal, &row.Votes)
		if err != nil {
			return []Row{}, err
		}
		animals = append(animals, row)
	}

	return animals, nil
}
