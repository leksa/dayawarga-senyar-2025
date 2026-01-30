# Testing & Development Guide

## Prerequisites

1. **Go installed** (1.21 or later)
2. **PostgreSQL database running** (local or via Docker)
3. **Environment variables configured** in `.env` file

## Running Unit Tests

### Run all tests
```bash
cd dayawarga-senyar-2025/services/api
make test
```

### Run tests with coverage report
```bash
make test-coverage
```

This will generate a `coverage.out` file. To view the coverage report:
```bash
go tool cover -html=coverage.out
```

### Run specific test package
```bash
go test -v ./internal/validator -timeout 30s
```

### Run specific test function
```bash
go test -v ./internal/validator -run TestValidatePagination -timeout 30s
```

## Database Migrations

### Setup environment variables
Create or update `.env` file:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=senyar
DB_PASSWORD=senyar123
DB_NAME=senyar
```

### Run all migrations
```bash
make migrate
# or
make migrate-up
```

This will apply all pending migrations including:
- Schema creation
- Performance indexes (CRITICAL for bbox queries!)

### Rollback last migration
```bash
make migrate-down
```

### Clean and recreate database (DEVELOPMENT ONLY!)
```bash
make migrate-clean
```

⚠️ **WARNING**: This will DROP the database! Only use in development!

## Running the Application

### Development mode
```bash
make run
```

### Build production binary
```bash
make build
```

Then run:
```bash
./bin/api
```

## Testing the API Locally

After starting the application, you can test it in your browser:

### 1. Test locations endpoint (with validation)
```
http://localhost:8080/api/v1/locations?page=1&limit=50
```

### 2. Test bounding box filter
```
http://localhost:8080/api/v1/locations?bbox=95,-6,98,5
```
This should return locations in Aceh region.

### 3. Test search with input sanitization
```
http://localhost:8080/api/v1/locations?search=Posko
```

### 4. Test invalid input (should return 400)
```
http://localhost:8080/api/v1/locations?page=0
http://localhost:8080/api/v1/locations?limit=500
http://localhost:8080/api/v1/locations?bbox=invalid
```

### 5. Test specific location by ID
```
http://localhost:8080/api/v1/locations/{uuid}
```

### 6. Test rate limiting (should get 429 after many requests)
```bash
# Run this multiple times quickly
for i in {1..600}; do curl http://localhost:8080/api/v1/locations; done
```

You should see:
```
X-RateLimit-Limit: 500
X-RateLimit-Remaining: [decreasing number]
```

After hitting the limit:
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMITED",
    "message": "Too many requests, please try again later"
  }
}
```

## Performance Indexes Impact

After running migrations, you should see significant performance improvements:

### Before indexes
- Bounding box queries: Slow (full table scan)
- Pagination with filters: Slow
- Type/Status filters: Slow

### After indexes
- ✅ Bounding box queries: **Fast** (PostGIS GIST index)
- ✅ Pagination with filters: **Fast** (composite indexes)
- ✅ Type/Status filters: **Fast** (partial indexes)
- ✅ Geo queries: **Fast** (spatial index)

### Check if indexes are created
```sql
SELECT indexname, tablename 
FROM pg_indexes 
WHERE schemaname = 'public' 
AND tablename IN ('locations', 'feeds', 'location_photos', 'sync_state')
ORDER BY tablename, indexname;
```

## Troubleshooting

### Test failures
```bash
# Run tests with verbose output
go test -v ./internal/...

# Run tests with race detection
go test -race ./internal/...
```

### Migration errors
```bash
# Check migration status
migrate -path ./infrastructure/database/migrations -database "postgresql://user:pass@host:port/dbname?sslmode=disable" version

# Force migration (DANGEROUS!)
migrate -path ./infrastructure/database/migrations -database "postgresql://user:pass@host:port/dbname?sslmode=disable" force 1
```

### Database connection issues
```bash
# Test connection
psql -h localhost -p 5432 -U senyar -d senyar

# Check if Postgres is running
docker ps | grep postgres
```

### Rate limiter not working
Check if you're seeing the correct headers:
```bash
curl -I http://localhost:8080/api/v1/locations
```

Look for:
```
X-RateLimit-Limit: 500
X-RateLimit-Remaining: 499
```

## Performance Testing

### Load test locations endpoint
```bash
# Install wrk (if not installed)
# macOS: brew install wrk
# Linux: sudo apt-get install wrk

# Run load test
wrk -t4 -c100 -d30s http://localhost:8080/api/v1/locations
```

### Test bbox query performance
```bash
time curl "http://localhost:8080/api/v1/locations?bbox=95,-6,98,5"
```

## Next Steps

After verifying local development works:

1. ✅ All tests passing
2. ✅ Database migrations applied
3. ✅ API endpoints working in browser
4. ✅ Rate limiting working
5. ✅ Performance indexes active

Then you're ready to:
- Add more unit tests
- Add integration tests
- Set up CI/CD pipeline
- Deploy to staging/production