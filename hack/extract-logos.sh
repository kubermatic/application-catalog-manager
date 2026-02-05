#!/bin/bash
# Script to extract logo and metadata from kubermatic ApplicationDefinition YAML files
# and format them for use in Go code.

set -euo pipefail

# Configuration
KUBERMATIC_DIR="../kubermatic/pkg/ee/default-application-catalog/applicationdefinitions"
APPS=("aikit" "k8sgpt-operator" "kube-vip" "kubevirt" "local-ai" "trivy-operator")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if yq is installed
if ! command -v yq &>/dev/null; then
	echo -e "${RED}Error: yq is not installed. Please install it first.${NC}"
	echo "  brew install yq"
	exit 1
fi

echo "=========================================="
echo "Extracting logos from kubermatic YAML files"
echo "=========================================="
echo ""

for app in "${APPS[@]}"; do
	yaml_file="${KUBERMATIC_DIR}/${app}-app.yaml"

	if [[ ! -f "$yaml_file" ]]; then
		echo -e "${RED}Warning: $yaml_file not found${NC}"
		continue
	fi

	echo -e "${GREEN}=== $app ===${NC}"

	# Extract displayName
	display_name=$(yq eval '.spec.displayName' "$yaml_file")
	echo "DisplayName: $display_name"

	# Extract description
	description=$(yq eval '.spec.description' "$yaml_file")
	echo "Description: $description"

	# Extract logoFormat
	logo_format=$(yq eval '.spec.logoFormat' "$yaml_file")
	echo "LogoFormat: $logo_format"

	# Extract logo and convert to single line (remove all newlines and spaces between chunks)
	logo=$(yq eval '.spec.logo' "$yaml_file" | tr -d '\n' | tr -d ' ')

	# Output in Go format for easy copy/paste
	echo ""
	echo -e "${YELLOW}Go code snippet:${NC}"
	echo "                Logo:             \"$logo\","
	echo "                LogoFormat:       \"$logo_format\","
	echo ""
	echo "----------------------------------------"
	echo ""

	# Save to file
	echo "$logo" >"logo_${app}.txt"
	echo "$logo_format" >"logoformat_${app}.txt"
done

echo ""
echo -e "${GREEN}Done! Copy the Logo and LogoFormat lines above into your Go code.${NC}"
echo ""
echo "Usage tips:"
echo "  1. Run this script from the application-catalog-manager directory"
echo "  2. Copy the Logo and LogoFormat lines for each app"
echo "  3. Paste them into the ChartMetadata struct in applicationcatalog.go"
echo ""
