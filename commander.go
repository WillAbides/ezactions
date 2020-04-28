package ezactions

import (
	"fmt"
)

// WorkflowCommander issues workflow commands
//
// https://help.github.com/en/actions/reference/workflow-commands-for-github-actions
type WorkflowCommander struct {
	Printer func(string)
}

// CommanderFileLocation is an optional file location used for setting workflow messages
type CommanderFileLocation struct {
	File string
	Line int
	Col  int
}

// SetEnvironmentVariable creates or updates an environment variable for any actions running next in a job. The action
// that creates or updates the environment variable does not have access to the new value, but all subsequent actions
// in a job will have access. Environment variables are case-sensitive and you can include punctuation.
//
// https://help.github.com/en/actions/reference/workflow-commands-for-github-actions#setting-an-environment-variable
func (w *WorkflowCommander) SetEnvironmentVariable(name, value string) {
	w.printf("::set-env name=%s::%s\n", name, value)
}

// SetOutputParameter sets an actions's output parameter.
//
// https://help.github.com/en/actions/reference/workflow-commands-for-github-actions#setting-an-output-parameter
func (w *WorkflowCommander) SetOutputParameter(name, value string) {
	w.printf("::set-output name=%s::%s\n", name, value)
}

// AddSystemPath prepends a directory to the system PATH variable for all subsequent actions in the current job.
// The currently running action cannot access the new path variable.
//
// https://help.github.com/en/actions/reference/workflow-commands-for-github-actions#adding-a-system-path
func (w *WorkflowCommander) AddSystemPath(path string) {
	w.printf("::add-path::%s\n", path)
}

// SetDebugMessage prints a debug message to the log. You must create a secret named ACTIONS_STEP_DEBUG with the value
// true to see the debug messages set by this command in the log. To learn more about creating secrets and using them in
// a step
//
// https://help.github.com/en/actions/reference/workflow-commands-for-github-actions#setting-a-debug-message
func (w *WorkflowCommander) SetDebugMessage(msg string, fileLoc *CommanderFileLocation) {
	w.log("debug", msg, fileLoc)
}

// SetWarningMessage creates a warning message and prints the message to the log. You can optionally provide a filename
// (file), line number (line), and column (col) number where the warning occurred.
//
// https://help.github.com/en/actions/reference/workflow-commands-for-github-actions#setting-a-warning-message
func (w *WorkflowCommander) SetWarningMessage(msg string, fileLoc *CommanderFileLocation) {
	w.log("warning", msg, fileLoc)
}

// SetErrorMessage creates an error message and prints the message to the log. You can optionally provide a filename
// (file), line number (line), and column (col) number where the warning occurred.
//
// https://help.github.com/en/actions/reference/workflow-commands-for-github-actions#setting-an-error-message
func (w *WorkflowCommander) SetErrorMessage(msg string, fileLoc *CommanderFileLocation) {
	w.log("error", msg, fileLoc)
}

// MaskValueInLog - Masking a value prevents a string or variable from being printed in the log. Each masked word
// separated by whitespace is replaced with the * character. You can use an environment variable or string for the mask's
// value.
//
// https://help.github.com/en/actions/reference/workflow-commands-for-github-actions#masking-a-value-in-log
func (w *WorkflowCommander) MaskValueInLog(value string) {
	w.printf("::add-mask::%s\n", value)
}

// StopWorkflowCommands stops processing any workflow commands. This special command allows you to log anything without
// accidentally running a workflow command. For example, you could stop logging to output an entire script that has
// comments.
//
// Returns a function to restart processing workflow commands.
//
// https://help.github.com/en/actions/reference/workflow-commands-for-github-actions#stopping-and-starting-workflow-commands
func (w *WorkflowCommander) StopWorkflowCommands(endtoken string) func() {
	w.printf("::stop-commands::%s\n", endtoken)
	return func() {
		w.printf("::%s::\n", endtoken)
	}
}

func (w *WorkflowCommander) printf(pattern string, args ...interface{}) {
	out := fmt.Sprintf(pattern, args...)
	if w.Printer == nil {
		fmt.Print(out)
	}
	w.Printer(out)
}

func (w *WorkflowCommander) log(logLevel string, msg string, fileLoc *CommanderFileLocation) {
	if fileLoc == nil {
		w.printf("::%s::%s\n", logLevel, msg)
		return
	}
	w.printf("::%s file=%s,line=%d,col=%d::%s\n", logLevel, fileLoc.File, fileLoc.Line, fileLoc.Col, msg)
}
