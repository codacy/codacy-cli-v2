module trivy-example

go 1.22.3

require (
	github.com/aquasecurity/trivy v0.49.1 // MEDIUM ERROR 
	github.com/spf13/cobra v1.8.0
    github.com/sirupsen/logrus v1.4.2    
    github.com/dexidp/dex v0.0.0-20200121184102-3b39c6440888 // CRITICAL ERROR - CVE-2020-26160 - Insecure JWT implementation 
)