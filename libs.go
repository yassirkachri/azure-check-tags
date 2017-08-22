package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// Init Loggers.
func initLog() {
	logsFile = getLogFile()

	Trace = log.New(logsFile,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(logsFile,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(logsFile,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(logsFile,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

// Get the Log file. One file per hour is created YYYYMMDD-HH-logs.log
func getLogFile() io.Writer {
	logsFilef, err := os.OpenFile(time.Now().Local().Format("20060102-15-")+"logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	onErrorFail(err, "OpenFile Logs failed")
	return (logsFilef)

}

//check if a directory exists if not create it
func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			onErrorFail(err, "CreateDirectory Output failed")

		}
	}
}

// write Output file form lines strucutre
func writeOutputFileFromLines(linesOut []azOutputLine, filename string, separator string) {
	if separator == "" {
		separator = ";"
	}
	//create  Output Dir if not exists
	createDirIfNotExist(outputsDir)
	lf, err := os.OpenFile(outputsDir+filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	onErrorFail(err, "WriteOutputFileFromLines OpenFile Ouptut failed")
	defer lf.Close()
	for i := 0; i < len(linesOut); i++ {
		l := linesOut[i]
		message := ""
		message = l.AccountName + separator + l.TenantID + separator + l.SubscriptionName + separator + l.ResourceGroupName + separator + l.ResourceID + separator + l.ResourceName + separator + l.ResourceType + separator + l.ResourceLocation + separator
		for _, xv := range l.Tags {
			message = message + xv + separator
		}
		_, err = lf.WriteString(message + "\n")
		onErrorFail(err, "WriteString Logs failed")

	}

}

// send a GET or POST request using http client
// sData map is used in GET method to build the querystring and for POST method to create the data to be sent
func sendRequest(sMethod string, sHost string, sResource string, sData map[string]string, sHeader map[string]string) []byte {
	apiURL := sHost
	resource := sResource
	u, err := url.ParseRequestURI(apiURL)
	onErrorFail(err, "ParseRequestURI failed")
	u.Path = resource
	urlStr := u.String()
	client := &http.Client{}
	var req *http.Request

	if sMethod == "POST" {
		Trace.Println("Write Datas for POST")
		ladata := url.Values{}
		i := 0
		for xkey, xvalue := range sData {
			if i == 0 {
				ladata.Set(xkey, xvalue)
				i = i + 1
			} else {
				ladata.Add(xkey, xvalue)
			}
		}
		req, err = http.NewRequest(sMethod, urlStr, bytes.NewBufferString(ladata.Encode()))
		onErrorFail(err, "Prepare Request http.NewRequest failed")
	}

	if sMethod == "GET" {
		req, err = http.NewRequest(sMethod, urlStr, nil)
		onErrorFail(err, "Prepare Request http.NewRequest GET Method failed")
		q := req.URL.Query()
		for xkey, xvalue := range sData {
			q.Add(xkey, xvalue)
		}
		req.URL.RawQuery = q.Encode()
		Trace.Println("URL ", req.URL.String())

	}
	// set Header if it's not empty
	if sHeader != nil {
		Trace.Println("Write Header")
		for xkey, xvalue := range sHeader {
			req.Header.Add(xkey, xvalue)
		}
	}

	resp, err := client.Do(req)
	onErrorFail(err, "client.Do(req) failed")

	defer resp.Body.Close()
	lebody, err := ioutil.ReadAll(resp.Body)
	onErrorFail(err, "ioutil.ReadAll(resp.Body) failed")
	// if OK
	if resp.StatusCode == 200 {
		return lebody
	}
	//else if
	Error.Println("Error  ", resp.StatusCode, resp.Status)
	Error.Println("Error  Body ", string(lebody))
	return nil

}

// onErrorFail prints a failure message and exits the program if err is not nil.
func onErrorFail(err error, message string) {
	if err != nil {
		fmt.Printf("%s: %s\n", message, err)
		Error.Println(message, err)
		os.Exit(1)
	}
}

// get the ressource group name from the resource ID
func getRessourceGroupFromID(resID string) string {
	r, _ := regexp.Compile(`.+/resourceGroups/(.+)/providers/.+`)
	result := r.FindStringSubmatch(resID)
	for _, v := range result {
		if v != resID {
			return (v)
		}
	}
	return ("")
}

// check if a Tag (key, value) is ok
// TAG not exists return  (NoTAG)
//  TAG Value not matching the regular expression return  TagValueKO
// IF OK return  value
func isValideTags(xkey string, xvalue string, pTags map[string]string) string {

	for xk, xv := range pTags {
		if strings.ToUpper(xkey) == strings.ToUpper(xk) && isValideTagsValue(xv, xvalue) {
			return xv
		}
		if strings.ToUpper(xkey) == strings.ToUpper(xk) && !isValideTagsValue(xv, xvalue) {
			return "TagValueKO"
		}
	}
	return "NoTAG"
}

// chek if a value is matching a regexp
func isValideTagsValue(xvalue string, xregx string) bool {
	r, _ := regexp.Compile(xregx)
	if r.MatchString(xvalue) {
		return true
	}
	return false
}

// a function used for GET request when a NEXLINK is provided by an API
func sendGetRequest(sRUL string, sHeader map[string]string) []byte {

	client := &http.Client{}
	req, _ := http.NewRequest("GET", sRUL, nil)

	if sHeader != nil {
		for xkey, xvalue := range sHeader {
			req.Header.Add(xkey, xvalue)
		}
	}
	resp, err := client.Do(req)
	onErrorFail(err, "NextLINK : client.Do(req) failed")
	defer resp.Body.Close()
	lebody, err := ioutil.ReadAll(resp.Body)
	onErrorFail(err, "NextLINK : ioutil.ReadAll(resp.Body) NextLINK failed")

	if resp.StatusCode == 200 {
		return lebody
	}
	Error.Println("Error  ", resp.StatusCode, resp.Status)
	Error.Println("Error  Body ", string(lebody))
	return nil

}
