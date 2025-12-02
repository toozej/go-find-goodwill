# go-find-goodwill

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/toozej/go-find-goodwill)
[![Go Report Card](https://goreportcard.com/badge/github.com/toozej/go-find-goodwill)](https://goreportcard.com/report/github.com/toozej/go-find-goodwill)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/toozej/go-find-goodwill/cicd.yaml)
![Docker Pulls](https://img.shields.io/docker/pulls/toozej/go-find-goodwill)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/toozej/go-find-goodwill/total)

## Database Migration to GORM

The go-find-goodwill application has transitioned from file-based SQL migrations to GORM's AutoMigrate system, providing automatic schema management and ORM capabilities.

### Key Changes

- **GORM ORM**: Full ORM support with struct-based schema definition
- **AutoMigrate**: Automatic table creation and schema updates
- **Transaction Support**: Built-in transaction management
- **Relationship Management**: Automatic handling of one-to-many and many-to-many relationships
- **Improved Developer Experience**: Type-safe queries and better error handling

### Documentation

For comprehensive GORM implementation details:

- **[DATABASE_SCHEMA.md](DATABASE_SCHEMA.md)**: Complete GORM schema specification
- **[GORM_USAGE_GUIDE.md](GORM_USAGE_GUIDE.md)**: Usage patterns and best practices
- **[internal/goodwill/db/README.md](internal/goodwill/db/README.md)**: Database layer documentation

### Migration Guide

The application now uses GORM for all database operations:

1. **Schema Definition**: Models defined as Go structs with GORM tags
2. **Automatic Migrations**: `AutoMigrate()` handles schema creation and updates
3. **Transaction Support**: Built-in transaction patterns for data integrity
4. **Legacy Compatibility**: Conversion functions for backward compatibility

### Setup Instructions

- set up new repository in quay.io web console
  - (DockerHub and GitHub Container Registry do this automatically on first push/publish)
  - name must match Git repo name
  - grant robot user with username stored in QUAY_USERNAME "write" permissions (your quay.io account should already have admin permissions)
- set built packages visibility in GitHub packages to public
  - navigate to https://github.com/users/$USERNAME/packages/container/$REPO/settings
  - scroll down to "Danger Zone"
  - change visibility to public

## changes required to update golang version

- `make update-golang-version`
