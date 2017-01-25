// sitemap.go - A site map generator for https://defacto2.net/detail/
//
// References:
// Wikipedia XML Sitemaps https://en.wikipedia.org/wiki/Site_map
// Package xml https://golang.org/pkg/encoding/xml/#example_Encoder
// Build Web Application with Golang - 7.1 XML https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/07.1.html

package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	dbName   string = "defacto2"                     // Database name
	dbServer string = "tcp(localhost:3306)"          // Database server connection, protocol(IP or domain address:port number)
	dbUser   string = "root"                         // Database username to login
	pwPath   string = "/path/to/password"            // The path to a secured text file containing the MySQL login password
	resource string = "https://defacto2.net/detail/" // The base URL used in the <loc> tag
)

func main() {

	// url are the <url></url> tags in the sitemap
	type url struct {
		Location     string `xml:"loc"`
		LastModified string `xml:"lastmod,omitempty"` // optional
	}
	// Urlset is the sitemap xml template
	type Urlset struct {
		XMLName xml.Name `xml:"urlset"`
		XMLNS   string   `xml:"xmlns,attr"`
		Svs     []url    `xml:"url"`
	}

	// fetch database password
	pwFile, err := os.Open(pwPath)
	checkErr(err)
	defer pwFile.Close()

	pw, err := ioutil.ReadAll(pwFile)
	checkErr(err)
	password := strings.TrimSpace(fmt.Sprintf("%s", pw))

	// connect to the database
	db, err := sql.Open("mysql", fmt.Sprintf("root:x%v@%v/%v", password, dbServer, dbName))
	checkErr(err)

	// query
	var id string
	var createdat sql.NullString
	var updatedat sql.NullString
	rows, err := db.Query("SELECT `id`,`createdat`,`updatedat` FROM `files` WHERE `deletedat` IS NULL")
	checkErr(err)

	defer db.Close()

	v := &Urlset{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}

	// handle query results
	for rows.Next() {
		err = rows.Scan(&id, &createdat, &updatedat)
		checkErr(err)
		// check for valid createdat and updatedat entries
		_, udErr := updatedat.Value()
		_, cdErr := createdat.Value()
		if udErr != nil || cdErr != nil {
			continue // skip record (log in future?)
		}
		// parse createdat and updatedat to use in the <lastmod> tag
		var lastmodString string
		if udValid := updatedat.Valid; udValid == true {
			lastmodString = updatedat.String
		} else if cdValid := createdat.Valid; cdValid == true {
			lastmodString = createdat.String
		}
		lastmodFields := strings.Fields(lastmodString)
		// append url
		var lastmodValue string // blank by default; <lastmod> tag has `omitempty` set, so it won't display if no value is given
		if len(lastmodFields) > 0 {
			lastmodValue = lastmodFields[0]
		}
		v.Svs = append(v.Svs, url{fmt.Sprintf("%v%v", resource, obfuscateParam(id)), lastmodValue})
	}
	output, err := xml.MarshalIndent(v, "  ", "    ")
	checkErr(err)

	os.Stdout.Write([]byte(xml.Header))
	os.Stdout.Write(output)

}

func checkErr(err error) {
	if err != nil {
		log.Fatal("ERROR:", err)
	}
}

// obfuscateParam is a port of the CFWheels, Global Helper Function, obfuscateParam() function.
// It takes database auto increment, primary key attributes and obfuscates their values.
// API reference http://docs.cfwheels.org/reference#obfuscateparam
// CFML source https://github.com/cfwheels/cfwheels/blob/1.4/wheels/global/public.cfm
func obfuscateParam(param string) string {
	rv := param // return value
	// check to make sure param doesn't begin with a 0 digit
	if rv0 := rv[0]; rv0 == '0' {
		return rv
	}
	paramInt, err := strconv.Atoi(param) // convert param to an int type
	if err != nil {
		return rv
	}
	iEnd := len(rv) // count the number of digits in param
	afloat64 := math.Pow10(iEnd) + float64(reverseInt(paramInt))
	// keep a and b as int type
	a := int(afloat64)
	b := 0
	for i := 1; i <= iEnd; i++ {
		// slice individual digits from param and sum them
		paramSlice, err := strconv.Atoi(string(param[iEnd-i]))
		if err != nil {
			return rv
		}
		b += paramSlice
	}
	// base 64 conversion
	a ^= 461
	b += 154
	return strconv.FormatInt(int64(b), 16) + strconv.FormatInt(int64(a), 16)
}

// reverseInt reverses an integer.
// reverseInt(12345) will return 54321.
// Credit, Wade73: http://stackoverflow.com/questions/35972561/reverse-int-golang
func reverseInt(value int) int {
	intString := strconv.Itoa(value)
	newString := ""
	for x := len(intString); x > 0; x-- {
		newString += string(intString[x-1])
	}
	newInt, err := strconv.Atoi(newString)
	checkErr(err)
	return newInt
}
