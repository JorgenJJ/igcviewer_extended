package main

import (
	"encoding/json"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/gorilla/mux"
	"github.com/marni/goigc"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Metadata - Struct for storing info about the API
type Metadata struct {
	Uptime string `json:"uptime,omitempty"`
	Desc string `json:"desc,omitempty"`
	Version string `json:"version,omitempty"`
}

// Track - Struct for storing basic info about a track
type Track struct {
	ID int `json:"id" bson:"id,omitempty"`
	URL string `json:"url" bson:"url,omitempty"`
}

// TrackInfo - Struct for storing detailed info about a track
type TrackInfo struct {
	FDate time.Time `json:"fdate,omitempty"`
	Pilot string `json:"pilot,omitempty"`
	Glider string `json:"glider,omitempty"`
	GliderID string `json:"glider_id,omitempty"`
	TrackLength int `json:"track_length,omitempty"`
}

// IDList - Struct for storing IDs
type IDList struct {
	ID int `json:"id,omitempty"`
}

type DB struct {
	Database	*mgo.Database
}
/*
const (
	MongoDBHosts = "paragliding-cluster-koft4.mongodb.net"
	AuthDatabase = "test"
	AuthUserName = "dbAdmin"
	AuthPassword = "WtpkGi1oSjfTcu4G"
)
*/
const (
	MongoDBHosts = "localhost:27017"
	AuthDatabase = "test"
	AuthUserName = ""
	AuthPassword = ""
)

var _init_ctx sync.Once
var _instance *DB

var idlist []IDList
var tracks []Track
var lastTrack = 0


var srvUrl = "mongodb+srv://dbAdmin:WtpkGi1oSjfTcu4G@paragliding-cluster-koft4.mongodb.net/test?retryWrites=true"
var stndrUrl = "mongodb://dbAdmin:WtpkGi1oSjfTcu4G@paragliding-cluster-shard-00-00-koft4.mongodb.net:27017,paragliding-cluster-shard-00-01-koft4.mongodb.net:27017,paragliding-cluster-shard-00-02-koft4.mongodb.net:27017/test/ssl=true&replicaSet=paragliding-cluster-shard-0&authSource=admin&retryWrites=true"

func main() {
	router := mux.NewRouter()
	port := os.Getenv("PORT")
/*
	db, err := mongo.NewClient("mongodb+srv://dbAdmin:WtpkGi1oSjfTcu4G@paragliding-cluster-koft4.mongodb.net/test?retryWrites=true")
	if err != nil { log.Fatal(err) }
	collection := db.Database("baz").Collection("qux")
	res, err := collection.InsertOne(context.Background(), map[string]string{"hello": "world"})
	if err != nil { log.Fatal(err) }
	id := res.InsertedID
	log.Println(id)
*/

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:		[]string{MongoDBHosts},
		Timeout:	600 * time.Second,
		Database:	AuthDatabase,
		Username:	AuthUserName,
		Password:	AuthPassword,
	}

	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		log.Fatal(err)
	}
	session.SetMode(mgo.Monotonic, true)

	router.HandleFunc("/paragliding/api", getMetadata).Methods("GET")
	router.HandleFunc("/paragliding/api/track", registerTrack).Methods("POST")
	router.HandleFunc("/paragliding/api/track", getIDs).Methods("GET")
	router.HandleFunc("/paragliding/api/track/{id}", getTrackMeta).Methods("GET")
	router.HandleFunc("/paragliding/api/track/{id}/{field}", getTrackMetaField).Methods("GET")
	router.HandleFunc("/paragliding", redirect).Methods("GET")
	router.HandleFunc("/paragliding/api/ticker/latest", getLatest).Methods("GET")
	router.HandleFunc("/paragliding/api/ticker", getTicker).Methods("GET")
	router.HandleFunc("/paragliding/api/ticker/{timestamp}", getTimestamped).Methods("GET")

	http.ListenAndServe(":"+port, router)
}
/*
func New() *mgo.Database {
	_init_ctx.Do(func() {
		_instance = new(DB)

		mongoDBDialInfo := &mgo.DialInfo{
			Addrs:		[]string{MongoDBHosts},
			Timeout:	600 * time.Second,
			Database:	AuthDatabase,
			Username:	AuthUserName,
			Password:	AuthPassword,
		}

		session, err := mgo.DialWithInfo(mongoDBDialInfo)

		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		_instance.Database = session.DB(AuthDatabase)
	})
	return _instance.Database
}*/

func getMetadata(w http.ResponseWriter, r *http.Request) {
	metadata := Metadata{"Yes", "Service for Paragliding tracks", "v1"}
	json.NewEncoder(w).Encode(metadata)
}

	// Reads a URL as a parameter, makes a new track for it in memory, and writes out the new id in json format
func registerTrack(w http.ResponseWriter, r *http.Request) {
	url, err := r.URL.Query()["url"]
	if !err || len(url[0]) < 1 {
		log.Println("URL parameter is missing")
	} else {	// If a URL is sent
		var track Track
		var id IDList
		_ = json.NewDecoder(r.Body).Decode(&track)
		track.URL = string(url[0])
		lastTrack++
		track.ID = lastTrack
		id.ID = lastTrack
		tracks = append(tracks, track)
		idlist = append(idlist, id)
		jsonConverter := fmt.Sprintf(`"{"id":%d}"`, track.ID)
		output := []byte(jsonConverter)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(output)

	}
}

	// Writes all the registered IDs
func getIDs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(idlist)
}

	// Writes information about a specific track registered in the memory
func getTrackMeta(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()
	_, input := path.Split(url)

	in, err := strconv.Atoi(input)
	if err != nil {
		log.Fatal(err)
	}

	if in <= lastTrack {	// If the ID exists in memory
		t, e := igc.ParseLocation(tracks[in - 1].URL)
		if e != nil {
			log.Fatal(e)
		}

		info := TrackInfo{t.Date, t.Pilot, t.GliderType, t.GliderID, 9}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(info)

	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

	// Writes a specific piece of information about a specific track
func getTrackMetaField(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()
	temp := strings.Split(url, "/")
	f := temp[5]
	t := temp[4]

	in, err := strconv.Atoi(t)
	if err != nil {
		log.Fatal(err)
	}
	if in <= lastTrack {	// If the ID exists in memory

		t, e := igc.ParseLocation(tracks[in - 1].URL)
		if e != nil {
			log.Fatal(e)
		}

		info := TrackInfo{t.Date, t.Pilot, t.GliderType, t.GliderID, 9}

		switch f {
		case "pilot":
			fmt.Fprintln(w, info.Pilot)
		case "glider":
			fmt.Fprintln(w, info.Glider)
		case "glider_id":
			fmt.Fprintln(w, info.GliderID)
		case "track_length":
			fmt.Fprintln(w, info.TrackLength)
		case "H_date":
			fmt.Fprintln(w, info.FDate)
		default:
			w.WriteHeader(http.StatusNotFound)

		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

}

func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/paragliding/api", http.StatusSeeOther)
}

func getLatest(w http.ResponseWriter, r *http.Request) {

}

func getTicker(w http.ResponseWriter, r *http.Request) {

}

func getTimestamped(w http.ResponseWriter, r *http.Request) {

}