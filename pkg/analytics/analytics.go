package analytics

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
)

// Variables that we'll set @ build time
var (
	Tier = "oss"
)

const (
	app              = "cli"
	analyticsType    = "track"
	commandEvent     = "Command"
	trackEventURL    = "https://app.pendo.io/data/track"
	trackEventSecret = "bf6ab33e-cd70-46e7-4b77-279f54cac447"
)

type analyticsRequestBody struct {
	AnalyticsType string           `json:"type"`
	Event         string           `json:"event"`
	VisitorID     string           `json:"visitorId"`
	Timestamp     int64            `json:"timestamp"`
	Properties    *eventProperties `json:"properties"`
}

type eventProperties struct {
	Tier    string `json:"tier"`
	Version string `json:"version"`
	App     string `json:"app"`
	Command string `json:"command"`
	Flags   string `json:"flags"`
}

// TrackCommand converts the provided command into an event
// and submits it to the analytics service
func TrackCommand(cmd *cobra.Command, userID string) error {
	cmdPath := getCommandPath(cmd)

	flags := getFlags(cmd)

	client := resty.New()

	reqBody := &analyticsRequestBody{
		AnalyticsType: analyticsType,
		Event:         commandEvent,
		VisitorID:     userID,
		Timestamp:     time.Now().UnixMilli(),
		Properties: &eventProperties{
			Tier:    Tier,
			Version: version.Version,
			App:     app,
			Command: cmdPath,
			Flags:   flags,
		},
	}

	reqBodyData, err := json.MarshalIndent(reqBody, "", "  ")
	if err != nil {
		reqBodyData = []byte("invalid request body")
	}

	_, _ = client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("x-pendo-integration-key", trackEventSecret).
		SetBody(string(reqBodyData)).
		Post(trackEventURL)

	return nil
}

func getCommandPath(cmd *cobra.Command) string {
	cmdPath := cmd.CommandPath()

	r, err := regexp.Compile(`^` + cmd.Root().Name() + ` +`)
	if err != nil {
		cmdPath = "unknown command path"
	}

	return r.ReplaceAllString(cmdPath, "")
}

func getFlags(cmd *cobra.Command) string {
	flags := []string{}

	allFlags := cmd.Flags()

	err := allFlags.ParseAll(os.Args[1:], func(flag *pflag.Flag, value string) error {

		flags = append(flags, flag.Name)

		return nil
	})

	if len(flags) > 0 {
		sort.Strings(flags)
	}

	if err != nil {
		flags = append(flags, "unknown flag")
	}

	return strings.Join(flags, " ")
}
