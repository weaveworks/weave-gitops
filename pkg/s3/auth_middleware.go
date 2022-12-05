package s3

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/minio/minio-go/v7/pkg/s3utils"
)

// Signature and API related constants.
const (
	unsignedPayload   = "UNSIGNED-PAYLOAD"
	signV4Algorithm   = "AWS4-HMAC-SHA256"
	iso8601DateFormat = "20060102T150405Z"
	yyyymmdd          = "20060102"
)

type credential struct {
	AccessKeyID string
	Time        string
	Location    string
	Service     string
	Request     string
}

func parseCredential(str string) (credential, error) {
	parts := strings.Split(str, "/")
	if len(parts) != 5 {
		return credential{}, fmt.Errorf("invalid credential format")
	}

	return credential{
		AccessKeyID: parts[0],
		Time:        parts[1],
		Location:    parts[2],
		Service:     parts[3],
		Request:     parts[4],
	}, nil
}

func AuthMiddleware(accessKeyID, secretAccessKey string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, rq *http.Request) {
		if err := verifySignature(*rq, accessKeyID, secretAccessKey); err != nil {
			authorizedError(w, err)
			return
		}

		handler.ServeHTTP(w, rq)
	})
}

func authorizedError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Set("Content-Type", "application/xml")

	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Error>
  <Code>Unauthorized</Code>
  <Message>%s</Message>
</Error>`, err.Error())

	if _, err := w.Write([]byte(xml)); err != nil {
		return
	}
}

// sum256 calculate sha256 sum for an input byte array.
func sum256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)

	return hash.Sum(nil)
}

// sumHMAC calculate hmac between two input byte array.
func sumHMAC(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)

	return hash.Sum(nil)
}

func getHashedPayload(req http.Request) string {
	hashedPayload := req.Header.Get("X-Amz-Content-Sha256")
	if hashedPayload == "" {
		// Presign does not have a payload, use S3 recommended value.
		hashedPayload = unsignedPayload
	}

	return hashedPayload
}

func headerExists(key string, headers []string) bool {
	for _, k := range headers {
		if k == key {
			return true
		}
	}

	return false
}

// getHostAddr returns host header if available, otherwise returns host from URL
func getHostAddr(req *http.Request) string {
	host := req.Header.Get("host")
	if host != "" && req.Host != host {
		return host
	}

	if req.Host != "" {
		return req.Host
	}

	return req.URL.Host
}

// Trim leading and trailing spaces and replace sequential spaces with one space, following Trimall()
// in http://docs.aws.amazon.com/general/latest/gr/sigv4-create-canonical-request.html
func signV4TrimAll(input string) string {
	// Compress adjacent spaces (a space is determined by
	// unicode.IsSpace() internally here) to one space and return
	return strings.Join(strings.Fields(input), " ")
}

func getCanonicalHeaders(req http.Request, signedHeaders string) string {
	headers := []string{}
	vals := make(map[string][]string)

	for _, k := range strings.Split(signedHeaders, ";") {
		lk := strings.ToLower(k)
		headers = append(headers, lk)
		vals[lk] = req.Header.Values(k)
	}

	if !headerExists("host", headers) {
		headers = append(headers, "host")
	}

	sort.Strings(headers)

	var buf bytes.Buffer
	// Save all the headers in canonical form <header>:<value> newline
	// separated for each header.
	for _, k := range headers {
		buf.WriteString(k)
		buf.WriteByte(':')

		switch {
		case k == "host":
			buf.WriteString(getHostAddr(&req))
			buf.WriteByte('\n')
		default:
			for idx, v := range vals[k] {
				if idx > 0 {
					buf.WriteByte(',')
				}

				buf.WriteString(signV4TrimAll(v))
			}

			buf.WriteByte('\n')
		}
	}

	return buf.String()
}

// getSigningKey hmac seed to calculate final signature.
func getSigningKey(secret, loc string, t time.Time, serviceType string) []byte {
	date := sumHMAC([]byte("AWS4"+secret), []byte(t.Format(yyyymmdd)))
	location := sumHMAC(date, []byte(loc))
	service := sumHMAC(location, []byte(serviceType))
	signingKey := sumHMAC(service, []byte("aws4_request"))

	return signingKey
}

// getSignature final signature in hexadecimal form.
func getSignature(signingKey []byte, stringToSign string) string {
	return hex.EncodeToString(sumHMAC(signingKey, []byte(stringToSign)))
}

// getScope generate a string of a specific date, an AWS region, and a
// service.
func getScope(location string, t time.Time, serviceType string) string {
	scope := strings.Join([]string{
		t.Format(yyyymmdd),
		location,
		serviceType,
		"aws4_request",
	}, "/")

	return scope
}

// getCredential generate a credential string.
func getCredential(accessKeyID, location string, t time.Time, serviceType string) string {
	scope := getScope(location, t, serviceType)
	return accessKeyID + "/" + scope
}

// getCanonicalRequest generate a canonical request of style.
//
// canonicalRequest =
//
//	<HTTPMethod>\n
//	<CanonicalURI>\n
//	<CanonicalQueryString>\n
//	<CanonicalHeaders>\n
//	<SignedHeaders>\n
//	<HashedPayload>
func getCanonicalRequest(req http.Request, signedHeaders string, hashedPayload string) string {
	req.URL.RawQuery = strings.ReplaceAll(req.URL.Query().Encode(), "+", "%20")
	canonicalRequest := strings.Join([]string{
		req.Method,
		s3utils.EncodePath(req.URL.Path),
		req.URL.RawQuery,
		getCanonicalHeaders(req, signedHeaders),
		signedHeaders,
		hashedPayload,
	}, "\n")

	return canonicalRequest
}

// getStringToSignV4 a string based on selected query values.
func getStringToSignV4(t time.Time, location, canonicalRequest, serviceType string) string {
	stringToSign := signV4Algorithm + "\n" + t.Format(iso8601DateFormat) + "\n"
	stringToSign = stringToSign + getScope(location, t, serviceType) + "\n"
	stringToSign += hex.EncodeToString(sum256([]byte(canonicalRequest)))

	return stringToSign
}

// verifySignature - verify signature for S3 version '4'
func verifySignature(req http.Request, accessKeyID string, secretAccessKey string) error {
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return fmt.Errorf("header Authorization is missing")
	}

	auth = strings.TrimPrefix(auth, signV4Algorithm+" ")

	parts := strings.Split(auth, ", ")
	if len(parts) != 3 {
		return fmt.Errorf("invalid Authorization header")
	}

	credentialStr := strings.SplitN(parts[0], "=", 2)[1]
	signedHeaders := strings.SplitN(parts[1], "=", 2)[1]
	parsedSignature := strings.SplitN(parts[2], "=", 2)[1]

	hashedPayload := getHashedPayload(req)

	amzDate := req.Header.Get("X-Amz-Date")
	if amzDate == "" {
		return fmt.Errorf("header X-Amz-Date is missing")
	}

	t, err := time.Parse(iso8601DateFormat, amzDate)
	if err != nil {
		return err
	}

	credential, err := parseCredential(credentialStr)
	if err != nil {
		return err
	}

	// Get canonical request
	canonicalRequest := getCanonicalRequest(req, signedHeaders, hashedPayload)

	// Get string to sign from canonical request.
	stringToSign := getStringToSignV4(t, credential.Location, canonicalRequest, credential.Service)

	// Get hmac signing key.
	signingKey := getSigningKey(secretAccessKey, credential.Location, t, credential.Service)

	// Get credential string.
	computedCredential := getCredential(accessKeyID, credential.Location, t, credential.Service)

	// Calculate parsedSignature.
	computedSignature := getSignature(signingKey, stringToSign)

	if computedCredential != credentialStr {
		return fmt.Errorf("access denied: credential does not match")
	}

	if computedSignature != parsedSignature {
		return fmt.Errorf("access denied: signature does not match")
	}

	return nil
}
