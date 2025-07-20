# bin/generate-docs Requirements

## Purpose

The `generate-docs` script automates documentation generation from templates and live cluster state, creating comprehensive component documentation, cluster inventories, and architecture guides.

## Functional Requirements

### Documentation Generation Categories

#### Component Documentation
- **ACM Documentation**: Generate Advanced Cluster Management component guide
- **Pipelines Documentation**: Generate OpenShift Pipelines component guide
- **Template-based**: Use embedded templates for consistent formatting
- **Directory Detection**: Only generate docs for components that exist in repository

#### Inventory Documentation
- **Cluster Inventory**: Generate live and configured cluster listings
- **Application Inventory**: Generate ArgoCD application status and configurations
- **Live Data Integration**: Include real-time cluster and application status when possible
- **Fallback Handling**: Gracefully handle offline/disconnected scenarios

#### Architecture Documentation
- **GitOps Flow**: Generate GitOps workflow and sync wave documentation
- **Mermaid Diagrams**: Include visual workflow representations
- **ApplicationSet Patterns**: Document multi-cluster application deployment patterns

### Output Requirements

#### File Structure
Generated documentation must follow this structure:
```
docs/
├── components/
│   ├── acm.md                    # ACM component guide
│   └── pipelines.md              # Pipelines component guide
├── reference/
│   ├── cluster-inventory.md      # Live cluster status
│   └── application-inventory.md  # Application status
└── architecture/
    └── gitops-flow.md           # GitOps workflow guide
```

#### Content Standards
- **Audience Metadata**: Include audience, complexity, time, prerequisites
- **Consistent Formatting**: Use standardized section headers and code blocks
- **Live Data Integration**: Include real-time status when cluster access available
- **Timestamp**: Include generation timestamp for inventory documents

### Command Line Interface

#### Usage Patterns
```bash
# Generate all documentation (default)
./bin/generate-docs

# Generate specific categories
./bin/generate-docs --components
./bin/generate-docs --inventory
./bin/generate-docs --architecture

# Show help
./bin/generate-docs --help
```

#### Option Requirements
- `--components`: Generate only component documentation
- `--inventory`: Generate only inventory documentation
- `--architecture`: Generate only architecture documentation
- `--all`: Generate all documentation (default behavior)
- `-h, --help`: Display usage information

### Live Data Integration

#### Cluster Access Detection
- **oc Command**: Verify `oc` command availability
- **Authentication**: Check cluster authentication via `oc auth can-i`
- **Graceful Degradation**: Generate static documentation when cluster unavailable

#### Live Data Sources
- **ManagedCluster**: Query ACM managed cluster status
- **Applications**: Query ArgoCD application status
- **Custom Columns**: Use specific output formats for consistent data presentation

### Error Handling Requirements

#### Access Failures
- **Missing oc**: Log warning and skip live data sections
- **Authentication**: Log warning when cluster access unavailable
- **Permission Denied**: Handle insufficient permissions gracefully

#### File System Operations
- **Directory Creation**: Ensure target directories exist before writing
- **Write Permissions**: Handle filesystem permission issues
- **Existing Files**: Overwrite existing documentation files

### Template Requirements

#### Component Templates
- **Metadata Header**: Audience, complexity, time, prerequisites
- **Overview Section**: Component purpose and capabilities
- **Deployment Structure**: Directory structure and organization
- **Key Resources**: Important Kubernetes resources
- **Common Operations**: Frequently used commands
- **Related Documentation**: Cross-references to other docs

#### Architecture Templates
- **Mermaid Diagrams**: Visual workflow representations
- **Sync Wave Details**: Detailed deployment ordering explanation
- **Repository Flow**: Git to cluster deployment flow
- **ApplicationSet Patterns**: Multi-cluster deployment patterns

### Dependencies

#### External Commands
- **oc**: OpenShift CLI for cluster data (optional)
- **find**: File system searching for repository structure
- **grep**: Text searching for index updates

#### Repository Structure
- **Directory Detection**: Check for component directories before generation
- **Configuration Discovery**: Find cluster and application configurations
- **Kustomization Files**: Identify deployable configurations

### Index Management

#### Documentation Index
- **Existing Index**: Check for existing `docs/INDEX.md`
- **Section Addition**: Add new sections for generated components
- **Cross-references**: Maintain links to generated documentation

#### File Discovery
- **Generated Files**: List newly created documentation
- **Modification Time**: Show files newer than script execution
- **Organized Display**: Present generated files in structured format

### Related Tools

### Documentation Pipeline
- **[update-dynamic-docs.md](./update-dynamic-docs.md)** - Updates dynamic content in the documentation generated by this tool
- **[validate-docs.md](./validate-docs.md)** - Validates the documentation created by this tool

### Source Integration
- **[status.md](./status.md)** - Provides live cluster data for documentation generation
- **[bootstrap.md](./bootstrap.md)** - Documents the infrastructure that this tool generates docs for

### Quality Assurance
- **[health-check.md](./health-check.md)** - Validates components that are documented by this tool

## Design Principles

*This tool enables **living documentation** - automatically synchronized with repository structure and live cluster state for accurate, up-to-date technical documentation.*