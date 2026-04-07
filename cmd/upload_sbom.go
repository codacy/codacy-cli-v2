package cmd

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"codacy/cli-v2/utils/logger"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	sbomAPIToken  string
	sbomProvider  string
	sbomOrg       string
	sbomImageName string
	sbomTag       string
	sbomRepoName  string
	sbomEnv       string
	sbomFormat string
)

func init() {
	uploadSBOMCmd.Flags().StringVarP(&sbomAPIToken, "api-token", "a", "", "API token for Codacy API (required)")
	uploadSBOMCmd.Flags().StringVarP(&sbomProvider, "provider", "p", "", "Git provider (gh, gl, bb) (required)")
	uploadSBOMCmd.Flags().StringVarP(&sbomOrg, "organization", "o", "", "Organization name on the Git provider (required)")
	uploadSBOMCmd.Flags().StringVarP(&sbomTag, "tag", "t", "", "Docker image tag (defaults to image tag or 'latest')")
	uploadSBOMCmd.Flags().StringVarP(&sbomRepoName, "repository", "r", "", "Repository name (optional)")
	uploadSBOMCmd.Flags().StringVarP(&sbomEnv, "environment", "e", "", "Environment where the image is deployed (optional)")
	uploadSBOMCmd.Flags().StringVar(&sbomFormat, "format", "cyclonedx", "SBOM format: cyclonedx or spdx-json (default cyclonedx, smaller output)")

	uploadSBOMCmd.MarkFlagRequired("api-token")
	uploadSBOMCmd.MarkFlagRequired("provider")
	uploadSBOMCmd.MarkFlagRequired("organization")

	rootCmd.AddCommand(uploadSBOMCmd)
}

var uploadSBOMCmd = &cobra.Command{
	Use:   "upload-sbom <IMAGE_NAME>",
	Short: "Generate and upload an SBOM for a Docker image to Codacy",
	Long: `Generate an SBOM (Software Bill of Materials) for a Docker image using Trivy
and upload it to Codacy for vulnerability tracking.

By default, Trivy generates a CycloneDX SBOM (smaller output). Use --format
to switch to spdx-json if needed. Both formats are accepted by the Codacy API.`,
	Example: `  # Generate and upload SBOM
  codacy-cli upload-sbom -a <api-token> -p gh -o my-org -r my-repo myapp:latest

  # Use SPDX format instead
  codacy-cli upload-sbom -a <api-token> -p gh -o my-org -r my-repo --format spdx-json myapp:v1.0.0`,
	Args: cobra.ExactArgs(1),
	Run:  runUploadSBOM,
}

func runUploadSBOM(_ *cobra.Command, args []string) {
	exitCode := executeUploadSBOM(args[0])
	exitFunc(exitCode)
}

// executeUploadSBOM generates (or reads) an SBOM and uploads it to Codacy. Returns exit code.
func executeUploadSBOM(imageRef string) int {
	if err := validateImageName(imageRef); err != nil {
		logger.Error("Invalid image name", logrus.Fields{"image": imageRef, "error": err.Error()})
		color.Red("Error: %v", err)
		return 2
	}

	if sbomFormat != "cyclonedx" && sbomFormat != "spdx-json" {
		color.Red("Error: --format must be 'cyclonedx' or 'spdx-json'")
		return 2
	}

	imageName, tag := parseImageRef(imageRef)
	if sbomTag != "" {
		tag = sbomTag
	}
	sbomImageName = imageName

	logger.Info("Starting SBOM upload", logrus.Fields{
		"image":    imageRef,
		"provider": sbomProvider,
		"org":      sbomOrg,
	})

	// Generate SBOM with Trivy
	trivyPath, err := getTrivyPath()
	if err != nil {
		handleTrivyNotFound(err)
		return 2
	}

	tmpFile, err := os.CreateTemp("", "codacy-sbom-*")
	if err != nil {
		logger.Error("Failed to create temp file", logrus.Fields{"error": err.Error()})
		color.Red("Error: Failed to create temporary file: %v", err)
		return 2
	}
	tmpFile.Close()
	sbomPath := tmpFile.Name()
	defer os.Remove(sbomPath)

	fmt.Printf("Generating SBOM for image: %s\n", imageRef)
	args := []string{"image", "--format", sbomFormat, "-o", sbomPath, imageRef}
	logger.Info("Running Trivy SBOM generation", logrus.Fields{"command": fmt.Sprintf("%s %v", trivyPath, args)})

	var stderrBuf bytes.Buffer
	if err := commandRunner.RunWithStderr(trivyPath, args, &stderrBuf); err != nil {
		if isScanFailure(stderrBuf.Bytes()) {
			color.Red("Error: Failed to generate SBOM (image not found or no container runtime)")
		} else {
			color.Red("Error: Failed to generate SBOM: %v", err)
		}
		logger.Error("Trivy SBOM generation failed", logrus.Fields{"error": err.Error()})
		return 2
	}
	fmt.Println("SBOM generated successfully")

	// Upload SBOM to Codacy
	fmt.Printf("Uploading SBOM to Codacy (org: %s/%s)...\n", sbomProvider, sbomOrg)
	if err := uploadSBOMToCodacy(sbomPath, sbomImageName, tag); err != nil {
		logger.Error("Failed to upload SBOM", logrus.Fields{"error": err.Error()})
		color.Red("Error: Failed to upload SBOM: %v", err)
		return 1
	}

	color.Green("Successfully uploaded SBOM for %s:%s", sbomImageName, tag)
	return 0
}

// parseImageRef splits an image reference into name and tag.
// e.g. "myapp:v1.0.0" -> ("myapp", "v1.0.0"), "myapp" -> ("myapp", "latest")
func parseImageRef(imageRef string) (string, string) {
	// Handle digest references (image@sha256:...)
	if idx := strings.Index(imageRef, "@"); idx != -1 {
		return imageRef[:idx], imageRef[idx+1:]
	}

	// Find the last colon that is part of the tag (not the registry port)
	lastSlash := strings.LastIndex(imageRef, "/")
	tagPart := imageRef
	if lastSlash != -1 {
		tagPart = imageRef[lastSlash:]
	}

	if idx := strings.LastIndex(tagPart, ":"); idx != -1 {
		absIdx := idx
		if lastSlash != -1 {
			absIdx = lastSlash + idx
		}
		return imageRef[:absIdx], imageRef[absIdx+1:]
	}

	return imageRef, "latest"
}

func uploadSBOMToCodacy(sbomPath, imageName, tag string) error {
	url := fmt.Sprintf("https://app.codacy.com/api/v3/organizations/%s/%s/image-sboms",
		sbomProvider, sbomOrg)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the SBOM file
	sbomFile, err := os.Open(sbomPath)
	if err != nil {
		return fmt.Errorf("failed to open SBOM file: %w", err)
	}
	defer sbomFile.Close()

	part, err := writer.CreateFormFile("sbom", filepath.Base(sbomPath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, sbomFile); err != nil {
		return fmt.Errorf("failed to write SBOM to form: %w", err)
	}

	// Add required fields
	writer.WriteField("imageName", imageName)
	writer.WriteField("tag", tag)

	// Add optional fields
	if sbomRepoName != "" {
		writer.WriteField("repositoryName", sbomRepoName)
	}
	if sbomEnv != "" {
		writer.WriteField("environment", sbomEnv)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("api-token", sbomAPIToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
