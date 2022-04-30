package hel

import (
	"bufio"
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

	"gorm.io/gorm"
)

const (
	// UserAgentCrawler generic crawler user agent
	UserAgentCrawler = "Crawler"

	// UserAgentSamsungS9 Samsung Galaxy S9
	UserAgentSamsungS9 = "Mozilla/5.0 (Linux; Android 8.0.0; SM-G960F Build/R16NW) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.84 Mobile Safari/537.36"

	// UserAgentIphoneXSChrome Apple iPhone XS (Chrome)
	UserAgentIphoneXSChrome = "Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/69.0.3497.105 Mobile/15E148 Safari/605.1"

	// UserAgentIphoneXRSafari Apple iPhone XR (Safari)
	UserAgentIphoneXRSafari = "Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.0 Mobile/15E148 Safari/604.1"

	// UserAgentIphoneXSMaxFirefox Apple iPhone XS Max (Firefox)
	UserAgentIphoneXSMaxFirefox = "Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/13.2b11866 Mobile/16A366 Safari/605.1.15"

	// UserAgentChrome79Windows Chrome 79 Windows
	UserAgentChrome79Windows = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.117 Safari/537.36"
)

// GormModel since *gorm.Model didn't set json keys
type GormModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"  json:"-"`
}

// GormModelv2 gorm model v2
type GormModelv2 struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GormModelv3 gorm model v3
type GormModelv3 struct {
	ID uint `gorm:"column:id;primarykey" json:"id"`
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
// return example: queryLikes `post_title` LIKE ? OR `post_title` LIKE ? OR `post_title` LIKE ? OR `post_content` LIKE ? OR `post_content` LIKE ? OR `post_content` LIKE ? OR `post_name` LIKE ? OR `post_name` LIKE ? OR `post_name` LIKE ?
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
	_, err = f.WriteString(str)
	f.Close()
	return err
}

// BytesToFile write byte to a file
func BytesToFile(outFilepath string, bytes []byte) error {
	f, err := os.Create(outFilepath)
	if err != nil {
		return err
	}
	_, err = f.Write(bytes)
	f.Close()
	return err
}

// URLResponse get full response of a url
// make sure to call `defer response.Body.Close()` in your caller function
func URLResponse(urlStr string, userAgent string) (*http.Response, error) {
	// fmt.Printf("HTML code of %s ...\n", urlStr)
	if len(userAgent) == 0 {
		userAgent = UserAgentCrawler
	}
	// Create HTTP client with timeout
	client := &http.Client{}

	// Create and modify HTTP request before sending
	request, err := http.NewRequest("GET", urlStr, nil)

	if err != nil {
		return &http.Response{}, err
	}

	// set user agent
	request.Header.Set("User-Agent", userAgent)

	// Make request
	response, err := client.Do(request)

	if err != nil {
		return &http.Response{}, err
	}

	// defer response.Body.Close()
	client.CloseIdleConnections()

	return response, nil
}

// URLContent get contents of a url
func URLContent(urlStr string, userAgent string) ([]byte, error) {

	// Make request
	response, err := URLResponse(urlStr, userAgent)

	if err != nil {
		return nil, err
	}

	htmlBytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	response.Body.Close()

	return htmlBytes, nil
}

// URLContentMust return []bytes
// panics if failed
func URLContentMust(urlStr string, userAgent string) []byte {

	htmlBytes, err := URLContent(urlStr, userAgent)

	if err != nil {
		panic("[URLContentMust] Error getting data - " + err.Error())
	}

	return htmlBytes
}

// URLStrMust return string of a url
// panics if failed
func URLStrMust(urlStr string, userAgent string) string {

	htmlBytes, err := URLContent(urlStr, userAgent)

	if err != nil {
		panic("[URLStrMust] Error getting data - " + err.Error())
	}

	return string(htmlBytes)
}

// FileRemoveIfExists removes a file if exists
func FileRemoveIfExists(path string) error {
	if FileExists(path) {
		return os.Remove(path)
	}
	return nil
}

// FileWordList returns a file.txt in array and count of arr
func FileWordList(path string) ([]string, int) {

	var lines []string
	var count = 0

	file, err := os.Open(path)

	if err != nil {
		return nil, 0
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		count++
		lines = append(lines, scanner.Text())
	}

	file.Close()

	return lines, count
}

// FileBytes get []byte of a file
func FileBytes(filePath string) ([]byte, error) {

	file, err := os.Open(filePath)

	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(file)

	if err != nil {
		return nil, err
	}

	file.Close()

	return b, nil
}

// FileBytesMust get []byte of a file
// panics if failed
func FileBytesMust(filePath string) []byte {
	bytes, err := FileBytes(filePath)
	if err != nil {
		panic("Error in FileBytesMust, filePath: " + filePath + ", err: " + err.Error())
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

// DirCreateIfNotExists creates a directory if not exists
func DirCreateIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, 0755)
	}
	return nil
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

// ArrIntSortAsc sorts a array of integer in asc order
func ArrIntSortAsc(ints []int) []int {
	sort.Slice(ints, func(i, j int) bool {
		return ints[i] < ints[j]
	})
	return ints
}

// ArrIntSortDesc sorts a array of integer in desc order
func ArrIntSortDesc(ints []int) []int {
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

// StrFilterToAlphabetsAndNumbersMust filters out to [a-zA-Z0-9] of a string
// panics if fails
func StrFilterToAlphabetsAndNumbersMust(str string) string {
	str, err := StrFilterToAlphabetsAndNumbers(str)
	if err != nil {
		panic("error StrFilterToAlphabetsAndNumbersMust: " + err.Error())
	}
	return str
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

// StrToArr string to array with TrimSpace
func StrToArr(str string, sep string) []string {
	return strings.Split(
		strings.TrimSpace(str),
		sep,
	)
}

//==============================================
//					ARRAY
//==============================================

// ArrIntContains check whether a interger contains in a interger array
func ArrIntContains(array []int, value int) bool {
	var contains = false
	for _, a := range array {
		if a == value {
			contains = true
			break
		}
	}
	return contains
}

// ArrStrUnique returns array with unique values of array from a array of string
func ArrStrUnique(array []string) []string {
	var uniques []string
	for _, v := range array {
		if !ArrStrContains(uniques, v) {
			uniques = append(uniques, v)
		}
	}
	return uniques
}

// ArrStrContains check whether a string contains in a string array
func ArrStrContains(array []string, value string) bool {
	var contains = false
	for _, v := range array {
		if v == value {
			contains = true
			break
		}
	}
	return contains
}

// ArrStrHasAnySuffix checkes whether a string has any suffixes
func ArrStrHasAnySuffix(arr []string, str string) bool {
	var has = false
	for i := range arr {
		if strings.HasSuffix(str, arr[i]) {
			has = true
			break
		}
	}
	return has
}

// ArrStrHasAnyPrefix checkes whether a string has any prefixes
func ArrStrHasAnyPrefix(arr []string, str string) bool {
	var has = false
	for i := range arr {
		if strings.HasPrefix(str, arr[i]) {
			has = true
			break
		}
	}
	return has
}

// ArrStrLimit splits a array in a certain limit
// limitZB is Zero Based
// Ex: arr = ["a", "b", "c", "d"], limitZB = 1
// returns ["a", "b"]
func ArrStrLimit(arr []string, limitZB int) []string {
	var new []string
	for i := range arr {
		if i > limitZB {
			break
		}
		new = append(new, arr[i])
	}
	return new
}

// ArrStrToStr array string to string
func ArrStrToStr(arr []string, sep string) string {
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
	p, err := json.MarshalIndent(data, "", "    ")
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

// PlP panic error
func PlP(str string, err error) {
	if err != nil {
		fmt.Println("#####################")
		panic("[hel.PlP] panicing, " + str + ", error details = " + err.Error())
	}
}
func lines(str string) {
	fmt.Println("-------------------------------------------------------" + str)
}
