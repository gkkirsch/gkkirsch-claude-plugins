---
name: file-uploads
description: >
  Implement file upload functionality — drag-and-drop, presigned URLs,
  multipart uploads, progress tracking, validation, and chunked resumable
  uploads. Covers both client and server implementations.
  Triggers: "file upload", "drag and drop upload", "presigned url",
  "multipart upload", "upload component", "file input", "dropzone".
  NOT for: image processing (use image-processing) or storage setup (use cloud-storage).
version: 1.0.0
argument-hint: "[presigned|multipart|dropzone|resumable]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# File Uploads

Implement file upload functionality with validation, progress tracking, and error handling.

## Upload Strategy Selection

| Strategy | File Size | Direct to Storage? | Progress? | Resumable? |
|----------|-----------|-------------------|-----------|------------|
| Form POST | < 10MB | No (through server) | Limited | No |
| Presigned URL | < 5GB | Yes (S3 direct) | Yes | No |
| Multipart (S3) | > 100MB | Yes | Per-part | Yes |
| tus protocol | Any size | Through tus server | Yes | Yes |

## React Dropzone Component

```bash
npm install react-dropzone
```

```tsx
// components/FileUpload.tsx
'use client';
import { useCallback, useState } from 'react';
import { useDropzone } from 'react-dropzone';

interface UploadedFile {
  id: string;
  name: string;
  size: number;
  url: string;
  type: string;
}

interface Props {
  accept?: Record<string, string[]>;
  maxSize?: number;          // bytes
  maxFiles?: number;
  onUpload: (files: UploadedFile[]) => void;
}

export function FileUpload({
  accept = { 'image/*': ['.jpg', '.jpeg', '.png', '.webp'] },
  maxSize = 10 * 1024 * 1024, // 10MB
  maxFiles = 5,
  onUpload,
}: Props) {
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState<Record<string, number>>({});
  const [errors, setErrors] = useState<string[]>([]);

  const uploadFile = async (file: File): Promise<UploadedFile> => {
    // 1. Get presigned URL from your API
    const res = await fetch('/api/uploads/presign', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        filename: file.name,
        contentType: file.type,
        size: file.size,
      }),
    });

    const { uploadUrl, fileId, publicUrl } = await res.json();

    // 2. Upload directly to storage with progress
    await new Promise<void>((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      xhr.open('PUT', uploadUrl);
      xhr.setRequestHeader('Content-Type', file.type);

      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) {
          setProgress(prev => ({ ...prev, [file.name]: (e.loaded / e.total) * 100 }));
        }
      };

      xhr.onload = () => (xhr.status < 400 ? resolve() : reject(new Error(`Upload failed: ${xhr.status}`)));
      xhr.onerror = () => reject(new Error('Upload failed'));
      xhr.send(file);
    });

    // 3. Confirm upload with your API
    await fetch('/api/uploads/confirm', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ fileId }),
    });

    return { id: fileId, name: file.name, size: file.size, url: publicUrl, type: file.type };
  };

  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    setUploading(true);
    setErrors([]);

    try {
      const results = await Promise.all(acceptedFiles.map(uploadFile));
      onUpload(results);
    } catch (err) {
      setErrors(prev => [...prev, (err as Error).message]);
    } finally {
      setUploading(false);
      setProgress({});
    }
  }, [onUpload]);

  const { getRootProps, getInputProps, isDragActive, fileRejections } = useDropzone({
    onDrop,
    accept,
    maxSize,
    maxFiles,
    disabled: uploading,
  });

  return (
    <div>
      <div
        {...getRootProps()}
        className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors
          ${isDragActive ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-gray-400'}
          ${uploading ? 'opacity-50 cursor-not-allowed' : ''}
        `}
      >
        <input {...getInputProps()} />
        <div className="space-y-2">
          <svg className="mx-auto h-12 w-12 text-gray-400" stroke="currentColor" fill="none" viewBox="0 0 48 48">
            <path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" />
          </svg>
          <p className="text-sm text-gray-600">
            {isDragActive ? 'Drop files here' : 'Drag & drop files, or click to browse'}
          </p>
          <p className="text-xs text-gray-500">
            Max {maxFiles} files, {(maxSize / 1024 / 1024).toFixed(0)}MB each
          </p>
        </div>
      </div>

      {/* Progress bars */}
      {Object.entries(progress).map(([name, pct]) => (
        <div key={name} className="mt-2">
          <div className="flex justify-between text-sm">
            <span className="truncate">{name}</span>
            <span>{Math.round(pct)}%</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div className="bg-blue-600 h-2 rounded-full transition-all" style={{ width: `${pct}%` }} />
          </div>
        </div>
      ))}

      {/* Errors */}
      {errors.map((err, i) => (
        <p key={i} className="mt-2 text-sm text-red-600">{err}</p>
      ))}

      {/* Rejection errors */}
      {fileRejections.map(({ file, errors }) => (
        <p key={file.name} className="mt-1 text-sm text-red-600">
          {file.name}: {errors.map(e => e.message).join(', ')}
        </p>
      ))}
    </div>
  );
}
```

## Server: Presigned URL Endpoint

```typescript
// api/uploads.ts
import { S3Client, PutObjectCommand, GetObjectCommand } from '@aws-sdk/client-s3';
import { getSignedUrl } from '@aws-sdk/s3-request-presigner';
import { randomUUID } from 'crypto';

const s3 = new S3Client({
  region: process.env.AWS_REGION!,
  credentials: {
    accessKeyId: process.env.AWS_ACCESS_KEY_ID!,
    secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY!,
  },
});

const BUCKET = process.env.S3_BUCKET!;

// Allowed MIME types
const ALLOWED_TYPES = new Set([
  'image/jpeg', 'image/png', 'image/webp', 'image/gif',
  'application/pdf',
  'video/mp4', 'video/webm',
]);

const MAX_SIZE = 50 * 1024 * 1024; // 50MB

// POST /api/uploads/presign
app.post('/api/uploads/presign', async (req, res) => {
  const { filename, contentType, size } = req.body;
  const userId = req.user!.id;

  // Validate
  if (!ALLOWED_TYPES.has(contentType)) {
    return res.status(400).json({ error: 'File type not allowed' });
  }
  if (size > MAX_SIZE) {
    return res.status(400).json({ error: 'File too large' });
  }

  // Generate unique key
  const ext = filename.split('.').pop()?.toLowerCase() || '';
  const fileId = randomUUID();
  const key = `uploads/${userId}/${fileId}.${ext}`;

  // Create presigned PUT URL (expires in 15 minutes)
  const command = new PutObjectCommand({
    Bucket: BUCKET,
    Key: key,
    ContentType: contentType,
    ContentLength: size,
    Metadata: {
      'uploaded-by': userId,
      'original-name': filename,
    },
  });

  const uploadUrl = await getSignedUrl(s3, command, { expiresIn: 900 });

  // Store pending upload in database
  await prisma.upload.create({
    data: {
      id: fileId,
      userId,
      key,
      filename,
      contentType,
      size,
      status: 'PENDING',
    },
  });

  res.json({
    uploadUrl,
    fileId,
    publicUrl: `${process.env.CDN_URL}/${key}`,
  });
});

// POST /api/uploads/confirm
app.post('/api/uploads/confirm', async (req, res) => {
  const { fileId } = req.body;
  const userId = req.user!.id;

  const upload = await prisma.upload.findFirst({
    where: { id: fileId, userId, status: 'PENDING' },
  });

  if (!upload) {
    return res.status(404).json({ error: 'Upload not found' });
  }

  // Verify file exists in S3
  try {
    await s3.send(new GetObjectCommand({ Bucket: BUCKET, Key: upload.key }));
  } catch {
    return res.status(400).json({ error: 'File not found in storage' });
  }

  // Mark as confirmed
  await prisma.upload.update({
    where: { id: fileId },
    data: { status: 'CONFIRMED' },
  });

  res.json({ success: true, url: `${process.env.CDN_URL}/${upload.key}` });
});
```

## Server-Side Upload (Express + Multer)

```bash
npm install multer @types/multer
```

```typescript
import multer from 'multer';
import { S3Client, PutObjectCommand } from '@aws-sdk/client-s3';

// Memory storage (streams to S3)
const upload = multer({
  storage: multer.memoryStorage(),
  limits: { fileSize: 10 * 1024 * 1024 }, // 10MB
  fileFilter: (req, file, cb) => {
    const allowed = ['image/jpeg', 'image/png', 'image/webp', 'application/pdf'];
    if (allowed.includes(file.mimetype)) {
      cb(null, true);
    } else {
      cb(new Error('File type not allowed'));
    }
  },
});

// Single file upload
app.post('/api/upload', upload.single('file'), async (req, res) => {
  if (!req.file) return res.status(400).json({ error: 'No file provided' });

  const key = `uploads/${req.user!.id}/${randomUUID()}-${req.file.originalname}`;

  await s3.send(new PutObjectCommand({
    Bucket: BUCKET,
    Key: key,
    Body: req.file.buffer,
    ContentType: req.file.mimetype,
  }));

  res.json({ url: `${process.env.CDN_URL}/${key}` });
});

// Multiple files
app.post('/api/upload-multiple', upload.array('files', 10), async (req, res) => {
  const files = req.files as Express.Multer.File[];
  const urls = await Promise.all(files.map(async (file) => {
    const key = `uploads/${req.user!.id}/${randomUUID()}-${file.originalname}`;
    await s3.send(new PutObjectCommand({
      Bucket: BUCKET, Key: key, Body: file.buffer, ContentType: file.mimetype,
    }));
    return `${process.env.CDN_URL}/${key}`;
  }));
  res.json({ urls });
});
```

## File Validation (Magic Bytes)

```typescript
// lib/file-validation.ts

// File signature (magic bytes) verification
const FILE_SIGNATURES: Record<string, number[][]> = {
  'image/jpeg': [[0xFF, 0xD8, 0xFF]],
  'image/png': [[0x89, 0x50, 0x4E, 0x47]],
  'image/gif': [[0x47, 0x49, 0x46, 0x38]],
  'image/webp': [[0x52, 0x49, 0x46, 0x46]], // RIFF header
  'application/pdf': [[0x25, 0x50, 0x44, 0x46]], // %PDF
  'video/mp4': [[0x00, 0x00, 0x00], [0x66, 0x74, 0x79, 0x70]], // ftyp (offset 4)
};

export function validateFileType(buffer: Buffer, claimedType: string): boolean {
  const signatures = FILE_SIGNATURES[claimedType];
  if (!signatures) return false;

  return signatures.some(sig =>
    sig.every((byte, i) => buffer[i] === byte)
  );
}

// Usage
app.post('/api/upload', upload.single('file'), async (req, res) => {
  if (!req.file) return res.status(400).json({ error: 'No file' });

  // Validate magic bytes match claimed MIME type
  if (!validateFileType(req.file.buffer, req.file.mimetype)) {
    return res.status(400).json({ error: 'File content does not match type' });
  }

  // Proceed with upload...
});
```

## Chunked/Resumable Uploads (tus Protocol)

```bash
npm install tus-js-client        # Client
npm install @tus/server @tus/s3-store  # Server
```

```typescript
// Server: tus upload endpoint
import { Server as TusServer } from '@tus/server';
import { S3Store } from '@tus/s3-store';

const tusServer = new TusServer({
  path: '/api/uploads/tus',
  datastore: new S3Store({
    bucket: BUCKET,
    s3ClientConfig: {
      region: process.env.AWS_REGION!,
      credentials: {
        accessKeyId: process.env.AWS_ACCESS_KEY_ID!,
        secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY!,
      },
    },
  }),
  maxSize: 5 * 1024 * 1024 * 1024, // 5GB
  generateUrl: (req, { id }) => `${req.headers.origin}/api/uploads/tus/${id}`,
});

// Mount with Express
app.all('/api/uploads/tus/*', (req, res) => tusServer.handle(req, res));
```

```typescript
// Client: resumable upload with tus
import * as tus from 'tus-js-client';

function resumableUpload(file: File, onProgress: (pct: number) => void): Promise<string> {
  return new Promise((resolve, reject) => {
    const upload = new tus.Upload(file, {
      endpoint: '/api/uploads/tus/',
      retryDelays: [0, 3000, 5000, 10000, 20000],
      chunkSize: 5 * 1024 * 1024, // 5MB chunks
      metadata: {
        filename: file.name,
        filetype: file.type,
      },
      onProgress: (bytesUploaded, bytesTotal) => {
        onProgress((bytesUploaded / bytesTotal) * 100);
      },
      onSuccess: () => {
        resolve(upload.url!);
      },
      onError: (error) => {
        reject(error);
      },
    });

    // Check for previous uploads to resume
    upload.findPreviousUploads().then((previousUploads) => {
      if (previousUploads.length > 0) {
        upload.resumeFromPreviousUpload(previousUploads[0]);
      }
      upload.start();
    });
  });
}
```

## Best Practices

1. **Presigned URLs for direct uploads** — keep file data off your server
2. **Validate on both client AND server** — client validation is for UX, server is for security
3. **Check magic bytes, not just extensions** — users can rename files
4. **Generate random filenames** — never use user-provided names for storage keys
5. **Set appropriate Content-Type** — prevents browser misinterpretation
6. **Add Content-Disposition** for downloads — `attachment; filename="original.pdf"`
7. **Clean up failed uploads** — lifecycle rules to delete incomplete multipart uploads
8. **Show progress for large files** — XHR `upload.onprogress` or tus for resumable
9. **Handle network failures** — retry with exponential backoff, or use tus for resumability
10. **Virus scan user uploads** — ClamAV, VirusTotal API, or AWS Macie
