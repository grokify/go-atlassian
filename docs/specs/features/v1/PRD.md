# Go-Atlassian v1.0 Product Requirements Document

## Overview

This document outlines the requirements and considerations for the Go-Atlassian v1.0 release, focusing on API stability, consistency, and developer experience.

## Goals

1. **API Stability**: Establish a stable public API with semantic versioning guarantees
2. **Consistency**: Ensure naming conventions and patterns are consistent across the codebase
3. **Testability**: Improve testability through interface definitions
4. **Documentation**: Provide clear documentation for all public types and functions

## Requirements

### P0 - Must Have (Blocking v1)

#### 1. Breaking API Changes

The following breaking changes must be completed before v1.0:

| Item | Status | Description |
|------|--------|-------------|
| Fix `ReadyforPlanningName` typo | ✅ Done | Renamed to `ReadyForPlanningName` |
| Fix `ReadyForDevlopmentName` typo | ✅ Done | Renamed to `ReadyForDevelopmentName` |
| Fix `TransitionsAPIReponse` typo | ✅ Done | Renamed to `TransitionsAPIResponse` |

#### 2. Code Quality

| Item | Status | Description |
|------|--------|-------------|
| Remove commented-out code | ✅ Done | Removed ~130 lines of dead code |
| Resolve TODO comments | ✅ Done | Implemented or documented all TODOs |
| Fix string literal typo in backlog_service.go | 🔲 Pending | Line 63: `"false) "` has extra characters |

### P1 - Should Have (Before v1)

#### 3. Naming Consistency

| Item | Status | Description |
|------|--------|-------------|
| Standardize Service/API terminology | 🔲 Pending | Choose `*Service` or `*API` consistently |
| Rename `CustomFieldAPI` field | 🔲 Pending | Client struct field should match type name |
| Rename `*_api.go` files | 🔲 Pending | Consider renaming to `*_service.go` for consistency |

#### 4. Type Consolidation

| Item | Status | Description |
|------|--------|-------------|
| Consolidate CLI IssueMeta | ✅ Done | CLI now uses shared `jira.IssueOutput` |
| Document custom field types | 🔲 Pending | Clarify `CustomField` vs `IssueCustomField` vs `CustomFieldOption` |

### P2 - Nice to Have (Post v1)

#### 5. Interface Definitions

| Item | Status | Description |
|------|--------|-------------|
| Define `IssueServicer` interface | 🔲 Pending | Enable mocking for tests |
| Define `ClientProvider` interface | 🔲 Pending | Abstract HTTP client for testing |

#### 6. Config Struct Refactoring

| Item | Status | Description |
|------|--------|-------------|
| Simplify `StageConfig` field names | 🔲 Pending | `StageNamePlanning` → `Planning` |
| Refactor `StatusCategoryConfig` | 🔲 Pending | Improve field naming consistency |

## Non-Goals

- Full backward compatibility with v0.x (breaking changes are acceptable)
- Comprehensive test coverage (can be improved incrementally post-v1)
- Performance optimizations (current performance is acceptable)

## Success Metrics

1. Zero compiler errors after upgrading from v0.x with documented migration path
2. All public API types and functions have documentation comments
3. Consistent naming patterns across all packages

## Timeline

- **v0.32.0**: Complete P1 items (naming consistency)
- **v1.0.0**: Release with stable API guarantees

## References

- [TRD.md](TRD.md) - Technical Requirements Document
- [CHANGELOG.md](/CHANGELOG.md) - Release history
