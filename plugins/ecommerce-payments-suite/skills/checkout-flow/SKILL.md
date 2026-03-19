---
name: checkout-flow
description: >
  Build complete e-commerce checkout flows with cart management, shipping calculation,
  tax handling, promo codes, and order confirmation. Includes React components,
  server-side validation, and conversion optimization patterns.
  Triggers: "checkout flow", "shopping cart", "build checkout", "cart page",
  "order flow", "purchase flow".
  NOT for: payment processor-specific setup (use stripe-integration skill).
version: 1.0.0
argument-hint: "[simple|multi-step|single-page]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Checkout Flow Builder

Build a complete e-commerce checkout experience.

## Step 1: Cart State Management

### React Context Cart

```tsx
interface CartItem {
  id: string;
  name: string;
  price: number;    // cents
  quantity: number;
  image?: string;
  variant?: string;
}

interface CartState {
  items: CartItem[];
  addItem: (item: Omit<CartItem, 'quantity'>) => void;
  removeItem: (id: string) => void;
  updateQuantity: (id: string, quantity: number) => void;
  clearCart: () => void;
  subtotal: number;
  itemCount: number;
}

const CartContext = createContext<CartState | null>(null);

export function CartProvider({ children }: { children: React.ReactNode }) {
  const [items, setItems] = useState<CartItem[]>(() => {
    if (typeof window === 'undefined') return [];
    const saved = localStorage.getItem('cart');
    return saved ? JSON.parse(saved) : [];
  });

  // Persist to localStorage
  useEffect(() => {
    localStorage.setItem('cart', JSON.stringify(items));
  }, [items]);

  const addItem = (item: Omit<CartItem, 'quantity'>) => {
    setItems(prev => {
      const existing = prev.find(i => i.id === item.id);
      if (existing) {
        return prev.map(i => i.id === item.id ? { ...i, quantity: i.quantity + 1 } : i);
      }
      return [...prev, { ...item, quantity: 1 }];
    });
  };

  const removeItem = (id: string) => {
    setItems(prev => prev.filter(i => i.id !== id));
  };

  const updateQuantity = (id: string, quantity: number) => {
    if (quantity <= 0) return removeItem(id);
    setItems(prev => prev.map(i => i.id === id ? { ...i, quantity } : i));
  };

  const clearCart = () => setItems([]);

  const subtotal = items.reduce((sum, item) => sum + item.price * item.quantity, 0);
  const itemCount = items.reduce((sum, item) => sum + item.quantity, 0);

  return (
    <CartContext.Provider value={{ items, addItem, removeItem, updateQuantity, clearCart, subtotal, itemCount }}>
      {children}
    </CartContext.Provider>
  );
}

export const useCart = () => {
  const context = useContext(CartContext);
  if (!context) throw new Error('useCart must be used within CartProvider');
  return context;
};
```

## Step 2: Cart UI

```tsx
function CartPage() {
  const { items, updateQuantity, removeItem, subtotal, itemCount } = useCart();

  if (items.length === 0) {
    return (
      <div className="text-center py-16">
        <p className="text-xl text-gray-500">Your cart is empty</p>
        <a href="/products" className="text-blue-600 underline mt-4 inline-block">
          Continue shopping
        </a>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6">Your Cart ({itemCount} items)</h1>

      <div className="space-y-4">
        {items.map(item => (
          <div key={item.id} className="flex items-center gap-4 border-b pb-4">
            {item.image && (
              <img src={item.image} alt={item.name} className="w-20 h-20 object-cover rounded" />
            )}
            <div className="flex-1">
              <h3 className="font-medium">{item.name}</h3>
              {item.variant && <p className="text-sm text-gray-500">{item.variant}</p>}
              <p className="font-medium">${(item.price / 100).toFixed(2)}</p>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => updateQuantity(item.id, item.quantity - 1)}
                className="w-8 h-8 border rounded"
                aria-label={`Decrease quantity of ${item.name}`}
              >
                -
              </button>
              <span className="w-8 text-center">{item.quantity}</span>
              <button
                onClick={() => updateQuantity(item.id, item.quantity + 1)}
                className="w-8 h-8 border rounded"
                aria-label={`Increase quantity of ${item.name}`}
              >
                +
              </button>
            </div>
            <p className="w-24 text-right font-medium">
              ${((item.price * item.quantity) / 100).toFixed(2)}
            </p>
            <button
              onClick={() => removeItem(item.id)}
              className="text-red-500 hover:text-red-700"
              aria-label={`Remove ${item.name} from cart`}
            >
              Remove
            </button>
          </div>
        ))}
      </div>

      <div className="mt-8 border-t pt-6">
        <div className="flex justify-between text-lg font-bold">
          <span>Subtotal</span>
          <span>${(subtotal / 100).toFixed(2)}</span>
        </div>
        <p className="text-sm text-gray-500 mt-1">Shipping and taxes calculated at checkout</p>
        <a
          href="/checkout"
          className="block w-full bg-blue-600 text-white text-center py-3 rounded-lg mt-4 font-medium hover:bg-blue-700"
        >
          Proceed to Checkout
        </a>
      </div>
    </div>
  );
}
```

## Step 3: Checkout Page

### Single-Page Checkout

```tsx
function CheckoutPage() {
  const { items, subtotal } = useCart();
  const [step, setStep] = useState<'info' | 'payment'>('info');
  const [shippingInfo, setShippingInfo] = useState({});
  const [shippingCost, setShippingCost] = useState(0);
  const [tax, setTax] = useState(0);
  const [promoDiscount, setPromoDiscount] = useState(0);

  const total = subtotal + shippingCost + tax - promoDiscount;

  return (
    <div className="max-w-6xl mx-auto p-6 grid grid-cols-1 lg:grid-cols-5 gap-8">
      {/* Left: Form */}
      <div className="lg:col-span-3">
        <ContactSection />
        <ShippingSection
          onCalculate={(cost) => setShippingCost(cost)}
          onInfoComplete={(info) => {
            setShippingInfo(info);
            calculateTax(info).then(setTax);
          }}
        />
        <PromoCodeSection onApply={(discount) => setPromoDiscount(discount)} />
        <PaymentSection total={total} shippingInfo={shippingInfo} />
      </div>

      {/* Right: Order Summary (sticky) */}
      <div className="lg:col-span-2">
        <div className="sticky top-6 border rounded-lg p-6">
          <h2 className="font-bold text-lg mb-4">Order Summary</h2>
          {items.map(item => (
            <div key={item.id} className="flex justify-between text-sm py-2">
              <span>{item.name} x{item.quantity}</span>
              <span>${((item.price * item.quantity) / 100).toFixed(2)}</span>
            </div>
          ))}
          <div className="border-t mt-4 pt-4 space-y-2 text-sm">
            <div className="flex justify-between">
              <span>Subtotal</span>
              <span>${(subtotal / 100).toFixed(2)}</span>
            </div>
            <div className="flex justify-between">
              <span>Shipping</span>
              <span>{shippingCost === 0 ? 'Calculated next' : `$${(shippingCost / 100).toFixed(2)}`}</span>
            </div>
            {promoDiscount > 0 && (
              <div className="flex justify-between text-green-600">
                <span>Discount</span>
                <span>-${(promoDiscount / 100).toFixed(2)}</span>
              </div>
            )}
            <div className="flex justify-between">
              <span>Tax</span>
              <span>{tax === 0 ? 'Calculated next' : `$${(tax / 100).toFixed(2)}`}</span>
            </div>
          </div>
          <div className="border-t mt-4 pt-4 flex justify-between font-bold text-lg">
            <span>Total</span>
            <span>${(total / 100).toFixed(2)}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
```

## Step 4: Promo Codes

### Server-Side Promo Validation

```typescript
// Database schema
interface PromoCode {
  code: string;
  type: 'percentage' | 'fixed';
  value: number; // percentage (0-100) or cents
  minOrderAmount?: number;
  maxUses?: number;
  currentUses: number;
  expiresAt?: Date;
  applicableProducts?: string[];
}

app.post('/api/promo/validate', async (req, res) => {
  const { code, subtotal, items } = req.body;

  const promo = await db.promoCode.findUnique({ where: { code: code.toUpperCase() } });

  if (!promo) return res.status(404).json({ error: 'Invalid promo code' });
  if (promo.expiresAt && promo.expiresAt < new Date()) return res.status(400).json({ error: 'Promo code expired' });
  if (promo.maxUses && promo.currentUses >= promo.maxUses) return res.status(400).json({ error: 'Promo code limit reached' });
  if (promo.minOrderAmount && subtotal < promo.minOrderAmount) {
    return res.status(400).json({ error: `Minimum order $${(promo.minOrderAmount / 100).toFixed(2)}` });
  }

  let discount: number;
  if (promo.type === 'percentage') {
    discount = Math.round(subtotal * (promo.value / 100));
  } else {
    discount = Math.min(promo.value, subtotal);
  }

  res.json({ discount, description: promo.type === 'percentage' ? `${promo.value}% off` : `$${(promo.value / 100).toFixed(2)} off` });
});
```

## Step 5: Order Confirmation

```tsx
function OrderConfirmation({ orderId }: { orderId: string }) {
  const [order, setOrder] = useState(null);

  useEffect(() => {
    fetch(`/api/orders/${orderId}`).then(r => r.json()).then(setOrder);
  }, [orderId]);

  if (!order) return <div>Loading...</div>;

  return (
    <div className="max-w-2xl mx-auto p-6 text-center">
      <div className="text-green-500 text-6xl mb-4">✓</div>
      <h1 className="text-2xl font-bold">Order Confirmed!</h1>
      <p className="text-gray-600 mt-2">Order #{order.id}</p>
      <p className="text-gray-600">A confirmation email has been sent to {order.email}</p>

      <div className="border rounded-lg p-6 mt-8 text-left">
        <h2 className="font-bold mb-4">Order Summary</h2>
        {order.items.map(item => (
          <div key={item.id} className="flex justify-between py-2">
            <span>{item.name} x{item.quantity}</span>
            <span>${((item.price * item.quantity) / 100).toFixed(2)}</span>
          </div>
        ))}
        <div className="border-t mt-4 pt-4 flex justify-between font-bold">
          <span>Total Charged</span>
          <span>${(order.total / 100).toFixed(2)}</span>
        </div>
      </div>

      <div className="mt-8">
        <a href="/products" className="text-blue-600 underline">Continue Shopping</a>
      </div>
    </div>
  );
}
```

## Conversion Optimization Checklist

- [ ] Guest checkout (no account required)
- [ ] Progress indicator (step 1 of 2)
- [ ] Sticky order summary on desktop
- [ ] Real-time form validation
- [ ] Autocomplete attributes on all fields
- [ ] Clear error messages
- [ ] Loading state during payment processing
- [ ] Trust badges near payment section
- [ ] Mobile-responsive layout
- [ ] Back button doesn't lose form data
- [ ] Promo code field collapsed by default
- [ ] Clear total breakdown (subtotal, shipping, tax, discount)
- [ ] "Secure checkout" messaging