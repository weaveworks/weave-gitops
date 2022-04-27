package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getMatchingLabels(appName string) client.MatchingLabels {
	return matchLabel(withPartOfLabel(appName))
}

func decodeFromBase64(v interface{}, enc string) error {
	return json.NewDecoder(base64.NewDecoder(base64.StdEncoding, strings.NewReader(enc))).Decode(v)
}

func encodeToBase64(v interface{}) (string, error) {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)

	err := json.NewEncoder(encoder).Encode(v)
	if err != nil {
		return "", err
	}

	encoder.Close()

	return buf.String(), nil
}
