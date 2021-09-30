package templates_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/pkg/templates"
)

func TestGetTemplates(t *testing.T) {
	tests := []struct {
		name             string
		ts               []templates.Template
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no templates",
			expected: "No templates found.\n",
		},
		{
			name: "templates includes just name",
			ts: []templates.Template{
				{
					Name: "template-a",
				},
				{
					Name: "template-b",
				},
			},
			expected: "NAME\tDESCRIPTION\ntemplate-a\ntemplate-b\n",
		},
		{
			name: "templates include all fields",
			ts: []templates.Template{
				{
					Name:        "template-a",
					Description: "a desc",
				},
				{
					Name:        "template-b",
					Description: "b desc",
				},
			},
			expected: "NAME\tDESCRIPTION\ntemplate-a\ta desc\ntemplate-b\tb desc\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve templates from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewFakeClient(tt.ts, tt.err)
			w := new(bytes.Buffer)
			err := templates.GetTemplates(c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

type FakeClient struct {
	ts  []templates.Template
	err error
}

func NewFakeClient(ts []templates.Template, err error) *FakeClient {
	return &FakeClient{
		ts:  ts,
		err: err,
	}
}

func (c *FakeClient) Source() string {
	return "In-memory fake"
}

func (c *FakeClient) RetrieveTemplates() ([]templates.Template, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.ts, nil
}
