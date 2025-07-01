package imds

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	TokenTTLSeconds = 21600
	defaultTokenURL = "http://169.254.169.254/latest/api/token"
	defaultMetadataURL = "http://169.254.169.254/latest/meta-data"
	defaultDynamicURL = "http://169.254.169.254/latest/dynamic"
	defaultUserDataURL = "http://169.254.169.254/latest/user-data"
)

var (
	TokenURL    = defaultTokenURL
	MetadataURL = defaultMetadataURL
	DynamicURL  = defaultDynamicURL
	UserDataURL = defaultUserDataURL
)

type Client struct {
	httpClient *http.Client
	token      string
	tokenExp   time.Time
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) getToken(ctx context.Context) error {
	if c.token != "" && time.Now().Before(c.tokenExp) {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", TokenURL, nil)
	if err != nil {
		return fmt.Errorf("creating token request: %w", err)
	}

	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", fmt.Sprintf("%d", TokenTTLSeconds))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("getting IMDSv2 token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	tokenBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading token response: %w", err)
	}

	c.token = string(tokenBytes)
	c.tokenExp = time.Now().Add(time.Duration(TokenTTLSeconds) * time.Second)

	return nil
}

func (c *Client) makeRequest(ctx context.Context, url string) (string, error) {
	if err := c.getToken(ctx); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-aws-ec2-metadata-token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	return string(body), nil
}

func (c *Client) GetMetadata(ctx context.Context, path string) (string, error) {
	path = strings.TrimPrefix(path, "/")
	url := fmt.Sprintf("%s/%s", MetadataURL, path)
	return c.makeRequest(ctx, url)
}

func (c *Client) GetDynamic(ctx context.Context, path string) (string, error) {
	path = strings.TrimPrefix(path, "/")
	url := fmt.Sprintf("%s/%s", DynamicURL, path)
	return c.makeRequest(ctx, url)
}

func (c *Client) GetUserData(ctx context.Context) (string, error) {
	return c.makeRequest(ctx, UserDataURL)
}

func (c *Client) GetInstanceID(ctx context.Context) (string, error) {
	return c.GetMetadata(ctx, "instance-id")
}

func (c *Client) GetInstanceType(ctx context.Context) (string, error) {
	return c.GetMetadata(ctx, "instance-type")
}

func (c *Client) GetLocalIPv4(ctx context.Context) (string, error) {
	return c.GetMetadata(ctx, "local-ipv4")
}

func (c *Client) GetPublicIPv4(ctx context.Context) (string, error) {
	return c.GetMetadata(ctx, "public-ipv4")
}

func (c *Client) GetRegion(ctx context.Context) (string, error) {
	az, err := c.GetMetadata(ctx, "placement/availability-zone")
	if err != nil {
		return "", err
	}
	return az[:len(az)-1], nil
}

func (c *Client) GetAvailabilityZone(ctx context.Context) (string, error) {
	return c.GetMetadata(ctx, "placement/availability-zone")
}

func (c *Client) GetIAMRole(ctx context.Context) (string, error) {
	return c.GetMetadata(ctx, "iam/security-credentials/")
}

func (c *Client) GetInstanceIdentityDocument(ctx context.Context) (string, error) {
	return c.GetDynamic(ctx, "instance-identity/document")
}

func (c *Client) ListMetadataPaths(ctx context.Context, path string) ([]string, error) {
	response, err := c.GetMetadata(ctx, path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(response), "\n")
	var paths []string
	for _, line := range lines {
		if line != "" {
			paths = append(paths, line)
		}
	}

	return paths, nil
}
