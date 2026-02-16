package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

func main() {
	var input io.Reader

	// Check if a file argument is provided
	if len(os.Args) > 1 {
		filename := os.Args[1]
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", filename, err)
			os.Exit(1)
		}
		defer file.Close()
		input = file
	} else {
		// Read from stdin
		input = os.Stdin
	}

	// Read all input
	data, err := io.ReadAll(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON
	var testResults []MCPTestResult
	if err := json.Unmarshal(data, &testResults); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Convert to JUnit XML
	junitXML := convertToJUnit(testResults)

	// Output XML
	output, err := xml.MarshalIndent(junitXML, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating XML: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(xml.Header + string(output))
}
