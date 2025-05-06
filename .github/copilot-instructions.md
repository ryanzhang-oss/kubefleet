# GitHub Copilot Instructions

This document provides guidance for GitHub Copilot when generating code for the KubeFleet project.

## Project Overview

KubeFleet is a Kubernetes-based fleet management system. Consider the following when generating code:

- This is a cloud-native application following microservices architecture
- Primary language: Go
- Kubernetes-native design principles are followed

## Code Style and Conventions

### Go

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use error handling with proper context
- Implement interfaces only when needed
- Prefer table-driven tests

## Project Architecture

- Controllers follow the Kubernetes Operator pattern
- APIs are RESTful and follow Kubernetes API conventions
- Use dependency injection for better testability
- Prefer immutable data structures

## Testing Practices

- Write unit tests for all business logic
- Integration tests should use testcontainers where applicable
- Aim for high test coverage in critical paths
- Mock external dependencies in unit tests

## Documentation

- Document all public APIs
- Include examples in documentation
- Keep comments up-to-date with code changes
- Document architecture decisions in ADRs

## Common Patterns

- Use context for cancellation and timeouts
- Follow the error handling patterns established in the codebase
- Use structured logging with appropriate log levels
- Implement graceful shutdown for all services

## Security Considerations

- Never hardcode secrets or credentials
- Validate all user inputs
- Follow the principle of least privilege
- Use proper RBAC for Kubernetes resources

## Performance Guidelines

- Consider resource constraints in Kubernetes environments
- Implement appropriate caching strategies
- Use pagination for listing APIs
- Be mindful of memory usage in data processing

## Compatibility

- Support multiple Kubernetes versions (specified in go.mod)
- Consider backward compatibility in API changes
- Follow semantic versioning
