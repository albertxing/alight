package main

import (
	"strings"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oschwald/geoip2-golang"
	"log"
	"net/http"
	"net"
	"os"
	"strconv"
)

var db *sql.DB
var visitorsStmt *sql.Stmt
var visitStmt *sql.Stmt

type Visit struct {
	timse string
	location string
	ip string
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" || r.Method == "" {
		getCount(w)
	} else if r.Method == "POST" {
		post(w, r)
	}
}

func get(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	rows, err := db.Query("select datetime(time, 'localtime'), location, ip from visits, visitors where visits.visitor = visitors.id;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	visits := []Visit{}

	// fmt.Fprint(w, "[")
	// first := true
	for rows.Next() {
		// if first {
		// 	first = false
		// } else {
		// 	fmt.Fprint(w, ",")
		// }

		var time string
		var location string
		var ip string

		rows.Scan(&time, &location, &ip)
		v := Visit{time, location, ip}
		visits = append(visits, v)

		// fmt.Fprintf(w, "{\"time\":\"%s\",\"location\":\"%s\",\"ip\":\"%s\"}", time, location, ip)
	}
	// fmt.Fprint(w, "]")
	b, _ := json.Marshal(visits)
	fmt.Fprint(w, b)

	rows.Close()
}

func getCount(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	rows, err := db.Query("select count(id), strftime(\"%Y-%m-%d %H:00:00\", datetime(time, 'localtime')) from visits where time > datetime('now', '-1000 hours') group by strftime(\"%Y%j%H\", time);")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// fmt.Fprint(w, "[")
	// first := true
	counts := []map[string]string{}
	for rows.Next() {
		// if first {
		// 	first = false
		// } else {
		// 	fmt.Fprint(w, ",")
		// }

		var count string
		var time string

		rows.Scan(&count, &time)
		counts = append(counts, map[string]string{
			"time": time,
			"count": count,
			})
		// fmt.Fprintf(w, "{\"time\":\"%s UTC\",\"count\":%s}", time, count)
	}
	b, _ := json.Marshal(counts)
	fmt.Fprintf(w, string(b))
	// fmt.Fprint(w, "]")

	rows.Close()
}

func post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.FormValue("action") == "enter" {
		var id int64
		cookie, err := r.Cookie("avid")

		if err != nil {
			ip := strings.Split(r.RemoteAddr, ":")[0]
			result, err := visitorsStmt.Exec(geo(ip), ip)

			if err != nil {
				log.Print(err)
				return
			}

			id, _ = result.LastInsertId()
			//fmt.Println(id)
		} else {
			id_s, _ := strconv.Atoi(cookie.Value)
			id = int64(id_s)
		}

		_, err = visitStmt.Exec(r.FormValue("url"), r.FormValue("referrer"), id)

		if err != nil {
			log.Print(err)
			return
		}
	}
}

func geo(ipstring string) string {
    db, err := geoip2.Open("GeoLite2-City.mmdb")
    if err != nil {
            log.Fatal(err)
    }
    defer db.Close()
    // If you are using strings that may be invalid, check that ip is not nil
    ip := net.ParseIP(ipstring)
    record, err := db.City(ip)
    if err != nil {
            log.Fatal(err)
    }

	return record.City.Names["en"] + " " + record.Country.IsoCode
}

func main() {
	isNew := false

	_, err := os.Open("./alight.db")
	if err != nil {
		isNew = true
	}

	db, err = sql.Open("sqlite3", "./alight.db")
	defer db.Close()

	if isNew {
		sqlStmt := `
		create table visits (id integer primary key, url text, time integer, referrer text, visitor integer);
		create table visitors (id integer primary key, location text, ip text);
		`

		_, err = db.Exec(sqlStmt)
		if err != nil {
			os.Remove("./alight.db")
			log.Printf("%q: %s\n", err, sqlStmt)
			return
		}
	}

	db.Exec("pragma synchronous = OFF")

	visitorsStmt, err = db.Prepare("insert into visitors values (null, '?', ?)")
	if err != nil {
		log.Fatal(err)
	}
	visitStmt, err = db.Prepare("insert into visits values (null, ?, datetime('now'), ?, ?);")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8000", nil)
}
