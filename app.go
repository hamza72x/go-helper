package hel

import (
	"fmt"
	"io/ioutil"
	"strings"
	"os"
	"net/http"
	"encoding/json"
	"regexp"
	"log"
	"strconv"
	"sort"
)

func GetURLContent(urlStr string, userAgent string) []byte {

	// fmt.Printf("HTML code of %s ...\n", urlStr)

	// Create HTTP client with timeout
	client := &http.Client{}

	// Create and modify HTTP request before sending
	request, err := http.NewRequest("GET", urlStr, nil)
	ErrOSExit("request, err := http.NewRequest: ", err)

	request.Header.Set("User-Agent", userAgent)

	// Make request
	response, err := client.Do(request)
	ErrOSExit("response, err := client.Do", err)

	htmlBytes, err := ioutil.ReadAll(response.Body)
	ErrOSExit("htmlBytes, err := ioutil.ReadAll", err)

	response.Body.Close()
	client.CloseIdleConnections()

	return htmlBytes
}

func GetFileBytes(filePath string) []byte {

	file, err := os.Open(filePath)
	PErr("reading file "+filePath, err)

	b, err := ioutil.ReadAll(file)
	PErr("ioutil.ReadAll: "+filePath, err)

	file.Close()
	return b
}

func GetFileStr(path string) string {
	return string(GetFileBytes(path))
}

func FileExists(filename string) bool {

	info, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func PathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func AZ_AND_NUMBER_ONLY(str string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(str, "")
}

func GetNonCreatedFileName(baseName string, ext string, i int) string {
	if !FileExists(baseName + ext) {
		return baseName + ext
	} else if !FileExists(baseName + "_" + strconv.Itoa(i) + ext) {
		return baseName + "_" + strconv.Itoa(i) + ext
	}
	return GetNonCreatedFileName(baseName, ext, i+1)
}

func PS(str string) {
	lines("")
	fmt.Println("+ " + str)
}
func PM(str string) {
	fmt.Println("+ " + str)
}
func PE(str string) {
	fmt.Println("+ " + str)
	lines("\n")
}
func P(str string) {
	lines("")
	fmt.Println("+ " + str)
	lines("\n")
}

func Pl(a ...interface{}) {
	lines("")
	fmt.Fprintln(os.Stdout, a...)
	lines("\n")
}
func lines(str string) {
	fmt.Println("-------------------------------------------------------" + str)
}
func ErrOSExit(title string, err error) {
	PErr(title, err)
	if err != nil {
		os.Exit(0)
	}
}

func ContainsInt(array []int, value int) bool {
	for _, a := range array {
		if a == value {
			return true
		}
	}
	return false
}
func PErr(title string, err error) {
	if err != nil {
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("! Error " + title + ": " + err.Error())
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
	}
}

func SortIntAsc(ints []int) []int {
	sort.Slice(ints, func(i, j int) bool {
		return ints[i] < ints[j]
	})
	return ints
}

func pf(a ...interface{}) (n int, err error) {
	lines("")
	return fmt.Fprintln(os.Stdout, a...)
}
func OSExit(str string) {
	PS(str)
	PE("I Quit :'(")
	os.Exit(0)
}
func StrContains(array []string, value string) bool {
	for _, a := range array {
		if a == value {
			return true
		}
	}
	return false
}

func TempWrite(path string, str string) {
	f, _ := os.Create(path)
	f.WriteString(str)
	f.Close()
}

func StrToArr(str string, sep string) []string {
	return strings.Split(
		strings.TrimSpace(str),
		sep,
	)
}

func LimitStrArr(strs []string, limit int) []string {
	var newArr []string
	for i, _ := range strs {
		if i+1 > limit {
			break
		}
		newArr = append(newArr, strs[i])
	}
	return newArr
}

func ArrToStr(strs []string, sep string) string {
	var str = ""
	t := len(strs) - 1
	for i, v := range strs {
		str += v
		if t != i {
			str += sep
		}
	}
	return str
}

func PrettyPrint(data interface{}) {
	var p []byte
	//    var err := error
	p, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s \n", p)
}
