//go:build e2e
// +build e2e

package test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/google/go-github/v43/github"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params/info"
	tgithub "github.com/openshift-pipelines/pipelines-as-code/test/pkg/github"
	"github.com/openshift-pipelines/pipelines-as-code/test/pkg/payload"
	twait "github.com/openshift-pipelines/pipelines-as-code/test/pkg/wait"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// {"severity":"ERROR","timestamp":"2022-03-18T06:58:29.133001347Z","logger":"pipelinesascode","caller":"adapter/sinker.go:34",
// "message":"failed to parse event: could not refresh installation id 348340043's token: received non 2xx response status
// \"404 Not Found\" when fetching https://api.github.com/app/installations/348340043/access_tokens","provider":"github","event":"",
// "stacktrace":"github.com/openshift-pipelines/pipelines-as-code/pkg/adapter.(*sinker).processEvent\n\tgithub.com/openshift-pipelines/pipelines-as-code/pkg/adapter/sinker.go:34\ngithub.com/openshift-pipelines/pipelines-as-code/pkg/adapter.listener.handleEvent.func1.1\n\tgithub.com/openshift-pipelines/pipelines-as-code/pkg/adapter/adapter.go:104"}

func TestGithubPullRequestOkToTest(t *testing.T) {
	t.Skip()
	ctx := context.TODO()
	runcnx, ghcnx, opts, targetNS, targetRefName, prNumber, sha := tgithub.RunPullRequest(ctx, t, "Github OkToTest comment", "testdata/pipelinerun.yaml", false)
	defer tgithub.TearDown(ctx, t, runcnx, ghcnx, prNumber, targetRefName, targetNS, opts)

	repoinfo, resp, err := ghcnx.Client.Repositories.Get(ctx, opts.Organization, opts.Repo)
	assert.NilError(t, err)
	if resp != nil && resp.Response.StatusCode == http.StatusNotFound {
		t.Errorf("Repository %s not found in %s", opts.Organization, opts.Repo)
	}

	runevent := info.Event{
		BaseBranch:    repoinfo.GetDefaultBranch(),
		DefaultBranch: repoinfo.GetDefaultBranch(),
		HeadBranch:    targetRefName,
		Organization:  opts.Organization,
		Repository:    opts.Repo,
		URL:           repoinfo.GetHTMLURL(),
		SHA:           sha,
		Sender:        opts.Organization,
	}

	installID, err := strconv.ParseInt(os.Getenv("TEST_GITHUB_REPO_INSTALLATION_ID"), 10, 64)
	assert.NilError(t, err)
	event := github.IssueCommentEvent{
		Comment: &github.IssueComment{
			Body: github.String(`/ok-to-test`),
		},
		Installation: &github.Installation{
			ID: &installID,
		},
		Action: github.String("created"),
		Issue: &github.Issue{
			State: github.String("open"),
			PullRequestLinks: &github.PullRequestLinks{
				HTMLURL: github.String(fmt.Sprintf("%s/%s/pull/%d",
					os.Getenv("TEST_GITHUB_API_URL"),
					os.Getenv("TEST_GITHUB_REPO_OWNER"), prNumber)),
			},
		},
		Repo: &github.Repository{
			DefaultBranch: &runevent.DefaultBranch,
			HTMLURL:       &runevent.URL,
			Name:          &runevent.Repository,
			Owner:         &github.User{Login: &runevent.Organization},
		},
		Sender: &github.User{
			Login: &runevent.Sender,
		},
	}

	err = payload.Send(ctx,
		runcnx,
		os.Getenv("TEST_EL_URL"),
		os.Getenv("TEST_EL_WEBHOOK_SECRET"),
		os.Getenv("TEST_GITHUB_API_URL"),
		os.Getenv("TEST_GITHUB_REPO_INSTALLATION_ID"),
		event,
		"issue_comment",
	)
	assert.NilError(t, err)

	runcnx.Clients.Log.Infof("Wait for the second repository update to be updated")
	waitOpts := twait.Opts{
		RepoName:        targetNS,
		Namespace:       targetNS,
		MinNumberStatus: 1,
		PollTimeout:     twait.DefaultTimeout,
		TargetSHA:       sha,
	}
	err = twait.UntilRepositoryUpdated(ctx, runcnx.Clients, waitOpts)
	assert.NilError(t, err)

	runcnx.Clients.Log.Infof("Check if we have the repository set as succeeded")
	repo, err := runcnx.Clients.PipelineAsCode.PipelinesascodeV1alpha1().Repositories(targetNS).Get(ctx, targetNS, metav1.GetOptions{})
	assert.NilError(t, err)
	assert.Assert(t, repo.Status[len(repo.Status)-1].Conditions[0].Status == corev1.ConditionTrue)
}
