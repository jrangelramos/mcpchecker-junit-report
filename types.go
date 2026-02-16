package main

import "encoding/xml"

// MCPTestResult represents a single test result from the MCP checker
type MCPTestResult struct {
	TaskName            string               `json:"taskName"`
	TaskPath            string               `json:"taskPath"`
	TaskPassed          bool                 `json:"taskPassed"`
	TaskOutput          string               `json:"taskOutput"`
	TaskError           string               `json:"taskError,omitempty"`
	Difficulty          string               `json:"difficulty"`
	AssertionResults    map[string]Assertion `json:"assertionResults"`
	AllAssertionsPassed bool                 `json:"allAssertionsPassed"`
	CallHistory         CallHistory          `json:"callHistory"`
	SetupOutput         PhaseOutput          `json:"setupOutput"`
	AgentOutput         PhaseOutput          `json:"agentOutput"`
	VerifyOutput        PhaseOutput          `json:"verifyOutput"`
	CleanupOutput       PhaseOutput          `json:"cleanupOutput"`
}

// Assertion represents an individual assertion result
type Assertion struct {
	Passed bool `json:"passed"`
}

// CallHistory represents the history of tool and resource calls
type CallHistory struct {
	ToolCalls     []ToolCall     `json:"ToolCalls"`
	ResourceReads []ResourceRead `json:"ResourceReads"`
}

// ToolCall represents a single tool invocation
type ToolCall struct {
	ServerName string                 `json:"serverName"`
	Success    bool                   `json:"success"`
	Name       string                 `json:"name"`
	Result     map[string]interface{} `json:"result"`
}

// ResourceRead represents a single resource read operation
type ResourceRead struct {
	ServerName string `json:"serverName"`
	Success    bool   `json:"success"`
	URI        string `json:"uri"`
}

// PhaseOutput represents output from a test phase
type PhaseOutput struct {
	Success bool   `json:"Success"`
	Error   string `json:"Error"`
}

// JUnitTestSuites represents the root element of JUnit XML output
type JUnitTestSuites struct {
	XMLName xml.Name `xml:"testsuites"`
	Suites  []JUnitTestSuite
}

// JUnitTestSuite represents a single test suite in JUnit XML format
type JUnitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Skipped   int             `xml:"skipped,attr"`
	TestCases []JUnitTestCase `xml:"testcase"`
}

// JUnitTestCase represents a single test case in JUnit XML format
type JUnitTestCase struct {
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
	Error     *JUnitError   `xml:"error,omitempty"`
	SystemOut string        `xml:"system-out,omitempty"`
	SystemErr string        `xml:"system-err,omitempty"`
}

// JUnitFailure represents a test failure in JUnit XML format
type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// JUnitError represents a test error in JUnit XML format
type JUnitError struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}
