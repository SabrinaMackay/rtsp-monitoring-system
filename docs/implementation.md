# Project Explanation: Camera Health Monitoring System

## What This Project Does

This is an **automated RTSP camera health monitoring system** written in Go. It continuously checks whether IP security cameras are online, accessible, and functioning properly by testing their RTSP (Real-Time Streaming Protocol) video streams.

## The Problem It Solves

In security and surveillance systems with dozens or hundreds of cameras, it's critical to know when cameras go offline, lose connectivity, or have authentication issues. Manually checking each camera is time-consuming and impractical. This system automates that process by:

1. **Continuously monitoring** all cameras in a database
2. **Detecting failures quickly** (checks every minute)
3. **Classifying problems** so operators know what type of issue occurred
4. **Processing cameras efficiently** using concurrent workers

## How It Works

### Architecture Overview

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│  PostgreSQL │────────▶│  Main Loop   │────────▶│   Worker    │
│  Database   │         │  (1 minute)  │         │    Pool     │
└─────────────┘         └──────────────┘         └─────────────┘
                                                         │
                                                         ▼
                                                  ┌─────────────┐
                                                  │   FFmpeg    │
                                                  │   Testing   │
                                                  └─────────────┘
                                                         │
                                                         ▼
                                                  ┌─────────────┐
                                                  │   Results   │
                                                  │   Console   │
                                                  └─────────────┘
```

### Component Breakdown

#### 1. **Main Loop** (`main.go`)
- Runs indefinitely with a 1-minute ticker
- Fetches all cameras from the database each cycle
- Passes cameras to the testing service
- Handles errors gracefully and continues running

#### 2. **Camera Testing Service** (`service/cameraHealth.go`)
- Creates a worker pool with 10 concurrent workers (configurable)
- Uses Go channels to distribute work and collect results
- Ensures all tests complete before moving to the next cycle
- Prevents overwhelming the system with too many simultaneous connections

**Why this design?**
- **Concurrency**: Testing 10 cameras serially at 10 seconds each ~1.7 minutes. With 10 workers, it takes ~10 seconds
- **Resource control**: Limits concurrent FFmpeg processes to prevent system overload
- **Non-blocking**: Uses goroutines and channels for efficient async processing

#### 3. **Camera Processor** (`producer/processor.go`)
This is where the actual health check happens. For each camera:

1. **Spawns FFmpeg** with specific parameters:
   - `-rtsp_transport tcp`: Uses TCP instead of UDP (more reliable)
   - `-timeout 5000000`: 5-second connection timeout
   - `-frames:v 1`: Only attempts to grab a single frame (fast test)
   - `-f null -`: Discards the frame (we only care about connectivity)

2. **Sets a 10-second deadline** to prevent hanging on unresponsive cameras

3. **Classifies errors** by parsing FFmpeg's output:
   - `401 / unauthorised` → Camera requires authentication or credentials are wrong
   - `context timeout` → Healthy check exceeded the 10-second limit
   - `offline` → Something else went wrong

**Why FFmpeg?**
- Industry-standard tool for video streaming
- Supports RTSP protocol natively
- Provides detailed error messages
- Lightweight frame capture test is fast and efficient

#### 4. **Result Consumer** (`consumer/output.go`)
- Runs as a goroutine listening on a channel
- Prints results as they come in (real-time feedback)
- Could easily be extended to:
  - Write to a database
  - Send alerts/notifications
  - Update a dashboard
  - Generate reports

#### 5. **Database Connection** (`utils/dbConn.go`)
- Establishes connection to PostgreSQL
- Uses environment variables for configuration
- Reads camera information (ID, name, RTSP URL)
- Simple query: `SELECT * FROM cameras`

### Data Flow

```
1. Timer triggers (every 60 seconds)
                ↓
2. Query database for all cameras
                ↓
3. For each camera:
   - Serialize to JSON
   - Add to worker queue
   - Worker picks up task
                ↓
4. Worker spawns FFmpeg subprocess
   - Attempt to connect to RTSP stream
   - Try to grab one frame
   - Wait up to 15 seconds
                ↓
5. Analyze FFmpeg output
   - Success → "healthy"
   - Error → Classify error type
                ↓
6. Send result to channel
                ↓
7. Consumer prints result to console
                ↓
8. Wait for next tick
```

## Key Design Decisions

### Why Go?
- **Excellent concurrency primitives**: Goroutines and channels make concurrent testing natural
- **Fast execution**: Compiled binary is fast and efficient
- **Great stdlib**: Built-in support for JSON, SQL, command execution
- **Simple deployment**: Single binary with no runtime dependencies (except FFmpeg)

### Why FFmpeg for Testing?
- **Standard tool**: Already used in video processing workflows
- **Reliable**: Battle-tested protocol implementation
- **Fast**: Can test connectivity by grabbing a single frame
- **Informative errors**: Provides detailed error messages for diagnosis

### Why Worker Pool Pattern?
- **Controlled concurrency**: Prevents spawning thousands of FFmpeg processes
- **Prevents resource exhaustion**: Limits memory and CPU usage
- **Backpressure handling**: Queue naturally throttles if workers are busy
- **Scalable**: Easy to adjust worker count based on system resources

### Why 1-Minute Intervals?
- **Balance**: Frequent enough to catch issues quickly, infrequent enough to not overwhelm cameras
- **Camera-friendly**: Most IP cameras can handle connection attempts every minute
- **Practical**: Most camera issues don't need sub-minute detection

## Potential Improvements

### Short-term
- **Persistence**: Store results in database instead of just printing
- **Alerting**: Send notifications when cameras go offline
- **Metrics**: Track uptime percentages and failure patterns
- **Web Dashboard**: Real-time visualization of camera status

### Medium-term
- **Adaptive testing**: Test failing cameras more frequently
- **Retry logic**: Attempt multiple times before marking as failed
- **Historical data**: Track trends and generate reports
- **Authentication management**: Securely store camera credentials

### Long-term
- **Distributed testing**: Run testers from multiple locations
- **Load balancing**: Distribute testing across multiple servers
- **AI anomaly detection**: Identify unusual patterns automatically
- **Integration**: API for other systems to query camera status

## Use Cases

This system is valuable for:

1. **Security Operations Centers**: Monitor hundreds of surveillance cameras
2. **Smart Buildings**: Ensure security infrastructure is operational
3. **Retail**: Track camera uptime across multiple store locations
4. **Critical Infrastructure**: Maintain 24/7 video surveillance
5. **Compliance**: Demonstrate security system reliability for audits

## Technical Requirements Explained

### Why PostgreSQL?
- The `docker-compose.yml` provides a pre-populated database with camera data
- Production-ready relational database
- Easy to query and manage camera information

### Why Environment Variables?
- **Security**: Don't hard-code credentials
- **Flexibility**: Easy to change configuration per environment
- **12-Factor App**: Follows modern application best practices

### Why Docker Compose?
- **Easy setup**: One command to start all dependencies
- **Consistent environment**: Same setup for all developers
- **Production-like**: Mirrors how it would be deployed

## Conclusion

This is a practical, production-oriented solution for camera health monitoring. It demonstrates:
- Concurrent programming patterns in Go
- Integration with external tools (FFmpeg)
- Database connectivity
- Error handling and classification
- Scalable architecture

The system is designed to be **reliable** (handles failures gracefully), **efficient** (concurrent processing), and **extensible** (easy to add features like alerting or dashboards).