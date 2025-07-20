# bin/requirements

This directory contains detailed functional requirements for command-line tools in the `bin/` directory.

## Organization

Requirements are separated from implementation documentation to provide:
- **Clear separation** between requirements and usage examples
- **Focused specifications** for tool behavior and constraints  
- **Centralized location** for requirement management
- **Reference documentation** for implementation details

## Naming Conventions

The project uses **semantic naming conventions** that distinguish between different types of tools:

### **Executable Script Naming**
```bash
# Infrastructure/Utility Scripts → .sh extension
bin/bootstrap.sh           # Complex orchestration script
bin/status.sh              # CRD monitoring utility
bin/wait.kube.sh           # Resource waiting utility

# User Commands → No extension  
bin/new-cluster            # User-facing cluster creation
bin/clean-aws              # User-facing cleanup command
bin/health-check           # User-facing monitoring command
bin/generate-cluster       # User-facing generation tool
```

### **Requirements File Naming**
```bash
# Rule: Transform dots to hyphens, always drop .sh extension
bin/bootstrap.sh           → bin/requirements/bootstrap.md
bin/status.sh              → bin/requirements/status.md
bin/wait.kube.sh           → bin/requirements/wait-kube.md
bin/new-cluster            → bin/requirements/new-cluster.md
```

### **Why This Convention Works**

**1. Semantic Clarity**
- **Infrastructure scripts** (`.sh`) are called by other scripts or systems
- **User commands** (no extension) are called directly by users
- The naming immediately indicates intended usage pattern

**2. User Experience**
- Commands like `./bin/new-cluster` feel natural vs `./bin/new-cluster.sh`
- Users don't need to remember implementation details in command names
- Follows Unix convention where user commands hide implementation

**3. Technical Benefits**
- IDE/editor recognition for shell scripts (`.sh` = syntax highlighting)
- Clear distinction between utility scripts and user-facing tools
- Maintains compatibility with existing documentation and workflows

**4. Requirements Mapping**
- Predictable transformation: dots become hyphens, `.sh` drops
- Consistent file discovery: `bin/tool-name` → `bin/requirements/tool-name.md`
- Maintains semantic relationship between executable and specification

## Files

### Core Cluster Management
- **[generate-cluster.md](./generate-cluster.md)** - Requirements for `bin/generate-cluster` overlay generator
- **[convert-cluster.md](./convert-cluster.md)** - Requirements for `bin/convert-cluster` regional specification converter  
- **[new-cluster.md](./new-cluster.md)** - Requirements for `bin/new-cluster` interactive cluster generator
- **[regenerate-all-clusters.md](./regenerate-all-clusters.md)** - Requirements for `bin/regenerate-all-clusters` bulk regeneration tool

### Bootstrap and Infrastructure
- **[bootstrap.md](./bootstrap.md)** - Requirements for `bin/bootstrap.sh` GitOps bootstrap orchestrator
- **[bootstrap-vault-integration.md](./bootstrap-vault-integration.md)** - Requirements for `bin/bootstrap.vault-integration.sh` Vault automation

### AWS Resource Management
- **[find-aws-resources.md](./find-aws-resources.md)** - Requirements for `bin/find-aws-resources` discovery tool
- **[clean-aws.md](./clean-aws.md)** - Requirements for `bin/clean-aws` AWS resource cleanup tool
- **[test-find-aws-resources.md](./test-find-aws-resources.md)** - Requirements for `bin/test-find-aws-resources` validation tool

### Monitoring and Operations
- **[health-check.md](./health-check.md)** - Requirements for `bin/health-check` comprehensive status reporter
- **[status.md](./status.md)** - Requirements for `bin/status.sh` CRD establishment monitor
- **[wait-kube.md](./wait-kube.md)** - Requirements for `bin/wait.kube.sh` resource readiness monitor

### Documentation Tools
- **[generate-docs.md](./generate-docs.md)** - Requirements for `bin/generate-docs` documentation generator
- **[update-dynamic-docs.md](./update-dynamic-docs.md)** - Requirements for `bin/update-dynamic-docs` dynamic content updater
- **[validate-docs.md](./validate-docs.md)** - Requirements for `bin/validate-docs` documentation validator

## Related Documentation

Each requirements file corresponds to implementation documentation in the parent `bin/` directory:

### Core Tools
- `bin/generate-cluster.md` → References `bin/requirements/generate-cluster.md`
- `bin/convert-cluster.md` → References `bin/requirements/convert-cluster.md`
- `bin/new-cluster.md` → References `bin/requirements/new-cluster.md`
- `bin/clean-aws.md` → References `bin/requirements/clean-aws.md`
- `bin/find-aws-resources.md` → References `bin/requirements/find-aws-resources.md`

### New Requirements Added
- `bin/bootstrap.sh` → `bin/requirements/bootstrap.md`
- `bin/bootstrap.vault-integration.sh` → `bin/requirements/bootstrap-vault-integration.md`
- `bin/regenerate-all-clusters` → `bin/requirements/regenerate-all-clusters.md`
- `bin/health-check` → `bin/requirements/health-check.md`
- `bin/status.sh` → `bin/requirements/status.md`
- `bin/wait.kube.sh` → `bin/requirements/wait-kube.md`
- `bin/generate-docs` → `bin/requirements/generate-docs.md`
- `bin/update-dynamic-docs` → `bin/requirements/update-dynamic-docs.md`
- `bin/validate-docs` → `bin/requirements/validate-docs.md`
- `bin/test-find-aws-resources` → `bin/requirements/test-find-aws-resources.md`

## Tool Workflows

### Bootstrap Sequence
```
bootstrap.sh → status.sh → bootstrap.vault-integration.sh
```

### Cluster Lifecycle
```
new-cluster → generate-cluster → [deploy] → health-check
                ↓
         regenerate-all-clusters (bulk operations)
```

### AWS Management
```
find-aws-resources → [analyze] → clean-aws
        ↑
test-find-aws-resources (validation)
```

### Documentation Pipeline
```
generate-docs → update-dynamic-docs → validate-docs
```

## Dependency Mapping

### Critical Path Dependencies
```
bootstrap (core infrastructure)
    ├── status (CRD readiness)
    ├── wait-kube (resource readiness)
    └── bootstrap-vault-integration (secret management)
```

### Cluster Provisioning Dependencies
```
new-cluster (interactive creation)
    └── generate-cluster (overlay generation)
        └── regenerate-all-clusters (bulk operations)
            └── health-check (validation)
```

### AWS Resource Dependencies
```
find-aws-resources (discovery)
    ├── clean-aws (cleanup)
    └── test-find-aws-resources (validation)
```

### Documentation Dependencies
```
generate-docs (static content)
    └── update-dynamic-docs (live content)
        └── validate-docs (quality assurance)
```

### Cross-Functional Dependencies
- All cluster operations depend on `bootstrap` completion
- `health-check` provides validation for all workflows
- `status` and `wait-kube` are used by multiple tools for synchronization

## Quick Reference

### Tool Purposes
| Tool | Purpose | Category |
|------|---------|----------|
| `bootstrap` | GitOps infrastructure orchestration | Bootstrap |
| `bootstrap-vault-integration` | Vault-based secret management | Bootstrap |
| `generate-cluster` | Regional spec to overlay conversion | Cluster Mgmt |
| `convert-cluster` | Existing cluster to regional spec | Cluster Mgmt |
| `new-cluster` | Interactive cluster creation | Cluster Mgmt |
| `regenerate-all-clusters` | Bulk cluster regeneration | Cluster Mgmt |
| `find-aws-resources` | AWS resource discovery | AWS Mgmt |
| `clean-aws` | AWS resource cleanup | AWS Mgmt |
| `test-find-aws-resources` | AWS discovery validation | AWS Mgmt |
| `health-check` | Comprehensive status reporting | Operations |
| `status` | CRD establishment monitoring | Operations |
| `wait-kube` | Generic resource readiness | Operations |
| `generate-docs` | Static documentation generation | Documentation |
| `update-dynamic-docs` | Live content synchronization | Documentation |
| `validate-docs` | Documentation quality validation | Documentation |

### Common Entry Points
- **Start here**: `bootstrap` - Sets up entire infrastructure
- **Create clusters**: `new-cluster` - Interactive cluster creation
- **Check status**: `health-check` - Comprehensive monitoring
- **Clean up AWS**: `clean-aws` - Resource removal
- **Generate docs**: `generate-docs` - Documentation creation

## Interactive Tool Selection

### **"I want to..."**

**Set up a new OpenShift environment**
```bash
./bin/bootstrap.sh               # Set up hub cluster infrastructure
./bin/new-cluster               # Create your first managed cluster
./bin/health-check              # Verify everything is working
```

**Create additional clusters**
```bash
./bin/new-cluster               # Interactive cluster creation
# OR
./bin/generate-cluster regions/us-west-2/my-cluster/  # From existing spec
./bin/health-check              # Verify cluster deployment
```

**Monitor and troubleshoot**
```bash
./bin/health-check              # Comprehensive status overview
./bin/status.sh <crd-name>      # Wait for specific CRDs
./bin/wait.kube.sh <resource>   # Wait for resource readiness
```

**Clean up resources**
```bash
./bin/find-aws-resources <cluster-id>  # Discover resources to clean
./bin/clean-aws                        # Clean up AWS resources
```

**Work with documentation**
```bash
./bin/generate-docs            # Generate static documentation
./bin/update-dynamic-docs      # Update live content
./bin/validate-docs           # Validate documentation quality
```

**Bulk operations**
```bash
./bin/regenerate-all-clusters  # Regenerate all cluster overlays
./bin/convert-cluster <path>   # Convert existing to regional spec
```

### **By User Role**

**Platform Administrator** (First-time setup)
1. `bootstrap` - Initialize GitOps infrastructure
2. `bootstrap-vault-integration` - Set up secret management
3. `new-cluster` - Create first managed cluster
4. `health-check` - Verify deployment

**Cluster Operator** (Day-to-day operations)
1. `new-cluster` - Add managed clusters
2. `health-check` - Monitor cluster status
3. `find-aws-resources` → `clean-aws` - Resource lifecycle

**Developer** (Documentation and automation)
1. `generate-docs` - Create documentation
2. `validate-docs` - Quality assurance
3. `update-dynamic-docs` - Sync live content

## End-to-End Workflows

For complete scenario documentation, see **[WORKFLOWS.md](./WORKFLOWS.md)**:

- **🚀 Complete Environment Setup** - First-time bootstrap deployment
- **🏗️ Adding New Clusters** - Interactive, regional spec, and conversion workflows  
- **🔧 Maintenance and Operations** - Daily operations and troubleshooting
- **🧹 Resource Cleanup** - Planned and emergency AWS resource removal
- **📚 Documentation Workflows** - Static generation and quality assurance
- **🔄 Migration Scenarios** - Legacy to semantic naming conversion
- **⚡ Advanced Automation** - Custom scripts and CI/CD integration

## Advanced Orchestration

For workflow automation patterns, see **[ORCHESTRATION.md](./ORCHESTRATION.md)**:

- **🔄 Core Patterns** - Sequential dependencies, parallel operations, conditional workflows
- **🛠️ Advanced Features** - State management, progress reporting, resource monitoring
- **🚀 Production Examples** - Full setup orchestration, error recovery, resume functionality
- **🔧 Tool Integration** - Workflow-aware modifications and cross-tool state sharing

## Usage

Requirements files contain:
- **Functional specifications** for tool behavior
- **Command-line interface definitions**
- **Input/output requirements**
- **Implementation constraints and patterns**
- **Validation and error handling requirements**

Implementation documentation in `bin/*.md` focuses on:
- **Usage examples** and command syntax
- **Workflow integration** with other tools
- **Status and validation reporting**
- **Troubleshooting guidance**