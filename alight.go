package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oschwald/geoip2-golang"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var db *sql.DB
var visitorsStmt *sql.Stmt
var visitStmt *sql.Stmt

type Visit struct {
	timse    string
	location string
	ip       string
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" || r.Method == "" {
		get(w)
	} else if r.Method == "POST" {
		post(w, r)
	}
}

func get(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	rows, err := db.Query("select count(id), strftime(\"%Y-%m-%d %H:00:00\", datetime(time, 'localtime')) from visits where time > datetime('now', '-500 hours') group by strftime(\"%Y%j%H\", time);")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	result := map[string][]map[string]string{}
	counts := []map[string]string{}
	for rows.Next() {
		var count string
		var time string

		rows.Scan(&count, &time)
		counts = append(counts, map[string]string{
			"time":  time,
			"count": count,
		})
	}
	result["counts"] = counts

	lrows, err := db.Query("select count(city), city, country, iso from visitors group by city, iso;")
	if err != nil {
		log.Fatal(err)
	}
	defer lrows.Close()
	locations := []map[string]string{}
	for lrows.Next() {
		var count string
		var city string
		var country string
		var iso string

		lrows.Scan(&count, &city, &country, &iso)
		locations = append(locations, map[string]string{
			"city":    city,
			"country": country,
			"iso":     iso,
			"count":   count,
		})
	}
	result["locations"] = locations

	b, _ := json.Marshal(result)
	fmt.Fprintf(w, string(b))

	rows.Close()
}

func post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if strings.Contains(r.UserAgent(), "Googlebot") {
		return
	}

	if r.FormValue("action") == "enter" {
		var id int64
		avid := r.FormValue("avid")

		if avid == "" {
			host, _, _ := net.SplitHostPort(r.RemoteAddr)
			if host != "" {
				gr := geo(host)
				result, err := visitorsStmt.Exec(gr["city"], gr["country"], gr["iso"], host, r.UserAgent())

				if err != nil {
					log.Print(err)
					return
				}

				id, _ = result.LastInsertId()
				response := map[string]string{}
				response["vid"] = strconv.FormatInt(id, 10)

				rj, _ := json.Marshal(response)
				fmt.Fprintf(w, string(rj))
			}
		} else {
			id_s, _ := strconv.Atoi(avid)
			id = int64(id_s)
		}

		_, err := visitStmt.Exec(r.FormValue("url"), r.FormValue("referrer"), id)

		if err != nil {
			log.Print(err)
			return
		}
	}
}

func geo(ipstring string) map[string]string {
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ip := net.ParseIP(ipstring)
	if ip != nil {
		record, err := db.City(ip)
		if err != nil {
			log.Fatal(err)
		}
		return map[string]string{
			"city":    record.City.Names["en"],
			"country": record.Country.Names["en"],
			"iso":     record.Country.IsoCode,
		}
	}

	return map[string]string{
		"city":    "",
		"country": "",
		"iso":     "",
	}
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
		create table visits (id integer primary key, url text, time integer, referrer text, vid integer references visitors);
		create table visitors (vid integer primary key, city text, country text, iso text, ip text, ua text);
		`

		_, err = db.Exec(sqlStmt)
		if err != nil {
			os.Remove("./alight.db")
			log.Printf("%q: %s\n", err, sqlStmt)
			return
		}
	}

	db.Exec("pragma synchronous = OFF")

	visitorsStmt, err = db.Prepare("insert into visitors values (null, ?, ?, ?, ?, ?)")
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
