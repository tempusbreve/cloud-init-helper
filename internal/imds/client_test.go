package imds

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_GetToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		if r.Header.Get("X-aws-ec2-metadata-token-ttl-seconds") != "21600" {
			t.Errorf("Expected TTL header to be 21600")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test-token"))
	}))
	defer server.Close()

	client := NewClient()
	client.httpClient = server.Client()

	originalTokenURL := TokenURL
	TokenURL = server.URL
	defer func() { TokenURL = originalTokenURL }()

	ctx := context.Background()
	err := client.getToken(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client.token != "test-token" {
		t.Errorf("Expected token 'test-token', got %s", client.token)
	}

	if client.tokenExp.Before(time.Now()) {
		t.Errorf("Token expiration should be in the future")
	}
}

func TestClient_MakeRequest(t *testing.T) {
	tokenCalled := false
	metadataCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/token") {
			tokenCalled = true
			if r.Method != "PUT" {
				t.Errorf("Expected PUT for token request, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test-token"))
			return
		}

		metadataCalled = true
		if r.Method != "GET" {
			t.Errorf("Expected GET for metadata request, got %s", r.Method)
		}
		if r.Header.Get("X-aws-ec2-metadata-token") != "test-token" {
			t.Errorf("Expected token header to be 'test-token'")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test-response"))
	}))
	defer server.Close()

	client := NewClient()
	client.httpClient = server.Client()

	originalTokenURL := TokenURL
	TokenURL = server.URL + "/token"
	defer func() { TokenURL = originalTokenURL }()

	ctx := context.Background()
	response, err := client.makeRequest(ctx, server.URL+"/metadata")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response != "test-response" {
		t.Errorf("Expected response 'test-response', got %s", response)
	}

	if !tokenCalled {
		t.Error("Token endpoint should have been called")
	}

	if !metadataCalled {
		t.Error("Metadata endpoint should have been called")
	}
}

func TestClient_GetMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/token") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test-token"))
			return
		}

		if r.URL.Path == "/meta-data/instance-id" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("i-1234567890abcdef0"))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient()
	client.httpClient = server.Client()

	originalTokenURL := TokenURL
	originalMetadataURL := MetadataURL
	TokenURL = server.URL + "/token"
	MetadataURL = server.URL + "/meta-data"
	defer func() {
		TokenURL = originalTokenURL
		MetadataURL = originalMetadataURL
	}()

	ctx := context.Background()
	instanceID, err := client.GetMetadata(ctx, "instance-id")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if instanceID != "i-1234567890abcdef0" {
		t.Errorf("Expected instance ID 'i-1234567890abcdef0', got %s", instanceID)
	}
}

func TestClient_GetRegion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/token") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test-token"))
			return
		}

		if r.URL.Path == "/meta-data/placement/availability-zone" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("us-west-2a"))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient()
	client.httpClient = server.Client()

	originalTokenURL := TokenURL
	originalMetadataURL := MetadataURL
	TokenURL = server.URL + "/token"
	MetadataURL = server.URL + "/meta-data"
	defer func() {
		TokenURL = originalTokenURL
		MetadataURL = originalMetadataURL
	}()

	ctx := context.Background()
	region, err := client.GetRegion(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if region != "us-west-2" {
		t.Errorf("Expected region 'us-west-2', got %s", region)
	}
}

func TestClient_ListMetadataPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/token") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test-token"))
			return
		}

		if r.URL.Path == "/meta-data/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("instance-id\ninstance-type\nlocal-ipv4\npublic-ipv4\n"))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient()
	client.httpClient = server.Client()

	originalTokenURL := TokenURL
	originalMetadataURL := MetadataURL
	TokenURL = server.URL + "/token"
	MetadataURL = server.URL + "/meta-data"
	defer func() {
		TokenURL = originalTokenURL
		MetadataURL = originalMetadataURL
	}()

	ctx := context.Background()
	paths, err := client.ListMetadataPaths(ctx, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := []string{"instance-id", "instance-type", "local-ipv4", "public-ipv4"}
	if len(paths) != len(expected) {
		t.Errorf("Expected %d paths, got %d", len(expected), len(paths))
	}

	for i, path := range paths {
		if path != expected[i] {
			t.Errorf("Expected path %s, got %s", expected[i], path)
		}
	}
}
