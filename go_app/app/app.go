package main

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type DataRow map[string]string

type AppConfig struct {
	pg struct {
		host     string
		port     string
		user     string
		password string
		dbname   string
	}
	http struct {
		httpport string
	}
}

var conf AppConfig

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func initConfig() {
	conf.pg.host = "localhost"
	conf.pg.port = "5432"
	conf.pg.user = "postgres"
	conf.pg.password = "posgtres"
	conf.pg.dbname = "appdb"
	conf.http.httpport = ":8080"

	PG_HOST, exists := os.LookupEnv("PG_HOST")
	if exists {
		conf.pg.host = PG_HOST
	}

	PG_PORT, exists := os.LookupEnv("PG_PORT")
	if exists {
		conf.pg.port = PG_PORT
	}
	PG_USER, exists := os.LookupEnv("PG_USER")
	if exists {
		conf.pg.user = PG_USER
	}
	PG_PASS, exists := os.LookupEnv("PG_PASS")
	if exists {
		conf.pg.password = PG_PASS
	}
	PG_DBNAME, exists := os.LookupEnv("PG_DBNAME")
	if exists {
		conf.pg.dbname = PG_DBNAME
	}

	HTTP_PORT, exists := os.LookupEnv("HTTP_PORT")
	if exists {
		conf.http.httpport = HTTP_PORT
	}

}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func getStringAndHash() []string {
	Row := make([]string, 0)
	JustString := String(8)
	Row = append(Row, JustString)
	Row = append(Row, GetMD5Hash(JustString))
	return Row
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func ParseTempl(tableData []DataRow) string {

	buffer := new(bytes.Buffer)

	templateParse, err := template.ParseFiles("./index.html")

	if err != nil {
		log.Fatal("Error: while parsin html: %v", err)
	}
	err = templateParse.ExecuteTemplate(buffer, "index", map[string]interface{}{"items": tableData})

	if err != nil {
		log.Fatal("Error: while parsin html: %v", err)
	}

	return buffer.String()
}

func main() {

	initConfig()

	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.pg.host, conf.pg.port, conf.pg.user, conf.pg.password, conf.pg.dbname)

	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	defer db.Close()

	err = db.Ping()
	CheckError(err)

	fmt.Println("DB Connected.")
	insertDynStmt := `insert into "apptable"("name", "hash") values($1, $2)`

	for i := 0; i < 10; i++ {
		Data := getStringAndHash()
		_, e := db.Exec(insertDynStmt, Data[0], Data[1])
		CheckError(e)
	}
	rows, err := db.Query(`SELECT "uuid", "ts", "name", "hash" FROM "apptable"`)
	CheckError(err)

	TableRows := []DataRow{}

	defer rows.Close()
	for rows.Next() {
		var uuid string
		var ts string
		var name string
		var hash string

		err = rows.Scan(&uuid, &ts, &name, &hash)
		CheckError(err)

		TableRow := DataRow{
			"uuid": uuid,
			"ts":   ts,
			"name": name,
			"hash": hash,
		}
		TableRows = append(TableRows, TableRow)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Fprintf(w, strings.Join(getStringAndHash(), ","))
		fmt.Fprintf(w, ParseTempl(TableRows))
	})

	fmt.Printf("Server running (port=8080), route: http://localhost:8080/\n")
	if err := http.ListenAndServe(conf.http.httpport, nil); err != nil {
		log.Fatal(err)
	}
}
