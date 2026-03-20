---
name: file-uploads
description: >
  File upload patterns — drag-and-drop, presigned URLs, progress tracking,
  image previews, chunked uploads, and validation.
  Triggers: "file upload", "drag and drop", "presigned URL", "image preview",
  "upload progress", "multipart upload", "dropzone".
  NOT for: form layout or validation (use react-hook-form or zod-schemas).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# File Uploads

## Quick Start

```bash
npm install react-dropzone
```

## Basic File Input with React Hook Form

```tsx
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';

const MAX_FILE_SIZE = 5 * 1024 * 1024; // 5 MB
const ACCEPTED_TYPES = ['image/jpeg', 'image/png', 'image/webp'];

const schema = z.object({
  name: z.string().min(1),
  avatar: z
    .instanceof(File, { message: 'Please upload a file' })
    .refine((f) => f.size <= MAX_FILE_SIZE, 'File must be under 5 MB')
    .refine((f) => ACCEPTED_TYPES.includes(f.type), 'Only JPEG, PNG, and WebP'),
});

type FormValues = z.infer<typeof schema>;

function ProfileForm() {
  const { register, handleSubmit, formState: { errors }, setValue, watch } = useForm<FormValues>({
    resolver: zodResolver(schema),
  });

  const avatar = watch('avatar');

  const onSubmit = async (data: FormValues) => {
    const formData = new FormData();
    formData.append('name', data.name);
    formData.append('avatar', data.avatar);
    await fetch('/api/profile', { method: 'POST', body: formData });
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register('name')} />

      <input
        type="file"
        accept={ACCEPTED_TYPES.join(',')}
        onChange={(e) => {
          const file = e.target.files?.[0];
          if (file) setValue('avatar', file, { shouldValidate: true });
        }}
      />
      {avatar && (
        <img
          src={URL.createObjectURL(avatar)}
          alt="Preview"
          className="h-20 w-20 rounded-full object-cover"
        />
      )}
      {errors.avatar && <p className="text-red-500">{errors.avatar.message}</p>}

      <button type="submit">Save</button>
    </form>
  );
}
```

## Drag & Drop with react-dropzone

```tsx
import { useDropzone } from 'react-dropzone';
import { useCallback, useState } from 'react';

function FileUploader() {
  const [files, setFiles] = useState<File[]>([]);

  const onDrop = useCallback((acceptedFiles: File[]) => {
    setFiles((prev) => [...prev, ...acceptedFiles]);
  }, []);

  const { getRootProps, getInputProps, isDragActive, isDragReject } = useDropzone({
    onDrop,
    accept: {
      'image/*': ['.jpeg', '.jpg', '.png', '.webp'],
      'application/pdf': ['.pdf'],
    },
    maxSize: 10 * 1024 * 1024, // 10 MB
    maxFiles: 5,
  });

  const removeFile = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index));
  };

  return (
    <div>
      <div
        {...getRootProps()}
        className={cn(
          'border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors',
          isDragActive && 'border-blue-500 bg-blue-50',
          isDragReject && 'border-red-500 bg-red-50',
          !isDragActive && 'border-gray-300 hover:border-gray-400',
        )}
      >
        <input {...getInputProps()} />
        {isDragActive ? (
          <p>Drop files here...</p>
        ) : (
          <p>Drag & drop files here, or click to browse</p>
        )}
        <p className="text-sm text-gray-500 mt-1">
          PNG, JPG, WebP, or PDF up to 10 MB (max 5 files)
        </p>
      </div>

      {files.length > 0 && (
        <ul className="mt-4 space-y-2">
          {files.map((file, i) => (
            <li key={i} className="flex items-center gap-2">
              {file.type.startsWith('image/') && (
                <img
                  src={URL.createObjectURL(file)}
                  alt=""
                  className="h-10 w-10 object-cover rounded"
                />
              )}
              <span className="text-sm">{file.name}</span>
              <span className="text-xs text-gray-400">
                ({(file.size / 1024).toFixed(0)} KB)
              </span>
              <button onClick={() => removeFile(i)} className="text-red-500 text-sm">
                Remove
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
```

## Dropzone + React Hook Form

```tsx
import { useDropzone } from 'react-dropzone';
import { Controller, useForm } from 'react-hook-form';

function UploadForm() {
  const { control, handleSubmit } = useForm<{ files: File[] }>({
    defaultValues: { files: [] },
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Controller
        name="files"
        control={control}
        rules={{ required: 'Please upload at least one file' }}
        render={({ field: { onChange, value }, fieldState: { error } }) => (
          <Dropzone
            files={value}
            onDrop={(accepted) => onChange([...value, ...accepted])}
            onRemove={(index) => onChange(value.filter((_, i) => i !== index))}
            error={error?.message}
          />
        )}
      />
      <button type="submit">Upload</button>
    </form>
  );
}
```

## Upload with Progress (XMLHttpRequest)

```tsx
function useFileUpload() {
  const [progress, setProgress] = useState(0);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const upload = useCallback(async (file: File, url: string) => {
    setUploading(true);
    setProgress(0);
    setError(null);

    return new Promise<string>((resolve, reject) => {
      const xhr = new XMLHttpRequest();

      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
          setProgress(Math.round((e.loaded / e.total) * 100));
        }
      });

      xhr.addEventListener('load', () => {
        setUploading(false);
        if (xhr.status >= 200 && xhr.status < 300) {
          resolve(xhr.responseText);
        } else {
          const msg = `Upload failed (${xhr.status})`;
          setError(msg);
          reject(new Error(msg));
        }
      });

      xhr.addEventListener('error', () => {
        setUploading(false);
        const msg = 'Upload failed — check your connection';
        setError(msg);
        reject(new Error(msg));
      });

      const formData = new FormData();
      formData.append('file', file);

      xhr.open('POST', url);
      xhr.send(formData);
    });
  }, []);

  return { upload, progress, uploading, error };
}

// Usage
function UploadButton() {
  const { upload, progress, uploading } = useFileUpload();

  return (
    <div>
      <input
        type="file"
        onChange={async (e) => {
          const file = e.target.files?.[0];
          if (file) await upload(file, '/api/upload');
        }}
        disabled={uploading}
      />
      {uploading && (
        <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
          <div
            className="bg-blue-600 h-2 rounded-full transition-all"
            style={{ width: `${progress}%` }}
          />
        </div>
      )}
    </div>
  );
}
```

## Presigned URL Upload (S3/R2/GCS)

```typescript
// Server — generate presigned URL
import { S3Client, PutObjectCommand } from '@aws-sdk/client-s3';
import { getSignedUrl } from '@aws-sdk/s3-request-presigner';

const s3 = new S3Client({ region: process.env.AWS_REGION });

app.post('/api/upload/presign', async (req, res) => {
  const { filename, contentType } = req.body;
  const key = `uploads/${Date.now()}-${filename}`;

  const command = new PutObjectCommand({
    Bucket: process.env.S3_BUCKET,
    Key: key,
    ContentType: contentType,
  });

  const url = await getSignedUrl(s3, command, { expiresIn: 300 }); // 5 min

  res.json({ url, key });
});
```

```tsx
// Client — upload directly to S3
async function uploadToS3(file: File) {
  // 1. Get presigned URL from your server
  const { url, key } = await fetch('/api/upload/presign', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ filename: file.name, contentType: file.type }),
  }).then((r) => r.json());

  // 2. Upload directly to S3
  await fetch(url, {
    method: 'PUT',
    body: file,
    headers: { 'Content-Type': file.type },
  });

  // 3. Return the key for saving in your database
  return key;
}
```

## Server-Side: Express + Multer

```typescript
import multer from 'multer';
import path from 'path';

const upload = multer({
  storage: multer.memoryStorage(),
  limits: {
    fileSize: 10 * 1024 * 1024, // 10 MB
    files: 5,
  },
  fileFilter: (_req, file, cb) => {
    const allowed = ['.jpg', '.jpeg', '.png', '.webp', '.pdf'];
    const ext = path.extname(file.originalname).toLowerCase();
    if (allowed.includes(ext)) {
      cb(null, true);
    } else {
      cb(new Error(`File type ${ext} not allowed`));
    }
  },
});

// Single file
app.post('/api/upload', upload.single('file'), (req, res) => {
  const file = req.file!;
  res.json({ filename: file.originalname, size: file.size });
});

// Multiple files
app.post('/api/upload/bulk', upload.array('files', 5), (req, res) => {
  const files = req.files as Express.Multer.File[];
  res.json({ count: files.length });
});

// Error handling
app.use((err: Error, req: any, res: any, next: any) => {
  if (err instanceof multer.MulterError) {
    if (err.code === 'LIMIT_FILE_SIZE') {
      return res.status(413).json({ error: 'File too large (max 10 MB)' });
    }
    if (err.code === 'LIMIT_FILE_COUNT') {
      return res.status(400).json({ error: 'Too many files (max 5)' });
    }
  }
  next(err);
});
```

## Image Preview and Compression

```tsx
function useImagePreview(file: File | null) {
  const [preview, setPreview] = useState<string | null>(null);

  useEffect(() => {
    if (!file) { setPreview(null); return; }
    const url = URL.createObjectURL(file);
    setPreview(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);

  return preview;
}

// Client-side image compression before upload
async function compressImage(file: File, maxWidth = 1200, quality = 0.8): Promise<Blob> {
  const img = new Image();
  const canvas = document.createElement('canvas');

  return new Promise((resolve) => {
    img.onload = () => {
      const scale = Math.min(1, maxWidth / img.width);
      canvas.width = img.width * scale;
      canvas.height = img.height * scale;

      const ctx = canvas.getContext('2d')!;
      ctx.drawImage(img, 0, 0, canvas.width, canvas.height);

      canvas.toBlob((blob) => resolve(blob!), 'image/jpeg', quality);
    };
    img.src = URL.createObjectURL(file);
  });
}
```

## Chunked Upload (Large Files)

```tsx
async function uploadInChunks(file: File, url: string, chunkSize = 5 * 1024 * 1024) {
  const totalChunks = Math.ceil(file.size / chunkSize);
  const uploadId = crypto.randomUUID();

  for (let i = 0; i < totalChunks; i++) {
    const start = i * chunkSize;
    const end = Math.min(start + chunkSize, file.size);
    const chunk = file.slice(start, end);

    const formData = new FormData();
    formData.append('chunk', chunk);
    formData.append('uploadId', uploadId);
    formData.append('chunkIndex', String(i));
    formData.append('totalChunks', String(totalChunks));
    formData.append('filename', file.name);

    const response = await fetch(url, { method: 'POST', body: formData });
    if (!response.ok) throw new Error(`Chunk ${i} failed`);
  }

  await fetch(`${url}/complete`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ uploadId, filename: file.name, totalChunks }),
  });
}
```

## Zod File Validation Schemas

```typescript
const fileSchema = {
  image: (maxMB = 5) =>
    z.instanceof(File)
      .refine((f) => f.size > 0, 'File is empty')
      .refine((f) => f.size <= maxMB * 1024 * 1024, `Must be under ${maxMB} MB`)
      .refine(
        (f) => ['image/jpeg', 'image/png', 'image/webp'].includes(f.type),
        'Only JPEG, PNG, and WebP images'
      ),

  pdf: (maxMB = 20) =>
    z.instanceof(File)
      .refine((f) => f.size <= maxMB * 1024 * 1024, `Must be under ${maxMB} MB`)
      .refine((f) => f.type === 'application/pdf', 'Must be a PDF'),

  any: (maxMB = 10, types?: string[]) =>
    z.instanceof(File)
      .refine((f) => f.size <= maxMB * 1024 * 1024, `Must be under ${maxMB} MB`)
      .refine(
        (f) => !types || types.includes(f.type),
        `Allowed types: ${types?.join(', ')}`
      ),

  images: (maxFiles = 5, maxMB = 5) =>
    z.array(z.instanceof(File))
      .min(1, 'Upload at least one image')
      .max(maxFiles, `Maximum ${maxFiles} images`)
      .refine(
        (files) => files.every((f) => f.size <= maxMB * 1024 * 1024),
        `Each file must be under ${maxMB} MB`
      )
      .refine(
        (files) => files.every((f) => f.type.startsWith('image/')),
        'All files must be images'
      ),
};
```

## Gotchas

1. **Memory leaks with `URL.createObjectURL`.** Always revoke with `URL.revokeObjectURL()` in a cleanup function or when the preview is no longer needed.

2. **`<input type="file">` is uncontrolled.** You can't set its value programmatically (security restriction). Use `onChange` + `setValue` with RHF, not `register` directly.

3. **FormData for multipart.** Don't set `Content-Type` header manually when using `FormData` — the browser sets the correct `multipart/form-data` boundary automatically.

4. **Presigned URLs expire.** Generate them right before upload, not when the form loads. 5 minutes is a safe default.

5. **Mobile file inputs.** On iOS, `accept="image/*"` opens the camera by default. Add `capture` attribute to control: `capture="environment"` for back camera, omit for photo library.

6. **Large file uploads need chunking.** Browsers and servers have upload size limits. For files over 50 MB, use chunked uploads or presigned multipart uploads.
