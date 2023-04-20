package kubeinteraction

import (
	"encoding/json"
	"fmt"

	"github.com/openshift-pipelines/pipelines-as-code/pkg/params/info"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
)

const (
	resGroupName            = "results.tekton.dev"
	recordSummaryAnnotation = "/recordSummaryAnnotations"
)

type ResultAnnotation struct {
	Repo          string `json:"repo"`
	Commit        string `json:"commit"`
	EventType     string `json:"eventType"`
	PullRequestID int    `json:"pull_request-id,omitempty"`
}

// Add annotation to PipelineRuns produced by PaC for TektonResults
// to capture data for summary and record.
func AddResultsAnnotation(event *info.Event, pipelineRun *tektonv1.PipelineRun) error {
	if event == nil {
		return fmt.Errorf("nil event")
	}
	resultAnnotation := ResultAnnotation{
		Repo:          event.Repository,
		Commit:        event.SHA,
		EventType:     event.EventType,
		PullRequestID: event.PullRequestNumber,
	}

	// convert the `resultAnnotation` sturct into JSON string
	resAnnotationJSON, err := json.Marshal(resultAnnotation)
	if err != nil {
		return err
	}
	// append the result annotation
	pipelineRun.Annotations[resGroupName+recordSummaryAnnotation] = string(resAnnotationJSON)

	return nil
}
