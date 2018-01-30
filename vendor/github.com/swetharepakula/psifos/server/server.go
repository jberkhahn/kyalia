package server

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
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

func (s *PsifosServer) Start(port int) {
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if e != nil {
		log.Fatal("listen error:", e)
	}

	http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		var err error

		if strings.Contains(path, "/put/cats") {
			_, err = s.Db.Exec("update pets set votes = votes + 1 where animal = ?", "cats")

		} else if strings.Contains(path, "/put/dogs") {
			_, err = s.Db.Exec("update pets set votes = votes + 1 where animal = ?", "dogs")

		} else if strings.Contains(path, "setup/database") {
			_, err = s.Db.Exec("truncate table pets")
			if err != nil {
				driverErr, ok := err.(*mysql.MySQLError)
				if !ok || driverErr.Number != 1146 {
					// 1146 is error code for Table doesn't exist
					RespondWithError(w, err)
					return
				}
			}

			_, err = s.Db.Exec("CREATE TABLE IF NOT EXISTS pets ( animal varchar(32), votes integer )")
			if err != nil {
				RespondWithError(w, err)
				return
			}

			_, err = s.Db.Exec("insert into pets (animal, votes) values (?, ?)", "cats", 0)
			if err != nil {
				RespondWithError(w, err)
				return
			}

			_, err = s.Db.Exec("insert into pets (animal, votes) values (?, ?)", "dogs", 0)
			if err != nil {
				RespondWithError(w, err)
				return
			}
		}

		if err != nil {
			RespondWithError(w, err)
		} else {
			rows, err := GetAllRows(s.Db)
			if err != nil {
				RespondWithError(w, err)
				return
			}
			w.Header().Set("Content-type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(fmt.Sprintf("%v", rows)))
		}
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
