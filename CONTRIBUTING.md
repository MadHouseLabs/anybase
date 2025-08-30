# Contributing to AnyBase

Thank you for your interest in contributing to AnyBase! We welcome contributions from the community and are grateful for any help you can provide.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Please be respectful and professional in all interactions.

## How to Contribute

### Reporting Issues

- Check existing issues before creating a new one
- Use clear, descriptive titles
- Provide detailed steps to reproduce the issue
- Include system information (OS, Go version, etc.)
- Add relevant logs or error messages

### Suggesting Features

- Open a discussion first to gauge interest
- Provide clear use cases and benefits
- Consider implementation complexity
- Be open to feedback and alternative approaches

### Pull Requests

1. **Fork the Repository**
   ```bash
   git clone https://github.com/MadHouseLabs/anybase.git
   cd anybase
   git remote add upstream https://github.com/MadHouseLabs/anybase.git
   ```

2. **Create a Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make Your Changes**
   - Write clean, readable code
   - Follow existing code style and conventions
   - Add tests for new functionality
   - Update documentation as needed

4. **Test Your Changes**
   ```bash
   # Run backend tests
   go test ./...
   
   # Run dashboard tests
   cd dashboard
   pnpm test
   ```

5. **Commit Your Changes**
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```
   
   Follow conventional commit format:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `style:` for code style changes
   - `refactor:` for code refactoring
   - `test:` for test additions/changes
   - `chore:` for maintenance tasks

6. **Push to Your Fork**
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Create a Pull Request**
   - Use a clear, descriptive title
   - Reference any related issues
   - Describe what changes you made and why
   - Include screenshots for UI changes

## Development Setup

### Backend Development

```bash
# Install dependencies
go mod download

# Run with hot reload
go run cmd/server/main.go

# Run tests
go test ./...

# Run linter
golangci-lint run

# Format code
go fmt ./...
```

### Dashboard Development

```bash
cd dashboard

# Install dependencies
pnpm install

# Run development server
pnpm dev

# Run tests
pnpm test

# Build for production
pnpm build

# Run linter
pnpm lint
```

## Code Style Guidelines

### Go Code
- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use meaningful variable and function names
- Keep functions small and focused
- Add comments for exported functions
- Handle errors explicitly
- Use dependency injection for testability

### TypeScript/React Code
- Use functional components with hooks
- Follow React best practices
- Use TypeScript types properly
- Keep components small and reusable
- Use meaningful prop names
- Add JSDoc comments for complex functions

### General Guidelines
- Write self-documenting code
- Avoid deep nesting
- DRY (Don't Repeat Yourself)
- KISS (Keep It Simple, Stupid)
- Test edge cases
- Consider performance implications

## Testing

### Writing Tests
- Write unit tests for all new functions
- Use table-driven tests in Go
- Mock external dependencies
- Test error cases
- Aim for >80% code coverage

### Running Tests
```bash
# Backend tests with coverage
go test -cover ./...

# Dashboard tests with coverage
cd dashboard
pnpm test --coverage
```

## Documentation

- Update README.md for user-facing changes
- Add inline comments for complex logic
- Update API documentation for endpoint changes
- Include examples for new features
- Keep documentation concise and clear

## Review Process

1. All PRs require at least one review
2. CI checks must pass
3. No merge conflicts
4. Documentation updated if needed
5. Tests added/updated as appropriate

## Release Process

1. Features are merged to `main` branch
2. Releases are tagged with semantic versioning
3. Changelog is automatically generated
4. Docker images are built and published

## Getting Help

- Open a [Discussion](https://github.com/MadHouseLabs/anybase/discussions) for questions
- Join our community chat (coming soon)
- Check existing issues and PRs
- Read the documentation thoroughly

## Recognition

Contributors will be recognized in:
- The project README
- Release notes
- Our contributors page

Thank you for contributing to AnyBase!