# Docker Reference

## CLI Commands

### Images

```bash
docker build -t myapp:latest .                      # build image
docker build -t myapp:latest -f Dockerfile.prod .   # custom Dockerfile
docker build --no-cache -t myapp:latest .           # rebuild without cache
docker build --target builder -t myapp:builder .    # build specific stage
DOCKER_BUILDKIT=1 docker build -t myapp:latest .    # use BuildKit

docker images                                       # list images
docker image ls --filter "dangling=true"            # untagged images
docker image prune                                  # remove dangling images
docker image prune -a                               # remove all unused images
docker rmi <image>                                  # remove image
docker tag myapp:latest registry/myapp:v1.0         # tag image
docker push registry/myapp:v1.0                     # push to registry
docker pull registry/myapp:v1.0                     # pull from registry
docker save myapp:latest | gzip > myapp.tar.gz      # export image
docker load < myapp.tar.gz                          # import image
docker history myapp:latest                         # show image layers
docker inspect myapp:latest                         # image metadata
```

### Containers

```bash
docker run -d --name myapp -p 3000:3000 myapp:latest    # run detached
docker run -it --rm myapp:latest /bin/sh                  # interactive, remove on exit
docker run -d --restart=unless-stopped myapp:latest       # auto-restart
docker run -d -e NODE_ENV=production myapp:latest         # with env var
docker run -d --env-file .env myapp:latest                # with env file
docker run -d -v $(pwd)/data:/app/data myapp:latest       # with bind mount
docker run -d -v mydata:/app/data myapp:latest            # with named volume
docker run -d --network mynet myapp:latest                # on custom network
docker run -d --memory=512m --cpus=1.5 myapp:latest       # with resource limits
docker run -d --read-only --tmpfs /tmp myapp:latest       # read-only filesystem

docker ps                                           # running containers
docker ps -a                                        # all containers
docker stop <container>                             # graceful stop
docker kill <container>                             # force stop
docker rm <container>                               # remove container
docker rm -f <container>                            # force remove
docker container prune                              # remove stopped containers
docker start <container>                            # start stopped container
docker restart <container>                          # restart container
docker rename old-name new-name                     # rename container
docker update --memory=1g <container>               # update limits
```

### Logs & Debugging

```bash
docker logs <container>                             # all logs
docker logs <container> --tail 100                  # last 100 lines
docker logs <container> -f                          # follow logs
docker logs <container> --since 1h                  # last hour
docker logs <container> --timestamps                # with timestamps

docker exec -it <container> /bin/sh                 # shell into container
docker exec <container> cat /etc/hosts              # run command
docker exec -u root <container> bash                # exec as root

docker inspect <container>                          # full metadata
docker inspect -f '{{.State.Health.Status}}' <c>    # health status
docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' <c>  # IP
docker stats                                        # live resource usage
docker stats --no-stream                            # snapshot
docker top <container>                              # processes in container
docker diff <container>                             # filesystem changes
docker cp <container>:/path/file ./local            # copy from container
docker cp ./local <container>:/path/file            # copy to container
```

### Volumes

```bash
docker volume create mydata                         # create volume
docker volume ls                                    # list volumes
docker volume inspect mydata                        # volume details
docker volume rm mydata                             # remove volume
docker volume prune                                 # remove unused volumes
```

### Networks

```bash
docker network create mynet                         # create bridge network
docker network create --driver overlay mynet        # overlay (swarm)
docker network ls                                   # list networks
docker network inspect mynet                        # network details
docker network connect mynet <container>            # attach container
docker network disconnect mynet <container>         # detach container
docker network rm mynet                             # remove network
docker network prune                                # remove unused networks
```

### System

```bash
docker system df                                    # disk usage
docker system prune                                 # cleanup everything
docker system prune -a --volumes                    # aggressive cleanup
docker system info                                  # system-wide info
docker version                                      # version info
```

## Dockerfile Instructions

| Instruction | Purpose | Example |
|-------------|---------|---------|
| `FROM` | Base image | `FROM node:20-alpine AS builder` |
| `WORKDIR` | Set working directory | `WORKDIR /app` |
| `COPY` | Copy files (respects .dockerignore) | `COPY package*.json ./` |
| `ADD` | Copy + extract archives + URLs | `ADD app.tar.gz /app/` |
| `RUN` | Execute command (creates layer) | `RUN npm ci --production` |
| `CMD` | Default command (overridable) | `CMD ["node", "server.js"]` |
| `ENTRYPOINT` | Fixed command (args appended) | `ENTRYPOINT ["node"]` |
| `ENV` | Set environment variable | `ENV NODE_ENV=production` |
| `ARG` | Build-time variable | `ARG VERSION=latest` |
| `EXPOSE` | Document port (does not publish) | `EXPOSE 3000` |
| `VOLUME` | Create mount point | `VOLUME ["/data"]` |
| `USER` | Set user for subsequent commands | `USER node` |
| `HEALTHCHECK` | Container health check | `HEALTHCHECK CMD curl -f http://localhost:3000/health` |
| `LABEL` | Add metadata | `LABEL version="1.0" maintainer="dev@example.com"` |
| `SHELL` | Override default shell | `SHELL ["/bin/bash", "-c"]` |
| `STOPSIGNAL` | Signal to stop container | `STOPSIGNAL SIGTERM` |

### CMD vs ENTRYPOINT

```dockerfile
# CMD only — fully overridable
CMD ["node", "server.js"]
# docker run myapp                    → node server.js
# docker run myapp node repl.js       → node repl.js (overridden)

# ENTRYPOINT only — args appended
ENTRYPOINT ["node"]
# docker run myapp                    → node
# docker run myapp server.js          → node server.js

# ENTRYPOINT + CMD — best pattern
ENTRYPOINT ["node"]
CMD ["server.js"]
# docker run myapp                    → node server.js
# docker run myapp repl.js            → node repl.js
```

### Shell Form vs Exec Form

```dockerfile
# Exec form (preferred) — runs directly, PID 1, receives signals
CMD ["node", "server.js"]
ENTRYPOINT ["python", "app.py"]
RUN ["apt-get", "install", "-y", "curl"]

# Shell form — wraps in /bin/sh -c, doesn't receive signals
CMD node server.js
RUN apt-get install -y curl

# Shell form is needed for:
RUN echo "building..." && npm run build    # shell features
RUN if [ "$ENV" = "prod" ]; then ...; fi   # conditionals
```

## Docker Compose Reference

### Service Configuration

```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      target: development              # multi-stage target
      args:
        NODE_ENV: development
    image: myapp:latest                 # or use build, not both usually
    container_name: myapp               # fixed name (bad for scaling)
    restart: unless-stopped             # always | on-failure | no
    ports:
      - "3000:3000"                     # host:container
      - "127.0.0.1:3000:3000"          # localhost only
    expose:
      - "3000"                          # internal only (no host mapping)
    environment:
      NODE_ENV: production
      DB_URL: postgres://db:5432/app
    env_file:
      - .env
      - .env.local
    volumes:
      - .:/app                          # bind mount
      - /app/node_modules               # anonymous volume (exclude)
      - pgdata:/var/lib/postgresql/data  # named volume
    depends_on:
      db:
        condition: service_healthy      # wait for health check
    networks:
      - backend
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 512M
          cpus: "0.5"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
    command: ["node", "server.js"]      # override CMD
    entrypoint: ["/entrypoint.sh"]      # override ENTRYPOINT
    working_dir: /app
    user: "1000:1000"
    stdin_open: true                    # -i
    tty: true                           # -t
    privileged: false
    read_only: true
    tmpfs:
      - /tmp
    cap_drop:
      - ALL
    cap_add:
      - NET_BIND_SERVICE
    security_opt:
      - no-new-privileges:true
```

### Compose Commands

```bash
docker compose up                       # start all services (foreground)
docker compose up -d                    # start detached
docker compose up --build               # rebuild images first
docker compose up app                   # start specific service
docker compose down                     # stop and remove containers
docker compose down -v                  # also remove volumes
docker compose down --rmi all           # also remove images
docker compose stop                     # stop without removing
docker compose start                    # start stopped services
docker compose restart                  # restart all
docker compose ps                       # list containers
docker compose logs                     # all logs
docker compose logs -f app              # follow specific service
docker compose exec app sh              # shell into running service
docker compose run app npm test         # run one-off command
docker compose build                    # build all images
docker compose pull                     # pull latest images
docker compose config                   # validate and show resolved config
docker compose top                      # show processes
```

## Best Practices Summary

### Image Size

| Strategy | Impact |
|----------|--------|
| Use Alpine or distroless base | 5-50x smaller images |
| Multi-stage builds | Remove build deps from final image |
| Combine RUN commands | Fewer layers |
| Clean up in same RUN | `apt-get install && ... && apt-get clean && rm -rf /var/lib/apt/lists/*` |
| Use .dockerignore | Don't copy node_modules, .git, etc. |
| Order layers by change frequency | Better cache hits |

### Security

| Practice | How |
|----------|-----|
| Don't run as root | `USER node` or `USER 1000` |
| Use read-only filesystem | `--read-only --tmpfs /tmp` |
| Drop capabilities | `cap_drop: [ALL]` then add only needed |
| Pin image versions | `node:20.11-alpine` not `node:latest` |
| Scan images | `docker scout cves myapp:latest` |
| Use secrets | `docker secret` or `--mount=type=secret` |
| No sensitive data in ENV | Use secrets or mounted files |
| Verify image signatures | `docker trust inspect` |

### Performance

| Technique | Effect |
|-----------|--------|
| Layer caching | Copy package.json before source for dependency caching |
| BuildKit cache mounts | `--mount=type=cache,target=/root/.npm` |
| Parallel multi-stage | Independent stages build in parallel |
| .dockerignore | Faster build context transfer |
| Slim base images | Faster pull, less memory |

## Common Port Mappings

| Service | Default Port |
|---------|-------------|
| HTTP | 80 |
| HTTPS | 443 |
| PostgreSQL | 5432 |
| MySQL | 3306 |
| MongoDB | 27017 |
| Redis | 6379 |
| Elasticsearch | 9200 |
| RabbitMQ | 5672 (AMQP), 15672 (management) |
| Kafka | 9092 |
| MinIO | 9000 (API), 9001 (console) |
| MailHog | 1025 (SMTP), 8025 (web) |
| Prometheus | 9090 |
| Grafana | 3000 |
| Nginx | 80, 443 |
| Node.js | 3000 |
| Python/Django | 8000 |
| Go | 8080 |
