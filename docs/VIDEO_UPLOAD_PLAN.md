# Video Upload Service - Implementation Plan

## 🎯 Goal
Build a robust video upload service that handles large video files efficiently with chunk uploading, metadata extraction, and cloud storage integration.

## 📋 Architecture Overview

```
Client (Browser/App)
        ↓
    [Upload API]
        ↓
   ┌────────────────┐
   │ Upload Handler │
   └────────────────┘
           ↓
    ┌──────┴──────┐
    ↓             ↓
[Chunk Manager] [Metadata Extractor]
    ↓             ↓
    └──────┬──────┘
           ↓
    [Storage Service]
           ↓
    MinIO/S3 Bucket
           ↓
    [Database Record]
```

## 🔧 Components to Build

### 1. Storage Setup (MinIO)
**Why MinIO?**
- S3-compatible (same code works with AWS S3 later)
- Easy to run locally for development
- Perfect for learning distributed storage

**What to do:**
- [ ] Install MinIO locally
- [ ] Create bucket for videos
- [ ] Setup access credentials
- [ ] Add to `.env` configuration

---

### 2. Video Model & Database Schema

**New Table: `videos`**
```sql
CREATE TABLE videos (
    id              INTEGER PRIMARY KEY,
    movie_id        INTEGER,  -- Foreign key to movies table
    title           TEXT,
    filename        TEXT,
    original_name   TEXT,
    file_size       INTEGER,  -- in bytes
    duration        FLOAT,    -- in seconds
    width           INTEGER,
    height          INTEGER,
    codec           TEXT,
    format          TEXT,     -- mp4, avi, etc.
    bitrate         INTEGER,
    fps             FLOAT,
    storage_path    TEXT,     -- S3/MinIO path
    upload_status   TEXT,     -- pending, uploading, completed, failed
    created_at      TIMESTAMP,
    updated_at      TIMESTAMP,
    FOREIGN KEY(movie_id) REFERENCES movies(id)
);
```

**New Table: `upload_sessions`** (for chunk tracking)
```sql
CREATE TABLE upload_sessions (
    id              TEXT PRIMARY KEY,  -- UUID
    video_id        INTEGER,
    total_chunks    INTEGER,
    uploaded_chunks INTEGER DEFAULT 0,
    chunk_size      INTEGER,
    status          TEXT,  -- in_progress, completed, failed
    created_at      TIMESTAMP,
    expires_at      TIMESTAMP,
    FOREIGN KEY(video_id) REFERENCES videos(id)
);
```

---

### 3. Chunked Upload Implementation

**Why Chunking?**
- Large videos (GB+) can't upload in one request
- Network failures - resume from last chunk
- Better user experience with progress bar
- Prevents timeouts

**Flow:**
```
1. Client: Initiate upload → Get upload session ID
2. Client: Upload chunk 1/N → Server stores temp
3. Client: Upload chunk 2/N → Server stores temp
   ...
N. Client: Upload chunk N/N → Server merges all chunks
N+1. Server: Extract metadata → Save to MinIO → Update DB
```

**Endpoints:**
- `POST /videos/upload/initiate` - Start upload session
- `POST /videos/upload/chunk` - Upload single chunk
- `POST /videos/upload/complete` - Finalize upload
- `GET /videos/upload/status/:id` - Check upload progress
- `DELETE /videos/upload/cancel/:id` - Cancel upload

---

### 4. Metadata Extraction

**Use FFprobe** (from FFmpeg suite)
- Get video duration
- Get resolution (width x height)
- Get codec (H.264, H.265, etc.)
- Get bitrate
- Get frame rate

**Go Library:** `github.com/xfrr/goffmpeg`

**Example:**
```go
func ExtractMetadata(filePath string) (*VideoMetadata, error) {
    // Run ffprobe command
    // Parse JSON output
    // Return structured metadata
}
```

---

### 5. Storage Integration (MinIO/S3)

**Operations:**
- Upload file to bucket
- Generate presigned URLs for download
- Delete file
- List files

**Go Library:** `github.com/minio/minio-go/v7`

**Storage Path Structure:**
```
bucket-name/
  videos/
    movie-{movie_id}/
      original/
        {video_id}.mp4
      thumbnails/
        {video_id}-thumb-001.jpg
        {video_id}-thumb-002.jpg
```

---

### 6. API Endpoints

**Public Endpoints:**
- `GET /videos/:id/stream` - Stream video (later with HLS)
- `GET /videos/:id/thumbnail` - Get video thumbnail

**Protected Endpoints (JWT required):**
- `POST /videos/upload/initiate` - Start upload
- `POST /videos/upload/chunk` - Upload chunk
- `POST /videos/upload/complete` - Complete upload
- `GET /videos/upload/status/:sessionId` - Upload progress
- `DELETE /videos/upload/cancel/:sessionId` - Cancel upload
- `GET /videos/:id` - Get video details
- `PUT /videos/:id` - Update video details
- `DELETE /videos/:id` - Delete video

---

### 7. Validation & Security

**File Validation:**
- Check file extension (mp4, avi, mov, mkv, webm)
- Check MIME type
- Maximum file size limit (e.g., 10GB)
- Virus scanning (optional, for production)

**Security:**
- JWT authentication required
- User can only upload to their movies
- Rate limiting on upload endpoints
- Signed URLs for video access

---

## 📦 Dependencies to Add

```bash
# MinIO client
go get github.com/minio/minio-go/v7

# FFmpeg wrapper for metadata extraction
go get github.com/xfrr/goffmpeg

# UUID for upload sessions
go get github.com/google/uuid

# File type detection
go get github.com/h2non/filetype
```

---

## 🚀 Implementation Order

### Phase 1: Basic Infrastructure (Today)
1. ✅ Setup MinIO locally with Docker
2. ✅ Create video model and migrations
3. ✅ Create upload session model
4. ✅ Setup S3/MinIO client
5. ✅ Create storage service wrapper

### Phase 2: Upload Handlers (Next)
6. ✅ Implement initiate upload endpoint
7. ✅ Implement chunk upload endpoint
8. ✅ Implement complete upload endpoint
9. ✅ Add validation and error handling

### Phase 3: Metadata & Storage (After)
10. ✅ Install FFmpeg/FFprobe
11. ✅ Implement metadata extraction
12. ✅ Upload to MinIO after assembly
13. ✅ Generate thumbnail from video

### Phase 4: Testing & Polish
14. ✅ Add Swagger documentation
15. ✅ Write integration tests
16. ✅ Add progress tracking
17. ✅ Error recovery mechanisms

---

## 🎯 Expected Features

After completion, you'll be able to:
- ✅ Upload videos up to 10GB
- ✅ Resume failed uploads
- ✅ See upload progress in real-time
- ✅ Automatically extract video metadata
- ✅ Store videos in MinIO (S3-compatible)
- ✅ Link videos to movies
- ✅ Generate video thumbnails
- ✅ Stream videos (basic, HLS in next phase)

---

## 📝 Example Usage Flow

```javascript
// Frontend code example
const uploadVideo = async (file, movieId) => {
  // 1. Initiate upload
  const { sessionId, chunkSize } = await fetch('/videos/upload/initiate', {
    method: 'POST',
    body: JSON.stringify({
      fileName: file.name,
      fileSize: file.size,
      movieId: movieId
    })
  }).then(r => r.json());

  // 2. Upload chunks
  const totalChunks = Math.ceil(file.size / chunkSize);
  for (let i = 0; i < totalChunks; i++) {
    const chunk = file.slice(i * chunkSize, (i + 1) * chunkSize);
    await fetch(`/videos/upload/chunk`, {
      method: 'POST',
      headers: {
        'X-Session-ID': sessionId,
        'X-Chunk-Index': i,
        'X-Total-Chunks': totalChunks
      },
      body: chunk
    });

    // Update progress: (i + 1) / totalChunks * 100
  }

  // 3. Complete upload
  const video = await fetch('/videos/upload/complete', {
    method: 'POST',
    body: JSON.stringify({ sessionId })
  }).then(r => r.json());

  console.log('Upload complete!', video);
};
```

---

## 🔍 Testing Strategy

1. **Unit Tests:**
   - Chunk assembly logic
   - Metadata extraction
   - Storage operations

2. **Integration Tests:**
   - Full upload flow
   - Resume interrupted upload
   - Error handling

3. **Manual Testing:**
   - Upload small video (< 100MB)
   - Upload large video (> 1GB)
   - Test with different formats
   - Test concurrent uploads

---

## 💡 Future Enhancements (Phase 2)

After basic upload works:
- Transcoding to multiple qualities
- HLS/DASH streaming
- Thumbnail generation at intervals
- Video preview generation
- Automatic backup to S3
- CDN integration

---

Let's start building! 🚀
