# rtsp-monitoring-system
# Camera Health Monitor

A Go-based service that continuously monitors the health of RTSP camera streams by testing their connectivity and availability.

## Prerequisites

- **Docker & Docker Compose** (recommended for easiest setup)
- **Go 1.19+** (if running locally without Docker)
- **FFmpeg** (if running locally without Docker)

## Quick Start with Docker

The easiest way to run this project is using the provided Docker Compose setup:

```bash
# Start the database and API services
docker compose -f docker-compose.yml up --build
```
IMPORTANT NOTE: 
make sure you are not running using the following ports: 5432, 8080, 8554

## Running the Camera Health Monitor

### Option 1: Using Docker (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd deepalerttest
   ```

2. **Start the infrastructure**
   ```bash
   docker compose -f docker-compose.yml up --build
   ```

3. **Build and run the monitor**
   
   Create a Dockerfile in the project root:
   ```dockerfile
   FROM golang:1.21-alpine
   
   RUN apk add --no-cache ffmpeg
   
   WORKDIR /app
   COPY . .
   
   RUN go mod download
   RUN go build -o camera-monitor .
   
   CMD ["./camera-monitor"]
   ```

4. **Set environment variables and run**
   ```bash
   docker build --load -t camera-monitor .
   
   docker run --network host \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=postgres \
     -e POSTGRES_HOST=localhost \
     -e POSTGRES_PORT=5432 \
     camera-monitor
   ```

   or 
   on Mac(Silicon)
    ```bash
   docker buildx build --load -t camera-monitor .
   ```

### Option 2: Running Locally

1. **Install FFmpeg**
   - **Ubuntu/Debian**: `sudo apt-get install ffmpeg`
   - **macOS**: `brew install ffmpeg`
   - **Windows**: Download from [ffmpeg.org](https://ffmpeg.org/download.html)

2. **Start the infrastructure**
   ```bash
    docker compose -f docker-compose.yml up --build
   ```

3. **Set environment variables**
   ```bash
   export POSTGRES_USER=postgres
   export POSTGRES_PASSWORD=postgres
   export POSTGRES_DB=postgres
   export POSTGRES_HOST=localhost
   export POSTGRES_PORT=5432
   ```

4. **Install Go dependencies**
   ```bash
   go mod download
   ```

5. **Run the application**
   ```bash
   go run main.go
   ```

## Project Structure

```
.
├── main.go              # Application entry point
├── service/
│   └── cameraHealth.go  # Camera health orchestration
├── producer/
│   └── processor.go     # RTSP stream camera logic
├── consumer/
│   └── output.go        # Results output handler
├── utils/
│   ├── types.go         # Data structures
│   ├── const.go         # Configuration constants
│   └── dbConn.go        # Database connection
└── docker-compose.yml   # Infrastructure setup
```

## Configuration

The application uses environment variables for configuration:

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_USER` | PostgreSQL username | (required) |
| `POSTGRES_PASSWORD` | PostgreSQL password | (required) |
| `POSTGRES_DB` | Database name | (required) |
| `POSTGRES_HOST` | Database host | `localhost` |
| `POSTGRES_PORT` | Database port | `5432` |

### Tuning Parameters

You can modify these constants in `utils/const.go`:

- `MaxWorkers`: Number of concurrent camera checks (default: 10)
- `FfmpegTimeout`: Maximum time to wait for FFmpeg (default: 15 seconds)
- `RtspTimeout`: RTSP connection timeout in microseconds (default: 5 seconds)

## How It Works

1. **Every minute**, the application queries the database for all cameras
2. **Concurrently checks** each camera's RTSP stream using FFmpeg (up to 10 at a time)
3. **Classifies** each camera's status as:
   - `healthy` - Stream is accessible and working
   - `unauthorised` - Authentication failure (401)
   - `offline` - Network/connection issues
   - `context_timeout` - Check exceeded time limit
4. **Outputs** results to console in real-time

## Expected Output

```
Starting camera health check run
message: ID: 1, Name: Front Door, Status: healthy
message: ID: 2, Name: Parking Lot, Status: offline
message: ID: 3, Name: Back Entrance, Status: unauthorised
...
```

## Database Schema

The application expects a PostgreSQL table named `cameras` with the following structure:

```sql
CREATE TABLE cameras (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    rtsp_url VARCHAR(512) NOT NULL
);
```

## Troubleshooting

### "failed to get camera data" error
- Ensure Docker Compose services are running: `docker-compose ps`
- Check database connectivity: `docker-compose logs db`
- Verify environment variables are set correctly

### FFmpeg not found
- Install FFmpeg using your package manager
- Verify installation: `ffmpeg -version`

### All cameras show "offline"
- Check if the API service is running on port 8080
- Verify cameras are accessible from your network
- Try testing a camera URL manually: `ffmpeg -rtsp_transport tcp -i rtsp://camera-url -frames:v 1 -f null -`

### High timeout rates
- Increase `FfmpegTimeout` in `utils/const.go`
- Check network latency to camera streams
- Reduce `MaxWorkers` to avoid overwhelming the network

## Development


### Building
```bash
go build -o camera-monitor .
```

### Adding Dependencies
```bash
go get <package-name>
go mod tidy
```
