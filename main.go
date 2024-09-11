package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
)

// Function to create a token (from previous implementation)
func getToken(userEmail, apiKey string) (string, error) {
	url := "https://ssoapi-ng.platform.verimatrixcloud.net/v1/token"

	payload := map[string]string{
		"userEmail": userEmail,
		"apiKey":    apiKey,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error creating request payload: %v", err)
	}
	requestBody := bytes.NewBuffer(payloadBytes)

	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading HTTP response: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing JSON response: %v", err)
	}

	if token, ok := result["token"].(string); ok {
		return token, nil
	}

	return "", fmt.Errorf("token not found in the response")
}

// Function to read a file and return its content as a single line
func readFileAsSingleLine(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("could not read file: %v", err)
	}
	return string(data), nil
}

// Function to read a file and return its Base64 encoded content
func readFileAsBase64(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("could not read file: %v", err)
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// Function to create a new application and return its ID
func createApplication(token, applicationPackageId, os, applicationName, subscriptionType string, permissionDelete, permissionUpload, permissionPrivate bool, certificate, certificateFileName, icon, iconMimeType string) (string, error) {
	url := "https://aps-api.appshield.verimatrixcloud.net/applications"

	payload := map[string]interface{}{
		"applicationPackageId": applicationPackageId,
		"os":                   os,
		"applicationName":      applicationName,
		"subscriptionType":     subscriptionType,
		"permissionDelete":     permissionDelete,
		"permissionUpload":     permissionUpload,
		"permissionPrivate":    permissionPrivate,
		"certificate":          certificate,
		"certificateFileName":  certificateFileName,
		"icon":                 icon,
		"iconMimeType":         iconMimeType,
		"additionalProp1":      map[string]interface{}{},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error creating request payload: %v", err)
	}
	requestBody := bytes.NewBuffer(payloadBytes)

	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading HTTP response: %v", err)
	}

	fmt.Println("Create Application Response:", string(body))

	// Parse the response and extract the applicationId
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing JSON response: %v", err)
	}

	if applicationId, ok := result["id"].(string); ok {
		return applicationId, nil
	}

	return "", fmt.Errorf("applicationId not found in the response")
}

// New function to create a build using the applicationId
func createBuild(token, applicationId, subscriptionType string) (string, error) {
	url := "https://aps-api.appshield.verimatrixcloud.net/builds"

	// Create the request payload
	payload := map[string]interface{}{
		"applicationId":    applicationId,
		"subscriptionType": subscriptionType,
		"additionalProp1":  map[string]interface{}{},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error creating request payload: %v", err)
	}
	requestBody := bytes.NewBuffer(payloadBytes)

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading HTTP response: %v", err)
	}

	fmt.Println("Build Response:", string(body))

	// Parse the response and extract the buildId
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing JSON response: %v", err)
	}

	if buildId, ok := result["id"].(string); ok {
		return buildId, nil
	}

	return "", fmt.Errorf("buildId not found in the response")
}

func updateBuildMetadata(token, buildId, os, androidManifest string) (string, error) {
	url := fmt.Sprintf("https://aps-api.appshield.verimatrixcloud.net/builds/%s/metadata", buildId)

	// Create the request payload
	payload := map[string]interface{}{
		"os": os,
		"osData": map[string]interface{}{
			"androidManifest": androidManifest,
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error creating request payload: %v", err)
	}
	requestBody := bytes.NewBuffer(payloadBytes)

	// Create the HTTP request
	req, err := http.NewRequest("PUT", url, requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading HTTP response: %v", err)
	}

	fmt.Println("Update Metadata Response:", string(body))

	return "", nil
}

func getBuildURL(token, buildId, fileName string) (string, error) {
	url := fmt.Sprintf("https://aps-api.appshield.verimatrixcloud.net/builds/%s/url?url=raw&uploadname=%s", buildId, fileName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func uploadFile(url string, filePath string) (string, error) {
	cmd := exec.Command("curl", "-X", "PUT", "-T", filePath, url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to execute command: %s\n", err)
	}
	/*
		file, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer file.Close()
		fmt.Println(url)
		req, err := http.NewRequest("PUT", url, file)
		if err != nil {
			return "", err
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusContinue {
			// Handle 100 Continue if needed
			// For example, you might want to log it or take some other action
			log.Println("Received 100 Continue status")
		} else if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to upload file: %s", resp.Status)
		}
	*/
	return string(output), nil
}

func patchRequest(token, cmd, buildId string) (string, error) {
	url := fmt.Sprintf("https://aps-api.appshield.verimatrixcloud.net/builds/%s?cmd=%s", buildId, cmd)
	req, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to patch request: %s", resp.Status)
	}

	return "", nil
}

func main() {
	// Token creation arguments
	userEmail := flag.String("user", "", "User email")
	apiKey := flag.String("api-key", "", "API key")

	// Application creation arguments
	applicationPackageId := flag.String("applicationPackageId", "", "Application package ID")
	os := flag.String("os", "android", "Operating system (e.g., android)")
	applicationName := flag.String("applicationName", "", "Application name (optional, defaults to applicationPackageId)")
	subscriptionType := flag.String("subscriptionType", "XTD_PLATFORM", "Subscription type")
	permissionDelete := flag.Bool("permissionDelete", true, "Permission to delete")
	permissionUpload := flag.Bool("permissionUpload", true, "Permission to upload")
	permissionPrivate := flag.Bool("permissionPrivate", false, "Permission to set private")
	certificateFile := flag.String("certificate", "", "Path to the certificate file (PEM format)")
	iconFile := flag.String("icon", "", "Path to the icon file")
	iconMimeType := flag.String("iconMimeType", "image/png", "MIME type of the icon file")
	androidManifestFile := flag.String("android-manifest", "", "Path to the AndroidManifest file")
	appFile := flag.String("appFile", "", "Path to the APK/IPA file")

	flag.Parse()

	// Validate token creation arguments
	if *userEmail == "" || *apiKey == "" {
		log.Fatal("Both --user and --api-key must be provided")
	}

	// Generate the token
	token, err := getToken(*userEmail, *apiKey)
	if err != nil {
		log.Fatalf("Error generating token: %v", err)
	}

	// Validate application creation arguments
	if *applicationPackageId == "" {
		log.Fatal("Application package ID must be provided")
	}

	// Use applicationPackageId as default applicationName if not provided
	if *applicationName == "" {
		*applicationName = *applicationPackageId
	}

	// Read certificate file content as a single line
	certificate, err := readFileAsSingleLine(*certificateFile)
	if err != nil {
		log.Fatalf("Error reading certificate file: %v", err)
	}

	manifest, err := readFileAsBase64(*androidManifestFile)
	if err != nil {
		log.Fatalf("Error reading certificate file: %v", err)
	}
	// Read icon file content as Base64 encoded
	icon, err := readFileAsBase64(*iconFile)
	if err != nil {
		log.Fatalf("Error reading icon file: %v", err)
	}

	// Extract certificate filename from file path
	certificateFileName := filepath.Base(*certificateFile)
	fileAppName := filepath.Base(*appFile)
	// Create the application and get the applicationId
	applicationId, err := createApplication(token, *applicationPackageId, *os, *applicationName, *subscriptionType, *permissionDelete, *permissionUpload, *permissionPrivate, certificate, certificateFileName, icon, *iconMimeType)
	if err != nil {
		log.Fatalf("Error creating application: %v", err)
	}

	// Create a build using the applicationId
	buildId, err := createBuild(token, applicationId, *subscriptionType)
	if err != nil {
		log.Fatalf("Error creating build: %v", err)
	}

	// Create a build using the applicationId
	buildManifest, err := updateBuildMetadata(token, buildId, *os, manifest)
	if err != nil {
		log.Fatalf("Error creating build: %v", err)
	}
	fmt.Printf(buildManifest)

	BuildURL, err := getBuildURL(token, buildId, fileAppName)
	if err != nil {
		log.Fatalf("Error creating build: %v", err)
	}
	fmt.Println("Build URL : ", BuildURL)

	uploadApp, err := uploadFile(BuildURL, *appFile)
	if err != nil {
		log.Fatalf("Error creating build: %v", err)
	}
	fmt.Printf(uploadApp)

	patchAPP, err := patchRequest(token, "upload-success", buildId)
	if err != nil {
		log.Fatalf("Error creating build: %v", err)
	}
	fmt.Printf(patchAPP)

	protectApp, err := patchRequest(token, "protect", buildId)
	if err != nil {
		log.Fatalf("Error creating build: %v", err)
	}
	fmt.Printf(protectApp)
}
