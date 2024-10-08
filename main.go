package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Location     []string
}

type Location struct {
	Locations []string `json:"locations"`
}

type Date struct {
	Dates []string `json:"dates"`
}

type Response struct {
	CreaDate1       string
	CreaDate2       string
	FirstAlbumDate1 string
	FirstAlbumDate2 string
	Location        string
	NMember         []string
	Searchword      string
}

type erreur struct {
	Titre string
	Text  string
}

var (
	urlArtist   = "https://groupietrackers.herokuapp.com/api/artists"
	urlLocation = "https://groupietrackers.herokuapp.com/api/locations"
	urlDate     = "https://groupietrackers.herokuapp.com/api/dates"
)

func main() {
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/info/", MoreInfoHandler)
	http.HandleFunc("/filters/", FiltersHandler)
	http.HandleFunc("/search/", SearchHandler)
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))
	// server
	fmt.Printf("Starting server at port 8080\n")
	fmt.Printf("http://localhost:8080/index\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// w: server to client / r: client to server
	if r.URL.Path != "/index" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	} else {
		artistData := ReadApi(urlArtist)
		var artistObject []Artist
		json.Unmarshal(artistData, &artistObject)
		locationData := ReadApi(urlLocation)
		locationData = locationData[9 : len(locationData)-2]
		var locationObject []Location
		json.Unmarshal(locationData, &locationObject)
		var locationsSlice []string
		for _, locations := range locationObject {
			locationsSlice = append(locationsSlice, locations.Locations...)
		}
		locationsSlice = removeDuplicateValues(locationsSlice)
		sort.Strings(locationsSlice)
		for i := range artistObject {
			artistObject[i].Location = append(artistObject[i].Location, locationObject[i].Locations...)
		}
		var DatesSlice []string
		for _, artist := range artistObject {
			DatesSlice = append(DatesSlice, strconv.Itoa(artist.CreationDate))
		}
		DatesSlice = removeDuplicateValues(DatesSlice)
		// Create the template
		tmpl, err := template.ParseFiles("./templates/index.html")
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		data := map[string]interface{}{
			"Artist":   artistObject,
			"Location": locationsSlice,
			"Date":     DatesSlice,
		}
		// execute the template with the structure
		err = tmpl.Execute(w, data)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}
}

func MoreInfoHandler(w http.ResponseWriter, r *http.Request) {
	index := strings.TrimPrefix(r.URL.Path, "/info")
	if r.URL.Path != "/info"+index {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	} else {
		// Get object of each structures
		artistData := ReadApi(urlArtist + index)
		var artistObject Artist
		json.Unmarshal(artistData, &artistObject)
		locationData := ReadApi(urlLocation + index)
		var locationObject Location
		json.Unmarshal(locationData, &locationObject)
		datesData := ReadApi(urlDate + index)
		var dateObject Date
		json.Unmarshal(datesData, &dateObject)
		// Create the template
		tmpl, err := template.ParseFiles("./templates/info.html")
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		data := map[string]interface{}{
			"Artist":   artistObject,
			"Location": locationObject,
			"Date":     dateObject,
		}
		// execute the template with the structure
		err = tmpl.Execute(w, data)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}
}

func FiltersHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	response := Response{
		CreaDate1:       r.FormValue("range1"),
		CreaDate2:       r.FormValue("range2"),
		FirstAlbumDate1: r.FormValue("fRange1"),
		FirstAlbumDate2: r.FormValue("fRange2"),
		Location:        r.FormValue("location"),
		NMember:         r.Form["nMembers"],
	}
	if r.URL.Path != "/filters/" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	} else {
		artistData := ReadApi(urlArtist)
		var artistObject []Artist
		json.Unmarshal(artistData, &artistObject)
		locationData := ReadApi(urlLocation)
		locationData = locationData[9 : len(locationData)-2]
		var locationObject []Location
		json.Unmarshal(locationData, &locationObject)
		var locationsSlice []string
		for _, locations := range locationObject {
			locationsSlice = append(locationsSlice, locations.Locations...)
		}
		locationsSlice = removeDuplicateValues(locationsSlice)
		sort.Strings(locationsSlice)
		for i := range locationObject {
			artistObject[i].Location = append(artistObject[i].Location, locationObject[i].Locations...)
		}
		NewArtistObject := Filters(artistObject, response)
		// Create the template
		tmpl, err := template.ParseFiles("./templates/filters.html")
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		data := map[string]interface{}{
			"Artist":   NewArtistObject,
			"Location": locationsSlice,
		}
		// execute the template with the structure
		err = tmpl.Execute(w, data)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}
}

func SearchHandler(w http.ResponseWriter, r *http.Request) { // handler of the page of search process/result
	searchword := Response{ // get the value of the balise searchword
		Searchword: r.FormValue("searchword"),
	}
	if r.URL.Path != "/search/" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	} else {
		artistData := ReadApi(urlArtist)
		var artistObject []Artist
		json.Unmarshal(artistData, &artistObject)
		locationData := ReadApi(urlLocation)
		locationData = locationData[9 : len(locationData)-2]
		var locationObject []Location
		json.Unmarshal(locationData, &locationObject)
		var locationsSlice []string
		for _, locations := range locationObject {
			locationsSlice = append(locationsSlice, locations.Locations...)
		}
		locationsSlice = removeDuplicateValues(locationsSlice)
		sort.Strings(locationsSlice)
		for i := range artistObject {
			artistObject[i].Location = append(artistObject[i].Location, locationObject[i].Locations...)
		}
		var DatesSlice []string
		for _, artist := range artistObject {
			DatesSlice = append(DatesSlice, strconv.Itoa(artist.CreationDate))
		}
		DatesSlice = removeDuplicateValues(DatesSlice)
		search := Search(artistObject, searchword.Searchword)
		tmpl, err := template.ParseFiles("./templates/search.html")
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		data := map[string]interface{}{
			"Artist":   search,
			"Artists":  artistObject,
			"Location": locationsSlice,
			"Date":     DatesSlice,
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}
}

func Search(artistData []Artist, searchword string) []Artist {
	var newArtistData []Artist
	for _, artist := range artistData { // loop to search the iteration and compare it with the searchword with every artist
		if strings.Contains(strings.ToLower(artist.Name), strings.ToLower(searchword)) { // if the searchword correspond with the artist name
			newArtistData = append(newArtistData, artist)
		} else if strings.Contains(strings.ToLower(artist.FirstAlbum), strings.ToLower(searchword)) { // if it correspond with the artist first album
			newArtistData = append(newArtistData, artist)
		} else if strings.Contains(strings.ToLower(strconv.Itoa(artist.CreationDate)), strings.ToLower(searchword)) { // if it correspond with the artist creation Date
			newArtistData = append(newArtistData, artist)
		} else {
			for _, member := range artist.Members { // loop to search the iteration and compare it with the searchword with every member
				if strings.Contains(strings.ToLower(member), strings.ToLower(searchword)) { // if it correspond with a member
					newArtistData = append(newArtistData, artist)
				}
			}
			for _, location := range artist.Location { // loop to search the iteration and compare it with the searchword with every location
				if strings.Contains(strings.ToLower(location), strings.ToLower(searchword)) { // if it correspond with a location
					newArtistData = append(newArtistData, artist)
				}
			}
		}
	}
	newArtistData = removeDuplicateArtists(newArtistData)
	return newArtistData
}

func ReadApi(url string) []byte {
	// Get api
	artists, err := http.Get(url)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	// Read api and put it in a byte array
	artistData, err := ioutil.ReadAll(artists.Body)
	if err != nil {
		log.Fatal(err)
	}
	return artistData
}

func Filters(artistData []Artist, response Response) []Artist {
	var newArtistData []Artist
	creaDate1, _ := strconv.Atoi(response.CreaDate1)
	creaDate2, _ := strconv.Atoi(response.CreaDate2)
	firstAlbumDate1, _ := strconv.Atoi(response.FirstAlbumDate1)
	firstAlbumDate2, _ := strconv.Atoi(response.FirstAlbumDate2)
	var firstAlbumYearInt int
	var firstAlbumYear string
	if creaDate1 > creaDate2 || firstAlbumDate1 > firstAlbumDate2 {
		fmt.Println("Error")
	} else {
		for _, artist := range artistData {
			firstAlbumYear = artist.FirstAlbum[len(artist.FirstAlbum)-4:]
			firstAlbumYearInt, _ = strconv.Atoi(firstAlbumYear)
			if len(response.NMember) > 0 {
				if response.Location != "any" {
					for _, location := range artist.Location {
						for _, v := range response.NMember {
							vInt, _ := strconv.Atoi(v)
							if creaDate1 <= artist.CreationDate && creaDate2 >= artist.CreationDate && vInt == len(artist.Members) && firstAlbumDate1 <= firstAlbumYearInt && firstAlbumDate2 >= firstAlbumYearInt && location == response.Location {
								newArtistData = append(newArtistData, artist)
							}
						}
					}
				} else {
					for _, v := range response.NMember {
						vInt, _ := strconv.Atoi(v)
						if creaDate1 <= artist.CreationDate && creaDate2 >= artist.CreationDate && vInt == len(artist.Members) && firstAlbumDate1 <= firstAlbumYearInt && firstAlbumDate2 >= firstAlbumYearInt {
							newArtistData = append(newArtistData, artist)
						}
					}
				}
			} else {
				if response.Location != "any" {
					for _, location := range artist.Location {
						if creaDate1 <= artist.CreationDate && creaDate2 >= artist.CreationDate && firstAlbumDate1 <= firstAlbumYearInt && firstAlbumDate2 >= firstAlbumYearInt && location == response.Location {
							newArtistData = append(newArtistData, artist)
						}
					}
				} else {
					if creaDate1 <= artist.CreationDate && creaDate2 >= artist.CreationDate && firstAlbumDate1 <= firstAlbumYearInt && firstAlbumDate2 >= firstAlbumYearInt {
						newArtistData = append(newArtistData, artist)
					}
				}
			}
		}
	}
	return newArtistData
}

func removeDuplicateValues(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func removeDuplicateArtists(intSlice []Artist) []Artist {
	keys := make(map[string]bool)
	list := []Artist{}
	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range intSlice {
		if _, value := keys[entry.Name]; !value {
			keys[entry.Name] = true
			list = append(list, entry)
		}
	}
	return list
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {
	erreur := erreur{}
	w.WriteHeader(status)
	tmpl, err := template.ParseFiles("./templates/erreur.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if status == http.StatusNotFound {
		erreur.Text = "404 NOT FOUND"
		err = tmpl.Execute(w, erreur)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if status == http.StatusBadRequest {
		erreur.Text = "BAD REQUEST"
		err = tmpl.Execute(w, erreur)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if status == http.StatusInternalServerError {
		erreur.Text = "INTERNAL SERVER ERROR"
		err = tmpl.Execute(w, erreur)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
