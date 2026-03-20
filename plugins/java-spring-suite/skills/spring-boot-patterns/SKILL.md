---
name: spring-boot-patterns
description: >
  Production Spring Boot patterns — REST controllers, service layer,
  JPA repositories, DTOs, validation, exception handling, configuration,
  and testing with JUnit 5 and MockMvc.
  Triggers: "spring boot", "spring rest", "spring controller", "spring jpa",
  "spring repository", "spring service", "spring dto", "spring validation".
  NOT for: Spring Security/auth patterns (use spring-security-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Spring Boot Production Patterns

## Project Structure

```
src/main/java/com/example/api/
  ApiApplication.java
  config/
    WebConfig.java
    OpenApiConfig.java
  controller/
    UserController.java
  dto/
    request/
      CreateUserRequest.java
      UpdateUserRequest.java
    response/
      UserResponse.java
      PagedResponse.java
  entity/
    User.java
    BaseEntity.java
  exception/
    ApiException.java
    GlobalExceptionHandler.java
    ErrorResponse.java
  repository/
    UserRepository.java
  service/
    UserService.java
    impl/
      UserServiceImpl.java
  mapper/
    UserMapper.java
```

## Base Entity

```java
// entity/BaseEntity.java
@MappedSuperclass
@EntityListeners(AuditingEntityListener.class)
public abstract class BaseEntity {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @CreatedDate
    @Column(nullable = false, updatable = false)
    private Instant createdAt;

    @LastModifiedDate
    @Column(nullable = false)
    private Instant updatedAt;

    @Version
    private Long version; // Optimistic locking
}
```

```java
// entity/User.java
@Entity
@Table(name = "users", indexes = {
    @Index(name = "idx_user_email", columnList = "email", unique = true),
    @Index(name = "idx_user_status", columnList = "status")
})
public class User extends BaseEntity {

    @Column(nullable = false, length = 100)
    private String name;

    @Column(nullable = false, unique = true)
    private String email;

    @Column(nullable = false)
    private String passwordHash;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 20)
    private UserStatus status = UserStatus.ACTIVE;

    @OneToMany(mappedBy = "user", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<Order> orders = new ArrayList<>();

    public enum UserStatus {
        ACTIVE, INACTIVE, SUSPENDED
    }

    // Getters, setters, builder pattern
}
```

## DTOs with Validation

```java
// dto/request/CreateUserRequest.java
public record CreateUserRequest(
    @NotBlank(message = "Name is required")
    @Size(min = 2, max = 100, message = "Name must be 2-100 characters")
    String name,

    @NotBlank(message = "Email is required")
    @Email(message = "Invalid email format")
    String email,

    @NotBlank(message = "Password is required")
    @Size(min = 8, max = 128, message = "Password must be 8-128 characters")
    String password
) {}

// dto/request/UpdateUserRequest.java
public record UpdateUserRequest(
    @Size(min = 2, max = 100, message = "Name must be 2-100 characters")
    String name,

    @Email(message = "Invalid email format")
    String email
) {}

// dto/response/UserResponse.java
public record UserResponse(
    UUID id,
    String name,
    String email,
    String status,
    Instant createdAt
) {}

// dto/response/PagedResponse.java
public record PagedResponse<T>(
    List<T> data,
    int page,
    int size,
    long totalElements,
    int totalPages,
    boolean hasNext
) {
    public static <T> PagedResponse<T> from(Page<T> page) {
        return new PagedResponse<>(
            page.getContent(),
            page.getNumber(),
            page.getSize(),
            page.getTotalElements(),
            page.getTotalPages(),
            page.hasNext()
        );
    }
}
```

## Mapper

```java
// mapper/UserMapper.java
@Component
public class UserMapper {

    public UserResponse toResponse(User user) {
        return new UserResponse(
            user.getId(),
            user.getName(),
            user.getEmail(),
            user.getStatus().name(),
            user.getCreatedAt()
        );
    }

    public User toEntity(CreateUserRequest request, String passwordHash) {
        User user = new User();
        user.setName(request.name());
        user.setEmail(request.email());
        user.setPasswordHash(passwordHash);
        return user;
    }

    public void updateEntity(User user, UpdateUserRequest request) {
        if (request.name() != null) {
            user.setName(request.name());
        }
        if (request.email() != null) {
            user.setEmail(request.email());
        }
    }
}
```

## Repository

```java
// repository/UserRepository.java
public interface UserRepository extends JpaRepository<User, UUID> {

    Optional<User> findByEmail(String email);

    boolean existsByEmail(String email);

    Page<User> findByStatus(User.UserStatus status, Pageable pageable);

    @Query("SELECT u FROM User u WHERE u.name LIKE %:search% OR u.email LIKE %:search%")
    Page<User> search(@Param("search") String search, Pageable pageable);

    @Modifying
    @Query("UPDATE User u SET u.status = :status WHERE u.id = :id")
    int updateStatus(@Param("id") UUID id, @Param("status") User.UserStatus status);

    // Projections — fetch only needed columns
    @Query("SELECT u.email FROM User u WHERE u.status = :status")
    List<String> findEmailsByStatus(@Param("status") User.UserStatus status);

    // Specification for dynamic queries
    default Page<User> findAll(UserFilter filter, Pageable pageable) {
        return findAll(toSpecification(filter), pageable);
    }
}
```

## Service Layer

```java
// service/UserService.java
public interface UserService {
    PagedResponse<UserResponse> list(int page, int size);
    UserResponse getById(UUID id);
    UserResponse create(CreateUserRequest request);
    UserResponse update(UUID id, UpdateUserRequest request);
    void delete(UUID id);
}

// service/impl/UserServiceImpl.java
@Service
@Transactional(readOnly = true)
@RequiredArgsConstructor
public class UserServiceImpl implements UserService {

    private final UserRepository userRepository;
    private final UserMapper userMapper;
    private final PasswordEncoder passwordEncoder;

    @Override
    public PagedResponse<UserResponse> list(int page, int size) {
        Pageable pageable = PageRequest.of(page, size, Sort.by("createdAt").descending());
        Page<UserResponse> users = userRepository.findAll(pageable)
            .map(userMapper::toResponse);
        return PagedResponse.from(users);
    }

    @Override
    public UserResponse getById(UUID id) {
        User user = userRepository.findById(id)
            .orElseThrow(() -> new ApiException(HttpStatus.NOT_FOUND,
                "User not found: " + id));
        return userMapper.toResponse(user);
    }

    @Override
    @Transactional
    public UserResponse create(CreateUserRequest request) {
        if (userRepository.existsByEmail(request.email())) {
            throw new ApiException(HttpStatus.CONFLICT,
                "Email already registered: " + request.email());
        }

        String hash = passwordEncoder.encode(request.password());
        User user = userMapper.toEntity(request, hash);
        user = userRepository.save(user);

        return userMapper.toResponse(user);
    }

    @Override
    @Transactional
    public UserResponse update(UUID id, UpdateUserRequest request) {
        User user = userRepository.findById(id)
            .orElseThrow(() -> new ApiException(HttpStatus.NOT_FOUND,
                "User not found: " + id));

        if (request.email() != null && !request.email().equals(user.getEmail())) {
            if (userRepository.existsByEmail(request.email())) {
                throw new ApiException(HttpStatus.CONFLICT,
                    "Email already taken: " + request.email());
            }
        }

        userMapper.updateEntity(user, request);
        user = userRepository.save(user);
        return userMapper.toResponse(user);
    }

    @Override
    @Transactional
    public void delete(UUID id) {
        if (!userRepository.existsById(id)) {
            throw new ApiException(HttpStatus.NOT_FOUND, "User not found: " + id);
        }
        userRepository.deleteById(id);
    }
}
```

## Controller

```java
// controller/UserController.java
@RestController
@RequestMapping("/api/users")
@RequiredArgsConstructor
@Tag(name = "Users", description = "User management")
public class UserController {

    private final UserService userService;

    @GetMapping
    public ResponseEntity<PagedResponse<UserResponse>> list(
        @RequestParam(defaultValue = "0") @Min(0) int page,
        @RequestParam(defaultValue = "20") @Min(1) @Max(100) int size
    ) {
        return ResponseEntity.ok(userService.list(page, size));
    }

    @GetMapping("/{id}")
    public ResponseEntity<UserResponse> get(@PathVariable UUID id) {
        return ResponseEntity.ok(userService.getById(id));
    }

    @PostMapping
    public ResponseEntity<UserResponse> create(@Valid @RequestBody CreateUserRequest request) {
        UserResponse user = userService.create(request);
        URI location = URI.create("/api/users/" + user.id());
        return ResponseEntity.created(location).body(user);
    }

    @PutMapping("/{id}")
    public ResponseEntity<UserResponse> update(
        @PathVariable UUID id,
        @Valid @RequestBody UpdateUserRequest request
    ) {
        return ResponseEntity.ok(userService.update(id, request));
    }

    @DeleteMapping("/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void delete(@PathVariable UUID id) {
        userService.delete(id);
    }
}
```

## Exception Handling

```java
// exception/ApiException.java
public class ApiException extends RuntimeException {
    private final HttpStatus status;

    public ApiException(HttpStatus status, String message) {
        super(message);
        this.status = status;
    }

    public HttpStatus getStatus() { return status; }
}

// exception/ErrorResponse.java
public record ErrorResponse(
    int status,
    String error,
    String message,
    Instant timestamp,
    Map<String, String> details
) {
    public ErrorResponse(HttpStatus status, String message) {
        this(status.value(), status.getReasonPhrase(), message, Instant.now(), null);
    }

    public ErrorResponse(HttpStatus status, String message, Map<String, String> details) {
        this(status.value(), status.getReasonPhrase(), message, Instant.now(), details);
    }
}

// exception/GlobalExceptionHandler.java
@RestControllerAdvice
@Slf4j
public class GlobalExceptionHandler {

    @ExceptionHandler(ApiException.class)
    public ResponseEntity<ErrorResponse> handleApiException(ApiException ex) {
        return ResponseEntity.status(ex.getStatus())
            .body(new ErrorResponse(ex.getStatus(), ex.getMessage()));
    }

    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<ErrorResponse> handleValidation(MethodArgumentNotValidException ex) {
        Map<String, String> errors = ex.getBindingResult().getFieldErrors().stream()
            .collect(Collectors.toMap(
                FieldError::getField,
                f -> f.getDefaultMessage() != null ? f.getDefaultMessage() : "invalid",
                (a, b) -> a
            ));

        return ResponseEntity.badRequest()
            .body(new ErrorResponse(HttpStatus.BAD_REQUEST, "Validation failed", errors));
    }

    @ExceptionHandler(DataIntegrityViolationException.class)
    public ResponseEntity<ErrorResponse> handleDataIntegrity(DataIntegrityViolationException ex) {
        return ResponseEntity.status(HttpStatus.CONFLICT)
            .body(new ErrorResponse(HttpStatus.CONFLICT, "Data integrity violation"));
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<ErrorResponse> handleGeneral(Exception ex) {
        log.error("Unhandled exception", ex);
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
            .body(new ErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, "Internal server error"));
    }
}
```

## Testing

```java
// UserControllerTest.java
@WebMvcTest(UserController.class)
class UserControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private UserService userService;

    @Autowired
    private ObjectMapper objectMapper;

    @Test
    void listUsers_returnsPagedResponse() throws Exception {
        var response = new PagedResponse<>(
            List.of(new UserResponse(UUID.randomUUID(), "Alice", "a@b.com", "ACTIVE", Instant.now())),
            0, 20, 1, 1, false
        );
        when(userService.list(0, 20)).thenReturn(response);

        mockMvc.perform(get("/api/users"))
            .andExpect(status().isOk())
            .andExpect(jsonPath("$.data[0].name").value("Alice"))
            .andExpect(jsonPath("$.totalElements").value(1));
    }

    @Test
    void createUser_withValidData_returns201() throws Exception {
        var request = new CreateUserRequest("Alice", "alice@test.com", "securepass123");
        var response = new UserResponse(UUID.randomUUID(), "Alice", "alice@test.com", "ACTIVE", Instant.now());
        when(userService.create(any())).thenReturn(response);

        mockMvc.perform(post("/api/users")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(request)))
            .andExpect(status().isCreated())
            .andExpect(jsonPath("$.name").value("Alice"))
            .andExpect(header().exists("Location"));
    }

    @Test
    void createUser_withInvalidEmail_returns400() throws Exception {
        var request = new CreateUserRequest("Alice", "invalid", "securepass123");

        mockMvc.perform(post("/api/users")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(request)))
            .andExpect(status().isBadRequest())
            .andExpect(jsonPath("$.details.email").exists());
    }

    @Test
    void getUser_notFound_returns404() throws Exception {
        UUID id = UUID.randomUUID();
        when(userService.getById(id))
            .thenThrow(new ApiException(HttpStatus.NOT_FOUND, "User not found"));

        mockMvc.perform(get("/api/users/{id}", id))
            .andExpect(status().isNotFound());
    }
}

// UserServiceTest.java (unit test)
@ExtendWith(MockitoExtension.class)
class UserServiceImplTest {

    @Mock private UserRepository userRepository;
    @Mock private UserMapper userMapper;
    @Mock private PasswordEncoder passwordEncoder;
    @InjectMocks private UserServiceImpl userService;

    @Test
    void create_duplicateEmail_throwsConflict() {
        var request = new CreateUserRequest("Alice", "dup@test.com", "pass1234");
        when(userRepository.existsByEmail("dup@test.com")).thenReturn(true);

        var ex = assertThrows(ApiException.class, () -> userService.create(request));
        assertEquals(HttpStatus.CONFLICT, ex.getStatus());
    }
}
```

## Gotchas

1. **`@Transactional` on private methods does nothing** — Spring AOP uses proxies. Transactions only work on public methods called from OUTSIDE the class. A public method calling a private `@Transactional` method within the same class bypasses the proxy.

2. **N+1 queries with JPA** — `user.getOrders()` triggers a lazy load per user. Use `@EntityGraph` or JPQL `JOIN FETCH` to eagerly load associations: `@Query("SELECT u FROM User u JOIN FETCH u.orders WHERE u.id = :id")`.

3. **`@RequestBody` validation requires `@Valid`** — A `@RequestBody CreateUserRequest request` without `@Valid` skips Bean Validation entirely. Always add `@Valid` before `@RequestBody` for validation annotations to fire.

4. **`save()` returns the managed entity** — `userRepository.save(user)` returns the managed entity with generated ID and timestamps. Always use the RETURNED entity: `user = userRepository.save(user)`. The input parameter may not have the ID set.

5. **`Optional.get()` without `isPresent()`** — `findById(id).get()` throws `NoSuchElementException`. Always use `orElseThrow(() -> new ApiException(...))` for meaningful error messages.

6. **`@Column(unique = true)` doesn't prevent race conditions** — Two concurrent requests can both check `existsByEmail` and get false, then both try to insert. The unique constraint catches this at the DB level, but you must handle `DataIntegrityViolationException` in addition to the pre-check.
