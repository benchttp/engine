package runner

import (
	"github.com/benchttp/engine/runner/internal/config"
	"github.com/benchttp/engine/runner/internal/output"
	"github.com/benchttp/engine/runner/internal/requester"
)

type (
	ConfigBody    = config.Body
	ConfigGlobal  = config.Global
	ConfigRequest = config.Request
	ConfigRunner  = config.Runner
	ConfigOutput  = config.Output

	Requester       = requester.Requester
	RequesterConfig = requester.Config
	RequesterState  = requester.State
	RequesterStatus = requester.Status

	OutputReport = output.Report
)

const (
	StatusRunning  RequesterStatus = "RUNNING"
	StatusCanceled RequesterStatus = "CANCELED"
	StatusTimeout  RequesterStatus = "TIMEOUT"
	StatusDone     RequesterStatus = "DONE"

	ConfigFieldMethod         = config.FieldMethod
	ConfigFieldURL            = config.FieldURL
	ConfigFieldHeader         = config.FieldHeader
	ConfigFieldBody           = config.FieldBody
	ConfigFieldRequests       = config.FieldRequests
	ConfigFieldConcurrency    = config.FieldConcurrency
	ConfigFieldInterval       = config.FieldInterval
	ConfigFieldRequestTimeout = config.FieldRequestTimeout
	ConfigFieldGlobalTimeout  = config.FieldGlobalTimeout
	ConfigFieldSilent         = config.FieldSilent
	ConfigFieldTemplate       = config.FieldTemplate
)

var (
	ConfigDefault     = config.Default
	ConfigFieldsUsage = config.FieldsUsage
	ConfigNewBody     = config.NewBody
	ConfigIsField     = config.IsField

	NewRequester = requester.New

	NewOutput = output.New
)
