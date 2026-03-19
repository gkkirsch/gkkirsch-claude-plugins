---
name: spring-project-setup
description: >
  Scaffolds a new Spring Boot 3.x project with Gradle, Docker Compose,
  PostgreSQL, testing infrastructure, and production-ready configuration.
---

# Spring Project Setup Skill

Sets up a complete Spring Boot project with production-ready defaults.

## What This Skill Does

1. Creates Gradle multi-module project structure (or single module)
2. Configures Spring Boot 3.x with Java 21
3. Sets up PostgreSQL with Flyway migrations
4. Configures Docker Compose for local development
5. Sets up testing infrastructure (JUnit 5, Testcontainers, MockMvc)
6. Configures application profiles (local, test, production)
7. Adds security, observability, and API documentation

## Usage

When the user wants to create a new Spring Boot project, gather these inputs:

1. **Project name**: kebab-case (e.g., `order-service`)
2. **Group ID**: reverse domain (e.g., `com.example`)
3. **Architecture**: single-module or multi-module
4. **Features needed**: Select from available starters

## Project Templates

### Single-Module Spring Boot Project

```
{project-name}/
├── build.gradle.kts
├── settings.gradle.kts
├── gradle/
│   └── wrapper/
│       ├── gradle-wrapper.jar
│       └── gradle-wrapper.properties
├── gradlew
├── gradlew.bat
├── docker-compose.yml
├── Dockerfile
├── .env.example
├── src/
│   ├── main/
│   │   ├── java/{group-path}/{project}/
│   │   │   ├── Application.java
│   │   │   ├── config/
│   │   │   │   ├── SecurityConfig.java
│   │   │   │   ├── WebConfig.java
│   │   │   │   └── OpenApiConfig.java
│   │   │   ├── domain/
│   │   │   │   ├── model/
│   │   │   │   └── port/
│   │   │   ├── application/
│   │   │   │   └── dto/
│   │   │   └── adapter/
│   │   │       ├── in/web/
│   │   │       └── out/persistence/
│   │   └── resources/
│   │       ├── application.yml
│   │       ├── application-local.yml
│   │       ├── application-test.yml
│   │       └── db/migration/
│   │           └── V1__initial_schema.sql
│   └── test/
│       └── java/{group-path}/{project}/
│           ├── IntegrationTestBase.java
│           ├── ArchitectureTest.java
│           └── adapter/in/web/
│               └── HealthCheckTest.java
└── .gitignore
```

### build.gradle.kts Template

```kotlin
plugins {
    java
    id("org.springframework.boot") version "3.3.0"
    id("io.spring.dependency-management") version "1.1.5"
}

group = "{group-id}"
version = "0.0.1-SNAPSHOT"

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(21)
    }
}

repositories {
    mavenCentral()
}

dependencies {
    // Spring Boot starters
    implementation("org.springframework.boot:spring-boot-starter-web")
    implementation("org.springframework.boot:spring-boot-starter-data-jpa")
    implementation("org.springframework.boot:spring-boot-starter-validation")
    implementation("org.springframework.boot:spring-boot-starter-actuator")
    implementation("org.springframework.boot:spring-boot-starter-security")

    // Database
    runtimeOnly("org.postgresql:postgresql")
    implementation("org.flywaydb:flyway-core")
    implementation("org.flywaydb:flyway-database-postgresql")

    // API Documentation
    implementation("org.springdoc:springdoc-openapi-starter-webmvc-ui:2.5.0")

    // Observability
    implementation("io.micrometer:micrometer-registry-prometheus")

    // Caching
    implementation("com.github.ben-manes.caffeine:caffeine")

    // Testing
    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testImplementation("org.springframework.security:spring-security-test")
    testImplementation("org.testcontainers:postgresql")
    testImplementation("org.testcontainers:junit-jupiter")
    testImplementation("com.tngtech.archunit:archunit-junit5:1.3.0")
}

tasks.withType<JavaCompile> {
    options.compilerArgs.addAll(listOf("-parameters"))
}

tasks.withType<Test> {
    useJUnitPlatform()
    jvmArgs("-XX:+EnableDynamicAgentLoading")
}
```

### application.yml Template

```yaml
spring:
  application:
    name: {project-name}

  # Database
  datasource:
    url: jdbc:postgresql://${DB_HOST:localhost}:${DB_PORT:5432}/${DB_NAME:{project-name}}
    username: ${DB_USER:{project-name}}
    password: ${DB_PASSWORD:secret}
    hikari:
      maximum-pool-size: 10
      minimum-idle: 5
      connection-timeout: 2000
      leak-detection-threshold: 60000

  # JPA
  jpa:
    open-in-view: false
    hibernate:
      ddl-auto: validate
    properties:
      hibernate:
        default_batch_fetch_size: 20
        jdbc:
          batch_size: 50
        order_inserts: true
        order_updates: true

  # Flyway
  flyway:
    enabled: true
    locations: classpath:db/migration

  # Virtual threads
  threads:
    virtual:
      enabled: true

# Server
server:
  shutdown: graceful
  error:
    include-message: always
    include-binding-errors: always

# Actuator
management:
  endpoints:
    web:
      exposure:
        include: health,info,metrics,prometheus
  endpoint:
    health:
      show-details: when_authorized

# Logging
logging:
  level:
    root: INFO
    "{group-id}": DEBUG
  pattern:
    console: "%d{HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n"

# API docs
springdoc:
  swagger-ui:
    path: /swagger-ui.html
  api-docs:
    path: /v3/api-docs
```

### application-local.yml Template

```yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/{project-name}
    username: {project-name}
    password: secret

logging:
  level:
    root: INFO
    "{group-id}": DEBUG
    org.hibernate.SQL: DEBUG
    org.hibernate.type.descriptor.sql.BasicBinder: TRACE
```

### application-test.yml Template

```yaml
spring:
  datasource:
    url: jdbc:tc:postgresql:16-alpine:///{project-name}
  jpa:
    hibernate:
      ddl-auto: create-drop
  flyway:
    enabled: false

logging:
  level:
    root: WARN
    "{group-id}": INFO
```

### docker-compose.yml Template

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: {project-name}
      POSTGRES_USER: {project-name}
      POSTGRES_PASSWORD: secret
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U {project-name}"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  pgdata:
```

### Dockerfile Template

```dockerfile
FROM eclipse-temurin:21-jdk-alpine AS builder
WORKDIR /app
COPY gradle/ gradle/
COPY gradlew build.gradle.kts settings.gradle.kts ./
RUN ./gradlew dependencies --no-daemon
COPY src/ src/
RUN ./gradlew bootJar --no-daemon -x test

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
RUN addgroup -S app && adduser -S app -G app
USER app
COPY --from=builder /app/build/libs/*.jar app.jar
ENV JAVA_OPTS="-XX:+UseZGC -XX:MaxRAMPercentage=75"
ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
```

### IntegrationTestBase Template

```java
package {group-id}.{project};

import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.web.client.TestRestTemplate;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;
import org.junit.jupiter.api.BeforeEach;

@Testcontainers
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
public abstract class IntegrationTestBase {

    @Container
    static final PostgreSQLContainer<?> postgres =
        new PostgreSQLContainer<>("postgres:16-alpine")
            .withDatabaseName("test")
            .withUsername("test")
            .withPassword("test")
            .withReuse(true);

    @DynamicPropertySource
    static void configureProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
        registry.add("spring.datasource.username", postgres::getUsername);
        registry.add("spring.datasource.password", postgres::getPassword);
    }

    @Autowired
    protected TestRestTemplate restTemplate;

    @Autowired
    protected JdbcTemplate jdbcTemplate;
}
```

### .gitignore Template

```
# Gradle
.gradle/
build/
!gradle/wrapper/gradle-wrapper.jar

# IDE
.idea/
*.iml
.vscode/
.classpath
.project
.settings/

# OS
.DS_Store
Thumbs.db

# Environment
.env
*.env.local

# Logs
*.log
logs/

# Application
*.jar
!gradle/wrapper/gradle-wrapper.jar
*.war
```

## Setup Procedure

1. Create project directory and navigate into it
2. Initialize Gradle wrapper: `gradle wrapper --gradle-version 8.7`
3. Create all template files with user's project name and group ID substituted
4. Create initial Flyway migration V1
5. Start Docker Compose: `docker compose up -d`
6. Run initial build: `./gradlew build`
7. Verify tests pass: `./gradlew test`
8. Initialize Git: `git init && git add -A && git commit -m "Initial Spring Boot project setup"`

## Checklist After Setup

- [ ] Application starts: `./gradlew bootRun`
- [ ] Health endpoint works: `curl localhost:8080/actuator/health`
- [ ] Database connects (check logs for Flyway migration)
- [ ] Tests pass: `./gradlew test`
- [ ] Swagger UI loads: `http://localhost:8080/swagger-ui.html`
- [ ] Metrics endpoint: `curl localhost:8080/actuator/metrics`
