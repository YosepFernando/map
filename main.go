package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type hospital struct{
    Nama string
    Latitude float64
    Longitude float64
}

func home(w http.ResponseWriter, r *http.Request){
    var tmplt = template.Must(template.ParseFiles("index.html"))
    db, err := sql.Open("mysql", "2105551091:2105551091@tcp(gis_2105551091.local.net:3306)/db_2105551091")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    rs, err := db.Query("Select nama_rumah_sakit, latitude, longitude FROM rumah_sakit")
    if err != nil {
        log.Fatal(err)
    }
    defer rs.Close()

    var hospitals []hospital

    for rs.Next() {
        var nama_rumah_sakit string
        var latitude, longitude float64
        err := rs.Scan(&nama_rumah_sakit, &latitude, &longitude)
        if err != nil {
            log.Fatal(err)
        }
        hospital := hospital{Nama: nama_rumah_sakit, Latitude: latitude, Longitude: longitude}
        hospitals = append(hospitals, hospital)
        // fmt.Printf("%f", longitude)
    }

    tmplt.Execute(w, hospitals)
}


func main()  {
    http.HandleFunc("/", home)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
        db, err := sql.Open("mysql", "2105551091:2105551091@tcp(gis_2105551091.local.net:3306)/db_2105551091")
        if err != nil {
            log.Fatal(err)
        }
        defer db.Close()
    
        rs, err := db.Query("Select nama_rumah_sakit, latitude, longitude FROM rumah_sakit")
        if err != nil {
            log.Fatal(err)
        }
        defer rs.Close()
    
        var hospitals []hospital
    
        for rs.Next() {
            var nama_rumah_sakit string
            var latitude, longitude float64
            err := rs.Scan(&nama_rumah_sakit, &latitude, &longitude)
            if err != nil {
                log.Fatal(err)
            }
            hospital := hospital{Nama: nama_rumah_sakit, Latitude: latitude, Longitude: longitude}
            hospitals = append(hospitals, hospital)
            // fmt.Printf("%f", longitude)
        }
    
        jsonData, err := json.Marshal(hospitals)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    
        w.Header().Set("Content-Type", "application/json")
        w.Write(jsonData)
    })

    serverport := ":9000"
    http.ListenAndServe(serverport, nil)
}

