package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/taskcluster/generic-worker/gwconfig"
	"github.com/taskcluster/httpbackoff"
	tcclient "github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/tcsecrets"
)

var (
	// not a const, because in testing we swap this out
	GCPMetadataBaseURL = "http://metadata.google.internal/computeMetadata/v1/"
)

type GCPUserData struct {
	WorkerType    string                       `json:"workerType"`
	WorkerGroup   string                       `json:"workerGroup"`
	ProvisionerID string                       `json:"provisionerId"`
	CredentialURL string                       `json:"credentialURL"`
	RootURL       string                       `json:"rootURL"`
	Data          WorkerTypeDefinitionUserData `json:"userData"`
}

type CredentialRequestData struct {
	Token string `json:"token"`
}

type TaskclusterCreds struct {
	AccessToken string `json:"accessToken"`
	ClientID    string `json:"clientId"`
	Certificate string `json:"certificate"`
}

func queryGCPMetaData(client *http.Client, path string) (string, error) {
	req, err := http.NewRequest("GET", GCPMetadataBaseURL+path, nil)

	if err != nil {
		return "", err
	}

	req.Header.Add("Metadata-Flavor", "Google")

	resp, _, err := httpbackoff.ClientDo(client, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	return string(content), err
}

func updateConfigWithGCPSettings(c *gwconfig.Config) error {
	log.Print("Querying GCP Metadata to get default worker type config settings...")
	// these are just default values, will be overwritten if set in worker type config
	c.ShutdownMachineOnInternalError = true
	c.ShutdownMachineOnIdle = true

	client := &http.Client{}
	userDataString, err := queryGCPMetaData(client, "instance/attributes/taskcluster")
	if err != nil {
		return err
	}

	var userData GCPUserData
	err = json.Unmarshal([]byte(userDataString), &userData)
	if err != nil {
		return err
	}

	c.ProvisionerID = userData.ProvisionerID
	c.WorkerType = userData.WorkerType
	c.WorkerGroup = userData.WorkerGroup

	c.RootURL = userData.RootURL

	// Now we get taskcluster credentials via instance identity
	// TODO: Disable getting instance identity after first run
	instanceIDPath := fmt.Sprintf("instance/service-accounts/default/identity?audience=%s&format=full", c.RootURL)
	instanceIDToken, err := queryGCPMetaData(client, instanceIDPath)
	if err != nil {
		return err
	}

	data := CredentialRequestData{Token: instanceIDToken}
	reqData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	dataBuffer := bytes.NewBuffer(reqData)

	credentialURL := userData.CredentialURL
	req, err := http.NewRequest("POST", credentialURL, dataBuffer)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, _, err := httpbackoff.ClientDo(client, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var creds TaskclusterCreds
	err = json.Unmarshal([]byte(content), &creds)
	if err != nil {
		return err
	}

	c.AccessToken = creds.AccessToken
	c.ClientID = creds.ClientID
	c.Certificate = creds.Certificate

	gcpMetadata := map[string]interface{}{}
	for _, path := range []string{
		"instance/image",
		"instance/id",
		"instance/machine-type",
		"instance/network-interfaces/0/access-configs/0/external-ip",
		"instance/zone",
		"instance/hostname",
		"instance/network-interfaces/0/ip",
	} {
		key := path[strings.LastIndex(path, "/")+1:]
		value, err := queryGCPMetaData(client, path)
		if err != nil {
			return err
		}
		gcpMetadata[key] = value
	}
	c.WorkerTypeMetadata["gcp"] = gcpMetadata
	c.WorkerID = "gcp-" + gcpMetadata["id"].(string)
	c.PublicIP = net.ParseIP(gcpMetadata["external-ip"].(string))
	c.PrivateIP = net.ParseIP(gcpMetadata["ip"].(string))
	c.InstanceID = gcpMetadata["id"].(string)
	c.InstanceType = gcpMetadata["machine-type"].(string)
	c.AvailabilityZone = gcpMetadata["zone"].(string)

	// TODO: These next two sections should be abstracted out into shared between gcp and aws

	// Host setup per worker type "userData" section.
	//
	// Note, we first update configuration from public host setup, before
	// calling tc-secrets to get private host setup, in case secretsBaseURL is
	// configured in userdata.
	err = c.MergeInJSON(userData.Data.GenericWorker, func(a map[string]interface{}) map[string]interface{} {
		if config, exists := a["config"]; exists {
			return config.(map[string]interface{})
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error applying /data/genericWorker/config from GCP userdata to config: %v", err)
	}

	// Fetch additional (secret) host setup from taskcluster-secrets service.
	// See: https://bugzil.la/1375200
	tcsec := c.Secrets()
	secretName := "worker-type:" + c.ProvisionerID + "/" + c.WorkerType
	sec, err := tcsec.Get(secretName)
	if err != nil {
		// 404 error is ok, since secrets aren't required. Anything else indicates there was a problem retrieving
		// secret or talking to secrets service, so they should return an error
		if apiCallException, isAPICallException := err.(*tcclient.APICallException); isAPICallException {
			rootCause := apiCallException.RootCause
			if badHTTPResponseCode, isBadHTTPResponseCode := rootCause.(httpbackoff.BadHttpResponseCode); isBadHTTPResponseCode {
				if badHTTPResponseCode.HttpResponseCode == 404 {
					log.Printf("WARNING: No worker secrets for worker type %v - secret %v does not exist.", c.WorkerType, secretName)
					err = nil
					sec = &tcsecrets.Secret{
						Secret: json.RawMessage(`{}`),
					}
				}
			}
		}
	}
	if err != nil {
		return fmt.Errorf("Error fetching secret %v from taskcluster-secrets service: %v", secretName, err)
	}
	b := bytes.NewBuffer([]byte(sec.Secret))
	d := json.NewDecoder(b)
	d.DisallowUnknownFields()
	var privateHostSetup PrivateHostSetup
	err = d.Decode(&privateHostSetup)
	if err != nil {
		return fmt.Errorf("Error converting secret %v from taskcluster-secrets service into config/files: %v", secretName, err)
	}

	// Apply config from secret
	err = c.MergeInJSON(sec.Secret, func(a map[string]interface{}) map[string]interface{} {
		if config, exists := a["config"]; exists {
			return config.(map[string]interface{})
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error applying config from secret %v to generic worker config: %v", secretName, err)
	}

	return nil
}
