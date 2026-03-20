---
name: express-api-patterns
description: >
  Production Express.js API patterns — routing, middleware, error handling,
  validation, authentication, file uploads, WebSockets, and graceful shutdown.
  Triggers: "express api", "express middleware", "express routes",
  "express error handling", "express validation", "express file upload".
  NOT for: Frontend React patterns (use react-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Express.js API Patterns

## Project Structure

```
src/
  index.ts          # Entry point, server startup
  app.ts            # Express app configuration
  routes/
    index.ts        # Route aggregator
    users.routes.ts
    posts.routes.ts
  controllers/
    users.controller.ts
    posts.controller.ts
  middleware/
    auth.ts
    validate.ts
    errorHandler.ts
    rateLimiter.ts
  services/
    users.service.ts
    posts.service.ts
  models/
    user.model.ts
    post.model.ts
  utils/
    AppError.ts
    asyncHandler.ts
    logger.ts
  types/
    express.d.ts    # Request type extensions
```

## App Configuration

```typescript
// src/app.ts
import express from "express";
import cors from "cors";
import helmet from "helmet";
import compression from "compression";
import { errorHandler, notFoundHandler } from "./middleware/errorHandler";
import { apiLimiter } from "./middleware/rateLimiter";
import routes from "./routes";

const app = express();

// Security & parsing
app.use(helmet());
app.use(cors({ origin: process.env.ALLOWED_ORIGINS?.split(","), credentials: true }));
app.use(compression());
app.use(express.json({ limit: "10kb" }));
app.use(express.urlencoded({ extended: true, limit: "10kb" }));

// Rate limiting
app.use("/api", apiLimiter);

// Health check (before auth)
app.get("/health", (_, res) => res.json({ status: "ok", uptime: process.uptime() }));

// Routes
app.use("/api", routes);

// Error handling (must be last)
app.use(notFoundHandler);
app.use(errorHandler);

export default app;
```

```typescript
// src/index.ts
import app from "./app";
import { logger } from "./utils/logger";

const PORT = process.env.PORT || 3000;

const server = app.listen(PORT, () => {
  logger.info(`Server running on port ${PORT}`);
});

// Graceful shutdown
function shutdown(signal: string) {
  logger.info(`${signal} received. Starting graceful shutdown...`);
  server.close(() => {
    logger.info("HTTP server closed");
    // Close DB connections, Redis, etc.
    process.exit(0);
  });
  // Force close after 10s
  setTimeout(() => {
    logger.error("Forced shutdown after timeout");
    process.exit(1);
  }, 10_000);
}

process.on("SIGTERM", () => shutdown("SIGTERM"));
process.on("SIGINT", () => shutdown("SIGINT"));
process.on("unhandledRejection", (reason) => {
  logger.error("Unhandled rejection:", reason);
  // Don't exit — let the error handler deal with it
});
```

## Custom Error Class

```typescript
// src/utils/AppError.ts
export class AppError extends Error {
  constructor(
    public statusCode: number,
    message: string,
    public code?: string,
    public isOperational = true
  ) {
    super(message);
    Object.setPrototypeOf(this, AppError.prototype);
  }

  static badRequest(message: string, code?: string) {
    return new AppError(400, message, code);
  }
  static unauthorized(message = "Not authenticated") {
    return new AppError(401, message, "UNAUTHORIZED");
  }
  static forbidden(message = "Not authorized") {
    return new AppError(403, message, "FORBIDDEN");
  }
  static notFound(resource = "Resource") {
    return new AppError(404, `${resource} not found`, "NOT_FOUND");
  }
  static conflict(message: string) {
    return new AppError(409, message, "CONFLICT");
  }
  static tooMany(message = "Too many requests") {
    return new AppError(429, message, "RATE_LIMITED");
  }
}
```

## Error Handling Middleware

```typescript
// src/middleware/errorHandler.ts
import { Request, Response, NextFunction } from "express";
import { AppError } from "../utils/AppError";
import { ZodError } from "zod";
import { logger } from "../utils/logger";

export function errorHandler(err: Error, req: Request, res: Response, _next: NextFunction) {
  // Known operational error
  if (err instanceof AppError) {
    return res.status(err.statusCode).json({
      error: { message: err.message, code: err.code },
    });
  }

  // Zod validation error
  if (err instanceof ZodError) {
    return res.status(400).json({
      error: {
        message: "Validation failed",
        code: "VALIDATION_ERROR",
        details: err.errors.map((e) => ({
          path: e.path.join("."),
          message: e.message,
        })),
      },
    });
  }

  // Prisma unique constraint
  if (err.constructor.name === "PrismaClientKnownRequestError" && (err as any).code === "P2002") {
    const field = (err as any).meta?.target?.[0] || "field";
    return res.status(409).json({
      error: { message: `${field} already exists`, code: "DUPLICATE" },
    });
  }

  // Unknown error
  logger.error("Unhandled error:", err);
  res.status(500).json({
    error: {
      message: process.env.NODE_ENV === "production" ? "Internal server error" : err.message,
      code: "INTERNAL_ERROR",
    },
  });
}

export function notFoundHandler(req: Request, res: Response) {
  res.status(404).json({
    error: { message: `Cannot ${req.method} ${req.path}`, code: "NOT_FOUND" },
  });
}
```

## Async Handler

```typescript
// src/utils/asyncHandler.ts
import { Request, Response, NextFunction, RequestHandler } from "express";

export function asyncHandler(fn: (req: Request, res: Response, next: NextFunction) => Promise<any>): RequestHandler {
  return (req, res, next) => {
    Promise.resolve(fn(req, res, next)).catch(next);
  };
}
```

## Validation Middleware

```typescript
// src/middleware/validate.ts
import { Request, Response, NextFunction } from "express";
import { z, ZodSchema } from "zod";

type ValidationTarget = "body" | "query" | "params";

export function validate(schema: ZodSchema, target: ValidationTarget = "body") {
  return (req: Request, _res: Response, next: NextFunction) => {
    const result = schema.safeParse(req[target]);
    if (!result.success) {
      next(result.error); // Caught by errorHandler
      return;
    }
    req[target] = result.data; // Replace with parsed + transformed data
    next();
  };
}

// Reusable schemas
export const paginationSchema = z.object({
  page: z.coerce.number().int().positive().default(1),
  limit: z.coerce.number().int().min(1).max(100).default(20),
  sort: z.enum(["createdAt", "updatedAt", "name"]).default("createdAt"),
  order: z.enum(["asc", "desc"]).default("desc"),
});

export const idParamSchema = z.object({
  id: z.string().uuid("Invalid ID format"),
});
```

## Route + Controller Pattern

```typescript
// src/routes/users.routes.ts
import { Router } from "express";
import * as users from "../controllers/users.controller";
import { validate } from "../middleware/validate";
import { requireAuth, requireRole } from "../middleware/auth";
import { idParamSchema, paginationSchema } from "../middleware/validate";
import { z } from "zod";

const router = Router();

const createUserSchema = z.object({
  name: z.string().min(2).max(100),
  email: z.string().email(),
  role: z.enum(["user", "admin"]).default("user"),
});

const updateUserSchema = createUserSchema.partial();

router.get("/",    requireAuth, validate(paginationSchema, "query"), users.list);
router.get("/:id", requireAuth, validate(idParamSchema, "params"),   users.getById);
router.post("/",   requireAuth, requireRole("admin"), validate(createUserSchema), users.create);
router.patch("/:id", requireAuth, validate(idParamSchema, "params"), validate(updateUserSchema), users.update);
router.delete("/:id", requireAuth, requireRole("admin"), validate(idParamSchema, "params"), users.remove);

export default router;
```

```typescript
// src/controllers/users.controller.ts
import { Request, Response } from "express";
import { asyncHandler } from "../utils/asyncHandler";
import { AppError } from "../utils/AppError";
import * as userService from "../services/users.service";

export const list = asyncHandler(async (req: Request, res: Response) => {
  const { page, limit, sort, order } = req.query as any;
  const { data, total } = await userService.findMany({ page, limit, sort, order });

  res.json({
    data,
    meta: {
      page, limit, total,
      totalPages: Math.ceil(total / limit),
    },
  });
});

export const getById = asyncHandler(async (req: Request, res: Response) => {
  const user = await userService.findById(req.params.id);
  if (!user) throw AppError.notFound("User");
  res.json({ data: user });
});

export const create = asyncHandler(async (req: Request, res: Response) => {
  const user = await userService.create(req.body);
  res.status(201).json({ data: user });
});

export const update = asyncHandler(async (req: Request, res: Response) => {
  const user = await userService.update(req.params.id, req.body);
  if (!user) throw AppError.notFound("User");
  res.json({ data: user });
});

export const remove = asyncHandler(async (req: Request, res: Response) => {
  await userService.remove(req.params.id);
  res.status(204).end();
});
```

## File Upload

```typescript
import multer from "multer";
import path from "path";
import crypto from "crypto";
import { AppError } from "../utils/AppError";

const storage = multer.diskStorage({
  destination: "uploads/",
  filename: (_, file, cb) => {
    const unique = crypto.randomBytes(16).toString("hex");
    cb(null, `${unique}${path.extname(file.originalname)}`);
  },
});

const upload = multer({
  storage,
  limits: { fileSize: 5 * 1024 * 1024 }, // 5MB
  fileFilter: (_, file, cb) => {
    const allowed = [".jpg", ".jpeg", ".png", ".webp", ".pdf"];
    const ext = path.extname(file.originalname).toLowerCase();
    if (allowed.includes(ext)) {
      cb(null, true);
    } else {
      cb(new AppError(400, `File type ${ext} not allowed`));
    }
  },
});

// Single file
router.post("/avatar", requireAuth, upload.single("avatar"), asyncHandler(async (req, res) => {
  if (!req.file) throw AppError.badRequest("No file uploaded");
  const url = `/uploads/${req.file.filename}`;
  await userService.updateAvatar(req.user!.id, url);
  res.json({ data: { url } });
}));

// Multiple files
router.post("/gallery", requireAuth, upload.array("images", 10), asyncHandler(async (req, res) => {
  const files = (req.files as Express.Multer.File[]) || [];
  const urls = files.map((f) => `/uploads/${f.filename}`);
  res.json({ data: { urls } });
}));
```

## WebSocket Integration

```typescript
import { WebSocketServer, WebSocket } from "ws";
import { Server } from "http";

export function setupWebSocket(server: Server) {
  const wss = new WebSocketServer({ server, path: "/ws" });

  const rooms = new Map<string, Set<WebSocket>>();

  wss.on("connection", (ws, req) => {
    const userId = authenticateWs(req); // Verify token from query string
    if (!userId) { ws.close(1008, "Unauthorized"); return; }

    ws.on("message", (raw) => {
      try {
        const msg = JSON.parse(raw.toString());
        switch (msg.type) {
          case "join":
            joinRoom(rooms, msg.room, ws);
            break;
          case "message":
            broadcastToRoom(rooms, msg.room, { type: "message", from: userId, text: msg.text }, ws);
            break;
        }
      } catch { ws.send(JSON.stringify({ type: "error", message: "Invalid message" })); }
    });

    ws.on("close", () => {
      rooms.forEach((clients) => clients.delete(ws));
    });
  });

  // Heartbeat to detect stale connections
  const interval = setInterval(() => {
    wss.clients.forEach((ws) => {
      if ((ws as any).isAlive === false) return ws.terminate();
      (ws as any).isAlive = false;
      ws.ping();
    });
  }, 30_000);

  wss.on("close", () => clearInterval(interval));
}

function broadcastToRoom(rooms: Map<string, Set<WebSocket>>, room: string, data: object, exclude?: WebSocket) {
  const clients = rooms.get(room);
  if (!clients) return;
  const msg = JSON.stringify(data);
  clients.forEach((client) => {
    if (client !== exclude && client.readyState === WebSocket.OPEN) {
      client.send(msg);
    }
  });
}
```

## Response Helpers

```typescript
// Consistent response format
interface ApiResponse<T> {
  data: T;
  meta?: { page: number; limit: number; total: number; totalPages: number };
}

// Paginated response builder
function paginate<T>(data: T[], total: number, page: number, limit: number): ApiResponse<T[]> {
  return {
    data,
    meta: { page, limit, total, totalPages: Math.ceil(total / limit) },
  };
}

// Created response
function created<T>(res: Response, data: T, location?: string) {
  if (location) res.setHeader("Location", location);
  return res.status(201).json({ data });
}

// No content
function noContent(res: Response) {
  return res.status(204).end();
}
```

## Gotchas

1. **Middleware order matters** — `helmet()`, `cors()`, `express.json()` BEFORE routes. Error handler AFTER routes. Rate limiter before auth. Getting order wrong causes silent failures.

2. **Don't throw in async route handlers** — Express doesn't catch async errors. Always use `asyncHandler` wrapper or express-async-errors package. Without it, unhandled rejections crash the server.

3. **Body size limits** — Default `express.json()` accepts 100KB. Set `{ limit: "10kb" }` explicitly. Without limits, attackers can send huge payloads that consume memory.

4. **Graceful shutdown** — Call `server.close()` on SIGTERM, then close DB/Redis connections. Without it, Kubernetes/Docker kills in-flight requests. Add a force-close timeout.

5. **Don't leak error details in production** — Stack traces, SQL errors, and internal paths help attackers. Return generic messages in production, detailed ones only in development.

6. **Validate params AND body** — Route params (`:id`) come as strings. Always validate and coerce with Zod. `parseInt(req.params.id)` returns NaN silently on bad input.
