package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	//"regexp"
	json "encoding/json"
	"strings"
	"sync"
	"time"

	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

//easyjson:json
type User struct {
	Browsers []string `json:"browsers"`
	Company  string   `json:"company"`
	Country  string   `json:"country"`
	Email    string   `json:"email"`
	Job      string   `json:"job"`
	Name     string   `json:"name"`
	Phone    string   `json:"phone"`
}

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

func FastSearch(out io.Writer) {
	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		panic(fileErr)
	}
	defer file.Close()

	writer := bytes.Buffer{}
	fileScanner := bufio.NewScanner(file)
	seenBrowsers := make(map[string]struct{}, 0)
	intBuff := make([]byte, 0, 8)

	writer.WriteString("found users:\n")
	var i int64
	user := &User{}
	for fileScanner.Scan() {
		err := user.UnmarshalJSON(fileScanner.Bytes())
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {
			if strings.Contains(browser, "Android") {
				isAndroid = true
				seenBrowsers[browser] = struct{}{}
			} else if strings.Contains(browser, "MSIE") {
				isMSIE = true
				seenBrowsers[browser] = struct{}{}
			}
		}

		if !(isAndroid && isMSIE) {
			i++
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		indx := strings.Index(user.Email, "@")
		writer.WriteByte('[')
		intBuff = intBuff[:0]
		writer.Write(strconv.AppendInt(intBuff, i, 10))
		writer.WriteString("] ")
		writer.WriteString(user.Name)
		writer.WriteString(" <")
		writer.WriteString(user.Email[:indx])
		writer.WriteString(" [at] ")
		writer.WriteString(user.Email[indx+1:])
		writer.WriteString(">\n")
		writer.WriteTo(out)
		i++
	}

	writer.WriteString("\nTotal unique browsers ")
	intBuff = intBuff[:0]
	writer.Write(strconv.AppendInt(intBuff, int64(len(seenBrowsers)), 10))
	writer.WriteByte('\n')
	writer.WriteTo(out)
}

func FastSearchOld(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	fmt.Fprint(out, "found users:\n")
	//out.Write("found users:\n")
	lines := strings.Split(string(fileContents), "\n")
	users := *getUsers(&lines)
	size := len(users)
	seenBrowsers := make(map[string]struct{}, size)
	writer := bytes.Buffer{}
	intBuff := make([]byte, 0, 8)

	isAndroid := false
	isMSIE := false
	foundAndroid := false
	foundMSIE := false

	for i, user := range users {

		isAndroid = false
		isMSIE = false

		for _, browser := range user.Browsers {

			foundAndroid = strings.Contains(browser, "Android")
			foundMSIE = strings.Contains(browser, "MSIE")

			if foundAndroid || foundMSIE {

				isAndroid = foundAndroid || isAndroid
				isMSIE = foundMSIE || isMSIE

				_, ok := seenBrowsers[browser]

				if !ok {
					seenBrowsers[browser] = struct{}{}
				}

			}

		}

		if isAndroid && isMSIE {
			//fmt.Fprintf(out, "[%d] %s <%s>\n", i, user.Name, ReplaceAll(user.Email, "@", " [at] "))
			indx := strings.Index(user.Email, "@")
			writer.WriteByte('[')
			intBuff = intBuff[:0]
			writer.Write(strconv.AppendInt(intBuff, int64(i), 10))
			writer.WriteString("] ")
			writer.WriteString(user.Name)
			writer.WriteString(" <")
			writer.WriteString(user.Email[:indx])
			writer.WriteString(" [at] ")
			writer.WriteString(user.Email[indx+1:])
			writer.WriteString(">\n")
			writer.WriteTo(out)
		}

	}

	//fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
	writer.WriteString("\nTotal unique browsers ")
	intBuff = intBuff[:0]
	writer.Write(strconv.AppendInt(intBuff, int64(len(seenBrowsers)), 10))
	writer.WriteByte('\n')
	writer.WriteTo(out)
}

// ReplaceAll returns a copy of the string s with all
// non-overlapping instances of old replaced by new.
// If old is empty, it matches at the beginning of the string
// and after each UTF-8 sequence, yielding up to k+1 replacements
// for a k-rune string.
func ReplaceAll(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

func getUsers(lines *[]string) *[]User {

	wg := &sync.WaitGroup{}
	size := len(*lines)
	users := make([]User, size, size)

	for i, line := range *lines {

		wg.Add(1)
		go func(i int, line string, users []User, wg *sync.WaitGroup) {
			defer wg.Done()
			err := users[i].UnmarshalJSON([]byte(line))
			if err != nil {
				panic(err)
			}
		}(i, line, users, wg)
	}

	wg.Wait()
	return &users
}

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson9f2eff5fDecodeJson(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "company":
			out.Company = string(in.String())
		case "country":
			out.Country = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "job":
			out.Job = string(in.String())
		case "name":
			out.Name = string(in.String())
		case "phone":
			out.Phone = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9f2eff5fEncodeJson(out *jwriter.Writer, in User) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"browsers\":"
		out.RawString(prefix[1:])
		if in.Browsers == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Browsers {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"company\":"
		out.RawString(prefix)
		out.String(string(in.Company))
	}
	{
		const prefix string = ",\"country\":"
		out.RawString(prefix)
		out.String(string(in.Country))
	}
	{
		const prefix string = ",\"email\":"
		out.RawString(prefix)
		out.String(string(in.Email))
	}
	{
		const prefix string = ",\"job\":"
		out.RawString(prefix)
		out.String(string(in.Job))
	}
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"phone\":"
		out.RawString(prefix)
		out.String(string(in.Phone))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v User) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9f2eff5fEncodeJson(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v User) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9f2eff5fEncodeJson(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9f2eff5fDecodeJson(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *User) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9f2eff5fDecodeJson(l, v)
}
