# bin/validate-docs Requirements

## Purpose

The `validate-docs` script provides comprehensive validation of markdown documentation, checking syntax, internal links, structure, and content quality across the entire repository.

## Functional Requirements

### Validation Categories

#### Markdown File Validation
- **Syntax Checking**: Validate markdown syntax and formatting
- **Header Validation**: Check header depth and formatting
- **Content Quality**: Identify common issues like trailing whitespace
- **Line Length**: Flag overly long lines (>120 characters)
- **Empty Files**: Detect and warn about empty documentation files

#### Link Validation
- **Internal Links**: Verify all internal markdown links resolve to existing files
- **Relative Path Resolution**: Support relative paths and directory traversal
- **Anchor Handling**: Skip anchor fragments for file existence checking
- **External Link Exclusion**: Skip validation of external URLs (https://)

#### Documentation Structure
- **Required Files**: Verify presence of critical documentation files
- **Documentation Index**: Check for proper documentation organization
- **Getting Started**: Validate presence of user onboarding documentation

#### Cross-Reference Analysis
- **Orphaned Files**: Identify files not referenced by other documentation
- **Duplicate Content**: Detect potential content duplication
- **Duplicate Headings**: Flag duplicate headings across files

### File Discovery and Processing

#### File Selection
- **Recursive Discovery**: Find all `.md` files throughout repository
- **Sorted Processing**: Process files in consistent order
- **Directory Support**: Support validation of specific directories
- **Individual Files**: Support validation of specific files

#### Context-Aware Processing
- **Relative Path Resolution**: Resolve links relative to file location
- **Repository Root**: Handle links relative to repository root
- **Complex Paths**: Support `../` and other relative path patterns

### Output Requirements

#### Progress Reporting
```
[INFO] Validating: docs/getting-started/README.md
[WARNING] Missing audience header in: docs/guides/monitoring.md
[ERROR] Broken link in docs/INDEX.md: guides/nonexistent.md
[SUCCESS] Documentation validation passed!
```

#### Validation Summary
```
===== VALIDATION SUMMARY =====
Files validated: 45
Links checked: 128
Broken links: 2
Errors: 3
Warnings: 8
```

#### Detailed Report Generation
Must generate `docs-validation-report.md` with:
- **Summary Statistics**: Files, links, errors, warnings
- **Status Assessment**: Pass/fail determination
- **Recommendations**: Actionable improvement suggestions

### Error Classification

#### Critical Errors (Fail Validation)
- **Broken Links**: Internal links to non-existent files
- **Missing Required Files**: Critical documentation files absent
- **Syntax Errors**: Malformed markdown that breaks rendering

#### Warnings (Pass with Notifications)
- **Format Issues**: Trailing whitespace, missing spaces in headers
- **Content Quality**: Long lines, missing audience headers
- **Structure Issues**: Orphaned files, duplicate headings

### Command Line Interface

#### Usage Patterns
```bash
# Validate all documentation
./bin/validate-docs

# Validate specific files
./bin/validate-docs README.md docs/INDEX.md

# Validate directory
./bin/validate-docs docs/

# Specialized validation
./bin/validate-docs --links-only
./bin/validate-docs --structure-only
./bin/validate-docs --quiet --no-report
```

#### Option Requirements
- `--links-only`: Check only internal links
- `--structure-only`: Check only documentation structure
- `--no-report`: Skip validation report generation
- `--quiet`: Suppress non-error output
- `-h, --help`: Display usage information

### Quality Assurance Checks

#### Documentation Standards
- **Audience Headers**: Require `**Audience**:` in docs/ files
- **Consistent Formatting**: Standardized header spacing and depth
- **Content Guidelines**: Flag overly long lines and empty files

#### Structure Requirements
Must verify presence of:
- **README.md**: Main repository documentation
- **ARCHITECTURE.md**: Technical architecture documentation
- **INSTALL.md**: Installation instructions
- **CLAUDE.md**: Claude Code integration documentation
- **docs/INDEX.md**: Documentation index (if docs/ exists)

### Advanced Analysis

#### Cross-Reference Validation
- **Orphan Detection**: Files not referenced elsewhere (except entry points)
- **Reference Counting**: Track how files link to each other
- **Basename Matching**: Support both full path and basename references

#### Content Analysis
- **Similar Files**: Detect files with similar names
- **Duplicate Headings**: Find repeated main headings across files
- **Content Organization**: Assess overall documentation structure

### Error Handling Requirements

#### File System Operations
- **Missing Files**: Handle non-existent files gracefully
- **Permission Issues**: Handle filesystem permission problems
- **Directory Traversal**: Safely navigate directory structures

#### Link Resolution
- **Path Normalization**: Handle various relative path formats
- **Realpath Resolution**: Resolve complex relative paths correctly
- **Context Preservation**: Maintain proper file context for link checking

### Performance Requirements

#### Efficient Processing
- **Incremental Validation**: Support validation of specific files/directories
- **Pattern Matching**: Use efficient regex patterns for content analysis
- **Memory Management**: Handle large documentation sets efficiently

#### Progress Tracking
- **Real-time Feedback**: Show validation progress for large sets
- **Error Counting**: Track issues across all files
- **Summary Statistics**: Provide comprehensive validation metrics

### Integration Requirements

#### CI/CD Integration
- **Exit Codes**: Return appropriate exit codes for automation
- **Report Generation**: Create machine-readable validation reports
- **Quiet Mode**: Support automated execution with minimal output

#### Development Workflow
- **Pre-commit Validation**: Enable validation before commits
- **Incremental Checks**: Support validation of changed files only
- **Quality Gates**: Provide clear pass/fail criteria

### Dependencies

#### External Commands
- **find**: File discovery and directory traversal
- **grep**: Pattern matching for content analysis
- **sed**: Text processing for link extraction
- **awk**: Line analysis for format checking

#### File System Access
- **Read Permissions**: Access all documentation files
- **Write Permissions**: Generate validation reports
- **Directory Navigation**: Traverse repository structure

### Report Generation

#### Validation Report Structure
```markdown
# Documentation Validation Report

*Generated on: [timestamp]*

## Summary
- Total Files Validated: X
- Total Links Checked: X
- Broken Links: X
- Errors: X
- Warnings: X

## Status
✅ PASSED - No critical issues found
❌ FAILED - Critical issues found
⚠️ WARNING - Minor issues found

## Recommendations
- Fix broken internal links
- Address formatting warnings
- Run validation regularly
```

#### Success Criteria
- **Pass**: No errors and no broken links
- **Fail**: Any errors or broken links present
- **Warning**: Only warnings present, no critical issues

## Related Tools

### Documentation Pipeline
- **[generate-docs.md](./generate-docs.md)** - Creates documentation that this tool validates
- **[update-dynamic-docs.md](./update-dynamic-docs.md)** - Updates documentation that this tool then validates

### Quality Assurance Integration
- **[health-check.md](./health-check.md)** - Validates system health, complementing documentation validation
- **[test-find-aws-resources.md](./test-find-aws-resources.md)** - Validates tool functionality, similar quality assurance approach

### Infrastructure Documentation
- **[bootstrap.md](./bootstrap.md)** - Critical documentation that requires validation
- **[status.md](./status.md)** - Status documentation that benefits from validation

## Design Principles

*This tool enables **documentation quality assurance** - ensuring comprehensive validation of documentation syntax, structure, and cross-references to maintain high-quality, navigable documentation throughout the repository.*