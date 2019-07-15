package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	//"regexp"
	"strings"
	"sync"
	"time"
)

func main() {

	SlowSearch(ioutil.Discard)
	FastSearch(ioutil.Discard)

	start := time.Now()
	slowOut := new(bytes.Buffer)
	SlowSearch(slowOut)
	slowResult := slowOut.String()
	end := time.Since(start)
	fmt.Println("SlowSearch: ", end)

	start = time.Now()
	fastOut := new(bytes.Buffer)
	FastSearch(fastOut)
	fastResult := fastOut.String()
	end = time.Since(start)
	fmt.Println("FastSearch: ", end)

	if slowResult != fastResult {
		fmt.Printf("results not match\nGot:\n%v\nExpected:\n%v", fastResult, slowResult)
	}
}

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	//regexpAt := regexp.MustCompile("@")
	var foundUsers bytes.Buffer
	foundUsers.WriteString("found users:\n")
	lines := strings.Split(string(fileContents), "\n")
	users := getUsers(lines)
	size := len(users)
	seenBrowsers := make(map[string]string, size)

	isAndroid := false
	isMSIE := false
	foundAndroid := false
	foundMSIE := false

	for i, user := range users {

		browsers, ok := user["browsers"].([]interface{})
		if !ok {
			// log.Println("cant cast browsers")
			continue
		}
		isAndroid = false
		isMSIE = false
		//isAndroidAndMSIE := false

		for _, browserRaw := range browsers {

			browser, ok := browserRaw.(string)
			if !ok {
				continue
			}
			foundAndroid = strings.Contains(browser, "Android")
			foundMSIE = strings.Contains(browser, "MSIE")
			if foundAndroid || foundMSIE {

				isAndroid = foundAndroid || isAndroid
				isMSIE = foundMSIE || isMSIE

				_, ok := seenBrowsers[browser]

				if !ok {
					seenBrowsers[browser] = ""
				}

			} else {
				continue
			}

		}

		if !(isAndroid && isMSIE) {
			continue
		} else {
			//email := regexpAt.ReplaceAllString(user["email"].(string), " [at] ")
			email := strings.ReplaceAll(user["email"].(string), "@", " [at] ")
			foundUser := fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
			foundUsers.WriteString(foundUser)
		}

	}

	fmt.Fprintln(out, foundUsers.String())
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}

func getUsers(lines []string) []map[string]interface{} {

	wg := &sync.WaitGroup{}
	size := len(lines)
	users := make([]map[string]interface{}, size, size)

	for i, line := range lines {

		wg.Add(1)
		go func(i int, line string, users []map[string]interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			user := make(map[string]interface{})
			// fmt.Printf("%v %v\n", err, line)
			err := json.Unmarshal([]byte(line), &user)
			if err != nil {
				panic(err)
			}
			//users = append(users, user)
			users[i] = user
		}(i, line, users, wg)
	}

	wg.Wait()
	return users
}
