package hel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GormModel since *gorm.Model didn't set json keys
type GormModel struct {
	ID        uint       `gorm:"column:id;primary_key" json:"id"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index" json:"-"`
}

// URLValid tests a string to determine if it is a well-structured url or not.
// valid: http://www.golangcode.com
// invalid: golangcode.com
func URLValid(toTest string) bool {

	_, err := url.ParseRequestURI(toTest)

	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)

	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// MixFile returns asset file with version, ex: /public/css/app.css?id=f1bbd1956
// make sure to set:  r.Static("public", "./public")
// asset: app.js / app.css, mixManifestPath: public/mix-manifest.json
// panics if can't get mix file or can't unmarshal!
func MixFile(asset string, mixManifestPath string) string {

	var mixManifest map[string]string
	fileBytes, err := FileBytes(mixManifestPath)

	if err != nil {
		panic("Error getting mix-manifest.json file")
	}

	if err := json.Unmarshal(fileBytes, &mixManifest); err != nil {
		panic("Error Unmarshal mix-manifest.json file")
	}

	var assetURL = "/public/"
	var subFolder = "css/"

	if strings.HasSuffix(asset, ".js") {
		subFolder = "js/"
	}

	// in case following array fails
	assetURL += subFolder + asset

	for key, value := range mixManifest {
		if strings.Contains(key, asset) {
			assetURL = "/public" + value
			break
		}
	}

	return assetURL
}

// GormSearchLikeQueryAndArgs returns queryString and queryArgs for gorm
// to search a query in multiple columns
// ex: columns := []string{"title", "long", "short"}
// then call like:  db.Where(queryStr, queryArgs...).Limit(20).Find(&model)
func GormSearchLikeQueryAndArgs(query string, columns []string) (queryStr string, queryArgs []interface{}) {

	likes := []string{"%" + query, query + "%", "%" + query + "%"}

	for i := range columns {
		for j := range likes {
			if j == 0 && i != 0 {
				queryStr += " OR "
			}
			queryStr += "`" + columns[i] + "`" + ` LIKE ?`
			queryArgs = append(queryArgs, likes[j])
			if len(likes)-1 != j {
				queryStr += " OR "
			}
		}
	}

	return
}

// StrToFile write string to a file
func StrToFile(outFilepath, str string) error {
	f, err := os.Create(outFilepath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(str)
	return err
}

// URLContent get contents of a url
func URLContent(urlStr string, userAgent string) ([]byte, error) {
	// fmt.Printf("HTML code of %s ...\n", urlStr)

	// Create HTTP client with timeout
	client := &http.Client{}
	defer client.CloseIdleConnections()

	// Create and modify HTTP request before sending
	request, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	// set user agent
	request.Header.Set("User-Agent", userAgent)

	// Make request
	response, err := client.Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	htmlBytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	return htmlBytes, nil
}

// FileBytes get []byte of a file
func FileBytes(filePath string) ([]byte, error) {

	file, err := os.Open(filePath)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	b, err := ioutil.ReadAll(file)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// FileBytesMust get []byte of a file
// panics if failed
func FileBytesMust(filePath string) []byte {
	bytes, err := FileBytes(filePath)
	if err != nil {
		panic("[panic] in FileBytesMust, filePath: " + filePath + ", err: " + err.Error())
	}
	return bytes
}

// FileStr get string content of a file
func FileStr(path string) (string, error) {
	bytes, err := FileBytes(path)
	return string(bytes), err
}

// FileStrMust get string of a file
// panics if failed
func FileStrMust(filePath string) string {
	bytes, err := FileBytes(filePath)
	if err != nil {
		panic("[panic] in FileStrMust, filePath: " + filePath + ", err: " + err.Error())
	}
	return string(bytes)
}

// FileExists checks is file exists and not directory?
func FileExists(filename string) bool {

	info, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

// PathExists checks if path exists, can be a directory or file
func PathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// NonCreatedFileName returns a unique file name
// if already file exists then -
// it returns by appending a number (and _) before the extension
// ex: config.ini / config_1.ini
func NonCreatedFileName(baseName string, ext string, i int) string {
	if !FileExists(baseName + ext) {
		return baseName + ext
	} else if !FileExists(baseName + "_" + strconv.Itoa(i) + ext) {
		return baseName + "_" + strconv.Itoa(i) + ext
	}
	return NonCreatedFileName(baseName, ext, i+1)
}

// IntContains check whether a interger contains in a interger array
func IntContains(array []int, value int) bool {
	var exists = false
	for _, a := range array {
		if a == value {
			exists = true
			break
		}
	}
	return exists
}

// IntSortAsc sorts a array of integer in asc order
func IntSortAsc(ints []int) []int {
	sort.Slice(ints, func(i, j int) bool {
		return ints[i] < ints[j]
	})
	return ints
}

// IntSortDesc sorts a array of integer in desc order
func IntSortDesc(ints []int) []int {
	sort.Slice(ints, func(i, j int) bool {
		return ints[i] > ints[j]
	})
	return ints
}

// StrFilterToAlphabetsAndNumbers filters out to [a-zA-Z0-9] of a string
func StrFilterToAlphabetsAndNumbers(str string) (string, error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return str, err
	}
	return reg.ReplaceAllString(str, ""), nil
}

// StrFilterToNumbers filters out to [0-9] of a string
func StrFilterToNumbers(str string) (string, error) {
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		return str, err
	}
	return reg.ReplaceAllString(str, ""), nil
}

// StrFilterToAlphabets filters out to [a-zA-Z] of a string
func StrFilterToAlphabets(str string) (string, error) {
	reg, err := regexp.Compile("[^a-zA-Z]+")
	if err != nil {
		return str, err
	}
	return reg.ReplaceAllString(str, ""), nil
}

// StrContains check whether a string contains in a string array
func StrContains(array []string, value string) bool {
	for _, a := range array {
		if a == value {
			return true
		}
	}
	return false
}

// StrToArr string to array with TrimSpace
func StrToArr(str string, sep string) []string {
	return strings.Split(
		strings.TrimSpace(str),
		sep,
	)
}

// StrArrLimit splits a array in a certain limit
// limitZB is Zero Based
// Ex: arr = ["a", "b", "c", "d"], limitZB = 1
// returns ["a", "b"]
func StrArrLimit(arr []string, limitZB int) []string {
	var new []string
	for i := range arr {
		if i > limitZB {
			break
		}
		new = append(new, arr[i])
	}
	return new
}

// StrArrToStr array string to string
func StrArrToStr(arr []string, sep string) string {
	var str = ""
	l := len(arr) - 1
	for i, v := range arr {
		str += v
		if l != i {
			str += sep
		}
	}
	return str
}

// PrettyPrint prints a interface with all fields
func PrettyPrint(data interface{}) {
	var p []byte
	p, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s \n", p)
}

// Pl prints interface with long dash
// ---------------
// interface print
// ---------------
func Pl(a ...interface{}) {
	lines("")
	fmt.Fprintln(os.Stdout, a...)
	lines("\n")
}

func lines(str string) {
	fmt.Println("-------------------------------------------------------" + str)
}
