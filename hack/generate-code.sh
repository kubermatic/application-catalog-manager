#!/bin/bash

OUTPUT="hack/new_charts.go.txt"
>"$OUTPUT"

# Function to generate chart struct
generate_chart() {
	local chart_name="$1"
	local app_name="$2"
	local display_name="$3"
	local description="$4"
	local doc_url="$5"
	local source_url="$6"
	local versions="$7"
	local default_values="$8"

	local logo=$(cat "logo_${chart_name}.txt")
	local logo_format=$(cat "logoformat_${chart_name}.txt")

	echo "		{" >>"$OUTPUT"
	echo "			ChartName: \"$chart_name\"," >>"$OUTPUT"
	echo "			Metadata: &catalogv1alpha1.ChartMetadata{" >>"$OUTPUT"
	echo "				AppName:          \"$app_name\"," >>"$OUTPUT"
	echo "				DisplayName:      \"$display_name\"," >>"$OUTPUT"
	echo "				Description:      \"$description\"," >>"$OUTPUT"
	echo "				DocumentationURL: \"$doc_url\"," >>"$OUTPUT"
	echo "				SourceURL:        \"$source_url\"," >>"$OUTPUT"
	echo "				Logo:             \"$logo\"," >>"$OUTPUT"
	echo "				LogoFormat:       \"$logo_format\"," >>"$OUTPUT"
	echo "			}," >>"$OUTPUT"
	echo "			ChartVersions: []catalogv1alpha1.ChartVersion{" >>"$OUTPUT"
	echo "$versions" >>"$OUTPUT"
	echo "			}," >>"$OUTPUT"

	if [ -n "$default_values" ]; then
		echo "			DefaultValuesBlock: \`$default_values\`," >>"$OUTPUT"
	fi

	echo "		}," >>"$OUTPUT"
}

# aikit
generate_chart "aikit" "aikit" "AIKit" \
	"Kubernetes Helm chart to deploy AIKit LLM images" \
	"https://sozercan.github.io/aikit/docs" \
	"https://github.com/sozercan/aikit" \
	"				{ChartVersion: \"0.18.0\", AppVersion: \"v0.18.0\"},
				{ChartVersion: \"0.16.0\", AppVersion: \"v0.16.0\"}," \
	""

# k8sgpt-operator
generate_chart "k8sgpt-operator" "k8sgpt-operator" "K8sGPT Operator" \
	"K8sGPT Operator is designed to enable K8sGPT within a Kubernetes cluster. It will allow you to create a custom resource that defines the behaviour and scope of a managed K8sGPT workload." \
	"https://docs.k8sgpt.ai/getting-started/in-cluster-operator/" \
	"https://github.com/k8sgpt-ai/k8sgpt-operator" \
	"				{ChartVersion: \"0.2.17\", AppVersion: \"0.0.26\"}," \
	""

# kube-vip
generate_chart "kube-vip" "kube-vip" "kube-vip" \
	"kube-vip provides Kubernetes clusters with a virtual IP and load balancer for both the control plane (for building a highly-available cluster) and Kubernetes Services of type LoadBalancer without relying on any external hardware or software." \
	"https://kube-vip.io/" \
	"https://github.com/kube-vip/helm-charts" \
	"				{ChartVersion: \"0.6.6\", AppVersion: \"v0.8.9\"},
				{ChartVersion: \"0.4.4\", AppVersion: \"v0.4.1\"}," \
	""

# kubevirt
generate_chart "kubevirt" "kubevirt" "KubeVirt" \
	"KubeVirt with Containerized Data Importer" \
	"https://kubevirt.io/" \
	"https://github.com/kubevirt/kubevirt" \
	"				{ChartVersion: \"v1.1.0\", AppVersion: \"v1.1.0\"}," \
	""

# local-ai
generate_chart "local-ai" "local-ai" "LocalAI" \
	"LocalAI is an open-source alternative to OpenAI's API, designed to run AI models on your own hardware." \
	"https://localai.io/docs/overview/" \
	"https://github.com/mudler/LocalAI" \
	"				{ChartVersion: \"3.4.2\", AppVersion: \"2.23\"}," \
	"service:
  # To Expose local-ai externally without ingress, set service type as \"LoadBalancer\". Default value is \"ClusterIP\".
  type: \"ClusterIP\"
  port: 8080
persistence:
  models:
    accessModes:
    - ReadWriteOnce
    annotations: {}
    enabled: true
    globalMount: /models
    size: 30Gi
  output:
    accessModes:
    - ReadWriteOnce
    annotations: {}
    enabled: true
    globalMount: /tmp/generated
    size: 30Gi
"

# trivy-operator
generate_chart "trivy-operator" "trivy-operator" "Trivy Operator" \
	"Trivy-Operator is a Kubernetes-native security toolkit." \
	"https://aquasecurity.github.io/trivy-operator/" \
	"https://github.com/aquasecurity/trivy-operator" \
	"				{ChartVersion: \"0.28.0\", AppVersion: \"0.26.0\"},
				{ChartVersion: \"0.25.0\", AppVersion: \"0.23.0\"},
				{ChartVersion: \"0.20.5\", AppVersion: \"0.18.4\"},
				{ChartVersion: \"0.15.1\", AppVersion: \"0.15.1\"}," \
	"trivy:
  # To specify that Trivy should ignore all unfixed vulnerabilities, set \`ignoredUnfixed\` flag to \`true\`
  ignoreUnfixed: true
"
