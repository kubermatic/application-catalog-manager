# Second-Day Operations for Application Catalog Manager

## Problem Statement

Currently, the Application Catalog Manager webhook injects default applications into ApplicationCatalog CRDs when `spec.helm.charts` is `nil`. This works well for initial setup but creates challenges for second-day operations:

1. **New applications:** When KKP adds support for a new application to `GetDefaultCharts()`, existing catalogs don't receive the update
2. **Deprecated versions:** When KKP deprecates an application version, existing catalogs retain the old version
3. **Manual updates:** Customers must manually update their catalogs to receive changes

## Design Goals

1. Enable automatic updates for catalogs that want them
2. Allow customers to opt-in to curated subsets of defaults
3. Support custom overrides (e.g., custom Helm registry)
4. Maintain backward compatibility with existing catalogs
5. Minimal code changes and complexity
6. Clear separation of customer vs KKP-managed resources

## Proposed Solution

### Overview

Introduce two new mechanisms:

1. **Annotation:** `defaultcatalog.k8c.io/include` - filters which defaults to include
2. **Spec field:** `spec.includeDefaults` - enables automatic synchronization with defaults

The webhook merges filtered defaults with user-defined charts, allowing for flexible configurations while maintaining control.

### API Changes

#### ApplicationCatalog CRD

Add new field to `spec`:

```go
type HelmSpec struct {
    // Existing fields
    Charts []ChartConfig `json:"charts,omitempty"`

    // IncludeDefaults indicates that the webhook should automatically
    // keep this catalog in sync with the default application catalog.
    // When true, the webhook will merge defaults on every UPDATE operation,
    // not just when charts is nil.
    //
    // Defaults to false (current behavior: one-time injection).
    IncludeDefaults bool `json:"includeDefaults,omitempty"`
}
```

#### KubermaticConfiguration

Add new section for controlling KKP's default catalog:

```go
type ApplicationCatalogSettings struct {
    // Default controls the KKP-managed default ApplicationCatalog.
    Default *DefaultApplicationCatalogSettings `json:"default,omitempty"`
}

type DefaultApplicationCatalogSettings struct {
    // Disable prevents creation of the default ApplicationCatalog.
    // When true, no default catalog is created by KKP.
    Disable bool `json:"disable,omitempty"`

    // Include is a list of chart names to include from the default catalog.
    // If empty, all default applications are included.
    // This allows administrators to curate which defaults are available.
    Include []string `json:"include,omitempty"`
}
```

#### Annotation

`defaultcatalog.k8c.io/include`: Comma-separated list of chart names to include from defaults.

Example:
```yaml
metadata:
  annotations:
    defaultcatalog.k8c.io/include: "nginx-ingress,cert-manager,argo-cd"
```

## Webhook Behavior

### Updated Logic

```go
func DefaultApplicationCatalog(catalog *catalogv1alpha1.ApplicationCatalog) {
    // Current behavior: one-time injection if includeDefaults is false
    if !catalog.Spec.IncludeDefaults {
        if catalog.Spec.Helm.Charts == nil {
            catalog.Spec.Helm.Charts = GetDefaultCharts()
            sortCharts(catalog.Spec.Helm.Charts)
        }
        return
    }

    // New behavior: automatic synchronization
    // Get all defaults
    allDefaults := GetDefaultCharts()

    // Filter by include annotation if present
    if includeAnnotation := catalog.Annotations["defaultcatalog.k8c.io/include"]; includeAnnotation != "" {
        includeList := strings.Split(includeAnnotation, ",")
        for i := range includeList {
            includeList[i] = strings.TrimSpace(includeList[i])
        }
        allDefaults = filterDefaultsByName(allDefaults, includeList)
    }

    // Sort defaults
    sortCharts(allDefaults)

    // Merge defaults with user's charts (user's charts take precedence)
    catalog.Spec.Helm.Charts = mergeCharts(catalog.Spec.Helm.Charts, allDefaults)
}

func mergeCharts(userCharts, defaults []catalogv1alpha1.ChartConfig) []catalogv1alpha1.ChartConfig {
    // Use map for easy lookup and override
    result := make(map[string]catalogv1alpha1.ChartConfig)

    // Add defaults first (lower priority)
    for _, chart := range defaults {
        result[chart.ChartName] = chart
    }

    // Override with user's charts (higher priority)
    for _, chart := range userCharts {
        result[chart.ChartName] = chart
    }

    // Convert back to sorted slice
    resultSlice := make([]catalogv1alpha1.ChartConfig, 0, len(result))
    for _, chart := range result {
        resultSlice = append(resultSlice, chart)
    }
    sort.Slice(resultSlice, func(i, j int) bool {
        return resultSlice[i].ChartName < resultSlice[j].ChartName
    })

    return resultSlice
}

func filterDefaultsByName(charts []catalogv1alpha1.ChartConfig, includeList []string) []catalogv1alpha1.ChartConfig {
    includeSet := make(map[string]struct{})
    for _, name := range includeList {
        includeSet[name] = struct{}{}
    }

    var filtered []catalogv1alpha1.ChartConfig
    for _, chart := range charts {
        if _, included := includeSet[chart.ChartName]; included {
            filtered = append(filtered, chart)
        }
    }
    return filtered
}
```

## User Scenarios

### Scenario 1: Curated Subset with Override

**User wants:** nginx-ingress, cert-manager, argo-cd only, with cert-manager from custom registry.

**Solution:**

```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: my-curated-catalog
  annotations:
    defaultcatalog.k8c.io/include: "nginx-ingress,cert-manager,argo-cd"
spec:
  includeDefaults: true
  helm:
    charts:
      - chartName: cert-manager
        spec:
          repositoryURL: "https://my-registry.com/charts"
```

**Result:**
- nginx-ingress (from defaults)
- cert-manager (user's registry override)
- argo-cd (from defaults)
- Other defaults (cilium, metallb, etc.) excluded by annotation

### Scenario 2: Automatic Updates from KKP

**User wants:** Stay in sync with all KKP defaults automatically.

**Solution:**

```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: my-synced-catalog
spec:
  includeDefaults: true
  helm: {}
```

**Result:**
- Webhook keeps catalog in sync with `GetDefaultCharts()`
- New applications automatically added
- Deprecated versions automatically updated
- No manual intervention needed

### Scenario 3: Defaults + Custom Apps

**User wants:** KKP defaults plus their own application-abc.

**Solution:** Create two separate catalogs.

**Catalog 1 (defaults):**
```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: my-defaults-catalog
spec:
  includeDefaults: true
  helm: {}
```

**Catalog 2 (custom):**
```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: my-custom-catalog
spec:
  helm:
    charts:
      - chartName: application-abc
        spec:
          repositoryURL: "https://my-company.com/charts"
```

**Result:**
- Two independent catalogs
- Conflict detection prevents appName conflicts
- First catalog gets automatic updates
- Second catalog contains only custom apps
- Users see apps from both catalogs

### Scenario 4: Fully Custom, No Defaults

**User wants:** Only their own applications, no KKP defaults.

**Solution A (customer level):**
```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: my-custom-catalog
spec:
  includeDefaults: false  # or omit, defaults to false
  helm:
    charts:
      - chartName: my-app-1
      - chartName: my-app-2
```

**Solution B (KKP level):**
```yaml
apiVersion: kubermatic.k8c.io/v1
kind: KubermaticConfiguration
metadata:
  name: kubermatic
spec:
  applicationCatalog:
    default:
      disable: true  # No default catalog created
```

### Scenario 5: Enterprise Filtered Defaults

**Admin wants:** Only "blessed" applications available in KKP default catalog.

**Solution:**

```yaml
apiVersion: kubermatic.k8c.io/v1
kind: KubermaticConfiguration
metadata:
  name: kubermatic
spec:
  applicationCatalog:
    default:
      disable: false
      include:
        - nginx-ingress
        - cert-manager
        - cilium
        # Only approved applications
```

KKP operator creates `kkp-defaults` ApplicationCatalog:
```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: kkp-defaults
  namespace: kubermatic
  annotations:
    defaultcatalog.k8c.io/include: "nginx-ingress,cert-manager,cilium"
spec:
  includeDefaults: true
  helm: {}
```

**Result:**
- kkp-defaults contains only filtered applications
- All users in this KKP installation see only approved defaults
- Centralized control for enterprise environments

## Implementation Steps

1. **Update ApplicationCatalog CRD**
   - Add `IncludeDefaults` field to `HelmSpec`
   - Generate CRD manifests

2. **Update webhook logic**
   - Modify `DefaultApplicationCatalog()` in `internal/pkg/defaulting/applicationcatalog.go`
   - Add `mergeCharts()` function
   - Add `filterDefaultsByName()` function
   - Update sorting logic

3. **Update KKP operator**
   - Add `DefaultApplicationCatalogSettings` to KubermaticConfiguration schema
   - Create `kkp-defaults` ApplicationCatalog based on configuration
   - Set annotations based on `include` list

4. **Testing**
   - Unit tests for merge and filter logic
   - Integration tests for webhook mutations
   - Test all user scenarios
   - Backward compatibility tests (existing catalogs)

5. **Documentation**
   - Update user documentation
   - Provide examples for each scenario
   - Document migration path (if needed)

## Backward Compatibility

- Existing catalogs without `includeDefaults: true` maintain current behavior
- One-time injection when `charts` is `nil`
- No automatic updates unless explicitly opted in
- No breaking changes to existing catalogs

## Future Enhancements

1. **Version tracking:** Add annotation to track which version of defaults was applied
2. **Diff view:** Show what changed between default versions
3. **Exclusion list:** Allow `exclude` annotation as alternative to `include`
4. **Per-version filtering:** Filter not just by chart name but by version ranges
5. **Validation warnings:** Warn customers when using deprecated versions

## Summary

This design provides:
- Automatic updates for catalogs that want them
- Flexible filtering for curated subsets
- Support for custom overrides
- Clear separation of concerns (KKP vs customer resources)
- Backward compatibility with existing catalogs
- Minimal code complexity

The solution follows KKP's existing patterns (allowlist filtering via KubermaticConfiguration) and leverages the existing conflict detection mechanism to prevent appName conflicts across catalogs.
