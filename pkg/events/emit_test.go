package events

import (
	"context"
	"testing"

	"github.com/openshift-pipelines/pipelines-as-code/pkg/apis/pipelinesascode/v1alpha1"
	testclient "github.com/openshift-pipelines/pipelines-as-code/pkg/test/clients"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	zapobserver "go.uber.org/zap/zaptest/observer"
	"gotest.tools/v3/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rtesting "knative.dev/pkg/reconciler/testing"
)

func TestEventEmitter_EmitMessage(t *testing.T) {
	observer, _ := zapobserver.New(zap.InfoLevel)
	fakelogger := zap.New(observer).Sugar()
	tests := []struct {
		name        string
		repo        *v1alpha1.Repository
		message     string
		logLevel    zapcore.Level
		expectEvent bool
	}{
		{
			name: "repo exists",
			repo: &v1alpha1.Repository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-repo",
					Namespace: "test-ns",
				},
				Spec: v1alpha1.RepositorySpec{},
			},
			message:     "info-message",
			logLevel:    zap.InfoLevel,
			expectEvent: true,
		},
		{
			name:        "repo doesn't exists",
			repo:        nil,
			message:     "error-message",
			logLevel:    zap.ErrorLevel,
			expectEvent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := rtesting.SetupFakeContext(t)
			stdata, _ := testclient.SeedTestData(t, ctx, testclient.Data{})

			// emit event
			NewEventEmitter(stdata.Kube, fakelogger).EmitMessage(tt.repo, tt.logLevel, tt.message)

			if tt.expectEvent {
				events, err := stdata.Kube.CoreV1().Events(tt.repo.Namespace).List(context.Background(), metav1.ListOptions{})
				assert.NilError(t, err)
				assert.Equal(t, events.Items[0].Message, tt.message)
				assert.Equal(t, events.Items[0].Type, v1.EventTypeNormal)
			}
		})
	}
}
