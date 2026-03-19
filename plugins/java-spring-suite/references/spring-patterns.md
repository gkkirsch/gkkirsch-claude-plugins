# Spring Patterns Reference

Comprehensive reference for Spring Boot design patterns, dependency injection, AOP, and
event-driven architecture patterns used in production Spring applications.

## Repository / Service / Controller Pattern

### The Standard Three-Layer Architecture

```java
// CONTROLLER — HTTP interface. No business logic here.
// Responsibilities: request validation, response mapping, HTTP status codes
@RestController
@RequestMapping("/api/v1/customers")
public class CustomerController {
    private final CustomerService customerService;

    public CustomerController(CustomerService customerService) {
        this.customerService = customerService;
    }

    @GetMapping("/{id}")
    public CustomerResponse getCustomer(@PathVariable UUID id) {
        Customer customer = customerService.findById(new CustomerId(id));
        return CustomerResponse.from(customer);  // Map to API response
    }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public CustomerResponse createCustomer(@Valid @RequestBody CreateCustomerRequest request) {
        Customer customer = customerService.create(request.toCommand());
        return CustomerResponse.from(customer);
    }
}

// SERVICE — Business logic orchestration. Transaction boundaries.
// Responsibilities: business rules, coordination, transaction management
@Service
@Transactional(readOnly = true)  // Default read-only for queries
public class CustomerService {
    private final CustomerRepository customerRepository;
    private final EmailService emailService;

    @Transactional  // Override for writes
    public Customer create(CreateCustomerCommand command) {
        // Business rule: email must be unique
        if (customerRepository.existsByEmail(command.email())) {
            throw new DuplicateEmailException(command.email());
        }

        Customer customer = Customer.create(command.name(), command.email());
        Customer saved = customerRepository.save(customer);

        emailService.sendWelcomeEmail(saved);
        return saved;
    }

    public Customer findById(CustomerId id) {
        return customerRepository.findById(id)
            .orElseThrow(() -> new CustomerNotFoundException(id));
    }
}

// REPOSITORY — Data access. Abstracts persistence mechanism.
// Responsibilities: CRUD operations, query execution, data mapping
public interface CustomerRepository {
    Optional<Customer> findById(CustomerId id);
    Customer save(Customer customer);
    boolean existsByEmail(String email);
    Page<Customer> findAll(CustomerQuery query, Pageable pageable);
}

// JPA implementation
@Repository
class JpaCustomerRepository implements CustomerRepository {
    private final CustomerJpaRepo jpaRepo;
    private final CustomerMapper mapper;

    @Override
    public Optional<Customer> findById(CustomerId id) {
        return jpaRepo.findById(id.value()).map(mapper::toDomain);
    }

    @Override
    public Customer save(Customer customer) {
        CustomerEntity entity = mapper.toEntity(customer);
        CustomerEntity saved = jpaRepo.save(entity);
        return mapper.toDomain(saved);
    }
}
```

### When to Break the Three-Layer Pattern

```java
// USE CASE pattern — when services get too large or have complex orchestration
// Each use case is a single-purpose service

public class TransferMoneyUseCase {
    private final AccountRepository accountRepository;
    private final TransactionRepository transactionRepository;
    private final FraudDetectionService fraudService;

    @Transactional
    public TransferResult execute(TransferCommand command) {
        Account source = accountRepository.findById(command.sourceId());
        Account target = accountRepository.findById(command.targetId());

        fraudService.validate(command);

        source.debit(command.amount());
        target.credit(command.amount());

        accountRepository.save(source);
        accountRepository.save(target);

        return new TransferResult(transactionRepository.save(
            new Transaction(source.id(), target.id(), command.amount())));
    }
}

// Register use cases as beans
@Configuration
public class UseCaseConfig {
    @Bean
    public TransferMoneyUseCase transferMoney(
            AccountRepository accounts,
            TransactionRepository transactions,
            FraudDetectionService fraud) {
        return new TransferMoneyUseCase(accounts, transactions, fraud);
    }
}
```

## Dependency Injection Patterns

### Constructor Injection (Preferred)

```java
// ALWAYS use constructor injection — immutable, testable, fails fast
@Service
public class OrderService {
    private final OrderRepository orderRepository;
    private final PaymentGateway paymentGateway;
    private final Clock clock;

    // Single constructor — @Autowired is optional
    public OrderService(OrderRepository orderRepository,
                        PaymentGateway paymentGateway,
                        Clock clock) {
        this.orderRepository = orderRepository;
        this.paymentGateway = paymentGateway;
        this.clock = clock;
    }
}

// With Lombok (if used in project)
@Service
@RequiredArgsConstructor
public class OrderService {
    private final OrderRepository orderRepository;
    private final PaymentGateway paymentGateway;
    private final Clock clock;
}
```

### Configuration Beans

```java
// Create beans for third-party classes and complex initialization
@Configuration
public class InfrastructureConfig {

    @Bean
    public Clock clock() {
        return Clock.systemUTC();  // Testable — inject Clock.fixed() in tests
    }

    @Bean
    public ObjectMapper objectMapper() {
        return new ObjectMapper()
            .registerModule(new JavaTimeModule())
            .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS)
            .disable(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES)
            .setPropertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE);
    }

    @Bean
    @Profile("!test")  // Only in non-test profiles
    public WebClient externalApiClient(WebClient.Builder builder,
                                        ExternalApiProperties props) {
        return builder
            .baseUrl(props.baseUrl())
            .defaultHeader("X-API-Key", props.apiKey())
            .build();
    }
}

// Conditional beans
@Configuration
public class StorageConfig {

    @Bean
    @ConditionalOnProperty(name = "app.storage.type", havingValue = "s3")
    public FileStorage s3Storage(S3Client s3Client, StorageProperties props) {
        return new S3FileStorage(s3Client, props.bucket());
    }

    @Bean
    @ConditionalOnProperty(name = "app.storage.type", havingValue = "local", matchIfMissing = true)
    public FileStorage localStorage(StorageProperties props) {
        return new LocalFileStorage(props.localPath());
    }
}
```

### Qualifier and Primary

```java
// When multiple implementations exist for an interface
public interface NotificationSender {
    void send(Notification notification);
}

@Component
@Qualifier("email")
public class EmailNotificationSender implements NotificationSender {
    @Override
    public void send(Notification notification) { /* email logic */ }
}

@Component
@Qualifier("sms")
public class SmsNotificationSender implements NotificationSender {
    @Override
    public void send(Notification notification) { /* sms logic */ }
}

@Primary  // Default when no qualifier specified
@Component
public class CompositeNotificationSender implements NotificationSender {
    private final List<NotificationSender> senders;

    public CompositeNotificationSender(
            @Qualifier("email") NotificationSender email,
            @Qualifier("sms") NotificationSender sms) {
        this.senders = List.of(email, sms);
    }

    @Override
    public void send(Notification notification) {
        senders.forEach(s -> s.send(notification));
    }
}
```

## Aspect-Oriented Programming (AOP)

### Logging Aspect

```java
@Aspect
@Component
@Order(1)  // Execute before other aspects
public class LoggingAspect {

    // Log all service method calls
    @Around("execution(* com.example..service.*.*(..))")
    public Object logServiceMethod(ProceedingJoinPoint joinPoint) throws Throwable {
        String methodName = joinPoint.getSignature().toShortString();
        Logger log = LoggerFactory.getLogger(joinPoint.getTarget().getClass());

        log.debug("Entering {} with args: {}", methodName, joinPoint.getArgs());
        long start = System.nanoTime();

        try {
            Object result = joinPoint.proceed();
            long duration = TimeUnit.NANOSECONDS.toMillis(System.nanoTime() - start);
            log.debug("Exiting {} — {}ms", methodName, duration);
            return result;
        } catch (Exception e) {
            long duration = TimeUnit.NANOSECONDS.toMillis(System.nanoTime() - start);
            log.error("Exception in {} — {}ms: {}", methodName, duration, e.getMessage());
            throw e;
        }
    }
}
```

### Retry Aspect

```java
// Custom @Retryable annotation (alternative to Spring Retry)
@Target(ElementType.METHOD)
@Retention(RetentionPolicy.RUNTIME)
public @interface Retryable {
    int maxAttempts() default 3;
    long delayMs() default 1000;
    double multiplier() default 2.0;
    Class<? extends Exception>[] retryOn() default {Exception.class};
    Class<? extends Exception>[] noRetryOn() default {};
}

@Aspect
@Component
public class RetryAspect {

    @Around("@annotation(retryable)")
    public Object retry(ProceedingJoinPoint joinPoint, Retryable retryable) throws Throwable {
        int attempts = 0;
        long delay = retryable.delayMs();
        Exception lastException = null;

        while (attempts < retryable.maxAttempts()) {
            try {
                return joinPoint.proceed();
            } catch (Exception e) {
                lastException = e;

                // Check if this exception type should be retried
                if (shouldNotRetry(e, retryable)) {
                    throw e;
                }

                attempts++;
                if (attempts < retryable.maxAttempts()) {
                    Logger log = LoggerFactory.getLogger(joinPoint.getTarget().getClass());
                    log.warn("Retry {}/{} for {} after {}ms: {}",
                        attempts, retryable.maxAttempts(),
                        joinPoint.getSignature().toShortString(), delay, e.getMessage());

                    Thread.sleep(delay);
                    delay = (long) (delay * retryable.multiplier());
                }
            }
        }

        throw lastException;
    }

    private boolean shouldNotRetry(Exception e, Retryable retryable) {
        for (Class<? extends Exception> noRetryClass : retryable.noRetryOn()) {
            if (noRetryClass.isInstance(e)) return true;
        }
        for (Class<? extends Exception> retryClass : retryable.retryOn()) {
            if (retryClass.isInstance(e)) return false;
        }
        return true;
    }
}

// Usage
@Service
public class PaymentService {

    @Retryable(maxAttempts = 3, delayMs = 500, retryOn = {IOException.class, TimeoutException.class})
    public PaymentResult processPayment(PaymentRequest request) {
        return paymentGateway.charge(request);
    }
}
```

### Caching Aspect

```java
@Aspect
@Component
public class CacheMetricsAspect {
    private final MeterRegistry meterRegistry;

    @Around("@annotation(org.springframework.cache.annotation.Cacheable)")
    public Object trackCacheMetrics(ProceedingJoinPoint joinPoint) throws Throwable {
        String methodName = joinPoint.getSignature().toShortString();
        Timer.Sample sample = Timer.start(meterRegistry);

        try {
            Object result = joinPoint.proceed();
            sample.stop(Timer.builder("cache.operation")
                .tag("method", methodName)
                .tag("result", "hit_or_miss")
                .register(meterRegistry));
            return result;
        } catch (Exception e) {
            sample.stop(Timer.builder("cache.operation")
                .tag("method", methodName)
                .tag("result", "error")
                .register(meterRegistry));
            throw e;
        }
    }
}
```

## Event-Driven Patterns in Spring

### Application Events (In-Process)

```java
// Domain event
public record OrderCreatedEvent(
    OrderId orderId,
    CustomerId customerId,
    Money totalAmount,
    Instant occurredAt
) {}

// Publishing events
@Service
@Transactional
public class OrderService {
    private final ApplicationEventPublisher eventPublisher;

    public Order createOrder(CreateOrderCommand command) {
        Order order = Order.create(command);
        orderRepository.save(order);

        // Published within the same transaction
        eventPublisher.publishEvent(new OrderCreatedEvent(
            order.id(),
            order.customerId(),
            order.calculateTotal(),
            Instant.now()
        ));

        return order;
    }
}

// Listening to events
@Component
public class OrderEventListeners {

    // Runs in the SAME transaction as publisher
    @TransactionalEventListener(phase = TransactionPhase.BEFORE_COMMIT)
    public void onOrderCreatedBeforeCommit(OrderCreatedEvent event) {
        // Validation that should fail the transaction if wrong
        inventoryService.reserve(event.orderId());
    }

    // Runs AFTER transaction commits successfully
    @TransactionalEventListener(phase = TransactionPhase.AFTER_COMMIT)
    public void onOrderCreatedAfterCommit(OrderCreatedEvent event) {
        // Side effects that shouldn't affect the main transaction
        emailService.sendOrderConfirmation(event.orderId());
        metricsService.recordOrderCreated(event.totalAmount());
    }

    // Async event processing
    @Async
    @TransactionalEventListener(phase = TransactionPhase.AFTER_COMMIT)
    public void onOrderCreatedAsync(OrderCreatedEvent event) {
        // Heavy processing that shouldn't block the response
        reportingService.generateOrderReport(event.orderId());
    }

    // Handle failure scenarios
    @TransactionalEventListener(phase = TransactionPhase.AFTER_ROLLBACK)
    public void onOrderCreationFailed(OrderCreatedEvent event) {
        alertService.notifyOrderFailure(event.orderId());
    }
}
```

### Event Sourcing Basics

```java
// Event store
public sealed interface OrderEvent permits
    OrderEvent.Created, OrderEvent.LineAdded, OrderEvent.Submitted,
    OrderEvent.Confirmed, OrderEvent.Cancelled {

    UUID eventId();
    UUID aggregateId();
    Instant occurredAt();
    int version();

    record Created(UUID eventId, UUID aggregateId, Instant occurredAt, int version,
                   UUID customerId) implements OrderEvent {}
    record LineAdded(UUID eventId, UUID aggregateId, Instant occurredAt, int version,
                     UUID productId, int quantity, long unitPrice) implements OrderEvent {}
    record Submitted(UUID eventId, UUID aggregateId, Instant occurredAt, int version)
        implements OrderEvent {}
    record Confirmed(UUID eventId, UUID aggregateId, Instant occurredAt, int version,
                     String transactionId) implements OrderEvent {}
    record Cancelled(UUID eventId, UUID aggregateId, Instant occurredAt, int version,
                     String reason) implements OrderEvent {}
}

// Aggregate rebuilt from events
public class Order {
    private UUID id;
    private UUID customerId;
    private List<OrderLine> lines = new ArrayList<>();
    private String status = "DRAFT";
    private int version = 0;

    // Rebuild state from event history
    public static Order fromEvents(List<OrderEvent> events) {
        Order order = new Order();
        events.forEach(order::apply);
        return order;
    }

    private void apply(OrderEvent event) {
        switch (event) {
            case OrderEvent.Created e -> {
                this.id = e.aggregateId();
                this.customerId = e.customerId();
                this.status = "DRAFT";
            }
            case OrderEvent.LineAdded e -> {
                this.lines.add(new OrderLine(e.productId(), e.quantity(), e.unitPrice()));
            }
            case OrderEvent.Submitted e -> {
                this.status = "SUBMITTED";
            }
            case OrderEvent.Confirmed e -> {
                this.status = "CONFIRMED";
            }
            case OrderEvent.Cancelled e -> {
                this.status = "CANCELLED";
            }
        }
        this.version = event.version();
    }
}

// Event store repository
@Repository
public class JdbcEventStore {
    private final JdbcClient jdbcClient;
    private final ObjectMapper objectMapper;

    public void append(UUID aggregateId, OrderEvent event) {
        jdbcClient.sql("""
                INSERT INTO event_store (event_id, aggregate_id, event_type, version, payload, occurred_at)
                VALUES (:eventId, :aggregateId, :eventType, :version, :payload::jsonb, :occurredAt)
                """)
            .param("eventId", event.eventId())
            .param("aggregateId", aggregateId)
            .param("eventType", event.getClass().getSimpleName())
            .param("version", event.version())
            .param("payload", objectMapper.writeValueAsString(event))
            .param("occurredAt", event.occurredAt())
            .update();
    }

    public List<OrderEvent> loadEvents(UUID aggregateId) {
        return jdbcClient.sql("""
                SELECT event_type, payload FROM event_store
                WHERE aggregate_id = :aggregateId
                ORDER BY version ASC
                """)
            .param("aggregateId", aggregateId)
            .query((rs, rowNum) -> deserializeEvent(
                rs.getString("event_type"),
                rs.getString("payload")))
            .list();
    }
}
```

## Specification Pattern

```java
// Type-safe query building with JPA Specifications
public class OrderSpecifications {

    public static Specification<OrderEntity> hasStatus(String status) {
        return (root, query, cb) -> cb.equal(root.get("status"), status);
    }

    public static Specification<OrderEntity> belongsToCustomer(UUID customerId) {
        return (root, query, cb) -> cb.equal(root.get("customerId"), customerId);
    }

    public static Specification<OrderEntity> createdAfter(Instant since) {
        return (root, query, cb) -> cb.greaterThanOrEqualTo(root.get("createdAt"), since);
    }

    public static Specification<OrderEntity> totalGreaterThan(BigDecimal amount) {
        return (root, query, cb) -> cb.greaterThan(root.get("totalAmount"), amount);
    }

    public static Specification<OrderEntity> searchByKeyword(String keyword) {
        return (root, query, cb) -> {
            String pattern = "%" + keyword.toLowerCase() + "%";
            return cb.or(
                cb.like(cb.lower(root.get("customerName")), pattern),
                cb.like(cb.lower(root.get("notes")), pattern)
            );
        };
    }
}

// Usage — compose specifications dynamically
@Service
public class OrderQueryService {
    private final OrderJpaRepository repository;

    public Page<OrderEntity> search(OrderSearchCriteria criteria, Pageable pageable) {
        Specification<OrderEntity> spec = Specification.where(null);

        if (criteria.status() != null) {
            spec = spec.and(OrderSpecifications.hasStatus(criteria.status()));
        }
        if (criteria.customerId() != null) {
            spec = spec.and(OrderSpecifications.belongsToCustomer(criteria.customerId()));
        }
        if (criteria.since() != null) {
            spec = spec.and(OrderSpecifications.createdAfter(criteria.since()));
        }
        if (criteria.minAmount() != null) {
            spec = spec.and(OrderSpecifications.totalGreaterThan(criteria.minAmount()));
        }
        if (criteria.keyword() != null) {
            spec = spec.and(OrderSpecifications.searchByKeyword(criteria.keyword()));
        }

        return repository.findAll(spec, pageable);
    }
}
```

## Template Method Pattern with Spring

```java
// Abstract base for batch processing jobs
public abstract class BatchProcessor<T> {
    private final MeterRegistry meterRegistry;

    @Transactional
    public BatchResult process() {
        Timer.Sample timer = Timer.start(meterRegistry);
        List<T> items = fetchItems();
        int success = 0, failed = 0;

        for (T item : items) {
            try {
                processItem(item);
                success++;
            } catch (Exception e) {
                handleError(item, e);
                failed++;
            }
        }

        onComplete(success, failed);
        timer.stop(Timer.builder("batch.processing")
            .tag("type", getProcessorName())
            .register(meterRegistry));

        return new BatchResult(success, failed);
    }

    protected abstract List<T> fetchItems();
    protected abstract void processItem(T item);
    protected abstract String getProcessorName();

    protected void handleError(T item, Exception e) {
        LoggerFactory.getLogger(getClass())
            .error("Failed to process {}: {}", item, e.getMessage());
    }

    protected void onComplete(int success, int failed) {
        LoggerFactory.getLogger(getClass())
            .info("{} complete: {} success, {} failed", getProcessorName(), success, failed);
    }
}

// Concrete implementation
@Component
public class ExpiredOrderProcessor extends BatchProcessor<Order> {
    private final OrderRepository orderRepository;

    @Override
    protected List<Order> fetchItems() {
        return orderRepository.findExpiredDraftOrders(Instant.now().minus(Duration.ofDays(7)));
    }

    @Override
    protected void processItem(Order order) {
        order.cancel("Expired — no activity for 7 days");
        orderRepository.save(order);
    }

    @Override
    protected String getProcessorName() { return "expired-orders"; }
}
```

## Strategy Pattern with Spring DI

```java
// Strategy interface
public interface PricingStrategy {
    Money calculatePrice(Product product, Customer customer);
    boolean supports(PricingType type);
}

// Multiple strategy implementations
@Component
public class StandardPricing implements PricingStrategy {
    @Override
    public Money calculatePrice(Product product, Customer customer) {
        return product.basePrice();
    }

    @Override
    public boolean supports(PricingType type) {
        return type == PricingType.STANDARD;
    }
}

@Component
public class VolumeDiscountPricing implements PricingStrategy {
    @Override
    public Money calculatePrice(Product product, Customer customer) {
        int totalOrders = customer.totalOrdersLastYear();
        double discount = totalOrders > 100 ? 0.15 : totalOrders > 50 ? 0.10 : totalOrders > 10 ? 0.05 : 0;
        return product.basePrice().applyDiscount((int) (discount * 100));
    }

    @Override
    public boolean supports(PricingType type) {
        return type == PricingType.VOLUME_DISCOUNT;
    }
}

// Strategy registry — auto-discovers all implementations
@Component
public class PricingStrategyRegistry {
    private final Map<PricingType, PricingStrategy> strategies;

    // Spring injects ALL PricingStrategy beans
    public PricingStrategyRegistry(List<PricingStrategy> strategyList) {
        this.strategies = strategyList.stream()
            .flatMap(s -> Arrays.stream(PricingType.values())
                .filter(s::supports)
                .map(type -> Map.entry(type, s)))
            .collect(Collectors.toMap(Map.Entry::getKey, Map.Entry::getValue));
    }

    public PricingStrategy getStrategy(PricingType type) {
        PricingStrategy strategy = strategies.get(type);
        if (strategy == null) {
            throw new UnsupportedPricingTypeException(type);
        }
        return strategy;
    }
}
```

## Converter / Mapper Pattern

```java
// MapStruct — compile-time mapper generation
@Mapper(componentModel = "spring", unmappedTargetPolicy = ReportingPolicy.ERROR)
public interface OrderMapper {

    @Mapping(target = "id", source = "id.value")
    @Mapping(target = "customerId", source = "customerId.value")
    @Mapping(target = "totalAmount", expression = "java(order.calculateTotal().amount())")
    @Mapping(target = "currency", expression = "java(order.calculateTotal().currency())")
    @Mapping(target = "createdAt", ignore = true)  // Set by JPA
    OrderEntity toEntity(Order order);

    @Mapping(target = "id", expression = "java(new OrderId(entity.getId()))")
    @Mapping(target = "customerId", expression = "java(new CustomerId(entity.getCustomerId()))")
    Order toDomain(OrderEntity entity);

    // Collection mappings auto-generated
    List<OrderResponse> toResponseList(List<Order> orders);

    // Custom mapping method
    default Money mapMoney(OrderEntity entity) {
        return new Money(entity.getTotalAmount().longValue(), entity.getCurrency());
    }
}

// Manual mapper when MapStruct is overkill
@Component
public class OrderManualMapper {

    public OrderResponse toResponse(Order order) {
        return new OrderResponse(
            order.id().value(),
            order.status().getClass().getSimpleName().toUpperCase(),
            order.calculateTotal().amount(),
            order.calculateTotal().currency(),
            order.lines().stream()
                .map(this::toLineResponse)
                .toList()
        );
    }

    private OrderLineResponse toLineResponse(OrderLine line) {
        return new OrderLineResponse(
            line.productId().value(),
            line.quantity(),
            line.unitPrice().amount()
        );
    }
}
```

## Scheduling Patterns

```java
@Configuration
@EnableScheduling
public class SchedulingConfig {
    // Use virtual threads for scheduled tasks (Spring Boot 3.2+)
    @Bean
    public TaskScheduler taskScheduler() {
        SimpleAsyncTaskScheduler scheduler = new SimpleAsyncTaskScheduler();
        scheduler.setVirtualThreads(true);
        scheduler.setThreadNamePrefix("scheduled-");
        return scheduler;
    }
}

@Component
public class ScheduledTasks {

    // Fixed-rate: run every 60 seconds regardless of execution time
    @Scheduled(fixedRate = 60_000)
    public void processExpiredOrders() {
        // If previous execution takes > 60s, next will start immediately after
    }

    // Fixed-delay: wait 60 seconds AFTER previous execution completes
    @Scheduled(fixedDelay = 60_000, initialDelay = 10_000)
    public void cleanupTempFiles() {
        // Guaranteed gap between executions
    }

    // Cron: specific schedule
    @Scheduled(cron = "0 0 2 * * MON-FRI")  // 2 AM weekdays
    public void generateDailyReport() {
        // Runs at specific times
    }

    // Configurable schedule from properties
    @Scheduled(cron = "${app.reports.schedule:0 0 3 * * *}")
    public void configurableTask() {
        // Default: 3 AM daily, overridable via config
    }

    // Distributed locking to prevent concurrent execution in clusters
    @Scheduled(fixedRate = 300_000)
    @SchedulerLock(name = "processPaymentRetries", lockAtLeastFor = "PT4M", lockAtMostFor = "PT30M")
    public void processPaymentRetries() {
        // Only one instance in the cluster executes this
        paymentRetryService.retryFailedPayments();
    }
}
```
