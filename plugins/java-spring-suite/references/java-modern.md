# Modern Java Reference (Java 21+)

Comprehensive reference for Java 21+ features including records, sealed classes, pattern matching,
virtual threads, and other modern Java capabilities for production applications.

## Records

### Basics

```java
// Records are immutable data carriers with auto-generated equals/hashCode/toString
public record Point(double x, double y) {}

// Usage
Point p = new Point(3.0, 4.0);
double x = p.x();        // accessor (no 'get' prefix)
String s = p.toString();  // "Point[x=3.0, y=4.0]"

// Records with validation (compact constructor)
public record Email(String value) {
    public Email {  // Compact constructor — no parameter list
        Objects.requireNonNull(value, "Email cannot be null");
        if (!value.matches("^[\\w.-]+@[\\w.-]+\\.[a-zA-Z]{2,}$")) {
            throw new IllegalArgumentException("Invalid email: " + value);
        }
        value = value.toLowerCase().trim();  // Normalize — reassigns parameter
    }
}

// Records with custom methods
public record Money(long amountInCents, String currency) {
    public static final Money ZERO = new Money(0, "USD");

    public Money {
        if (amountInCents < 0) throw new IllegalArgumentException("Negative amount");
        Objects.requireNonNull(currency);
    }

    public Money add(Money other) {
        if (!this.currency.equals(other.currency)) {
            throw new CurrencyMismatchException(this.currency, other.currency);
        }
        return new Money(this.amountInCents + other.amountInCents, this.currency);
    }

    public Money multiply(int factor) {
        return new Money(this.amountInCents * factor, this.currency);
    }

    public String formatted() {
        return "%s %.2f".formatted(currency, amountInCents / 100.0);
    }
}
```

### Records as DTOs

```java
// Perfect for API requests/responses — immutable, concise
public record CreateOrderRequest(
    @NotEmpty List<OrderLineRequest> lines,
    @NotBlank String currency
) {
    public record OrderLineRequest(
        @NotNull UUID productId,
        @Min(1) int quantity,
        @Positive long unitPrice
    ) {}
}

public record OrderResponse(
    UUID id,
    String status,
    List<OrderLineResponse> lines,
    long totalAmount,
    String currency,
    Instant createdAt
) {
    public static OrderResponse from(Order order) {
        return new OrderResponse(
            order.id().value(),
            order.status().name(),
            order.lines().stream().map(OrderLineResponse::from).toList(),
            order.calculateTotal().amountInCents(),
            order.calculateTotal().currency(),
            order.createdAt()
        );
    }

    public record OrderLineResponse(UUID productId, int quantity, long unitPrice) {
        public static OrderLineResponse from(OrderLine line) {
            return new OrderLineResponse(
                line.productId().value(), line.quantity(), line.unitPrice().amountInCents());
        }
    }
}

// Records as value objects in DDD
public record OrderId(UUID value) {
    public OrderId { Objects.requireNonNull(value); }
    public static OrderId generate() { return new OrderId(UUID.randomUUID()); }
    @Override public String toString() { return value.toString(); }
}

public record CustomerId(UUID value) {
    public CustomerId { Objects.requireNonNull(value); }
}
```

## Sealed Classes and Interfaces

### Exhaustive Type Hierarchies

```java
// Sealed types restrict which classes can extend/implement them
// The compiler knows ALL subtypes — enables exhaustive switch

public sealed interface Shape
    permits Circle, Rectangle, Triangle {
}

public record Circle(double radius) implements Shape {}
public record Rectangle(double width, double height) implements Shape {}
public record Triangle(double base, double height) implements Shape {}

// Exhaustive switch — compiler error if a case is missing
public double area(Shape shape) {
    return switch (shape) {
        case Circle c -> Math.PI * c.radius() * c.radius();
        case Rectangle r -> r.width() * r.height();
        case Triangle t -> 0.5 * t.base() * t.height();
        // No default needed — compiler knows all cases
    };
}
```

### Sealed Classes for Domain Modeling

```java
// Model a finite state machine with sealed types
public sealed interface OrderStatus permits
    OrderStatus.Draft,
    OrderStatus.Submitted,
    OrderStatus.PaymentPending,
    OrderStatus.Confirmed,
    OrderStatus.Shipped,
    OrderStatus.Delivered,
    OrderStatus.Cancelled {

    record Draft() implements OrderStatus {}
    record Submitted(Instant at) implements OrderStatus {}
    record PaymentPending(String paymentIntentId) implements OrderStatus {}
    record Confirmed(Instant at, String transactionId) implements OrderStatus {}
    record Shipped(Instant at, String trackingNumber) implements OrderStatus {}
    record Delivered(Instant at, String signedBy) implements OrderStatus {}
    record Cancelled(Instant at, String reason) implements OrderStatus {}

    // Business logic driven by status
    default boolean canBeCancelled() {
        return switch (this) {
            case Draft d -> true;
            case Submitted s -> true;
            case PaymentPending p -> true;
            case Confirmed c, Shipped s, Delivered d, Cancelled ca -> false;
        };
    }

    default String displayName() {
        return switch (this) {
            case Draft d -> "Draft";
            case Submitted s -> "Submitted";
            case PaymentPending p -> "Payment Pending";
            case Confirmed c -> "Confirmed";
            case Shipped s -> "Shipped";
            case Delivered d -> "Delivered";
            case Cancelled c -> "Cancelled — " + c.reason();
        };
    }
}

// Sealed class for API responses
public sealed interface ApiResponse<T> permits
    ApiResponse.Success, ApiResponse.Error {

    record Success<T>(T data, Map<String, String> metadata) implements ApiResponse<T> {}
    record Error<T>(int code, String message, List<String> details) implements ApiResponse<T> {}

    static <T> ApiResponse<T> success(T data) {
        return new Success<>(data, Map.of());
    }

    static <T> ApiResponse<T> error(int code, String message) {
        return new Error<>(code, message, List.of());
    }
}

// Sealed class for computation results
public sealed interface Result<T> permits Result.Ok, Result.Err {
    record Ok<T>(T value) implements Result<T> {}
    record Err<T>(String error, Throwable cause) implements Result<T> {}

    static <T> Result<T> ok(T value) { return new Ok<>(value); }
    static <T> Result<T> err(String error) { return new Err<>(error, null); }
    static <T> Result<T> err(String error, Throwable cause) { return new Err<>(error, cause); }

    default T orElseThrow() {
        return switch (this) {
            case Ok<T> ok -> ok.value();
            case Err<T> err -> throw new RuntimeException(err.error(), err.cause());
        };
    }

    default <U> Result<U> map(Function<T, U> mapper) {
        return switch (this) {
            case Ok<T> ok -> Result.ok(mapper.apply(ok.value()));
            case Err<T> err -> Result.err(err.error(), err.cause());
        };
    }

    default <U> Result<U> flatMap(Function<T, Result<U>> mapper) {
        return switch (this) {
            case Ok<T> ok -> mapper.apply(ok.value());
            case Err<T> err -> Result.err(err.error(), err.cause());
        };
    }
}
```

## Pattern Matching

### instanceof Pattern Matching

```java
// Pattern matching replaces casting boilerplate
// Before Java 16:
if (obj instanceof String) {
    String s = (String) obj;
    System.out.println(s.length());
}

// Java 16+:
if (obj instanceof String s) {
    System.out.println(s.length());  // s already cast and in scope
}

// Works with &&
if (obj instanceof String s && s.length() > 5) {
    processLongString(s);
}

// Negation pattern
if (!(obj instanceof String s)) {
    return;  // Early return
}
// s is in scope here
process(s);
```

### Switch Pattern Matching (Java 21)

```java
// Pattern matching in switch — the most powerful Java 21 feature
public String describe(Object obj) {
    return switch (obj) {
        case Integer i when i < 0 -> "negative integer: " + i;
        case Integer i -> "positive integer: " + i;
        case String s when s.isBlank() -> "blank string";
        case String s -> "string of length " + s.length();
        case List<?> list when list.isEmpty() -> "empty list";
        case List<?> list -> "list with " + list.size() + " elements";
        case null -> "null value";
        default -> "unknown: " + obj.getClass().getSimpleName();
    };
}

// Guarded patterns with 'when' clause
public BigDecimal calculateDiscount(Customer customer, Order order) {
    return switch (customer) {
        case PremiumCustomer p when order.total().compareTo(BigDecimal.valueOf(1000)) > 0 ->
            BigDecimal.valueOf(0.20);  // 20% for premium with large orders
        case PremiumCustomer p ->
            BigDecimal.valueOf(0.10);  // 10% for premium
        case RegularCustomer r when r.ordersThisYear() > 10 ->
            BigDecimal.valueOf(0.05);  // 5% for loyal regular customers
        case RegularCustomer r ->
            BigDecimal.ZERO;
        case NewCustomer n ->
            BigDecimal.valueOf(0.15);  // Welcome discount
    };
}

// Record pattern matching (Java 21) — deconstruct records in patterns
public record Address(String street, String city, String state, String zip) {}
public record Customer(String name, Address address) {}

public String formatForShipping(Object obj) {
    return switch (obj) {
        case Customer(var name, Address(var street, var city, var state, var zip))
            when state.equals("CA") ->
                "%s\n%s\n%s, %s %s\n(California sales tax applies)".formatted(
                    name, street, city, state, zip);

        case Customer(var name, Address(var street, var city, var state, var zip)) ->
                "%s\n%s\n%s, %s %s".formatted(name, street, city, state, zip);

        default -> obj.toString();
    };
}

// Nested record patterns
public record Order(OrderId id, Customer customer, List<OrderLine> lines, Money total) {}

public void processOrder(Order order) {
    switch (order) {
        case Order(var id, Customer(var name, Address(_, _, var state, _)), var lines, var total)
            when state.equals("NY") && total.amountInCents() > 10000 ->
                applyNYLuxuryTax(id, total);

        case Order(_, _, var lines, _) when lines.isEmpty() ->
            throw new EmptyOrderException();

        case Order(var id, _, _, var total) ->
            processStandard(id, total);
    }
}
```

### Unnamed Variables (Java 22+)

```java
// Use _ for variables you don't need
try {
    processOrder(order);
} catch (OrderNotFoundException _) {
    return OrderResult.notFound();
} catch (PaymentException _) {
    return OrderResult.paymentFailed();
}

// In enhanced for loops
int count = 0;
for (var _ : collection) {
    count++;
}

// In switch patterns
return switch (shape) {
    case Circle(var radius) -> Math.PI * radius * radius;
    case Rectangle(var w, var h) -> w * h;
    case Triangle(var b, var h) -> 0.5 * b * h;
};

// In lambda parameters
map.forEach((_, value) -> process(value));
```

## Virtual Threads (Java 21)

### Basic Usage

```java
// Create a virtual thread
Thread vt = Thread.ofVirtual().start(() -> {
    System.out.println("Running on virtual thread: " + Thread.currentThread());
});

// Named virtual threads (for debugging)
Thread.ofVirtual().name("order-processor-", 0).start(() -> {
    // Thread name: order-processor-0
});

// Virtual thread executor — one virtual thread per task
try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
    // Submit 10,000 tasks — each gets its own virtual thread
    List<Future<String>> futures = IntStream.range(0, 10_000)
        .mapToObj(i -> executor.submit(() -> fetchUrl("https://api.example.com/" + i)))
        .toList();

    for (Future<String> future : futures) {
        String result = future.get();
        process(result);
    }
}
```

### Structured Concurrency (Preview)

```java
// Structured concurrency — child tasks are scoped to parent
public OrderDetails fetchOrderDetails(UUID orderId) throws Exception {
    try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {
        // Fork subtasks — each runs on its own virtual thread
        Subtask<Order> orderTask = scope.fork(() -> orderService.findById(orderId));
        Subtask<Customer> customerTask = scope.fork(() -> customerService.findByOrderId(orderId));
        Subtask<List<Product>> productsTask = scope.fork(() -> productService.findByOrderId(orderId));
        Subtask<ShippingInfo> shippingTask = scope.fork(() -> shippingService.getInfo(orderId));

        scope.join();           // Wait for all subtasks
        scope.throwIfFailed();  // Propagate any exceptions

        // All subtasks completed successfully
        return new OrderDetails(
            orderTask.get(),
            customerTask.get(),
            productsTask.get(),
            shippingTask.get()
        );
    }
    // If any subtask fails, all others are automatically cancelled
}

// ShutdownOnSuccess — return first successful result
public String fetchFromMirrors(String path) throws Exception {
    try (var scope = new StructuredTaskScope.ShutdownOnSuccess<String>()) {
        scope.fork(() -> fetchFrom("https://mirror1.example.com" + path));
        scope.fork(() -> fetchFrom("https://mirror2.example.com" + path));
        scope.fork(() -> fetchFrom("https://mirror3.example.com" + path));

        scope.join();

        return scope.result();  // First successful result
    }
}
```

### Scoped Values (Java 21 Preview)

```java
// ScopedValue — immutable, inheritable context for virtual threads
// Replacement for ThreadLocal in virtual thread world
public class RequestContext {
    public static final ScopedValue<String> REQUEST_ID = ScopedValue.newInstance();
    public static final ScopedValue<String> USER_ID = ScopedValue.newInstance();
    public static final ScopedValue<String> TENANT_ID = ScopedValue.newInstance();
}

// Set scoped values
public void handleRequest(HttpRequest request) {
    ScopedValue.where(RequestContext.REQUEST_ID, request.header("X-Request-ID"))
        .where(RequestContext.USER_ID, request.authenticatedUser())
        .where(RequestContext.TENANT_ID, request.header("X-Tenant-ID"))
        .run(() -> {
            processOrder();  // All code within can read these values
        });
}

// Read scoped values — available in same thread and child virtual threads
public void processOrder() {
    String requestId = RequestContext.REQUEST_ID.get();
    String userId = RequestContext.USER_ID.get();
    log.info("[{}] Processing order for user {}", requestId, userId);
}
```

## Text Blocks and String Templates

```java
// Text blocks (Java 15+) — multi-line strings
String json = """
    {
        "name": "%s",
        "email": "%s",
        "roles": ["USER"]
    }
    """.formatted(name, email);

// SQL queries
String sql = """
    SELECT o.id, o.status, o.total_amount,
           c.name AS customer_name
    FROM orders o
    JOIN customers c ON c.id = o.customer_id
    WHERE o.status = ?
      AND o.created_at > ?
    ORDER BY o.created_at DESC
    LIMIT ?
    """;

// HTML templates
String html = """
    <html>
    <body>
        <h1>Order Confirmation</h1>
        <p>Order #%s has been confirmed.</p>
        <p>Total: %s</p>
    </body>
    </html>
    """.formatted(orderId, total.formatted());
```

## Collections and Stream Enhancements

```java
// Immutable collections factory methods (Java 9+)
List<String> list = List.of("a", "b", "c");
Set<String> set = Set.of("x", "y", "z");
Map<String, Integer> map = Map.of("one", 1, "two", 2, "three", 3);

// Map.ofEntries for larger maps
Map<String, String> config = Map.ofEntries(
    Map.entry("host", "localhost"),
    Map.entry("port", "8080"),
    Map.entry("protocol", "https"),
    Map.entry("timeout", "30s")
);

// Stream.toList() (Java 16+) — shorter than Collectors.toList()
List<String> names = customers.stream()
    .map(Customer::name)
    .toList();  // Returns unmodifiable list

// Stream.mapMulti() (Java 16+) — one-to-many without flatMap overhead
List<OrderLine> allLines = orders.stream()
    .<OrderLine>mapMulti((order, consumer) -> {
        for (OrderLine line : order.lines()) {
            if (line.quantity() > 0) {
                consumer.accept(line);
            }
        }
    })
    .toList();

// Gatherers (Java 24, preview in 22) — custom intermediate operations
// Example: sliding window
List<List<Integer>> windows = numbers.stream()
    .gather(Gatherers.windowSliding(3))
    .toList();
// [1,2,3,4,5] -> [[1,2,3], [2,3,4], [3,4,5]]

// SequencedCollections (Java 21) — ordered access for all collections
SequencedCollection<String> names = new ArrayList<>(List.of("Alice", "Bob", "Charlie"));
String first = names.getFirst();       // "Alice"
String last = names.getLast();         // "Charlie"
names.addFirst("Zara");               // Add to beginning
SequencedCollection<String> reversed = names.reversed();  // View in reverse order

SequencedMap<String, Integer> map = new LinkedHashMap<>();
map.putFirst("first", 1);
map.putLast("last", 99);
Map.Entry<String, Integer> firstEntry = map.firstEntry();
Map.Entry<String, Integer> lastEntry = map.lastEntry();
```

## Enhanced Switch Expressions

```java
// Switch expressions return values (Java 14+)
String result = switch (dayOfWeek) {
    case MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY -> "Weekday";
    case SATURDAY, SUNDAY -> "Weekend";
};

// Block form with yield
int numLetters = switch (dayOfWeek) {
    case MONDAY, FRIDAY, SUNDAY -> 6;
    case TUESDAY -> 7;
    case WEDNESDAY -> 9;
    case THURSDAY -> {
        log.debug("Calculating for Thursday");
        yield 8;
    }
    case SATURDAY -> 8;
};

// Exhaustive switch with sealed types — compiler enforces all cases
public double shippingCost(ShippingMethod method, double weight) {
    return switch (method) {
        case Standard s -> weight * 0.50;
        case Express e -> weight * 1.50 + 5.99;
        case Overnight o -> weight * 3.00 + 15.99;
        case InStorePickup p -> 0.0;
        // No default — compiler error if new ShippingMethod added
    };
}
```

## Helpful NullPointerExceptions and Optional Improvements

```java
// Helpful NPEs (Java 14+) — JVM tells you exactly what was null
// Before: NullPointerException
// After:  NullPointerException: Cannot invoke "String.length()" because
//         "customer.getAddress().getCity()" is null

// Optional improvements
Optional<String> name = Optional.of("Alice");

// or() (Java 9) — lazy fallback
Optional<Customer> customer = findByEmail(email)
    .or(() -> findByUsername(username))     // Try alternate lookup
    .or(() -> findByPhoneNumber(phone));    // Another fallback

// ifPresentOrElse() (Java 9)
findCustomer(id).ifPresentOrElse(
    customer -> sendEmail(customer),
    () -> log.warn("Customer {} not found", id)
);

// stream() (Java 9) — integrate with Stream API
List<String> emails = customerIds.stream()
    .map(this::findCustomer)       // Stream<Optional<Customer>>
    .flatMap(Optional::stream)      // Stream<Customer> — empties filtered out
    .map(Customer::email)
    .toList();

// Improved Optional with pattern matching (idiom)
public String getDisplayName(Optional<Customer> customer) {
    return customer
        .map(c -> switch (c) {
            case PremiumCustomer p -> "★ " + p.name();
            case RegularCustomer r -> r.name();
            case NewCustomer n -> n.name() + " (new)";
        })
        .orElse("Guest");
}
```

## Foreign Function & Memory API (Java 22+)

```java
// Call native C functions without JNI
// Example: calling strlen from libc
public class NativeExample {

    public static long strlen(String str) throws Throwable {
        // Look up the native function
        Linker linker = Linker.nativeLinker();
        SymbolLookup stdlib = linker.defaultLookup();

        MethodHandle strlen = linker.downcallHandle(
            stdlib.find("strlen").orElseThrow(),
            FunctionDescriptor.of(ValueLayout.JAVA_LONG, ValueLayout.ADDRESS)
        );

        // Allocate off-heap memory for the string
        try (Arena arena = Arena.ofConfined()) {
            MemorySegment cString = arena.allocateFrom(str);
            return (long) strlen.invoke(cString);
        }
    }
}

// Off-heap memory management
public class OffHeapBuffer {
    public static void example() {
        try (Arena arena = Arena.ofConfined()) {
            // Allocate 1MB off-heap
            MemorySegment segment = arena.allocate(1024 * 1024);

            // Write data
            segment.setAtIndex(ValueLayout.JAVA_INT, 0, 42);
            segment.setAtIndex(ValueLayout.JAVA_INT, 1, 100);

            // Read data
            int value = segment.getAtIndex(ValueLayout.JAVA_INT, 0);  // 42

            // Memory automatically freed when arena closes
        }
    }
}
```
