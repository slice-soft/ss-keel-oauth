
# Contributing to ss-keel-oauth

The base contributing guide — workflow, commit conventions, PR guidelines, and community standards — lives in the [ss-community](https://github.com/slice-soft/ss-community/blob/main/CONTRIBUTING.md) repository. Read it first.

This document covers only what is specific to this repository.

## Getting Started

> Requirements
>- Go 1.25+
>- Git


1. **Fork the repository**
2. **Clone your repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/ss-keel-oauth.git
   cd ss-keel-oauth
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Create a branch**
   ```bash
   git checkout -b feat/your-feature-name
   ```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

```

