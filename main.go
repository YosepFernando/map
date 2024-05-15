package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)



type hospital struct{
    id int
    Nama string
    Latitude float64
    Longitude float64
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var store = sessions.NewCookieStore([]byte("secret"))

const dbConnectionString = "2105551091:2105551091@tcp(gis_2105551091.local.net:3306)/db_2105551091"


func signin(w http.ResponseWriter, r *http.Request){
    var tmplt = template.Must(template.ParseFiles("templates/login.html"))
    tmplt.Execute(w, r)
}

func add(w http.ResponseWriter, r *http.Request){
    var tmplt = template.Must(template.ParseFiles("templates/add.html"))
    tmplt.Execute(w, r)
}
func updatePage(w http.ResponseWriter, r *http.Request){
    parts := strings.Split(r.URL.Path, "/")
    nama := parts[len(parts)-1]

    // Mendapatkan data rumah sakit dari database berdasarkan nama
    db, err := sql.Open("mysql", dbConnectionString)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    var hospital hospital
    err = db.QueryRow("SELECT id, nama_rumah_sakit, latitude, longitude FROM rumah_sakit WHERE nama_rumah_sakit = ?", nama).
        Scan(&hospital.id, &hospital.Nama, &hospital.Latitude, &hospital.Longitude)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Mengirim data rumah sakit ke template update.html
    tmpl := template.Must(template.ParseFiles("templates/update.html"))
    tmpl.Execute(w, hospital)
}

func home(w http.ResponseWriter, r *http.Request){
    session, _ := store.Get(r, "session-name")

	username, ok := session.Values["username"].(string)
	if !ok || username == "" {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

    var tmplt = template.Must(template.ParseFiles("templates/index.html"))
    db, err := sql.Open("mysql", dbConnectionString)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    rs, err := db.Query("Select id, nama_rumah_sakit, latitude, longitude FROM rumah_sakit")
    if err != nil {
        log.Fatal(err)
    }
    defer rs.Close()

    var hospitals []hospital

    for rs.Next() {
        var id int
        var nama_rumah_sakit string
        var latitude, longitude float64
        err := rs.Scan(&id, &nama_rumah_sakit, &latitude, &longitude)
        if err != nil {
            log.Fatal(err)
        }
        hospital := hospital{id: id, Nama: nama_rumah_sakit, Latitude: latitude, Longitude: longitude}
        hospitals = append(hospitals, hospital)
        // fmt.Printf("%f", longitude)
    }

    data := struct {
		Hospitals []hospital
		Username  string
	}{
		Hospitals: hospitals,
		Username:  username,
	}

    tmplt.Execute(w, data)
}


func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("mysql", dbConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, hashedPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", dbConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var storedHashedPassword string
	err = db.QueryRow("SELECT password FROM users WHERE username = ?", user.Username).Scan(&storedHashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(user.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	session, _ := store.Get(r, "session-name")
	session.Values["username"] = user.Username
	session.Save(r, w)

	w.WriteHeader(http.StatusOK)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	delete(session.Values, "username")
	session.Save(r, w)
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

func addmark(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parsing form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Mengambil data dari form
	nama := r.FormValue("username")
	latitude := r.FormValue("latitude")
	longitude := r.FormValue("longitude")

	// Konversi latitude dan longitude ke float64
	lat, err := strconv.ParseFloat(latitude, 64)
	if err != nil {
		return
	}
	long, err := strconv.ParseFloat(longitude, 64)
	if err != nil {
		return
	}

	// Membuka koneksi ke database
	db, err := sql.Open("mysql", dbConnectionString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Menyimpan data ke database
	_, err = db.Exec("INSERT INTO rumah_sakit (nama_rumah_sakit, latitude, longitude) VALUES (?, ?, ?)", nama, lat, long)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect ke halaman utama setelah data disimpan
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
    err := r.ParseForm()
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Mendapatkan data rumah sakit yang akan diupdate
    nama := r.FormValue("nama")
    latitude := r.FormValue("latitude")
    longitude := r.FormValue("longitude")

    // Konversi latitude dan longitude ke float64
    lat, err := strconv.ParseFloat(latitude, 64)
    if err != nil {
        http.Error(w, "Invalid latitude", http.StatusBadRequest)
        return
    }
    long, err := strconv.ParseFloat(longitude, 64)
    if err != nil {
        http.Error(w, "Invalid longitude", http.StatusBadRequest)
        return
    }

    // Update data rumah sakit di database
    db, err := sql.Open("mysql", dbConnectionString)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    _, err = db.Exec("UPDATE rumah_sakit SET latitude=?, longitude=? WHERE nama_rumah_sakit=?", lat, long, nama)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Redirect ke halaman utama setelah data diupdate
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
    nama := strings.TrimPrefix(r.URL.Path, "/delete/")
    nama = strings.TrimSpace(nama)
    
    // Membuka koneksi ke database
    db, err := sql.Open("mysql", dbConnectionString)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()
    
    // Menghapus data di database
    _, err = db.Exec("DELETE FROM rumah_sakit WHERE nama_rumah_sakit = ?", nama)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Mengirim respons sukses
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Marker deleted successfully"))
}


func main()  {
    http.HandleFunc("/", home)
    http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/signin", signin)
    http.HandleFunc("/logout", logoutHandler)
    http.HandleFunc("/add", add)
    http.HandleFunc("/addmark", addmark)
    http.HandleFunc("/updatepage/{nama}", updatePage)
    http.HandleFunc("/update/", updateHandler)
    http.HandleFunc("/delete/", deleteHandler)

    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
        db, err := sql.Open("mysql", dbConnectionString)
        if err != nil {
            log.Fatal(err)
        }
        defer db.Close()
    
        rs, err := db.Query("Select id, nama_rumah_sakit, latitude, longitude FROM rumah_sakit")
        if err != nil {
            log.Fatal(err)
        }
        defer rs.Close()
    
        var hospitals []hospital
    
        for rs.Next() {
            var id int
            var nama_rumah_sakit string
            var latitude, longitude float64
            err := rs.Scan(&id, &nama_rumah_sakit, &latitude, &longitude)
            if err != nil {
                log.Fatal(err)
            }
            hospital := hospital{id: id, Nama: nama_rumah_sakit, Latitude: latitude, Longitude: longitude}
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
    log.Println("Server started on localhost:9000")
    http.ListenAndServe(serverport, nil)

}

