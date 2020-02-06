// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	file, err := ioutil.ReadFile("../tags.json")
	checkError(err)
	var data []map[string]string

	err = json.Unmarshal(file, &data)
	checkError(err)

	output := []string{`// Code generated by go generate; DO NOT EDIT.
// This file was generated automatically based
// on tags.json, available at the repository's 
// root. 

package meta

func init() {
	Tags = []TagInfo {
`}

	for _, entry := range data {
		tagBytes := len(entry["tag"])
		bytes := make([]string, tagBytes/2)
		for i := 0; i < tagBytes; i += 2 {
			bytes[i/2] = "0x" + entry["tag"][i:i+2]
		}
		output = append(output, "		TagInfo { ")
		output = append(output, "[]byte{ "+strings.Join(bytes, ", ")+" }, ")
		output = append(output, "\""+entry["name"]+"\", ")
		output = append(output, "\""+entry["description"]+"\" },\n")
	}
	output = append(output, "	}\n}\n")

	fmt.Println(strings.Join(output, ""))

	err = ioutil.WriteFile("./tags_generated.go", []byte(strings.Join(output, "")), 0600)
	checkError(err)
}