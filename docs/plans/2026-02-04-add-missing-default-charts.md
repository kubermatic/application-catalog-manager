# Add Missing Default Charts Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add missing charts (aikit, k8sgpt-operator, kube-vip, kubevirt, local-ai, trivy-operator) to GetDefaultCharts() so they are included in the default application catalog.

**Architecture:** Charts are explicitly defined in GetDefaultCharts() function. Each chart requires ChartName, Metadata (AppName, DisplayName, Description, URLs, Logo (base64 encoded), LogoFormat), ChartVersions, and optional DefaultValuesBlock. The function returns a slice of ChartConfig structs that are used by the mutation webhook to populate the default-catalog ApplicationCatalog.

**Tech Stack:** Go 1.23, Kubernetes API, Controller-Runtime

---

## Context

**Problem:** The TestAppCatalogMigration test shows that 7 charts are missing from the new External Application Catalog Manager:
- aikit
- k8sgpt-operator
- kube-vip
- kubevirt
- local-ai
- trivy-operator (note: there is already a "trivy" chart, different from "trivy-operator")

These charts exist in the application-catalog/ directory but are not defined in GetDefaultCharts().

**Reference Research:** `thoughts/shared/research/2026-02-04-application-catalog-procedure.md`

**Key Files:**
- `internal/pkg/defaulting/applicationcatalog.go:51` - GetDefaultCharts() function
- `internal/pkg/defaulting/applicationcatalog_test.go:549` - TestGetDefaultChartNames()
- `pkg/apis/applicationcatalog/v1alpha1/common_types.go` - ChartConfig types

---

## Task 0: Extract Logos from Kubermatic YAML Files

**Purpose:** The logo fields are base64-encoded images stored in kubermatic YAML files. They need to be converted from multi-line YAML format to single-line Go strings. This task creates and uses a helper script to automate this extraction.

**Files:**
- Create: `hack/extract-logos.sh`

**Script Overview:**

The script (`hack/extract-logos.sh`) extracts logos from kubermatic ApplicationDefinition YAML files and formats them for use in Go code.

**Prerequisites:**

- `yq` must be installed: `brew install yq` (already installed in this environment)

**Step 1: Verify the script exists**

```bash
ls -la hack/extract-logos.sh
```

Expected: The script file exists

**Step 2: Run the script to extract all logos**

```bash
cd /Users/buraksekili/projects/w2/application-catalog-manager
bash hack/extract-logos.sh
```

The script will output formatted Go code snippets for each missing chart:

```
=== aikit ===
DisplayName: AIKit
Description: AIKit is a comprehensive platform to quickly get started to host, deploy, build and fine-tune large language models (LLMs).
LogoFormat: png

Go code snippet:
                Logo:             "iVBORw0KGgo...(full base64)...",
                LogoFormat:       "png",

----------------------------------------
```

**Step 3: Copy the Logo and LogoFormat lines**

For each chart, copy the two lines:
- `Logo:             "..."`
- `LogoFormat:       "..."`

These will be pasted into the ChartMetadata struct for each chart in the following tasks.

**Usage Tips:**

1. Run the script once to get all the logos
2. Keep the terminal output available for reference
3. Copy the Logo and LogoFormat lines when implementing each chart task below
4. The script handles the conversion from multi-line YAML format to single-line Go string format
5. The script validates that source YAML files exist before extraction

**Benefits:**

- Eliminates manual copy/paste errors with huge base64 strings
- Ensures correct formatting (single-line for Go vs multi-line in YAML)
- Reusable for future chart additions
- Can be version controlled

**Note:** The LOGO field in the kubermatic YAML files uses multi-line format (`|+`), but Go requires single-line strings. The script automatically handles this conversion by removing all newlines and spaces.

---

## Task 1: Add aikit Chart

**Files:**
- Modify: `internal/pkg/defaulting/applicationcatalog.go:250` (before closing brace of GetDefaultCharts)
- Modify: `internal/pkg/defaulting/applicationcatalog_test.go:564` (update expected names)

**Chart Metadata:**
- Chart.yaml location: `../application-catalog/charts/aikit/0.18.0/Chart.yaml`
- Versions available: 0.16.0, 0.18.0
- Description: Kubernetes Helm chart to deploy AIKit LLM images
- Logo: Use output from `hack/extract-logos.sh` (Task 0)

**Step 1: Get logo from script output**

From the output of Task 0, copy the Logo and LogoFormat lines for aikit.

**Step 2: Write the failing test**

First, update the test to expect the new chart name:

```go
// In internal/pkg/defaulting/applicationcatalog_test.go:553
expectedNames := []string{
    "aikit",           // NEW
    "argo-cd",
    "cert-manager",
    "cilium",
    "cluster-autoscaler",
    "falco",
    "flux2",
    "gpu-operator",
    "ingress-nginx",
    "k8sgpt-operator",  // NEW
    "kube-vip",         // NEW
    "kubevirt",         // NEW
    "kueue",
    "local-ai",         // NEW
    "metallb",
    "trivy",
    "trivy-operator",   // NEW
}
```

**Step 2: Run test to verify it fails**

```bash
cd /Users/buraksekili/projects/w2/application-catalog-manager
go test -v ./internal/pkg/defaulting/... -run TestGetDefaultChartNames
```

Expected: FAIL - missing chart names in actual result

**Step 3: Add aikit to GetDefaultCharts()**

Add before the closing `}` of `GetDefaultCharts()` (around line 250):

```go
        {
            ChartName: "aikit",
            Metadata: &catalogv1alpha1.ChartMetadata{
                AppName:          "aikit",
                DisplayName:      "AIKit",
                Description:      "Kubernetes Helm chart to deploy AIKit LLM images",
                DocumentationURL: "https://sozercan.github.io/aikit/docs",
                SourceURL:        "https://github.com/sozercan/aikit",
                Logo:             <LOGO_OBTAINED_WITH_BASH_SCRIPT>
                LogoFormat:       "png",
            },
            ChartVersions: []catalogv1alpha1.ChartVersion{
                {ChartVersion: "0.18.0", AppVersion: "v0.18.0"},
                {ChartVersion: "0.16.0", AppVersion: "v0.16.0"},
            },
        },
```

**Step 4: Run test to verify progress**

```bash
go test -v ./internal/pkg/defaulting/... -run TestGetDefaultChartNames
```

Expected: FAIL - still missing other charts (k8sgpt-operator, kube-vip, kubevirt, local-ai, trivy-operator)

---

## Task 2: Add k8sgpt-operator Chart

**Files:**
- Modify: `internal/pkg/defaulting/applicationcatalog.go:250` (before closing brace)

**Chart Metadata:**
- Chart.yaml location: `../application-catalog/charts/k8sgpt-operator/0.2.17/Chart.yaml`
- Versions available: 0.2.17
- Description: Automatic SRE Superpowers within your Kubernetes cluster
- Source: https://github.com/k8sgpt-ai/k8sgpt-operator

**Step 1: Add k8sgpt-operator to GetDefaultCharts()**

Add after the aikit entry:

```go
        {
            ChartName: "k8sgpt-operator",
            Metadata: &catalogv1alpha1.ChartMetadata{
                AppName:          "k8sgpt-operator",
                DisplayName:      "K8sGPT Operator",
                Description:      "K8sGPT Operator is designed to enable K8sGPT within a Kubernetes cluster. It will allow you to create a custom resource that defines the behaviour and scope of a managed K8sGPT workload.",
                DocumentationURL: "https://docs.k8sgpt.ai/getting-started/in-cluster-operator/",
                SourceURL:        "https://github.com/k8sgpt-ai/k8sgpt-operator",
                Logo:             <LOGO_OBTAINED_WITH_BASH_SCRIPT>
                LogoFormat:       "png",
            },
            ChartVersions: []catalogv1alpha1.ChartVersion{
                {ChartVersion: "0.2.17", AppVersion: "0.0.26"},
            },
        },
```

**Step 2: Run test to verify progress**

```bash
go test -v ./internal/pkg/defaulting/... -run TestGetDefaultChartNames
```

Expected: FAIL - still missing kube-vip, kubevirt, local-ai, trivy-operator

---

## Task 3: Add kube-vip Chart

**Files:**
- Modify: `internal/pkg/defaulting/applicationcatalog.go:250` (before closing brace)

**Chart Metadata:**
- Chart.yaml location: `../application-catalog/charts/kube-vip/0.6.6/Chart.yaml`
- Versions available: 0.4.4, 0.6.6
- Description: kube-vip provides Kubernetes clusters with a virtual IP and load balancer

**Step 1: Add kube-vip to GetDefaultCharts()**

```go
        {
            ChartName: "kube-vip",
            Metadata: &catalogv1alpha1.ChartMetadata{
                AppName:          "kube-vip",
                DisplayName:      "kube-vip",
                Description:      "kube-vip provides Kubernetes clusters with a virtual IP and load balancer for both the control plane (for building a highly-available cluster) and Kubernetes Services of type LoadBalancer without relying on any external hardware or software.",
                DocumentationURL: "https://kube-vip.io/",
                SourceURL:        "https://github.com/kube-vip/helm-charts",
                Logo:             <LOGO_OBTAINED_WITH_BASH_SCRIPT>
                LogoFormat:       "png",
            },
            ChartVersions: []catalogv1alpha1.ChartVersion{
                {ChartVersion: "0.6.6", AppVersion: "v0.8.9"},
                {ChartVersion: "0.4.4", AppVersion: "v0.4.1"},
            },
        },
```

**Step 2: Run test to verify progress**

```bash
go test -v ./internal/pkg/defaulting/... -run TestGetDefaultChartNames
```

Expected: FAIL - still missing kubevirt, local-ai, trivy-operator

---

## Task 4: Add kubevirt Chart

**Files:**
- Modify: `internal/pkg/defaulting/applicationcatalog.go:250` (before closing brace)

**Chart Metadata:**
- Chart.yaml location: `../application-catalog/charts/kubevirt/v1.1.0/Chart.yaml`
- Versions available: v1.1.0
- Description: KubeVirt with Containerized Data Importer

**Step 1: Add kubevirt to GetDefaultCharts()**

```go
        {
            ChartName: "kubevirt",
            Metadata: &catalogv1alpha1.ChartMetadata{
                AppName:          "kubevirt",
                DisplayName:      "KubeVirt",
                Description:      "KubeVirt with Containerized Data Importer",
                DocumentationURL: "https://kubevirt.io/",
                SourceURL:        "https://github.com/kubevirt/kubevirt",
                Logo:             <LOGO_OBTAINED_WITH_BASH_SCRIPT>
                LogoFormat:       "png",
            },
            ChartVersions: []catalogv1alpha1.ChartVersion{
                {ChartVersion: "v1.1.0", AppVersion: "v1.1.0"},
            },
        },
```

**Step 2: Run test to verify progress**

```bash
go test -v ./internal/pkg/defaulting/... -run TestGetDefaultChartNames
```

Expected: FAIL - still missing local-ai, trivy-operator

---

## Task 5: Add local-ai Chart

**Files:**
- Modify: `internal/pkg/defaulting/applicationcatalog.go:250` (before closing brace)

**Chart Metadata:**
- Chart.yaml location: `../application-catalog/charts/local-ai/3.4.2/Chart.yaml`
- Versions available: 3.4.2
- Description: LocalAI is an open-source alternative to OpenAI's API, designed to run AI models on your own hardware.
- LogoFormat: svg+xml (base64 encoded SVG)

**Step 1: Add local-ai to GetDefaultCharts()**

```go
        {
            ChartName: "local-ai",
            Metadata: &catalogv1alpha1.ChartMetadata{
                AppName:          "local-ai",
                DisplayName:      "LocalAI",
                Description:      "LocalAI is an open-source alternative to OpenAI's API, designed to run AI models on your own hardware.",
                DocumentationURL: "https://localai.io/docs/overview/",
                SourceURL:        "https://github.com/mudler/LocalAI",
                Logo:             <LOGO_OBTAINED_WITH_BASH_SCRIPT>
                LogoFormat:       "svg+xml",
            },
            ChartVersions: []catalogv1alpha1.ChartVersion{
                {ChartVersion: "3.4.2", AppVersion: "2.23"},
            },
            DefaultValuesBlock: `service:
  # To Expose local-ai externally without ingress, set service type as "LoadBalancer". Default value is "ClusterIP".
  type: "ClusterIP"
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
`,
        },
```

**Step 2: Run test to verify progress**

```bash
go test -v ./internal/pkg/defaulting/... -run TestGetDefaultChartNames
```

Expected: FAIL - still missing trivy-operator

---

## Task 6: Add trivy-operator Chart

**Files:**
- Modify: `internal/pkg/defaulting/applicationcatalog.go:250` (before closing brace)

**Chart Metadata:**
- Chart.yaml location: `../application-catalog/charts/trivy-operator/0.28.0/Chart.yaml`
- Versions available: 0.15.1, 0.20.5, 0.25.0, 0.28.0
- Description: Trivy-Operator is a Kubernetes-native security toolkit

**Note:** There is already a "trivy" chart in the defaults. "trivy-operator" is a separate chart.

**Step 1: Add trivy-operator to GetDefaultCharts()**

```go
        {
            ChartName: "trivy-operator",
            Metadata: &catalogv1alpha1.ChartMetadata{
                AppName:          "trivy-operator",
                DisplayName:      "Trivy Operator",
                Description:      "Trivy-Operator is a Kubernetes-native security toolkit.",
                DocumentationURL: "https://aquasecurity.github.io/trivy-operator/",
                SourceURL:        "https://github.com/aquasecurity/trivy-operator",
                Logo:             <LOGO_OBTAINED_WITH_BASH_SCRIPT>
                LogoFormat:       "png",
            },
            ChartVersions: []catalogv1alpha1.ChartVersion{
                {ChartVersion: "0.28.0", AppVersion: "0.26.0"},
                {ChartVersion: "0.25.0", AppVersion: "0.23.0"},
                {ChartVersion: "0.20.5", AppVersion: "0.18.4"},
                {ChartVersion: "0.15.1", AppVersion: "0.15.1"},
            },
            DefaultValuesBlock: `trivy:
  # To specify that Trivy should ignore all unfixed vulnerabilities, set \`ignoredUnfixed\` flag to \`true\`
  ignoreUnfixed: true
`,
        },
```

**Step 2: Run test to verify all charts pass**

```bash
go test -v ./internal/pkg/defaulting/... -run TestGetDefaultChartNames
```

Expected: PASS - All 18 chart names present

---

## Task 7: Run All Defaulting Package Tests

**Files:**
- Test: `internal/pkg/defaulting/applicationcatalog_test.go`

**Step 1: Run all tests in defaulting package**

```bash
cd /Users/buraksekili/projects/w2/application-catalog-manager
go test -v ./internal/pkg/defaulting/...
```

Expected: PASS - All tests pass

**Step 2: Run all tests in the repository**

```bash
go test -v ./...
```

Expected: PASS - All tests pass

---

## Task 8: Run E2E Test to Verify Migration

**Files:**
- Test: `../kubermatic/pkg/test/e2e/appcatalogmanager/deployment_test.go`

**Step 1: Run the migration test**

```bash
cd /Users/buraksekili/projects/w2/kubermatic
go test -v ./pkg/test/e2e/appcatalogmanager/... -run TestAppCatalogMigration -tags=e2e
```

Expected: PASS - Old-style app count (16) should now match new-style app count (16)

**Note:** This test requires:
- Running Kubernetes cluster
- Kubermatic installed
- Application catalog manager deployed

---

## Task 9: Update ValidateIncludeAnnotation Tests

**Files:**
- Modify: `internal/pkg/defaulting/applicationcatalog_test.go`

**Step 1: Check if ValidateIncludeAnnotation tests exist**

```bash
grep -n "ValidateIncludeAnnotation" internal/pkg/defaulting/applicationcatalog_test.go
```

If tests exist, verify they pass with new chart names.

**Step 2: Run validation tests**

```bash
go test -v ./internal/pkg/defaulting/... -run ValidateIncludeAnnotation
```

Expected: PASS

---

## Task 10: Update E2E Annotation Test

**Files:**
- Modify: `tests/e2e/applicationcatalog_test.go`

**Step 1: Check if the annotation test needs updating**

The test at `tests/e2e/applicationcatalog_test.go` was added in commit 9e20d09 to test the `defaultcatalog.k8c.io/include` annotation. Verify it still works with the new chart names.

**Step 2: Run E2E tests**

```bash
go test -v ./tests/e2e/... -run TestApplicationCatalog_IncludeAnnotation
```

Expected: PASS

---

## Task 11: Verify Chart Sorting

**Files:**
- Test: `internal/pkg/defaulting/applicationcatalog_test.go`

**Step 1: Verify GetDefaultChartNames returns sorted names**

The `GetDefaultChartNames()` function should return alphabetically sorted names. Verify:

```go
func TestGetDefaultChartNamesSorted(t *testing.T) {
    names := GetDefaultChartNames()
    for i := 1; i < len(names); i++ {
        if names[i] < names[i-1] {
            t.Errorf("Names not sorted: %q comes before %q", names[i-1], names[i])
        }
    }
}
```

**Step 2: Run test**

```bash
go test -v ./internal/pkg/defaulting/... -run TestGetDefaultChartNamesSorted
```

Expected: PASS

---

## Task 12: Commit Changes

**Files:**
- Modified: `internal/pkg/defaulting/applicationcatalog.go`
- Modified: `internal/pkg/defaulting/applicationcatalog_test.go`

**Step 1: Review changes**

```bash
git diff internal/pkg/defaulting/applicationcatalog.go
git diff internal/pkg/defaulting/applicationcatalog_test.go
```

**Step 2: Stage and commit**

```bash
git add internal/pkg/defaulting/applicationcatalog.go
git add internal/pkg/defaulting/applicationcatalog_test.go
git commit -s -m "feat: add missing default charts to GetDefaultCharts()

Add aikit, k8sgpt-operator, kube-vip, kubevirt, local-ai,
and trivy-operator to the default application catalog.

This resolves the TestAppCatalogMigration failure where these
charts were missing from the new External Application Catalog
Manager."
```

---

## Summary of New Charts

| Chart Name | App Name | Versions | Description |
|------------|----------|----------|-------------|
| aikit | aikit | 0.18.0, 0.16.0 | Kubernetes Helm chart to deploy AIKit LLM images |
| k8sgpt-operator | k8sgpt-operator | 0.2.17 | K8sGPT Operator for Kubernetes cluster |
| kube-vip | kube-vip | 0.6.6, 0.4.4 | Kubernetes Virtual IP and Load Balancer |
| kubevirt | kubevirt | v1.1.0 | KubeVirt with Containerized Data Importer |
| local-ai | local-ai | 3.4.2 | OpenAI compatible API for local AI inference (SVG logo) |
| trivy-operator | trivy-operator | 0.28.0, 0.25.0, 0.20.5, 0.15.1 | Kubernetes-native security scanner |

**Total new charts:** 6
**Total default charts after change:** 18 (was 11)

---

## Verification Checklist

After implementation, verify:

- [ ] `go test ./internal/pkg/defaulting/...` passes
- [ ] `go test ./...` passes
- [ ] `gimps .` to sort the Go dependencies

---

## Post-Implementation Notes

1. **Logo Encoding**: The `Logo` field contains base64 encoded image data (not URLs). The `LogoFormat` must be either "png" or "svg+xml". These logos were copied from the existing ApplicationDefinition YAML files in `../kubermatic/pkg/ee/default-application-catalog/applicationdefinitions/`.

2. **Documentation URLs**: Verify documentation URLs are correct and accessible.

3. **Source URLs**: All source URLs should point to the official repositories.

4. **Future Charts**: To add more charts in the future, follow the same pattern:
   - Add ChartConfig entry to GetDefaultCharts()
   - Update TestGetDefaultChartNames expected names
   - Run tests to verify
