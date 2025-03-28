---
title: PipelineRun status
weight: 6
---
# Status

## GitHub

When the pipeline finishes, the status will be added in the GitHub Check tabs
with a short recap of how long each task of your pipeline took and the output of
`tkn pr describe`.

## Webhook

On webhook if it's a pull request

## Failures

If a namespace has been matched to a Repository, Pipelines As Code will emit its log messages in the kubernetes events inside the `Repository`'s namespace.

## CRD

Status of your pipeline execution is stored inside the Repo CustomResource :

```console
% kubectl get repo -n pipelines-as-code-ci
NAME                  URL                                                        NAMESPACE             SUCCEEDED   REASON      STARTTIME   COMPLETIONTIME
pipelines-as-code-ci   https://github.com/openshift-pipelines/pipelines-as-code   pipelines-as-code-ci   True        Succeeded   59m         56m
```

The last 5 status are stored inside the CustomResource and can be accessed
directly like this :

```console
% kubectl get repo -n pipelines-as-code-ci -o json|jq .items[].pipelinerun_status
[
  {
    "completionTime": "2021-05-05T11:00:05Z",
    "conditions": [
      {
        "lastTransitionTime": "2021-05-05T11:00:05Z",
        "message": "Tasks Completed: 3 (Failed: 0, Cancelled 0), Skipped: 0",
        "reason": "Succeeded",
        "status": "True",
        "type": "Succeeded"
      }
    ],
    "pipelineRunName": "pipelines-as-code-test-run-7tr84",
    "startTime": "2021-05-05T10:53:43Z"
  },
  {
    "completionTime": "2021-05-05T11:20:18Z",
    "conditions": [
      {
        "lastTransitionTime": "2021-05-05T11:20:18Z",
        "message": "Tasks Completed: 3 (Failed: 0, Cancelled 0), Skipped: 0",
        "reason": "Succeeded",
        "status": "True",
        "type": "Succeeded"
      }
    ],
    "pipelineRunName": "pipelines-as-code-test-run-2fhhg",
    "startTime": "2021-05-05T11:11:20Z"
  },
  […]
```

## Notifications

Notifications are not handled by Pipelines as Code, the only place where we
notify a status in an interface is when we do a Pull Request on for example the
GitHub checks interface to show the results of the pull request.

If you need some other type of notification you can use
the [finally feature of tekton pipeline](https://github.com/tektoncd/pipeline/blob/main/docs/pipelines.md#adding-finally-to-the-pipeline)
.

Here is an example task to send a Slack message on failures (or success if you
like) :

<https://github.com/chmouel/tekton-slack-task-status>

The push pipeline of Pipelines as Code use this task, you can see the example
here :

[.tekton/push.yaml](https://github.com/openshift-pipelines/pipelines-as-code/blob/7b41cc3f769af40a84b7ead41c6f037637e95070/.tekton/push.yaml#L116)
