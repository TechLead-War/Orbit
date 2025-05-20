# ORBIT - LeetCode Rating Review System

## Overview
A scalable system for tracking and reviewing students' LeetCode ratings and performance.

## Tech Stack
- Go 1.21
- PostgreSQL
- Gin Web Framework
- GORM
- Docker

## Prerequisites
- Go 1.21 or higher
- PostgreSQL
- Docker & Docker Compose
- golang-migrate
- Postman (for API testing)

## Quick Start
1. Clone the repository
```bash
git clone https://github.com/ayush/ORBIT.git
cd ORBIT
```

2. Install golang-migrate
```bash
# MacOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
mv migrate /usr/local/bin/migrate
```

3. Setup environment
```bash
cp .env.example .env
```

4. Start PostgreSQL
```bash
make docker-up
```

5. Run migrations
```bash
make migrate-up
```

6. Start server
```bash
make run
```

## Database Migrations
Create new migration:
```bash
make migrate-create
```

Apply migrations:
```bash
make migrate-up
```

Rollback migrations:
```bash
make migrate-down
```

## API Testing with Postman
1. Import the Postman collection:
   - Open Postman
   - Click "Import"
   - Select `postman/ORBIT.postman_collection.json`
   - Also import `postman/ORBIT.postman_environment.json`

2. Available Endpoints:
   - Create Student: `POST /api/v1/students`
   - List Students: `GET /api/v1/students`
   - Get Student: `GET /api/v1/students/:id`
   - Update Rating: `PUT /api/v1/students/:id/rating`
   - Get Statistics: `GET /api/v1/students/:id/stats`

3. Example Requests:
```json
// Create Student
POST /api/v1/students
{
    "name": "John Doe",
    "email": "john@example.com",
    "leetcode_id": "johndoe123"
}

// Update Rating
PUT /api/v1/students/1/rating
{
    "rating": 1500,
    "problems_count": 100,
    "easy_count": 40,
    "medium_count": 40,
    "hard_count": 20,
    "contest_rating": 1600,
    "global_rank": 10000
}
```

## Make Commands
- `make build` - Build binary
- `make run` - Run server
- `make test` - Run tests
- `make clean` - Clean build files
- `make deps` - Download dependencies
- `make docker-up` - Start containers
- `make docker-down` - Stop containers

## Project Structure
```
.
├── cmd/
│   └── api/            # Application entrypoint
├── internal/
│   ├── api/           # HTTP handlers
│   ├── models/        # Data models
│   ├── repository/    # Data access
│   ├── service/       # Business logic
│   └── config/        # Configuration
├── migrations/        # Database migrations
├── postman/          # Postman collection and environment
├── scripts/          # Utility scripts
└── api/              # API documentation
```

## Environment Variables
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=Orbit
SERVER_PORT=8080
```

## Features
- Student profile management
- LeetCode rating tracking
- Performance analytics
- Review system

## Setup
1. Clone the repository
2. Copy `.env.example` to `.env` and configure your environment variables
3. Run `go mod download` to install dependencies
4. Run `go run cmd/api/main.go` to start the server

## API Documentation
API documentation is available in the `/api` directory.

## Development
1. Follow Go best practices and coding standards
2. Write tests for new features
3. Update documentation as needed

## License
MIT License

This is backend for leetcode ranker, that I am re-writing in GO.

Project Guide

Step1: Run the docker-desktop<br>
<br>
Step2: Run this command, otherwise if you can see a erd folder already then skip this step.
```bash
docker run --rm -v "$PWD/erd:/output" schemaspy/schemaspy \
-t pgsql \
-host host.docker.internal \
-port 5432 \
-db Orbit \
-u postgres \
-p password \
-s public
```
OR<br>
Run ``./generate_erd.sh``

Step3: You can see a ``erd`` folder, navigate to ``erd/index.html`` to see the database schema.

