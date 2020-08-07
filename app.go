package hel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DBModel since *gorm.Model didn't set json keys
type DBModel struct {
	ID        uint       `gorm:"column:id;primary_key" json:"id"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index" json:"deleted_at"`
}

// GetSearchLikeQueryAndArgs returns queryString and queryArgs for gorm
// to search a query in multiple columns
// ex: columns := []string{"title", "long", "short"}
func GetSearchLikeQueryAndArgs(query string, columns []string) (queryStr string, queryArgs []interface{}) {

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

// GetURLContent get contents of a url
func GetURLContent(urlStr string, userAgent string) ([]byte, error) {
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

// GetFileBytes get []byte of a file
func GetFileBytes(filePath string) ([]byte, error) {

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

// GetFileStr get string content of a file
func GetFileStr(path string) (string, error) {
	bytes, err := GetFileBytes(path)
	return string(bytes), err
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

// FilterToAlphabetsAndNumbers filters out to [a-zA-Z0-9] of a string
func FilterToAlphabetsAndNumbers(str string) (string, error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return str, err
	}
	return reg.ReplaceAllString(str, ""), nil
}

// FilterToNumbers filters out to [0-9] of a string
func FilterToNumbers(str string) (string, error) {
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		return str, err
	}
	return reg.ReplaceAllString(str, ""), nil
}

// FilterToAlphabets filters out to [a-zA-Z] of a string
func FilterToAlphabets(str string) (string, error) {
	reg, err := regexp.Compile("[^a-zA-Z]+")
	if err != nil {
		return str, err
	}
	return reg.ReplaceAllString(str, ""), nil
}

// GetNonCreatedFileName returns a unique file name
// if already file exists then -
// it returns by appending a number (and _) before the extension
// ex: config.ini / config_1.ini
func GetNonCreatedFileName(baseName string, ext string, i int) string {
	if !FileExists(baseName + ext) {
		return baseName + ext
	} else if !FileExists(baseName + "_" + strconv.Itoa(i) + ext) {
		return baseName + "_" + strconv.Itoa(i) + ext
	}
	return GetNonCreatedFileName(baseName, ext, i+1)
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

// StrLimitArr splits a array in a certain limit
// limitZB is Zero Based
// Ex: arr = ["a", "b", "c", "d"], limitZB = 1
// returns ["a", "b"]
func StrLimitArr(arr []string, limitZB int) []string {
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
func ArrStrToStr(strs []string, sep string) string {
	var str = ""
	l := len(strs) - 1
	for i, v := range strs {
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
