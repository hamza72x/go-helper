package hel

import (
	"fmt"
	"unicode/utf8"
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

func CheckDomain(name string) error {

	switch {
	case len(name) == 0:
		return nil // an empty domain name will result in a cookie without a domain restriction
	case len(name) > 255:
		return fmt.Errorf("cookie domain: name length is %d, can't exceed 255", len(name))
	}
	var l int
	for i := 0; i < len(name); i++ {
		b := name[i]
		if b == '.' {
			// check domain labels validity
			switch {
			case i == l:
				return fmt.Errorf("cookie domain: invalid character '%c' at offset %d: label can't begin with a period", b, i)
			case i-l > 63:
				return fmt.Errorf("cookie domain: byte length of label '%s' is %d, can't exceed 63", name[l:i], i-l)
			case name[l] == '-':
				return fmt.Errorf("cookie domain: label '%s' at offset %d begins with a hyphen", name[l:i], l)
			case name[i-1] == '-':
				return fmt.Errorf("cookie domain: label '%s' at offset %d ends with a hyphen", name[l:i], l)
			}
			l = i + 1
			continue
		}
		// test label character validity, note: tests are ordered by decreasing validity frequency
		if !(b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || b == '-' || b >= 'A' && b <= 'Z') {
			// show the printable unicode character starting at byte offset i
			c, _ := utf8.DecodeRuneInString(name[i:])
			if c == utf8.RuneError {
				return fmt.Errorf("cookie domain: invalid rune at offset %d", i)
			}
			return fmt.Errorf("cookie domain: invalid character '%c' at offset %d", c, i)
		}
	}
	// check top level domain validity
	switch {
	case l == len(name):
		return fmt.Errorf("cookie domain: missing top level domain, domain can't end with a period")
	case len(name)-l > 63:
		return fmt.Errorf("cookie domain: byte length of top level domain '%s' is %d, can't exceed 63", name[l:], len(name)-l)
	case name[l] == '-':
		return fmt.Errorf("cookie domain: top level domain '%s' at offset %d begins with a hyphen", name[l:], l)
	case name[len(name)-1] == '-':
		return fmt.Errorf("cookie domain: top level domain '%s' at offset %d ends with a hyphen", name[l:], l)
	case name[l] >= '0' && name[l] <= '9':
		return fmt.Errorf("cookie domain: top level domain '%s' at offset %d begins with a digit", name[l:], l)
	}
	return nil
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
	fmt.Println("-------------------------------------------------------")
	fmt.Println("+ " + str)
}
func PM(str string) {
	fmt.Println("+ " + str)
}
func PE(str string) {
	fmt.Println("+ " + str)
	fmt.Println("-------------------------------------------------------\n")
}
func P(str string) {
	fmt.Println("-------------------------------------------------------")
	fmt.Println("+ " + str)
	fmt.Println("-------------------------------------------------------\n")
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
	fmt.Println("-------------------------------------------------------")
	return fmt.Fprintln(os.Stdout, a...)
}
func OSExit(str string) {
	PS(str)
	PE("I Quit :'(")
	os.Exit(0)
}
func ContainsStr(array []string, value string) bool {
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
