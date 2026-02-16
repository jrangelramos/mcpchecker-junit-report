package main

import (
	"fmt"
	"strings"
)

// convertToJUnit converts MCP test results to JUnit XML format
func convertToJUnit(results []MCPTestResult) JUnitTestSuites {
	suites := JUnitTestSuites{}

	// Group tests by difficulty
	testsByDifficulty := make(map[string][]MCPTestResult)
	for _, result := range results {
		difficulty := result.Difficulty
		if difficulty == "" {
			difficulty = "unknown"
		}
		testsByDifficulty[difficulty] = append(testsByDifficulty[difficulty], result)
	}

	// Create a test suite for each difficulty level
	for difficulty, tests := range testsByDifficulty {
		suite := JUnitTestSuite{
			Name:      fmt.Sprintf("MCP Checker Tests - %s", difficulty),
			Tests:     len(tests),
			Failures:  0,
			Errors:    0,
			Skipped:   0,
			TestCases: make([]JUnitTestCase, 0, len(tests)),
		}

		for _, test := range tests {
			testCase := convertTestCase(test)
			suite.TestCases = append(suite.TestCases, testCase)

			// Count failures and errors
			if testCase.Failure != nil {
				suite.Failures++
			}
			if testCase.Error != nil {
				suite.Errors++
			}
		}

		suites.Suites = append(suites.Suites, suite)
	}

	return suites
}

// convertTestCase converts a single MCP test result to a JUnit test case
func convertTestCase(test MCPTestResult) JUnitTestCase {
	testCase := JUnitTestCase{
		Name:      test.TaskName,
		Classname: extractClassname(test.TaskPath, test.Difficulty),
		SystemOut: formatHumanReadableOutput(test),
	}

	// Determine if test failed and why
	if !test.TaskPassed {
		// Test execution failed
		testCase.Error = &JUnitError{
			Message: "Test execution failed",
			Type:    "ExecutionError",
			Content: test.TaskError,
		}
		if test.TaskError != "" {
			testCase.SystemErr = test.TaskError
		}
	} else if !test.AllAssertionsPassed {
		// Assertions failed
		failedAssertions := getFailedAssertions(test.AssertionResults)
		testCase.Failure = &JUnitFailure{
			Message: fmt.Sprintf("Assertion failures: %s", strings.Join(failedAssertions, ", ")),
			Type:    "AssertionFailure",
			Content: buildFailureContent(test, failedAssertions),
		}
	}

	// Check phase failures
	phaseErrors := collectPhaseErrors(test)
	if phaseErrors != "" {
		if testCase.Error != nil {
			testCase.Error.Content += "\n\nPhase Errors:\n" + phaseErrors
		} else if testCase.Failure != nil {
			testCase.Failure.Content += "\n\nPhase Errors:\n" + phaseErrors
		} else {
			// Phase failed but test reported as passed - treat as error
			testCase.Error = &JUnitError{
				Message: "Phase execution failed",
				Type:    "PhaseError",
				Content: phaseErrors,
			}
		}
		if testCase.SystemErr == "" {
			testCase.SystemErr = phaseErrors
		} else {
			testCase.SystemErr += "\n\n" + phaseErrors
		}
	}

	return testCase
}

// extractClassname extracts a meaningful classname from the task path
func extractClassname(taskPath string, difficulty string) string {
	if taskPath == "" {
		return difficulty
	}
	// Extract meaningful parts from path
	// e.g., "/home/.../tasks/create-function/create-function.yaml" -> "tasks.create-function"
	parts := strings.Split(taskPath, "/")
	for i, part := range parts {
		if part == "tasks" && i+1 < len(parts) {
			return fmt.Sprintf("tasks.%s", parts[i+1])
		}
	}
	return difficulty
}

// getFailedAssertions returns a list of failed assertion names
func getFailedAssertions(assertions map[string]Assertion) []string {
	var failed []string
	for name, assertion := range assertions {
		if !assertion.Passed {
			failed = append(failed, name)
		}
	}
	return failed
}

// buildFailureContent creates detailed failure content for failed assertions
func buildFailureContent(test MCPTestResult, failedAssertions []string) string {
	var content strings.Builder

	content.WriteString("Failed Assertions:\n")
	for _, assertion := range failedAssertions {
		content.WriteString(fmt.Sprintf("  - %s\n", assertion))
	}

	if test.TaskError != "" {
		content.WriteString("\nError Details:\n")
		content.WriteString(test.TaskError)
	}

	return content.String()
}

// collectPhaseErrors gathers all phase errors from a test result
func collectPhaseErrors(test MCPTestResult) string {
	var errors strings.Builder

	if !test.SetupOutput.Success && test.SetupOutput.Error != "" {
		errors.WriteString("Setup Phase Error:\n")
		errors.WriteString(test.SetupOutput.Error)
		errors.WriteString("\n\n")
	}

	if !test.AgentOutput.Success && test.AgentOutput.Error != "" {
		errors.WriteString("Agent Phase Error:\n")
		errors.WriteString(test.AgentOutput.Error)
		errors.WriteString("\n\n")
	}

	if !test.VerifyOutput.Success && test.VerifyOutput.Error != "" {
		errors.WriteString("Verify Phase Error:\n")
		errors.WriteString(test.VerifyOutput.Error)
		errors.WriteString("\n\n")
	}

	if !test.CleanupOutput.Success && test.CleanupOutput.Error != "" {
		errors.WriteString("Cleanup Phase Error:\n")
		errors.WriteString(test.CleanupOutput.Error)
		errors.WriteString("\n\n")
	}

	return strings.TrimSpace(errors.String())
}

// formatHumanReadableOutput creates a human-readable summary of test results
func formatHumanReadableOutput(test MCPTestResult) string {
	var output strings.Builder

	// Header with test status
	output.WriteString(fmt.Sprintf("Task: %s\n", test.TaskName))
	output.WriteString(fmt.Sprintf("Path: %s\n", test.TaskPath))
	output.WriteString(fmt.Sprintf("Difficulty: %s\n", test.Difficulty))

	status := "PASSED"
	if !test.TaskPassed {
		status = "FAILED"
	}
	output.WriteString(fmt.Sprintf("Status: %s\n", status))

	// Assertions summary
	passedCount := countPassedAssertions(test.AssertionResults)
	totalCount := len(test.AssertionResults)
	output.WriteString(fmt.Sprintf("Assertions: %d/%d passed\n", passedCount, totalCount))

	// Call history summary
	if test.CallHistory.ToolCalls != nil || test.CallHistory.ResourceReads != nil {
		toolCount := len(test.CallHistory.ToolCalls)
		resourceCount := len(test.CallHistory.ResourceReads)

		toolsByServer := groupToolCallsByServer(test.CallHistory.ToolCalls)
		var serverSummaries []string
		for server, count := range toolsByServer {
			serverSummaries = append(serverSummaries, fmt.Sprintf("%s:%d ok", server, count))
		}

		if toolCount > 0 || resourceCount > 0 {
			output.WriteString(fmt.Sprintf("Call history: tools=%d", toolCount))
			if len(serverSummaries) > 0 {
				output.WriteString(fmt.Sprintf(" (%s)", strings.Join(serverSummaries, ", ")))
			}
			if resourceCount > 0 {
				output.WriteString(fmt.Sprintf(" resources=%d", resourceCount))
			}
			output.WriteString("\n")
		}

		// Tool outputs
		if len(test.CallHistory.ToolCalls) > 0 {
			output.WriteString("  Tool output:\n")
			for _, toolCall := range test.CallHistory.ToolCalls {
				statusMarker := "ok"
				if !toolCall.Success {
					statusMarker = "failed"
				}
				output.WriteString(fmt.Sprintf("    • %s::%s (%s)\n", toolCall.ServerName, toolCall.Name, statusMarker))

				// Extract structured content if available
				if toolCall.Result != nil {
					if structuredContent, ok := toolCall.Result["structuredContent"].(map[string]interface{}); ok {
						if message, ok := structuredContent["message"].(string); ok && message != "" {
							// Truncate long messages
							if len(message) > 200 {
								lines := strings.Split(message, "\n")
								if len(lines) > 3 {
									output.WriteString(fmt.Sprintf("      %s\n", strings.TrimSpace(lines[0])))
									output.WriteString(fmt.Sprintf("      … (+%d lines)\n", len(lines)-1))
								} else {
									output.WriteString(fmt.Sprintf("      %s... (truncated)\n", message[:200]))
								}
							} else {
								// Show full message for short outputs
								formattedMsg := strings.ReplaceAll(strings.TrimSpace(message), "\n", "\n      ")
								output.WriteString(fmt.Sprintf("      %s\n", formattedMsg))
							}
						}
					}
				}
			}
		}
	}

	// Timeline (from taskOutput - split into bullet points)
	if test.TaskOutput != "" {
		output.WriteString("Timeline:\n")

		// Split output into paragraphs/sentences
		lines := strings.Split(test.TaskOutput, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Wrap long lines
			if len(line) > 100 {
				wrapped := wrapText(line, 100)
				for i, wrappedLine := range wrapped {
					if i == 0 {
						output.WriteString(fmt.Sprintf("  - note: %s\n", wrappedLine))
					} else {
						output.WriteString(fmt.Sprintf("    %s\n", wrappedLine))
					}
				}
			} else {
				output.WriteString(fmt.Sprintf("  - note: %s\n", line))
			}
		}
	}

	// Error details if test failed
	if test.TaskError != "" {
		output.WriteString("\nError:\n")
		errorLines := strings.Split(test.TaskError, "\n")
		for _, line := range errorLines {
			if line != "" {
				output.WriteString(fmt.Sprintf("  %s\n", line))
			}
		}
	}

	return output.String()
}

// countPassedAssertions counts how many assertions passed
func countPassedAssertions(assertions map[string]Assertion) int {
	count := 0
	for _, assertion := range assertions {
		if assertion.Passed {
			count++
		}
	}
	return count
}

// groupToolCallsByServer groups successful tool calls by server name
func groupToolCallsByServer(toolCalls []ToolCall) map[string]int {
	groups := make(map[string]int)
	for _, call := range toolCalls {
		if call.Success {
			groups[call.ServerName]++
		}
	}
	return groups
}

// wrapText wraps text to fit within a maximum width
func wrapText(text string, maxWidth int) []string {
	var lines []string
	words := strings.Fields(text)

	if len(words) == 0 {
		return lines
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= maxWidth {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
