# Spring Data Reference

Comprehensive reference for Spring Data JPA, query strategies, transactions, Spring Data
repositories, QueryDSL, and database access patterns for production Spring Boot applications.

## Spring Data JPA Repository Hierarchy

```
Repository<T, ID>                    ← Marker interface
  └── CrudRepository<T, ID>         ← Basic CRUD
       └── ListCrudRepository<T, ID> ← Returns List instead of Iterable
            └── JpaRepository<T, ID> ← JPA-specific (flush, batch, etc.)

PagingAndSortingRepository<T, ID>    ← Pagination + sorting
  └── ListPagingAndSortingRepository<T, ID>

JpaSpecificationExecutor<T>          ← Dynamic queries with Specifications
```

### Repository Interface Definition

```java
// Basic repository — extend JpaRepository for full JPA support
public interface OrderRepository extends JpaRepository<OrderEntity, UUID>,
                                          JpaSpecificationExecutor<OrderEntity> {

    // === DERIVED QUERIES (method name parsing) ===

    // Simple field match
    List<OrderEntity> findByStatus(String status);

    // Multiple conditions
    List<OrderEntity> findByStatusAndCustomerId(String status, UUID customerId);

    // Comparison operators
    List<OrderEntity> findByTotalAmountGreaterThan(BigDecimal amount);
    List<OrderEntity> findByCreatedAtBetween(Instant start, Instant end);

    // Collection membership
    List<OrderEntity> findByStatusIn(List<String> statuses);

    // String matching
    List<OrderEntity> findByCustomerNameContainingIgnoreCase(String name);

    // Existence checks
    boolean existsByCustomerIdAndStatus(UUID customerId, String status);

    // Count queries
    long countByStatus(String status);

    // Delete
    void deleteByStatusAndCreatedAtBefore(String status, Instant before);

    // Top/First
    List<OrderEntity> findTop10ByStatusOrderByCreatedAtDesc(String status);
    Optional<OrderEntity> findFirstByCustomerIdOrderByCreatedAtDesc(UUID customerId);

    // With pagination and sorting
    Page<OrderEntity> findByCustomerId(UUID customerId, Pageable pageable);
    Slice<OrderEntity> findByStatus(String status, Pageable pageable);

    // === JPQL QUERIES ===

    @Query("SELECT o FROM OrderEntity o WHERE o.status = :status AND o.totalAmount > :minAmount")
    List<OrderEntity> findLargeOrdersByStatus(
        @Param("status") String status,
        @Param("minAmount") BigDecimal minAmount);

    @Query("""
        SELECT o FROM OrderEntity o
        JOIN FETCH o.lines
        WHERE o.id = :id
        """)
    Optional<OrderEntity> findByIdWithLines(@Param("id") UUID id);

    // Projection — return only needed fields
    @Query("""
        SELECT new com.example.order.dto.OrderSummary(
            o.id, o.status, o.totalAmount, o.createdAt)
        FROM OrderEntity o
        WHERE o.customerId = :customerId
        ORDER BY o.createdAt DESC
        """)
    Page<OrderSummary> findSummariesByCustomerId(
        @Param("customerId") UUID customerId, Pageable pageable);

    // Update query
    @Modifying
    @Query("UPDATE OrderEntity o SET o.status = :status WHERE o.id = :id")
    int updateStatus(@Param("id") UUID id, @Param("status") String status);

    // Bulk delete
    @Modifying
    @Query("DELETE FROM OrderEntity o WHERE o.status = 'DRAFT' AND o.createdAt < :before")
    int deleteStaleOrders(@Param("before") Instant before);

    // === NATIVE QUERIES ===

    @Query(value = """
        SELECT o.* FROM orders o
        WHERE o.status = :status
        AND o.total_amount > (
            SELECT AVG(total_amount) FROM orders WHERE status = :status
        )
        ORDER BY o.total_amount DESC
        """, nativeQuery = true)
    List<OrderEntity> findAboveAverageOrders(@Param("status") String status);

    // Native query with pagination
    @Query(value = "SELECT * FROM orders WHERE customer_id = :customerId",
           countQuery = "SELECT COUNT(*) FROM orders WHERE customer_id = :customerId",
           nativeQuery = true)
    Page<OrderEntity> findByCustomerIdNative(
        @Param("customerId") UUID customerId, Pageable pageable);
}
```

### Interface-Based Projections

```java
// Closed projection — only specified fields loaded from DB
public interface OrderSummaryProjection {
    UUID getId();
    String getStatus();
    BigDecimal getTotalAmount();
    Instant getCreatedAt();

    // Computed value using SpEL
    @Value("#{target.totalAmount.multiply(new java.math.BigDecimal('0.1'))}")
    BigDecimal getTaxAmount();
}

// Use in repository
public interface OrderRepository extends JpaRepository<OrderEntity, UUID> {
    List<OrderSummaryProjection> findProjectedByCustomerId(UUID customerId);

    // Dynamic projection — caller chooses return type
    <T> List<T> findByStatus(String status, Class<T> type);
}

// Usage
List<OrderSummaryProjection> summaries =
    orderRepository.findProjectedByCustomerId(customerId);

// Or use the dynamic projection
List<OrderSummaryProjection> summaries =
    orderRepository.findByStatus("SUBMITTED", OrderSummaryProjection.class);

// Open projection with default method
public interface OrderDetailProjection {
    UUID getId();
    String getStatus();
    BigDecimal getTotalAmount();
    String getCurrency();
    Instant getCreatedAt();

    default String getFormattedTotal() {
        return getCurrency() + " " + getTotalAmount().setScale(2);
    }
}
```

## JPA Entity Design

### Entity Best Practices

```java
@Entity
@Table(name = "orders", indexes = {
    @Index(name = "idx_orders_customer_id", columnList = "customer_id"),
    @Index(name = "idx_orders_status", columnList = "status"),
    @Index(name = "idx_orders_created_at", columnList = "created_at DESC")
})
public class OrderEntity {

    @Id
    @Column(columnDefinition = "uuid")
    private UUID id;

    @Column(name = "customer_id", nullable = false)
    private UUID customerId;

    @Column(nullable = false, length = 20)
    @Enumerated(EnumType.STRING)  // Always STRING, never ORDINAL
    private OrderStatus status;

    @Column(name = "total_amount", precision = 19, scale = 2)
    private BigDecimal totalAmount;

    @Column(length = 3)
    private String currency;

    @Column(name = "created_at", nullable = false, updatable = false)
    private Instant createdAt;

    @Column(name = "updated_at", nullable = false)
    private Instant updatedAt;

    // Optimistic locking — prevents concurrent modification
    @Version
    private Long version;

    // One-to-Many with proper cascade and orphan removal
    @OneToMany(mappedBy = "order",
               cascade = CascadeType.ALL,
               orphanRemoval = true,
               fetch = FetchType.LAZY)  // ALWAYS lazy for collections
    @OrderBy("createdAt ASC")
    private List<OrderLineEntity> lines = new ArrayList<>();

    // Audit fields
    @CreatedDate
    @Column(name = "created_at", updatable = false)
    private Instant createdAt;

    @LastModifiedDate
    @Column(name = "updated_at")
    private Instant updatedAt;

    @CreatedBy
    @Column(name = "created_by", updatable = false)
    private String createdBy;

    @LastModifiedBy
    @Column(name = "updated_by")
    private String updatedBy;

    // Helper methods for bidirectional relationships
    public void addLine(OrderLineEntity line) {
        lines.add(line);
        line.setOrder(this);
    }

    public void removeLine(OrderLineEntity line) {
        lines.remove(line);
        line.setOrder(null);
    }

    // equals/hashCode based on ID — critical for JPA correctness
    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof OrderEntity other)) return false;
        return id != null && id.equals(other.id);
    }

    @Override
    public int hashCode() {
        return getClass().hashCode();  // Constant — safe for detached entities
    }
}
```

### Embedded Types and Value Objects

```java
@Embeddable
public class Address {
    @Column(nullable = false)
    private String street;

    @Column(nullable = false)
    private String city;

    @Column(nullable = false, length = 2)
    private String state;

    @Column(name = "zip_code", nullable = false, length = 10)
    private String zipCode;

    @Column(length = 2)
    private String country;
}

@Entity
@Table(name = "customers")
public class CustomerEntity {
    @Id
    private UUID id;

    @Embedded
    @AttributeOverrides({
        @AttributeOverride(name = "street", column = @Column(name = "shipping_street")),
        @AttributeOverride(name = "city", column = @Column(name = "shipping_city")),
        @AttributeOverride(name = "state", column = @Column(name = "shipping_state")),
        @AttributeOverride(name = "zipCode", column = @Column(name = "shipping_zip")),
        @AttributeOverride(name = "country", column = @Column(name = "shipping_country"))
    })
    private Address shippingAddress;

    @Embedded
    @AttributeOverrides({
        @AttributeOverride(name = "street", column = @Column(name = "billing_street")),
        @AttributeOverride(name = "city", column = @Column(name = "billing_city")),
        @AttributeOverride(name = "state", column = @Column(name = "billing_state")),
        @AttributeOverride(name = "zipCode", column = @Column(name = "billing_zip")),
        @AttributeOverride(name = "country", column = @Column(name = "billing_country"))
    })
    private Address billingAddress;
}

// JSON column for flexible data
@Entity
@Table(name = "orders")
public class OrderEntity {
    // Store arbitrary metadata as JSON
    @JdbcTypeCode(SqlTypes.JSON)
    @Column(columnDefinition = "jsonb")
    private Map<String, Object> metadata;

    // Typed JSON column with custom converter
    @Convert(converter = OrderPreferencesConverter.class)
    @Column(columnDefinition = "jsonb")
    private OrderPreferences preferences;
}

@Converter
public class OrderPreferencesConverter implements AttributeConverter<OrderPreferences, String> {
    private static final ObjectMapper mapper = new ObjectMapper();

    @Override
    public String convertToDatabaseColumn(OrderPreferences prefs) {
        try {
            return mapper.writeValueAsString(prefs);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Failed to serialize preferences", e);
        }
    }

    @Override
    public OrderPreferences convertToEntityAttribute(String json) {
        try {
            return json == null ? null : mapper.readValue(json, OrderPreferences.class);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Failed to deserialize preferences", e);
        }
    }
}
```

## Transaction Management

### Declarative Transactions

```java
@Service
@Transactional(readOnly = true)  // Default: read-only for all methods
public class OrderService {

    // Write operation — override the class-level readOnly
    @Transactional
    public Order createOrder(CreateOrderCommand command) {
        // Full read-write transaction
        Order order = new Order(command);
        return orderRepository.save(order);
    }

    // Read-only query — uses class-level @Transactional(readOnly = true)
    public Order findById(OrderId id) {
        return orderRepository.findById(id.value())
            .orElseThrow(() -> new OrderNotFoundException(id));
    }

    // Specific isolation level
    @Transactional(isolation = Isolation.SERIALIZABLE)
    public void transferInventory(UUID fromWarehouse, UUID toWarehouse, int quantity) {
        // Serializable prevents phantom reads in inventory counting
    }

    // Custom rollback rules
    @Transactional(
        rollbackFor = BusinessException.class,        // Rollback on this checked exception
        noRollbackFor = NotificationException.class   // Don't rollback for notification failures
    )
    public OrderConfirmation submitOrder(OrderId orderId) {
        Order order = findById(orderId);
        order.submit();
        orderRepository.save(order);

        try {
            notificationService.notifyOrderSubmitted(order);  // Non-critical
        } catch (NotificationException e) {
            log.warn("Failed to send notification: {}", e.getMessage());
            // Transaction NOT rolled back
        }

        return new OrderConfirmation(order);
    }

    // Propagation: REQUIRES_NEW — always create new transaction
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void auditOrderAction(OrderId orderId, String action) {
        // Runs in separate transaction — committed even if outer transaction rolls back
        auditRepository.save(new AuditEntry(orderId, action, Instant.now()));
    }
}
```

### Transaction Pitfalls

```java
// PITFALL 1: Self-invocation bypasses proxy
@Service
public class OrderService {
    // BAD — calling @Transactional method from within same class bypasses proxy
    public void processOrders(List<OrderId> orderIds) {
        for (OrderId id : orderIds) {
            processOrder(id);  // NOT transactional! Calls method directly, not through proxy
        }
    }

    @Transactional
    public void processOrder(OrderId id) {
        // This @Transactional is ignored when called from processOrders
    }

    // SOLUTION 1: Inject self (Spring will inject the proxy)
    @Autowired
    private OrderService self;  // Proxy reference

    public void processOrdersFixed(List<OrderId> orderIds) {
        for (OrderId id : orderIds) {
            self.processOrder(id);  // Goes through proxy — @Transactional works
        }
    }

    // SOLUTION 2: Move to separate bean
    // OrderProcessingService.processOrder() calls OrderService.processOrder()
}

// PITFALL 2: readOnly doesn't prevent writes in all DBs
// readOnly = true is a HINT — some JDBC drivers optimize reads
// PostgreSQL: routes to read replica, optimizes query plans
// But Hibernate won't prevent you from calling save() — it just won't flush

// PITFALL 3: Transaction scope too large
// BAD — holding transaction open during HTTP call
@Transactional
public void processOrderBad(OrderId id) {
    Order order = orderRepository.findById(id);
    PaymentResult result = paymentClient.charge(order);  // HTTP call inside transaction!
    order.confirm(result.transactionId());
    orderRepository.save(order);
}

// GOOD — narrow transaction scope
public void processOrderGood(OrderId id) {
    Order order = orderRepository.findById(id);                    // Read (no tx needed)
    PaymentResult result = paymentClient.charge(order);             // HTTP call (no tx)
    confirmOrder(id, result.transactionId());                       // Write (transactional)
}

@Transactional
public void confirmOrder(OrderId id, String transactionId) {
    Order order = orderRepository.findById(id);
    order.confirm(transactionId);
    orderRepository.save(order);
}
```

## JdbcClient (Spring 6.1+)

```java
// JdbcClient — fluent API for JDBC operations
// Lighter than JPA for read queries and bulk operations
@Repository
public class OrderQueryRepository {
    private final JdbcClient jdbcClient;

    // Simple query
    public Optional<OrderView> findById(UUID id) {
        return jdbcClient.sql("""
                SELECT id, customer_id, status, total_amount, currency, created_at
                FROM orders WHERE id = :id
                """)
            .param("id", id)
            .query(OrderView.class)  // Auto-maps to record
            .optional();
    }

    // List query with parameters
    public List<OrderView> findByCustomer(UUID customerId, String status) {
        return jdbcClient.sql("""
                SELECT id, customer_id, status, total_amount, currency, created_at
                FROM orders
                WHERE customer_id = :customerId
                  AND (:status IS NULL OR status = :status)
                ORDER BY created_at DESC
                """)
            .param("customerId", customerId)
            .param("status", status)
            .query(OrderView.class)
            .list();
    }

    // Insert
    public void insert(OrderEntity order) {
        jdbcClient.sql("""
                INSERT INTO orders (id, customer_id, status, total_amount, currency, created_at)
                VALUES (:id, :customerId, :status, :totalAmount, :currency, :createdAt)
                """)
            .param("id", order.getId())
            .param("customerId", order.getCustomerId())
            .param("status", order.getStatus())
            .param("totalAmount", order.getTotalAmount())
            .param("currency", order.getCurrency())
            .param("createdAt", order.getCreatedAt())
            .update();
    }

    // Aggregate query
    public OrderStats getStats(UUID customerId) {
        return jdbcClient.sql("""
                SELECT
                    COUNT(*) AS total_orders,
                    COALESCE(SUM(total_amount), 0) AS total_spent,
                    COALESCE(AVG(total_amount), 0) AS avg_order_value,
                    MAX(created_at) AS last_order_date
                FROM orders
                WHERE customer_id = :customerId
                """)
            .param("customerId", customerId)
            .query(OrderStats.class)
            .single();
    }

    // Custom row mapper for complex mappings
    public List<OrderWithCustomer> findWithCustomerDetails() {
        return jdbcClient.sql("""
                SELECT o.id, o.status, o.total_amount,
                       c.name AS customer_name, c.email AS customer_email
                FROM orders o
                JOIN customers c ON c.id = o.customer_id
                ORDER BY o.created_at DESC
                LIMIT 100
                """)
            .query((rs, rowNum) -> new OrderWithCustomer(
                rs.getObject("id", UUID.class),
                rs.getString("status"),
                rs.getBigDecimal("total_amount"),
                rs.getString("customer_name"),
                rs.getString("customer_email")
            ))
            .list();
    }
}
```

## N+1 Query Problem

### Detection and Solutions

```java
// THE PROBLEM: Loading orders, then lazily loading lines for each order
// Bad: 1 query for orders + N queries for lines = N+1 queries

// SOLUTION 1: JOIN FETCH in JPQL
@Query("SELECT o FROM OrderEntity o JOIN FETCH o.lines WHERE o.customerId = :customerId")
List<OrderEntity> findByCustomerIdWithLines(@Param("customerId") UUID customerId);
// Warning: JOIN FETCH with pagination doesn't work well — Hibernate loads ALL data

// SOLUTION 2: @EntityGraph — declarative eager loading
@EntityGraph(attributePaths = {"lines", "lines.product"})
List<OrderEntity> findByCustomerId(UUID customerId);

// Named entity graph
@Entity
@NamedEntityGraph(name = "Order.withLinesAndProducts",
    attributeNodes = {
        @NamedAttributeNode(value = "lines", subgraph = "lines-subgraph")
    },
    subgraphs = {
        @NamedSubgraph(name = "lines-subgraph",
            attributeNodes = @NamedAttributeNode("product"))
    })
public class OrderEntity { /* ... */ }

// Use named graph
@EntityGraph("Order.withLinesAndProducts")
Optional<OrderEntity> findById(UUID id);

// SOLUTION 3: @BatchSize — batch lazy loading
@Entity
public class OrderEntity {
    @OneToMany(mappedBy = "order", fetch = FetchType.LAZY)
    @BatchSize(size = 20)  // Load lines in batches of 20 orders
    private List<OrderLineEntity> lines;
}
// Result: 1 query for orders + ceil(N/20) queries for lines

// SOLUTION 4: Hibernate default_batch_fetch_size (global)
// spring.jpa.properties.hibernate.default_batch_fetch_size=20

// SOLUTION 5: Use JdbcClient for read models (skip JPA entirely)
public List<OrderWithLines> findOrdersWithLines(UUID customerId) {
    Map<UUID, OrderWithLines> orderMap = new LinkedHashMap<>();

    jdbcClient.sql("""
            SELECT o.id, o.status, o.total_amount,
                   ol.id AS line_id, ol.product_id, ol.quantity, ol.unit_price
            FROM orders o
            LEFT JOIN order_lines ol ON ol.order_id = o.id
            WHERE o.customer_id = :customerId
            ORDER BY o.created_at DESC, ol.created_at ASC
            """)
        .param("customerId", customerId)
        .query((rs, rowNum) -> {
            UUID orderId = rs.getObject("id", UUID.class);
            orderMap.computeIfAbsent(orderId, id -> new OrderWithLines(
                id, rs.getString("status"), rs.getBigDecimal("total_amount"), new ArrayList<>()
            ));

            UUID lineId = rs.getObject("line_id", UUID.class);
            if (lineId != null) {
                orderMap.get(orderId).lines().add(new OrderLineView(
                    lineId, rs.getObject("product_id", UUID.class),
                    rs.getInt("quantity"), rs.getBigDecimal("unit_price")
                ));
            }
            return null;
        })
        .list();

    return new ArrayList<>(orderMap.values());
}
```

## Database Migrations with Flyway

```sql
-- V1__initial_schema.sql
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    total_amount DECIMAL(19,2),
    currency VARCHAR(3) DEFAULT 'USD',
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_metadata ON orders USING gin(metadata);

CREATE TABLE order_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(19,2) NOT NULL CHECK (unit_price >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_lines_order_id ON order_lines(order_id);

-- V2__add_outbox_table.sql
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY,
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    published BOOLEAN NOT NULL DEFAULT FALSE,
    published_at TIMESTAMP
);

CREATE INDEX idx_outbox_unpublished ON outbox_events(created_at) WHERE published = FALSE;
```

## Pagination Best Practices

```java
// Pageable request handling
@GetMapping("/orders")
public Page<OrderSummary> listOrders(
    @RequestParam(defaultValue = "0") int page,
    @RequestParam(defaultValue = "20") int size,
    @RequestParam(defaultValue = "createdAt,desc") String[] sort) {

    // Cap page size to prevent abuse
    int cappedSize = Math.min(size, 100);

    Pageable pageable = PageRequest.of(page, cappedSize, parseSort(sort));
    return orderRepository.findSummaries(pageable);
}

// Keyset/cursor pagination — much better performance for large datasets
@GetMapping("/orders")
public CursorPage<OrderSummary> listOrdersByCursor(
    @RequestParam(required = false) String cursor,
    @RequestParam(defaultValue = "20") int limit) {

    int cappedLimit = Math.min(limit, 100);

    if (cursor != null) {
        CursorInfo info = decodeCursor(cursor);
        List<OrderSummary> orders = orderRepository.findAfterCursor(
            info.createdAt(), info.id(), cappedLimit + 1);

        boolean hasMore = orders.size() > cappedLimit;
        List<OrderSummary> page = hasMore ? orders.subList(0, cappedLimit) : orders;
        String nextCursor = hasMore ? encodeCursor(page.getLast()) : null;

        return new CursorPage<>(page, nextCursor, hasMore);
    } else {
        List<OrderSummary> orders = orderRepository.findFirstPage(cappedLimit + 1);
        boolean hasMore = orders.size() > cappedLimit;
        List<OrderSummary> page = hasMore ? orders.subList(0, cappedLimit) : orders;
        String nextCursor = hasMore ? encodeCursor(page.getLast()) : null;

        return new CursorPage<>(page, nextCursor, hasMore);
    }
}

// Cursor-based query — uses index efficiently (no OFFSET!)
@Query(value = """
    SELECT id, status, total_amount, created_at
    FROM orders
    WHERE (created_at, id) < (:cursorDate, :cursorId)
    ORDER BY created_at DESC, id DESC
    LIMIT :limit
    """, nativeQuery = true)
List<OrderSummary> findAfterCursor(
    @Param("cursorDate") Instant cursorDate,
    @Param("cursorId") UUID cursorId,
    @Param("limit") int limit);
```

## Auditing

```java
// Enable JPA auditing
@Configuration
@EnableJpaAuditing(auditorAwareRef = "springSecurityAuditorAware")
public class JpaAuditConfig {

    @Bean
    public AuditorAware<String> springSecurityAuditorAware() {
        return () -> Optional.ofNullable(SecurityContextHolder.getContext().getAuthentication())
            .filter(Authentication::isAuthenticated)
            .map(Authentication::getName);
    }
}

// Auditable base entity
@MappedSuperclass
@EntityListeners(AuditingEntityListener.class)
public abstract class AuditableEntity {

    @CreatedDate
    @Column(name = "created_at", nullable = false, updatable = false)
    private Instant createdAt;

    @LastModifiedDate
    @Column(name = "updated_at", nullable = false)
    private Instant updatedAt;

    @CreatedBy
    @Column(name = "created_by", updatable = false)
    private String createdBy;

    @LastModifiedBy
    @Column(name = "updated_by")
    private String updatedBy;
}

// Envers for full audit history
@Entity
@Audited  // Hibernate Envers — stores all changes in _AUD table
@Table(name = "orders")
public class OrderEntity extends AuditableEntity {
    @Id
    private UUID id;

    @Audited
    private String status;

    @Audited
    private BigDecimal totalAmount;

    @NotAudited  // Skip auditing for this field
    private String internalNotes;
}

// Query audit history
@Service
public class OrderAuditService {
    @PersistenceContext
    private EntityManager entityManager;

    public List<OrderRevision> getOrderHistory(UUID orderId) {
        AuditReader auditReader = AuditReaderFactory.get(entityManager);

        List<Number> revisions = auditReader.getRevisions(OrderEntity.class, orderId);

        return revisions.stream()
            .map(rev -> {
                OrderEntity entity = auditReader.find(OrderEntity.class, orderId, rev);
                RevisionType type = auditReader.findRevision(
                    DefaultRevisionEntity.class, rev).getClass() != null
                    ? RevisionType.MOD : RevisionType.ADD;

                return new OrderRevision(rev.longValue(), entity, type);
            })
            .toList();
    }
}
```
