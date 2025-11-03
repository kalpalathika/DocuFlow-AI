package main

import (
	"fmt"
	"os"

	"github.com/you/lexsy-mvp/server/docx"
)

func main() {
	path := "/Users/kalpalathikaramanujam/Downloads/Service_Agreement_Template.docx"
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	fmt.Printf("File size: %d bytes\n", len(data))
	fmt.Println("Attempting to detect fields...")

	fields, err := docx.DetectFields(data)
	if err != nil {
		fmt.Println("Error detecting fields:", err)
		return
	}

	fmt.Printf("Success! Found %d fields:\n", len(fields))
	for _, f := range fields {
		fmt.Println("  -", f)
	}
}
