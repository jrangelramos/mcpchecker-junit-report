# MCP Checker JUnit Report Converter

A CLI tool that converts MCP Checker test result JSON files to JUnit XML format for integration with CI/CD systems.

## Features

- Converts MCP Checker JSON test results to JUnit XML format
- Supports reading from file argument or stdin
- Groups tests by difficulty level (easy, medium, hard)
- Captures assertion failures and phase errors
- **Human-readable output format** inspired by `mcpchecker view`
  - Task summary with status and difficulty
  - Assertion pass/fail counts
  - Tool call history and summaries
  - Formatted timeline with bullet points
  - Tool output with intelligent truncation for long messages
- Preserves all test output and error messages

## Installation

### Install directly using Go (recommended)

Install the latest version directly from the repository:

```bash
go install github.com/jrangelramos/mcpchecker-junit-report@latest
```

The binary will be installed to `$GOPATH/bin` (usually `$HOME/go/bin`). Make sure this directory is in your `PATH`.

### Build from source

Clone the repository and build manually:

```bash
git clone https://github.com/jrangelramos/mcpchecker-junit-report.git
cd mcpchecker-junit-report
go build -o mcpchecker-junit-report
```

Or using Make:

```bash
make build
```

## Usage

### Read from file
```bash
mcpchecker-junit-report mcpchecker-eval-all-out.json > junit-report.xml
```

### Read from stdin
```bash
cat mcpchecker-eval-all-out.json | mcpchecker-junit-report > junit-report.xml
```

**Note:** If you built from source and didn't install to your PATH, use `./mcpchecker-junit-report` instead of `mcpchecker-junit-report`.

## JSON to JUnit Mapping

The tool maps MCP Checker test results to JUnit XML as follows:

| MCP Field | JUnit Element | Description |
|-----------|---------------|-------------|
| `taskName` | `testcase.name` | Name of the test |
| `taskPath` | `testcase.classname` | Extracted from path (e.g., "tasks.create-function") |
| `difficulty` | `testsuite.name` | Tests grouped by difficulty level |
| `taskPassed` | `error` element | If false, test execution failed |
| `allAssertionsPassed` | `failure` element | If false, assertions failed |
| `taskOutput` | `system-out` | Standard output from test |
| `taskError` | `system-err` | Error messages |
| `assertionResults` | `failure.content` | Details of failed assertions |
| Phase outputs | `system-err` | Errors from setup/agent/verify/cleanup phases |

## JUnit XML Output Structure

```xml
<testsuites>
  <testsuite name="MCP Checker Tests - easy" tests="3" failures="0" errors="0">
    <testcase name="create-function" classname="tasks.create-function">
      <system-out>Perfect! I've successfully created...</system-out>
    </testcase>
    <!-- More test cases -->
  </testsuite>
  <testsuite name="MCP Checker Tests - medium" tests="2" failures="1" errors="0">
    <!-- Medium difficulty tests -->
  </testsuite>
</testsuites>
```

## Test Result Categories

- **Pass**: `taskPassed=true` and `allAssertionsPassed=true`
- **Failure**: `taskPassed=true` but `allAssertionsPassed=false` (assertion failures)
- **Error**: `taskPassed=false` (execution errors)

## Output Format

The `system-out` field in the JUnit XML is formatted for human readability, similar to `mcpchecker view`:

```
Task: create-function
Path: /home/.../tasks/create-function/create-function.yaml
Difficulty: easy
Status: PASSED
Assertions: 3/3 passed
Call history: tools=1 (func-mcp:1 ok) resources=3
  Tool output:
    • func-mcp::create (ok)
      Created node function in /tmp/fn-create
Timeline:
  - note: Perfect! I've successfully created a Node.js Function named 'fn-create'
    at `/tmp/fn-create` using the default http template.
  - note: The Function has been initialized and is ready for development.
```

This structured format makes it easy to:
- Quickly identify test status and difficulty
- See assertion results at a glance
- Understand which tools were called
- Follow the test execution timeline
- Review error details in context

## Example

Given the input JSON:
```json
[
  {
    "taskName": "create-function",
    "taskPath": "/path/to/tasks/create-function/task.yaml",
    "taskPassed": true,
    "difficulty": "easy",
    "allAssertionsPassed": true,
    "assertionResults": {
      "toolsUsed": {"passed": true},
      "minToolCalls": {"passed": true}
    },
    "callHistory": {
      "ToolCalls": [{"serverName": "func-mcp", "name": "create", "success": true}]
    },
    "taskOutput": "Successfully created function"
  }
]
```

The tool generates JUnit XML with human-readable system-out content (see format above).
