package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"encoding/json"
	"net/http"
	"strings"
)

type TestCase struct {
	Description string `json:"description"`
	Calls       []Call `json:"calls"`
}

type Call struct {
	On     On `json:"on"`
	Expect Expect `json:"expect"`
	Remember map[string]string `json:"remember"`
}

type On struct {
	Method string `json:"method"`
	Url    string `json:"url"`
	Headers map[string]string `json:"headers"`
	Params map[string]string `json:"params"`
}

type Expect struct {
	StatusCode  int    `json:"statusCode"`
	ContentType string `json:"contentType"`
	Body        map[string]interface{} `json:"body"`
}

func main() {

	testCases, err := readTestCases()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Test Cases: %v\n", testCases)

	call(testCases[0].Calls[0]) //TODO cycle
}

func readTestCases() (testCases []TestCase, err error) {
	file, e := ioutil.ReadFile("C:/_ws_go/src/trest/tests.json") //TODO extract file(s) to param

	if e != nil {
		fmt.Printf("File error: %v\n", e)
		return
	}

	err = json.Unmarshal(file, &testCases)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	return
}

func call(call Call)(rememberMap map[string]string, failedExpectations []string) {
	on := call.On

	req, _ := http.NewRequest(on.Method, "http://localhost:8080"+on.Url, nil) //TODO extract url to param

	for key, value := range on.Headers {
		req.Header.Add(key, value)
	}

	q := req.URL.Query()
	for key, value := range on.Params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	fmt.Println(req)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error when sending request")
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response")
		return
	}

	//fmt.Printf("Code: %v\n", resp.Status)
	fmt.Printf("Resp: %v\n", string(body))

	var bodyMap map[string]interface{}
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		fmt.Println("Error parsing body")
		return
	}

	//fmt.Printf("bm: %v\n", bodyMap)

	//v := getByPath(bodyMap, "token")
	//fmt.Printf("v: %v\n", v)

	rememberMap = remember(bodyMap, call.Remember)
	fmt.Printf("rememberMap: %v\n", rememberMap)


	return
}

func remember (bodyMap map[string]interface{}, remember map[string]string) map[string]string {
	var rememberedMap = make(map[string]string)

	for varName, path := range remember {
		if strings.HasPrefix(path, "body.") {
			bodyPath := strings.TrimPrefix(path, "body.")
			splitPath := strings.Split(bodyPath, ".")

			b := make([]interface{}, len(splitPath))
			for i := range splitPath {
				b[i] = splitPath[i]
			}

			rememberedMap[varName]=getByPath(bodyMap, b...).(string)
			//fmt.Printf("v: %v\n", getByPath(bodyMap, b...))
		}
	}

	return rememberedMap
}

func getByPath(m interface{}, path ...interface{}) interface{} {

	for _, p := range path {
		switch idx := p.(type) {
		case string:
			m = m.(map[string]interface{})[idx]
		case int:
			m = m.([]interface{})[idx]
		}
	}
	return m
}

// TODO remember
// TODO resp matches schema
// TODO expect matchers: equal, anyOf, arrHasSize, arrHasItems
// TODO expect matchers without <any> indexes
// TODO expect: header, statusCode, body
// TODO add company name to test case (track snapshot usage)