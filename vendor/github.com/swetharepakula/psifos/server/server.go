package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
)

func NewPsifosServer() *PsifosServer {
	return &PsifosServer{}
}

type PsifosServer struct {
	listener net.Listener
	Db       *sql.DB
}

type VcapServices struct {
	ClearDBVcapServices      []ServiceInstances             `json:"clearDB"`
	UserProvidedVcapServices []UserProvidedServiceInstances `json:"user-provided"`
}

type ServiceInstances struct {
	Credentials Credentials `json:"credentials"`
}

type UserProvidedServiceInstances struct {
	Credentials UPCredentials `json:"credentials"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Hostname string `json:"hostname"`
	Name     string `json:"name"`
	Port     string `json:"port"`
}

type UPCredentials struct {
	ConnString string `json:"connstring"`
}

type Row struct {
	Animal string
	Votes  int
}

func (s *VcapServices) GetCreds() (string, error) {

	if len(s.ClearDBVcapServices) > 0 {
		creds := s.ClearDBVcapServices[0].Credentials
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", creds.Username, creds.Password, creds.Hostname, creds.Port, creds.Name), nil
	}

	if len(s.UserProvidedVcapServices) > 0 {
		return s.UserProvidedVcapServices[0].Credentials.ConnString, nil
	}

	return "", errors.New("No Suitable Database")
}

func GetVcapServicesCreds() string {

	connBytes := os.Getenv("VCAP_SERVICES")

	myServices := &VcapServices{}
	err := json.Unmarshal([]byte(connBytes), myServices)
	FreakOut(err)

	connString, err := myServices.GetCreds()
	FreakOut(err)

	return connString
}

func (s *PsifosServer) Start(port int) {
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if e != nil {
		log.Fatal("listen error:", e)
	}

	http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		var err error

		if strings.Contains(path, "/put/cats") {
			err = s.PutCategoryOne()

		} else if strings.Contains(path, "/put/dogs") {
			err = s.PutCategoryTwo()

		} else if strings.Contains(path, "setup/database") {
			err = s.SetupDatabase()
		}

		if err != nil {
			RespondWithError(w, err)
			return
		}

		rows, err := GetAllRows(s.Db)
		if err != nil {
			RespondWithError(w, err)
			return
		}

		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf("%v", rows)))
	}))
}

func (s *PsifosServer) Stop() {
	s.listener.Close()
}

func FreakOut(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *PsifosServer) PutCategoryOne() error {
	_, err := s.Db.Exec("update pets set votes = votes + 1 where animal = ?", "dogs")
	return err
}

func (s *PsifosServer) PutCategoryTwo() error {
	_, err := s.Db.Exec("update pets set votes = votes + 1 where animal = ?", "cats")
	return err
}

func (s *PsifosServer) SetupDatabase() error {

	_, err := s.Db.Exec("truncate table pets")
	if err != nil {
		driverErr, ok := err.(*mysql.MySQLError)
		if !ok || driverErr.Number != 1146 {
			// 1146 is error code for Table doesn't exist
			return err
		}
	}
	_, err = s.Db.Exec("CREATE TABLE IF NOT EXISTS pets ( animal varchar(32), votes integer )")
	if err != nil {
		return err
	}

	_, err = s.Db.Exec("insert into pets (animal, votes) values (?, ?)", "cats", 0)
	if err != nil {
		return err
	}

	_, err = s.Db.Exec("insert into pets (animal, votes) values (?, ?)", "dogs", 0)
	if err != nil {
		return err
	}
	return nil
}

func RespondWithError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

func GetAllRows(db *sql.DB) ([]Row, error) {
	rows, err := db.Query("select * from pets")
	if err != nil {
		return []Row{}, err
	}

	defer rows.Close()

	pets := []Row{}

	for rows.Next() {

		row := Row{}
		err = rows.Scan(&row.Animal, &row.Votes)
		if err != nil {
			return []Row{}, err
		}
		pets = append(pets, row)
	}

	return pets, nil
}
