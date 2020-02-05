// (C) Copyright 2019 Hewlett Packard Enterprise Development LP

package greenlake

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Client - wrapper class for greenlake api's
type Client struct {
	GrantType    string
	ClientID     string
	ClientSecret string
	TenantID     string
	Host         string
	APIKey       string
}

// Token structure
type Token struct {
	AccessToken     string `json:"access_token"`
	Scope           string `json:"scope"`
	TokenType       string `json:"token_type"`
	Expiry          string `json:"expiry"`
	ExpiresIn       int    `json:"expires_in"`
	AccessTokenOnly bool   `json:"accessTokenOnly"`
}

// User structure
type User struct {
	Active      bool   `json:"active"`
	DisplayName string `json:"displayName"`
	UserName    string `json:"userName"`
	Name        Name   `json:"name"`
}

// Name structure
type Name struct {
	FamilyName string `json:"familyName"`
	GivenName  string `json:"givenName"`
}

// NewGreenLakeClient create
func NewGreenLakeClient(grantType, clientID, secretKey, tenantID, host string) *Client {
	return &Client{
		GrantType:    grantType,
		ClientID:     clientID,
		ClientSecret: secretKey,
		TenantID:     tenantID,
		Host:         host,
		APIKey:       "",
	}
}

// NewGLClientFromAPIKey creates a new GreenLake Client from existing API sessions key
func NewGLClientFromAPIKey(host, tenantID, apikey string) *Client {
	return &Client{
		GrantType:    "client_credentials",
		ClientID:     "",
		ClientSecret: "LOCAL",
		APIKey:       apikey,
		TenantID:     tenantID,
		Host:         host,
	}
}

// GetToken api
func (c *Client) GetToken() (Token, error) {
	var result Token

	url := fmt.Sprintf(c.Host + "/identity/v1/token")
	jsonData := map[string]string{"grant_type": c.GrantType,
		"client_id":     c.ClientID,
		"client_secret": c.ClientSecret,
		"tenant_id":     c.TenantID}
	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))

	if err != nil {
		return result, err
	}

	request.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(request)

	if err != nil {
		return result, err
	}

	if err = json.Unmarshal(body, &result); err != nil {
		return result, err
	}

	return result, err
}

// GetUsers to list users
func (c *Client) GetUsers(path string) ([]byte, error) {
	url := fmt.Sprintf(c.Host + "/scim/v1/tenant/" + c.TenantID + "/" + path)
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "application/scim+json")
	request.Header.Set("Authorization", "Bearer "+c.APIKey)
	body, err := c.doRequest(request)

	return body, err
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	// Ignore invalid certificate
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error in response and response status: %s", response.Status)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	return body, err
}
