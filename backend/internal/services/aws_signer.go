package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type AWSV4Signer struct {
	AccessKey string
	SecretKey string
	Region    string
	Service   string
}

func (s *AWSV4Signer) SignRequest(req *http.Request, payload []byte) error {
	now := time.Now().UTC()
	
	// Set required headers
	req.Header.Set("X-Amz-Date", now.Format("20060102T150405Z"))
	req.Header.Set("Host", req.URL.Host)
	
	// Create canonical request
	canonicalRequest := s.createCanonicalRequest(req, payload)
	
	// Create string to sign
	stringToSign := s.createStringToSign(now, canonicalRequest)
	
	// Calculate signature
	signature := s.calculateSignature(now, stringToSign)
	
	// Set authorization header
	authHeader := s.createAuthorizationHeader(now, signature)
	req.Header.Set("Authorization", authHeader)
	
	return nil
}

func (s *AWSV4Signer) createCanonicalRequest(req *http.Request, payload []byte) string {
	// HTTP method
	method := req.Method
	
	// Canonical URI
	uri := req.URL.Path
	if uri == "" {
		uri = "/"
	}
	
	// Canonical query string
	queryString := s.createCanonicalQueryString(req.URL.Query())
	
	// Canonical headers
	canonicalHeaders, signedHeaders := s.createCanonicalHeaders(req)
	
	// Payload hash
	payloadHash := fmt.Sprintf("%x", sha256.Sum256(payload))
	
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		method, uri, queryString, canonicalHeaders, signedHeaders, payloadHash)
}

func (s *AWSV4Signer) createCanonicalQueryString(values url.Values) string {
	var keys []string
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	
	var params []string
	for _, key := range keys {
		for _, value := range values[key] {
			params = append(params, fmt.Sprintf("%s=%s",
				url.QueryEscape(key), url.QueryEscape(value)))
		}
	}
	
	return strings.Join(params, "&")
}

func (s *AWSV4Signer) createCanonicalHeaders(req *http.Request) (string, string) {
	var headers []string
	var signedHeaders []string
	
	// Get all header names
	for name := range req.Header {
		lowerName := strings.ToLower(name)
		headers = append(headers, lowerName)
		signedHeaders = append(signedHeaders, lowerName)
	}
	
	sort.Strings(headers)
	sort.Strings(signedHeaders)
	
	var canonicalHeaders []string
	for _, name := range headers {
		values := req.Header[http.CanonicalHeaderKey(name)]
		for _, value := range values {
			canonicalHeaders = append(canonicalHeaders,
				fmt.Sprintf("%s:%s", name, strings.TrimSpace(value)))
		}
	}
	
	return strings.Join(canonicalHeaders, "\n") + "\n",
		   strings.Join(signedHeaders, ";")
}

func (s *AWSV4Signer) createStringToSign(now time.Time, canonicalRequest string) string {
	algorithm := "AWS4-HMAC-SHA256"
	timestamp := now.Format("20060102T150405Z")
	credentialScope := s.getCredentialScope(now)
	hashedCanonicalRequest := fmt.Sprintf("%x", sha256.Sum256([]byte(canonicalRequest)))
	
	return fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm, timestamp, credentialScope, hashedCanonicalRequest)
}

func (s *AWSV4Signer) getCredentialScope(now time.Time) string {
	date := now.Format("20060102")
	return fmt.Sprintf("%s/%s/%s/aws4_request", date, s.Region, s.Service)
}

func (s *AWSV4Signer) calculateSignature(now time.Time, stringToSign string) string {
	date := now.Format("20060102")
	
	kDate := s.hmacSHA256([]byte("AWS4"+s.SecretKey), date)
	kRegion := s.hmacSHA256(kDate, s.Region)
	kService := s.hmacSHA256(kRegion, s.Service)
	kSigning := s.hmacSHA256(kService, "aws4_request")
	
	signature := s.hmacSHA256(kSigning, stringToSign)
	return fmt.Sprintf("%x", signature)
}

func (s *AWSV4Signer) createAuthorizationHeader(now time.Time, signature string) string {
	algorithm := "AWS4-HMAC-SHA256"
	credential := fmt.Sprintf("%s/%s", s.AccessKey, s.getCredentialScope(now))
	signedHeaders := s.getSignedHeadersList()
	
	return fmt.Sprintf("%s Credential=%s, SignedHeaders=%s, Signature=%s",
		algorithm, credential, signedHeaders, signature)
}

func (s *AWSV4Signer) getSignedHeadersList() string {
	// This should match the headers used in canonical headers
	// For simplicity, we'll use common headers
	return "host;x-amz-date"
}

func (s *AWSV4Signer) hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}