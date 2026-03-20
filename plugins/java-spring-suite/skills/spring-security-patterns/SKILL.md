---
name: spring-security-patterns
description: >
  Spring Security patterns — JWT authentication, OAuth2, role-based access,
  method security, CORS, CSRF, password encoding, and security testing.
  Triggers: "spring security", "spring jwt", "spring oauth", "spring auth",
  "spring rbac", "spring cors", "spring csrf", "spring password".
  NOT for: General Spring Boot patterns (use spring-boot-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Spring Security Patterns

## Security Configuration (Spring Boot 3.x)

```java
@Configuration
@EnableWebSecurity
@EnableMethodSecurity
public class SecurityConfig {

    private final JwtAuthFilter jwtAuthFilter;
    private final UserDetailsService userDetailsService;

    @Bean
    public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
        return http
            .csrf(csrf -> csrf.disable()) // Disable for stateless JWT APIs
            .cors(cors -> cors.configurationSource(corsConfigSource()))
            .sessionManagement(session ->
                session.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
            .authorizeHttpRequests(auth -> auth
                .requestMatchers("/auth/**", "/public/**", "/health").permitAll()
                .requestMatchers("/api/admin/**").hasRole("ADMIN")
                .requestMatchers(HttpMethod.GET, "/api/users/**").hasAnyRole("USER", "ADMIN")
                .requestMatchers(HttpMethod.POST, "/api/users/**").hasRole("ADMIN")
                .anyRequest().authenticated()
            )
            .exceptionHandling(ex -> ex
                .authenticationEntryPoint((req, res, e) -> {
                    res.setContentType("application/json");
                    res.setStatus(401);
                    res.getWriter().write("{\"error\":\"Unauthorized\"}");
                })
                .accessDeniedHandler((req, res, e) -> {
                    res.setContentType("application/json");
                    res.setStatus(403);
                    res.getWriter().write("{\"error\":\"Forbidden\"}");
                })
            )
            .addFilterBefore(jwtAuthFilter, UsernamePasswordAuthenticationFilter.class)
            .build();
    }

    @Bean
    public PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder(12);
    }

    @Bean
    public AuthenticationManager authenticationManager(AuthenticationConfiguration config)
        throws Exception {
        return config.getAuthenticationManager();
    }

    @Bean
    CorsConfigurationSource corsConfigSource() {
        CorsConfiguration config = new CorsConfiguration();
        config.setAllowedOrigins(List.of("https://myapp.com", "http://localhost:3000"));
        config.setAllowedMethods(List.of("GET", "POST", "PUT", "DELETE", "OPTIONS"));
        config.setAllowedHeaders(List.of("Authorization", "Content-Type"));
        config.setMaxAge(3600L);

        UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
        source.registerCorsConfiguration("/api/**", config);
        return source;
    }
}
```

## JWT Service

```java
@Service
public class JwtService {

    @Value("${jwt.secret}")
    private String secret;

    @Value("${jwt.expiration:86400000}") // 24h default
    private long expiration;

    private SecretKey getSigningKey() {
        byte[] keyBytes = Decoders.BASE64.decode(secret);
        return Keys.hmacShaKeyFor(keyBytes);
    }

    public String generateToken(UserDetails userDetails) {
        return generateToken(Map.of(), userDetails);
    }

    public String generateToken(Map<String, Object> extraClaims, UserDetails userDetails) {
        return Jwts.builder()
            .claims(extraClaims)
            .subject(userDetails.getUsername())
            .issuedAt(new Date())
            .expiration(new Date(System.currentTimeMillis() + expiration))
            .signWith(getSigningKey())
            .compact();
    }

    public String generateRefreshToken(UserDetails userDetails) {
        return Jwts.builder()
            .subject(userDetails.getUsername())
            .issuedAt(new Date())
            .expiration(new Date(System.currentTimeMillis() + expiration * 7)) // 7x access
            .signWith(getSigningKey())
            .compact();
    }

    public String extractUsername(String token) {
        return extractClaim(token, Claims::getSubject);
    }

    public <T> T extractClaim(String token, Function<Claims, T> resolver) {
        Claims claims = Jwts.parser()
            .verifyWith(getSigningKey())
            .build()
            .parseSignedClaims(token)
            .getPayload();
        return resolver.apply(claims);
    }

    public boolean isTokenValid(String token, UserDetails userDetails) {
        String username = extractUsername(token);
        return username.equals(userDetails.getUsername()) && !isTokenExpired(token);
    }

    private boolean isTokenExpired(String token) {
        return extractClaim(token, Claims::getExpiration).before(new Date());
    }
}
```

## JWT Authentication Filter

```java
@Component
@RequiredArgsConstructor
public class JwtAuthFilter extends OncePerRequestFilter {

    private final JwtService jwtService;
    private final UserDetailsService userDetailsService;

    @Override
    protected void doFilterInternal(
        HttpServletRequest request,
        HttpServletResponse response,
        FilterChain chain
    ) throws ServletException, IOException {

        String authHeader = request.getHeader("Authorization");
        if (authHeader == null || !authHeader.startsWith("Bearer ")) {
            chain.doFilter(request, response);
            return;
        }

        String token = authHeader.substring(7);

        try {
            String username = jwtService.extractUsername(token);

            if (username != null && SecurityContextHolder.getContext().getAuthentication() == null) {
                UserDetails userDetails = userDetailsService.loadUserByUsername(username);

                if (jwtService.isTokenValid(token, userDetails)) {
                    UsernamePasswordAuthenticationToken authToken =
                        new UsernamePasswordAuthenticationToken(
                            userDetails, null, userDetails.getAuthorities()
                        );
                    authToken.setDetails(new WebAuthenticationDetailsSource().buildDetails(request));
                    SecurityContextHolder.getContext().setAuthentication(authToken);
                }
            }
        } catch (JwtException e) {
            // Invalid token — continue without authentication
            // The security filter chain will reject if auth is required
        }

        chain.doFilter(request, response);
    }
}
```

## Auth Controller

```java
@RestController
@RequestMapping("/auth")
@RequiredArgsConstructor
public class AuthController {

    private final AuthenticationManager authManager;
    private final JwtService jwtService;
    private final UserDetailsService userDetailsService;
    private final UserService userService;

    public record LoginRequest(
        @NotBlank String email,
        @NotBlank String password
    ) {}

    public record AuthResponse(
        String accessToken,
        String refreshToken,
        long expiresIn
    ) {}

    public record RegisterRequest(
        @NotBlank @Size(min = 2) String name,
        @NotBlank @Email String email,
        @NotBlank @Size(min = 8) String password
    ) {}

    @PostMapping("/login")
    public ResponseEntity<AuthResponse> login(@Valid @RequestBody LoginRequest request) {
        authManager.authenticate(
            new UsernamePasswordAuthenticationToken(request.email(), request.password())
        );

        UserDetails user = userDetailsService.loadUserByUsername(request.email());
        String accessToken = jwtService.generateToken(user);
        String refreshToken = jwtService.generateRefreshToken(user);

        return ResponseEntity.ok(new AuthResponse(accessToken, refreshToken, 86400));
    }

    @PostMapping("/register")
    public ResponseEntity<AuthResponse> register(@Valid @RequestBody RegisterRequest request) {
        UserResponse user = userService.create(new CreateUserRequest(
            request.name(), request.email(), request.password()
        ));

        UserDetails userDetails = userDetailsService.loadUserByUsername(request.email());
        String accessToken = jwtService.generateToken(userDetails);
        String refreshToken = jwtService.generateRefreshToken(userDetails);

        return ResponseEntity.status(HttpStatus.CREATED)
            .body(new AuthResponse(accessToken, refreshToken, 86400));
    }

    @PostMapping("/refresh")
    public ResponseEntity<AuthResponse> refresh(@RequestHeader("Authorization") String header) {
        String refreshToken = header.substring(7);
        String username = jwtService.extractUsername(refreshToken);
        UserDetails user = userDetailsService.loadUserByUsername(username);

        if (!jwtService.isTokenValid(refreshToken, user)) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }

        String newAccessToken = jwtService.generateToken(user);
        return ResponseEntity.ok(new AuthResponse(newAccessToken, refreshToken, 86400));
    }
}
```

## Method-Level Security

```java
@Service
public class DocumentService {

    // Only users with ADMIN role
    @PreAuthorize("hasRole('ADMIN')")
    public void deleteAll() { /* ... */ }

    // Owner or admin
    @PreAuthorize("@authz.isOwner(#id) or hasRole('ADMIN')")
    public Document getById(UUID id) { /* ... */ }

    // Filter return values
    @PostFilter("filterObject.isPublic or filterObject.ownerId == authentication.name")
    public List<Document> listDocuments() { /* ... */ }

    // Custom security expression
    @PreAuthorize("@documentAuthz.canEdit(#id, authentication)")
    public Document update(UUID id, UpdateDocumentRequest request) { /* ... */ }
}

// Custom authorization bean
@Component("documentAuthz")
public class DocumentAuthorization {

    private final DocumentRepository documentRepo;

    public boolean canEdit(UUID documentId, Authentication auth) {
        Document doc = documentRepo.findById(documentId).orElse(null);
        if (doc == null) return false;

        String userId = auth.getName();
        if (doc.getOwnerId().equals(userId)) return true;

        return auth.getAuthorities().stream()
            .anyMatch(a -> a.getAuthority().equals("ROLE_EDITOR"));
    }
}
```

## UserDetails Implementation

```java
@Entity
@Table(name = "users")
public class User implements UserDetails {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    private String name;

    @Column(unique = true)
    private String email;

    private String password;

    @ElementCollection(fetch = FetchType.EAGER)
    @CollectionTable(name = "user_roles", joinColumns = @JoinColumn(name = "user_id"))
    @Column(name = "role")
    @Enumerated(EnumType.STRING)
    private Set<Role> roles = new HashSet<>();

    private boolean enabled = true;

    public enum Role {
        USER, EDITOR, ADMIN
    }

    @Override
    public Collection<? extends GrantedAuthority> getAuthorities() {
        return roles.stream()
            .map(role -> new SimpleGrantedAuthority("ROLE_" + role.name()))
            .collect(Collectors.toSet());
    }

    @Override
    public String getUsername() { return email; }

    @Override
    public boolean isAccountNonExpired() { return true; }

    @Override
    public boolean isAccountNonLocked() { return true; }

    @Override
    public boolean isCredentialsNonExpired() { return true; }

    @Override
    public boolean isEnabled() { return enabled; }
}
```

## Security Testing

```java
@WebMvcTest(UserController.class)
@Import(SecurityConfig.class)
class UserControllerSecurityTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private UserService userService;

    @MockBean
    private JwtService jwtService;

    @MockBean
    private UserDetailsService userDetailsService;

    // Test unauthenticated access
    @Test
    void listUsers_unauthenticated_returns401() throws Exception {
        mockMvc.perform(get("/api/users"))
            .andExpect(status().isUnauthorized());
    }

    // Test with mock user
    @Test
    @WithMockUser(roles = "USER")
    void listUsers_authenticated_returns200() throws Exception {
        when(userService.list(0, 20)).thenReturn(new PagedResponse<>(
            List.of(), 0, 20, 0, 0, false
        ));

        mockMvc.perform(get("/api/users"))
            .andExpect(status().isOk());
    }

    // Test role-based access
    @Test
    @WithMockUser(roles = "USER")
    void createUser_withoutAdmin_returns403() throws Exception {
        mockMvc.perform(post("/api/users")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"name\":\"A\",\"email\":\"a@b.com\",\"password\":\"12345678\"}"))
            .andExpect(status().isForbidden());
    }

    @Test
    @WithMockUser(roles = "ADMIN")
    void createUser_withAdmin_returns201() throws Exception {
        var response = new UserResponse(UUID.randomUUID(), "Alice", "a@b.com", "ACTIVE", Instant.now());
        when(userService.create(any())).thenReturn(response);

        mockMvc.perform(post("/api/users")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"name\":\"Alice\",\"email\":\"a@b.com\",\"password\":\"securepass123\"}"))
            .andExpect(status().isCreated());
    }

    // Test with custom authentication
    @Test
    void adminEndpoint_withJwtToken_works() throws Exception {
        String token = "valid.jwt.token";
        UserDetails userDetails = User.builder()
            .username("admin@test.com")
            .password("")
            .roles("ADMIN")
            .build();

        when(jwtService.extractUsername(token)).thenReturn("admin@test.com");
        when(jwtService.isTokenValid(eq(token), any())).thenReturn(true);
        when(userDetailsService.loadUserByUsername("admin@test.com")).thenReturn(userDetails);

        mockMvc.perform(get("/api/admin/stats")
                .header("Authorization", "Bearer " + token))
            .andExpect(status().isOk());
    }
}
```

## Rate Limiting

```java
// Using Bucket4j with Spring
@Configuration
public class RateLimitConfig {

    @Bean
    public FilterRegistrationBean<RateLimitFilter> rateLimitFilter() {
        FilterRegistrationBean<RateLimitFilter> bean = new FilterRegistrationBean<>();
        bean.setFilter(new RateLimitFilter());
        bean.addUrlPatterns("/api/*");
        return bean;
    }
}

public class RateLimitFilter extends OncePerRequestFilter {
    private final Map<String, Bucket> buckets = new ConcurrentHashMap<>();

    @Override
    protected void doFilterInternal(HttpServletRequest req, HttpServletResponse res,
                                     FilterChain chain) throws ServletException, IOException {
        String ip = req.getRemoteAddr();
        Bucket bucket = buckets.computeIfAbsent(ip, k -> createBucket());

        if (bucket.tryConsume(1)) {
            chain.doFilter(req, res);
        } else {
            res.setStatus(429);
            res.setContentType("application/json");
            res.getWriter().write("{\"error\":\"Too many requests\"}");
        }
    }

    private Bucket createBucket() {
        return Bucket.builder()
            .addLimit(Bandwidth.classic(100, Refill.intervally(100, Duration.ofMinutes(1))))
            .build();
    }
}
```

## Gotchas

1. **`ROLE_` prefix** — Spring Security automatically prepends `ROLE_` to role names. `hasRole("ADMIN")` checks for authority `ROLE_ADMIN`. `hasAuthority("ADMIN")` checks for exactly `ADMIN`. If your DB stores `ADMIN`, your `getAuthorities()` must return `ROLE_ADMIN`.

2. **Filter ordering** — `addFilterBefore(jwtFilter, UsernamePasswordAuthenticationFilter.class)` places your filter BEFORE Spring's default auth filter. If placed after, the default filter may reject the request before your JWT filter runs.

3. **`@PreAuthorize` needs `@EnableMethodSecurity`** — Without `@EnableMethodSecurity` on a `@Configuration` class, `@PreAuthorize`, `@PostAuthorize`, and `@Secured` annotations are silently ignored. Older Spring uses `@EnableGlobalMethodSecurity`.

4. **CSRF disabled for APIs** — Disable CSRF only for stateless REST APIs using JWTs. If your app uses session cookies, CSRF protection is critical. CSRF is enabled by default in Spring Security.

5. **`SecurityContextHolder` is thread-local** — In async or reactive code, the security context doesn't propagate to new threads automatically. Use `DelegatingSecurityContextRunnable` or `SecurityContextHolder.setStrategyName(MODE_INHERITABLETHREADLOCAL)`.

6. **JWT secret key length** — HMAC-SHA256 requires a 256-bit (32-byte) key minimum. Shorter keys silently weaken the signature. Use `Keys.secretKeyFor(SignatureAlgorithm.HS256)` to generate a proper key, then Base64-encode it for your config file.
