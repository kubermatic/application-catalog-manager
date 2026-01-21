# Second-Day Operations Guide

## Overview

The Application Catalog Manager webhook provides automatic synchronization with KKP's default application catalog. This guide explains how to keep your ApplicationCatalog resources up-to-date with KKP-managed applications.

## What This Does

When KKP adds new applications or updates existing ones in the default catalog, your ApplicationCatalog resources can automatically receive these changes. The webhook merges KKP's defaults with your custom applications on every UPDATE operation.

## Quick Start

### Stay in Sync with All KKP Defaults

Create an ApplicationCatalog with `includeDefaults: true`:

```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: my-synced-catalog
spec:
  helm:
    includeDefaults: true
```

The webhook automatically keeps this catalog synchronized with all KKP default applications.

## How It Works

### The includeDefaults Field

When `spec.helm.includeDefaults` is set to `true`, the webhook automatically merges KKP's default applications into your catalog on every UPDATE operation.

**Important:** You must explicitly opt-in. When `includeDefaults` is `false` or omitted, the webhook leaves your catalog completely unchanged.

### Webhook Behavior

The webhook processes UPDATE operations as follows:

1. Check if `includeDefaults` is `true`
2. If `false` or omitted, return immediately (no changes)
3. Get all KKP default applications
4. Filter by annotation if present
5. Sort defaults alphabetically
6. Merge with your charts (your charts take precedence)
7. Update the catalog spec

This happens on every UPDATE operation to the ApplicationCatalog resource.

### Include Annotation

The `defaultcatalog.k8c.io/include` annotation filters which default applications to include. Use comma-separated chart names.

```yaml
metadata:
  annotations:
    defaultcatalog.k8c.io/include: "ingress-nginx,cert-manager,argo-cd"
```

**Note:** Whitespace around chart names is automatically trimmed. Invalid chart names are silently ignored.

### Merge Behavior

When both defaults and your charts exist, your charts take precedence. The webhook:

1. Adds all filtered defaults to a map by chart name
2. Overrides with your charts (same chart name)
3. Sorts results alphabetically

**Important:** Your chart completely replaces the default chart with the same name. This is a struct-level replacement, not a field-level merge.

## Usage Examples

### Example 1: Curated Subset

You want only specific applications from KKP defaults (ingress-nginx, cert-manager, argo-cd).

```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: curated-catalog
  annotations:
    defaultcatalog.k8c.io/include: "ingress-nginx,cert-manager,argo-cd"
spec:
  helm:
    includeDefaults: true
```

**Result:** Only ingress-nginx, cert-manager, and argo-cd. Other KKP defaults excluded. Automatically stays in sync.

### Example 2: Override Default Registry

You want KKP defaults but use your own Helm registry for cert-manager.

```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: custom-registry-catalog
spec:
  helm:
    includeDefaults: true
    charts:
      - chartName: cert-manager
        repositorySettings:
          baseURL: "https://my-registry.com/charts"
```

**Result:** All KKP defaults included, but cert-manager uses your custom registry. Your override persists through updates.

### Example 3: Defaults Plus Custom Applications

You want KKP defaults plus your own custom applications.

```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: extended-catalog
spec:
  helm:
    includeDefaults: true
    charts:
      - chartName: my-custom-app-1
      - chartName: my-custom-app-2
```

**Result:** All KKP defaults plus my-custom-app-1 and my-custom-app-2.

### Example 4: Fully Custom Catalog

You want only your own applications, no KKP defaults.

**Option A: Omit includeDefaults**
```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: custom-only-catalog
spec:
  helm:
    charts:
      - chartName: my-app-1
      - chartName: my-app-2
```

**Option B: Explicitly disable**
```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: custom-only-catalog
spec:
  helm:
    includeDefaults: false
    charts:
      - chartName: my-app-1
      - chartName: my-app-2
```

**Result:** Only your applications. Webhook leaves catalog unchanged.

### Example 5: Separate Catalogs (Recommended)

For clarity, create separate catalogs for defaults and custom applications.

**Catalog 1: Defaults**
```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: kkp-defaults
spec:
  helm:
    includeDefaults: true
```

**Catalog 2: Custom**
```yaml
apiVersion: applicationcatalog.k8c.io/v1alpha1
kind: ApplicationCatalog
metadata:
  name: company-apps
spec:
  helm:
    charts:
      - chartName: application-abc
        repositorySettings:
          baseURL: "https://my-company.com/charts"
```

**Result:** Two independent catalogs. Users see applications from both.

## Chart States

The webhook handles different chart states:

### Charts is nil

```yaml
spec:
  helm:
    includeDefaults: true
    # charts: nil (omitted)
```

**Result:** All filtered defaults injected.

### Charts is empty array

```yaml
spec:
  helm:
    includeDefaults: true
    charts: []
```

**Result:** All filtered defaults injected.

### Charts has entries

```yaml
spec:
  helm:
    includeDefaults: true
    charts:
      - chartName: my-app
```

**Result:** my-app plus all filtered defaults.

## Behavior Matrix

| includeDefaults | charts value | annotation | Result |
|----------------|--------------|------------|--------|
| false/omitted | nil | any | No changes, webhook returns immediately |
| false/omitted | [] | any | Empty array preserved |
| false/omitted | [custom] | any | Custom charts only, unchanged |
| true | nil | omitted | All defaults injected |
| true | [] | omitted | All defaults injected |
| true | nil | "app1,app2" | Only app1, app2 from defaults |
| true | [] | "app1,app2" | Only app1, app2 from defaults |
| true | [custom] | omitted | All defaults plus custom |
| true | [default-app] | omitted | Default app (your chart overrides default) |
| true | [default-app with override] | omitted | Your override wins |
| true | [custom] | "app1,app2" | Filtered defaults plus custom |

## Annotation Details

### Empty Annotation

An empty annotation is treated as "no filter" (all defaults included):

```yaml
metadata:
  annotations:
    defaultcatalog.k8c.io/include: ""
```

**Result:** Same as omitting the annotation.

### Non-Existent Applications

Invalid chart names in the annotation are silently ignored:

```yaml
metadata:
  annotations:
    defaultcatalog.k8c.io/include: "ingress-nginx,non-existent,cert-manager"
```

**Result:** Only ingress-nginx and cert-manager included. No error.

### Whitespace Handling

Whitespace around chart names is automatically trimmed:

```yaml
metadata:
  annotations:
    defaultcatalog.k8c.io/include: "ingress-nginx , cert-manager , argo-cd"
```

Treated as: "ingress-nginx,cert-manager,argo-cd"

## Important Notes

### Webhook Triggering

The webhook only runs on CREATE and UPDATE operations. It does not continuously monitor for changes to `GetDefaultCharts()`. If the default application list changes in the code, existing catalogs are updated on their next UPDATE operation.

### No Automatic Creation

Existing catalogs do not automatically receive defaults. You must add `includeDefaults: true` to opt-in.

### Complete Struct Replacement

When you override a default application, your entire ChartConfig struct replaces the default. You don't get a field-level merge.

**Example:** If the default has metadata, versions, and values, and you only specify a custom baseURL, you lose the default metadata unless you also include it.

### Annotation Ignored Without Opt-In

The `defaultcatalog.k8c.io/include` annotation is ignored unless `includeDefaults: true` is set.

## Best Practices

1. **Use separate catalogs** for defaults and custom applications for clarity
2. **Override specific defaults** instead of copying all charts
3. **Use annotations** to curate which defaults you need
4. **Set includeDefaults: true** only on catalogs that should auto-update
5. **Keep custom-only catalogs** without `includeDefaults` to avoid confusion
6. **Always test** your catalog configuration in a non-production environment first

## Troubleshooting

### Expected Defaults Not Appearing

**Check:**
- `includeDefaults` is `true`
- Annotation doesn't exclude the application
- Application exists in KKP's defaults
- Catalog was updated after adding `includeDefaults`

### Custom Override Not Persisting

**Ensure:**
- Your chart is in `spec.helm.charts`
- Chart name matches the default exactly (case-sensitive)
- `includeDefaults` is `true`

**Remember:** Your chart completely replaces the default. Include all fields you need.

### Annotation Not Working

**Verify:**
- Annotation key is exactly `defaultcatalog.k8c.io/include`
- Chart names are spelled correctly
- Chart names are separated by commas
- `includeDefaults` is `true`

### Changes Not Taking Effect

**Remember:** The webhook only runs on CREATE and UPDATE operations. To trigger an update, you can:

```bash
kubectl patch applicationcatalog my-catalog --type merge -p '{"metadata":{"annotations":{"triggered-update":"'$(date +%s)"'}}}'
```

## Summary

- `includeDefaults: true` enables automatic synchronization on UPDATE
- Annotation `defaultcatalog.k8c.io/include` filters defaults
- Your charts override defaults by name (complete replacement)
- Separate catalogs for defaults and custom apps recommended
- Explicit opt-in required
- Webhook only runs on CREATE and UPDATE operations
