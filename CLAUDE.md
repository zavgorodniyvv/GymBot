# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GymBot is a Telegram bot for tracking push-up workouts written in Go. The bot helps users log workout sets, generates training plans, and provides statistics. The project is currently migrating from file-based storage to MongoDB.

## Architecture

The codebase follows a clean architecture pattern with these main packages:

- `cmd/bot/` - Main entry point and application startup
- `internal/bot/` - Bot logic and message handlers  
- `internal/storage/` - Data persistence layer (file-based and MongoDB implementations)
- `internal/planner/` - Workout planning algorithms

### Key Components

- **Bot Handler**: Processes Telegram updates and commands (`/start`, `/plan`, `/stats`, `/end`, `/reset`)
- **Storage Layer**: Dual implementation supporting both file-based (legacy) and MongoDB storage
- **Workout Planner**: Generates adaptive training plans based on user performance history
- **User Data Model**: Tracks sessions, current workouts, and training plans

## Development Commands

### Building and Running
```bash
# Build the application
go build ./cmd/bot

# Run directly with Go
go run ./cmd/bot

# Build with specific output name
go build -o gymbot ./cmd/bot
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./internal/storage
```

### Dependencies
```bash
# Download dependencies
go mod download

# Clean up unused dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Environment Setup

Required environment variables:
- `TELEGRAM_TOKEN` - Bot token from @BotFather
- `MONGO_URI` - MongoDB connection string (for new storage implementation)

## Storage Migration

The project is currently migrating from file-based storage (`storage.go`) to MongoDB (`mongo_storage.go`). Both implementations coexist:

- File storage uses JSON files in `data/` directory
- MongoDB storage uses `gymbot` database with `users` collection
- Both implement the same interface for `SaveUser()`, `LoadUser()`, and `FinishWorkout()`

## Code Patterns

- Error handling follows Go conventions with explicit error returns
- Storage operations use context with timeouts (5-10 seconds)
- User data is loaded/saved on each operation (no in-memory caching)
- Workout planning uses adaptive algorithms based on user performance