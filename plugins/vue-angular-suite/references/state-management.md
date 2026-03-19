# State Management Reference: Vue & Angular

Quick-reference guide for state management across Vue (Pinia, Vuex) and Angular (NgRx Store, Component Store, SignalStore).

---

## State Management Overview

### When You Need State Management vs When You Don't

| Scenario | State Management Needed? | Recommended Approach |
|---|---|---|
| Single component form | No | Local component state |
| Parent-child data sharing (1-2 levels) | No | Props / `@Input()` + events |
| Shared state across sibling components | Maybe | Provide/Inject (Vue) or Service (Angular) |
| App-wide auth / user session | Yes | Global store |
| Complex data with derived views | Yes | Store with selectors |
| Server cache (API responses) | Depends | TanStack Query / Apollo may be better |
| Theme / locale preferences | Maybe | Simple reactive state or store |

### Local vs Shared vs Global State

- **Local state**: Owned by a single component. Use `ref()` / `reactive()` in Vue or component properties / signals in Angular.
- **Shared state**: Needed by a subtree of components. Use composables (Vue) or injectable services (Angular).
- **Global state**: Application-wide data accessible from any component. Use Pinia (Vue) or NgRx / SignalStore (Angular).

### Decision Tree: Which Solution to Pick

```
Is the state used by only one component?
  YES -> Use local component state (ref/reactive or signal)
  NO ->
    Is it shared between a parent and direct children?
      YES -> Use props/events (Vue) or @Input/@Output (Angular)
      NO ->
        Is it shared across unrelated components?
          YES ->
            Is the state complex with many derived values?
              YES -> Pinia Setup Store (Vue) / NgRx Store or SignalStore (Angular)
              NO -> Composable with ref (Vue) / Injectable service with signals (Angular)
          NO ->
            Is it server/cache state?
              YES -> TanStack Query, Apollo, or SWR
              NO -> Local state is fine
```

---

## Pinia (Vue)

### Setup Store Pattern (Preferred)

The setup store pattern uses a function that returns reactive state, computed getters, and action functions. It mirrors the Composition API.

#### Full Example with State, Getters, Actions

```typescript
// stores/useProductStore.ts
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { Product } from '@/types';
import { productApi } from '@/api/products';

export const useProductStore = defineStore('products', () => {
  // State
  const products = ref<Product[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const selectedCategory = ref<string | null>(null);

  // Getters (computed)
  const filteredProducts = computed(() => {
    if (!selectedCategory.value) return products.value;
    return products.value.filter(p => p.category === selectedCategory.value);
  });

  const totalPrice = computed(() =>
    products.value.reduce((sum, p) => sum + p.price, 0)
  );

  const productCount = computed(() => products.value.length);

  const productById = computed(() => {
    return (id: string) => products.value.find(p => p.id === id);
  });

  // Actions
  async function fetchProducts() {
    loading.value = true;
    error.value = null;
    try {
      products.value = await productApi.getAll();
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch products';
    } finally {
      loading.value = false;
    }
  }

  function addProduct(product: Product) {
    products.value.push(product);
  }

  function removeProduct(id: string) {
    products.value = products.value.filter(p => p.id !== id);
  }

  function setCategory(category: string | null) {
    selectedCategory.value = category;
  }

  return {
    products,
    loading,
    error,
    selectedCategory,
    filteredProducts,
    totalPrice,
    productCount,
    productById,
    fetchProducts,
    addProduct,
    removeProduct,
    setCategory,
  };
});
```

#### Composing Stores

```typescript
// stores/useCartStore.ts
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { useProductStore } from './useProductStore';
import { useUserStore } from './useUserStore';

export const useCartStore = defineStore('cart', () => {
  const productStore = useProductStore();
  const userStore = useUserStore();

  const cartItemIds = ref<Map<string, number>>(new Map());

  const cartItems = computed(() => {
    const items: Array<{ product: Product; quantity: number }> = [];
    for (const [id, quantity] of cartItemIds.value) {
      const product = productStore.productById(id);
      if (product) items.push({ product, quantity });
    }
    return items;
  });

  const cartTotal = computed(() =>
    cartItems.value.reduce(
      (sum, item) => sum + item.product.price * item.quantity,
      0
    )
  );

  const discountedTotal = computed(() => {
    const discount = userStore.isPremium ? 0.9 : 1;
    return cartTotal.value * discount;
  });

  function addToCart(productId: string, quantity = 1) {
    const current = cartItemIds.value.get(productId) ?? 0;
    cartItemIds.value.set(productId, current + quantity);
  }

  function removeFromCart(productId: string) {
    cartItemIds.value.delete(productId);
  }

  return { cartItemIds, cartItems, cartTotal, discountedTotal, addToCart, removeFromCart };
});
```

#### Async Actions with Error Handling

```typescript
// stores/useOrderStore.ts
import { defineStore } from 'pinia';
import { ref } from 'vue';
import { orderApi } from '@/api/orders';

export const useOrderStore = defineStore('orders', () => {
  const orders = ref<Order[]>([]);
  const submitting = ref(false);
  const error = ref<string | null>(null);

  async function submitOrder(orderData: CreateOrderDto) {
    submitting.value = true;
    error.value = null;
    try {
      const newOrder = await orderApi.create(orderData);
      orders.value.push(newOrder);
      return newOrder;
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Order submission failed';
      throw e; // Re-throw so the component can handle it too
    } finally {
      submitting.value = false;
    }
  }

  async function cancelOrder(orderId: string) {
    try {
      await orderApi.cancel(orderId);
      const index = orders.value.findIndex(o => o.id === orderId);
      if (index !== -1) {
        orders.value[index] = { ...orders.value[index], status: 'cancelled' };
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Cancellation failed';
      throw e;
    }
  }

  return { orders, submitting, error, submitOrder, cancelOrder };
});
```

### Options Store Pattern

```typescript
// stores/useCounterStore.ts
import { defineStore } from 'pinia';

export const useCounterStore = defineStore('counter', {
  state: () => ({
    count: 0,
    lastUpdated: null as Date | null,
  }),

  getters: {
    doubleCount: (state) => state.count * 2,
    isPositive: (state) => state.count > 0,
    // Getter that receives other getters via `this`
    summary(): string {
      return `Count: ${this.count}, Double: ${this.doubleCount}`;
    },
  },

  actions: {
    increment() {
      this.count++;
      this.lastUpdated = new Date();
    },
    decrement() {
      this.count--;
      this.lastUpdated = new Date();
    },
    async incrementAsync() {
      await new Promise(resolve => setTimeout(resolve, 500));
      this.increment();
    },
  },
});
```

#### When to Use Options vs Setup

| Criteria | Options Store | Setup Store |
|---|---|---|
| Familiarity | Vuex-like API | Composition API style |
| Composing stores | Awkward (`this` context) | Natural (call other stores directly) |
| TypeScript inference | Good but requires annotations for `this` in getters | Excellent, fully inferred |
| Flexibility | Structured | Freeform |
| Recommendation | Legacy or simple stores | Default choice for new projects |

### Pinia Plugins

#### Persisted State Plugin

```typescript
// plugins/piniaPersistedState.ts
import { PiniaPluginContext } from 'pinia';

export function piniaPersistedState({ store }: PiniaPluginContext) {
  const storedState = localStorage.getItem(`pinia-${store.$id}`);
  if (storedState) {
    store.$patch(JSON.parse(storedState));
  }

  store.$subscribe((_mutation, state) => {
    localStorage.setItem(`pinia-${store.$id}`, JSON.stringify(state));
  });
}

// main.ts
import { createPinia } from 'pinia';
import { piniaPersistedState } from './plugins/piniaPersistedState';

const pinia = createPinia();
pinia.use(piniaPersistedState);
```

#### Custom Plugin Creation

```typescript
// plugins/piniaTimestampPlugin.ts
import { PiniaPluginContext } from 'pinia';

declare module 'pinia' {
  export interface PiniaCustomProperties {
    lastModified: Date | null;
  }
}

export function timestampPlugin({ store }: PiniaPluginContext) {
  store.lastModified = null;

  store.$subscribe(() => {
    store.lastModified = new Date();
  });
}
```

#### Logger Plugin Example

```typescript
// plugins/piniaLogger.ts
import { PiniaPluginContext } from 'pinia';

export function piniaLogger({ store }: PiniaPluginContext) {
  store.$onAction(({ name, args, after, onError }) => {
    const startTime = performance.now();
    console.log(`[Pinia] Action "${store.$id}.${name}" started`, args);

    after((result) => {
      const duration = performance.now() - startTime;
      console.log(
        `[Pinia] Action "${store.$id}.${name}" completed in ${duration.toFixed(2)}ms`,
        result
      );
    });

    onError((error) => {
      const duration = performance.now() - startTime;
      console.error(
        `[Pinia] Action "${store.$id}.${name}" failed after ${duration.toFixed(2)}ms`,
        error
      );
    });
  });
}
```

### Pinia with Nuxt

#### Auto-imports

In Nuxt 3, Pinia stores placed in `stores/` are auto-imported. No manual registration needed.

```typescript
// stores/useAuthStore.ts — auto-imported by Nuxt
export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null);
  const token = ref<string | null>(null);

  const isAuthenticated = computed(() => !!token.value);

  async function login(credentials: LoginDto) {
    const response = await $fetch('/api/auth/login', {
      method: 'POST',
      body: credentials,
    });
    user.value = response.user;
    token.value = response.token;
  }

  function logout() {
    user.value = null;
    token.value = null;
    navigateTo('/login');
  }

  return { user, token, isAuthenticated, login, logout };
});
```

#### SSR Considerations

```typescript
// stores/useDataStore.ts
export const useDataStore = defineStore('data', () => {
  const items = ref<Item[]>([]);

  async function fetchItems() {
    // useAsyncData prevents duplicate fetches during SSR/hydration
    const { data } = await useAsyncData('items', () =>
      $fetch('/api/items')
    );
    if (data.value) items.value = data.value;
  }

  return { items, fetchItems };
});
```

#### Hydration

Pinia handles hydration automatically in Nuxt. State set during SSR is serialized into the HTML payload and rehydrated on the client. Avoid storing non-serializable values (functions, class instances) in state.

### Store Subscriptions

#### $subscribe for State Changes

```typescript
const productStore = useProductStore();

// Watch for any state change
productStore.$subscribe((mutation, state) => {
  console.log('Mutation type:', mutation.type); // 'direct' | 'patch object' | 'patch function'
  console.log('Store ID:', mutation.storeId);
  console.log('New state:', state);
});

// Detach when component unmounts (default in setup)
// Or keep alive beyond component:
productStore.$subscribe(callback, { detached: true });
```

#### $onAction for Action Tracking

```typescript
const store = useProductStore();

const unsubscribe = store.$onAction(({ name, store, args, after, onError }) => {
  console.log(`Action "${name}" called with args:`, args);

  after((result) => {
    console.log(`Action "${name}" resolved with:`, result);
  });

  onError((error) => {
    console.warn(`Action "${name}" threw:`, error);
  });
});

// Later: unsubscribe();
```

#### Debugging with Subscriptions

```typescript
// Debug plugin that logs all state changes with diffs
function debugPlugin({ store }: PiniaPluginContext) {
  let previousState = JSON.parse(JSON.stringify(store.$state));

  store.$subscribe((_mutation, state) => {
    const current = JSON.parse(JSON.stringify(state));
    const changes: Record<string, { from: unknown; to: unknown }> = {};

    for (const key of Object.keys(current)) {
      if (JSON.stringify(previousState[key]) !== JSON.stringify(current[key])) {
        changes[key] = { from: previousState[key], to: current[key] };
      }
    }

    if (Object.keys(changes).length > 0) {
      console.table(changes);
    }
    previousState = current;
  });
}
```

### Testing Pinia

#### Testing Stores in Isolation

```typescript
// stores/__tests__/useProductStore.spec.ts
import { setActivePinia, createPinia } from 'pinia';
import { useProductStore } from '../useProductStore';
import { vi, describe, it, expect, beforeEach } from 'vitest';

vi.mock('@/api/products', () => ({
  productApi: {
    getAll: vi.fn(),
  },
}));

import { productApi } from '@/api/products';

describe('useProductStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('fetches products', async () => {
    const mockProducts = [{ id: '1', name: 'Widget', price: 9.99, category: 'tools' }];
    vi.mocked(productApi.getAll).mockResolvedValue(mockProducts);

    const store = useProductStore();
    await store.fetchProducts();

    expect(store.products).toEqual(mockProducts);
    expect(store.loading).toBe(false);
    expect(store.error).toBeNull();
  });

  it('handles fetch errors', async () => {
    vi.mocked(productApi.getAll).mockRejectedValue(new Error('Network error'));

    const store = useProductStore();
    await store.fetchProducts();

    expect(store.products).toEqual([]);
    expect(store.error).toBe('Network error');
  });

  it('filters products by category', () => {
    const store = useProductStore();
    store.products = [
      { id: '1', name: 'Widget', price: 10, category: 'tools' },
      { id: '2', name: 'Gadget', price: 20, category: 'electronics' },
    ];
    store.setCategory('tools');
    expect(store.filteredProducts).toHaveLength(1);
    expect(store.filteredProducts[0].name).toBe('Widget');
  });
});
```

#### Testing Components with Stores

```typescript
// components/__tests__/ProductList.spec.ts
import { mount } from '@vue/test-utils';
import { createTestingPinia } from '@pinia/testing';
import ProductList from '../ProductList.vue';

describe('ProductList', () => {
  it('displays products from the store', () => {
    const wrapper = mount(ProductList, {
      global: {
        plugins: [
          createTestingPinia({
            initialState: {
              products: {
                products: [
                  { id: '1', name: 'Widget', price: 9.99 },
                  { id: '2', name: 'Gadget', price: 19.99 },
                ],
                loading: false,
              },
            },
          }),
        ],
      },
    });

    expect(wrapper.findAll('[data-testid="product-item"]')).toHaveLength(2);
  });

  it('shows loading state', () => {
    const wrapper = mount(ProductList, {
      global: {
        plugins: [
          createTestingPinia({
            initialState: {
              products: { loading: true, products: [] },
            },
          }),
        ],
      },
    });

    expect(wrapper.find('[data-testid="loading"]').exists()).toBe(true);
  });
});
```

---

## Vuex (Legacy)

### Brief Overview

Vuex is the predecessor to Pinia for Vue state management. It uses a single store with modules, mutations, actions, and getters. New projects should use Pinia instead.

### Vuex to Pinia Migration Guide

| Vuex Concept | Pinia Equivalent | Notes |
|---|---|---|
| `state` | `ref()` / `reactive()` in setup store, or `state()` in options store | Same role, different API |
| `getters` | `computed()` in setup store, or `getters` in options store | No root getter access needed |
| `mutations` | Removed entirely | Mutate state directly in actions |
| `actions` | Functions in setup store, or `actions` in options store | No `commit`, no `dispatch` |
| `modules` | Separate stores | Each store is independent, compose by importing |
| `mapState` | `storeToRefs()` | Preserves reactivity for state/getters |
| `mapActions` | Destructure directly | `const { fetchData } = useMyStore()` |
| `createStore()` | `createPinia()` | Installed as a Vue plugin |
| `this.$store` | `useMyStore()` | Called inside `setup()` or `<script setup>` |
| `namespaced: true` | Not needed | Every store is namespaced by its ID |
| `rootState` / `rootGetters` | Import the other store | `const other = useOtherStore()` |

### Key Differences

- Pinia has no mutations; modify state directly in actions.
- Pinia stores are independent; no module nesting.
- Full TypeScript support without extra type annotations.
- Pinia is lighter (~1KB vs ~5KB for Vuex).
- Pinia works with Vue 2 (via `@vue/composition-api`) and Vue 3.

---

## NgRx Store (Angular)

### Feature Store Setup

#### createActionGroup

```typescript
// store/products.actions.ts
import { createActionGroup, emptyProps, props } from '@ngrx/store';
import { Product } from '../models/product.model';

export const ProductActions = createActionGroup({
  source: 'Products',
  events: {
    'Load Products': emptyProps(),
    'Load Products Success': props<{ products: Product[] }>(),
    'Load Products Failure': props<{ error: string }>(),
    'Add Product': props<{ product: Product }>(),
    'Remove Product': props<{ id: string }>(),
    'Select Product': props<{ id: string }>(),
    'Update Product': props<{ product: Product }>(),
  },
});
```

#### createReducer with on()

```typescript
// store/products.reducer.ts
import { createReducer, on } from '@ngrx/store';
import { ProductActions } from './products.actions';
import { Product } from '../models/product.model';

export interface ProductState {
  products: Product[];
  selectedProductId: string | null;
  loading: boolean;
  error: string | null;
}

export const initialState: ProductState = {
  products: [],
  selectedProductId: null,
  loading: false,
  error: null,
};

export const productReducer = createReducer(
  initialState,
  on(ProductActions.loadProducts, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),
  on(ProductActions.loadProductsSuccess, (state, { products }) => ({
    ...state,
    products,
    loading: false,
  })),
  on(ProductActions.loadProductsFailure, (state, { error }) => ({
    ...state,
    loading: false,
    error,
  })),
  on(ProductActions.addProduct, (state, { product }) => ({
    ...state,
    products: [...state.products, product],
  })),
  on(ProductActions.removeProduct, (state, { id }) => ({
    ...state,
    products: state.products.filter(p => p.id !== id),
  })),
  on(ProductActions.selectProduct, (state, { id }) => ({
    ...state,
    selectedProductId: id,
  })),
  on(ProductActions.updateProduct, (state, { product }) => ({
    ...state,
    products: state.products.map(p => p.id === product.id ? product : p),
  }))
);
```

#### createFeature Pattern

```typescript
// store/products.feature.ts
import { createFeature } from '@ngrx/store';
import { productReducer } from './products.reducer';

export const productFeature = createFeature({
  name: 'products',
  reducer: productReducer,
});

// Auto-generated selectors:
// productFeature.selectProductsState
// productFeature.selectProducts
// productFeature.selectLoading
// productFeature.selectError
// productFeature.selectSelectedProductId
```

#### createSelector

```typescript
// store/products.selectors.ts
import { createSelector } from '@ngrx/store';
import { productFeature } from './products.feature';

export const selectFilteredProducts = (category: string) =>
  createSelector(productFeature.selectProducts, (products) =>
    products.filter(p => p.category === category)
  );

export const selectSelectedProduct = createSelector(
  productFeature.selectProducts,
  productFeature.selectSelectedProductId,
  (products, selectedId) => products.find(p => p.id === selectedId) ?? null
);

export const selectProductCount = createSelector(
  productFeature.selectProducts,
  (products) => products.length
);

export const selectTotalValue = createSelector(
  productFeature.selectProducts,
  (products) => products.reduce((sum, p) => sum + p.price, 0)
);
```

#### createEffect

```typescript
// store/products.effects.ts
import { inject, Injectable } from '@angular/core';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { ProductActions } from './products.actions';
import { ProductService } from '../services/product.service';
import { catchError, map, mergeMap, of } from 'rxjs';

@Injectable()
export class ProductEffects {
  private actions$ = inject(Actions);
  private productService = inject(ProductService);

  loadProducts$ = createEffect(() =>
    this.actions$.pipe(
      ofType(ProductActions.loadProducts),
      mergeMap(() =>
        this.productService.getAll().pipe(
          map(products => ProductActions.loadProductsSuccess({ products })),
          catchError(error =>
            of(ProductActions.loadProductsFailure({ error: error.message }))
          )
        )
      )
    )
  );
}
```

### Full NgRx Example

#### Component Integration

```typescript
// products.component.ts
import { Component, inject, OnInit } from '@angular/core';
import { Store } from '@ngrx/store';
import { ProductActions } from './store/products.actions';
import { productFeature } from './store/products.feature';
import { selectSelectedProduct, selectProductCount } from './store/products.selectors';
import { AsyncPipe } from '@angular/common';

@Component({
  selector: 'app-products',
  standalone: true,
  imports: [AsyncPipe],
  template: `
    <div>
      <h2>Products ({{ productCount$ | async }})</h2>
      <div *ngIf="loading$ | async">Loading...</div>
      <div *ngIf="error$ | async as error" class="error">{{ error }}</div>
      <ul>
        <li *ngFor="let product of products$ | async"
            (click)="selectProduct(product.id)">
          {{ product.name }} - \${{ product.price }}
        </li>
      </ul>
      <div *ngIf="selectedProduct$ | async as selected">
        <h3>Selected: {{ selected.name }}</h3>
      </div>
    </div>
  `,
})
export class ProductsComponent implements OnInit {
  private store = inject(Store);

  products$ = this.store.select(productFeature.selectProducts);
  loading$ = this.store.select(productFeature.selectLoading);
  error$ = this.store.select(productFeature.selectError);
  selectedProduct$ = this.store.select(selectSelectedProduct);
  productCount$ = this.store.select(selectProductCount);

  ngOnInit() {
    this.store.dispatch(ProductActions.loadProducts());
  }

  selectProduct(id: string) {
    this.store.dispatch(ProductActions.selectProduct({ id }));
  }
}
```

#### Module Registration

```typescript
// products.module.ts (or in app.config.ts for standalone)
import { provideStore } from '@ngrx/store';
import { provideEffects } from '@ngrx/effects';
import { provideStoreDevtools } from '@ngrx/store-devtools';
import { productFeature } from './store/products.feature';
import { ProductEffects } from './store/products.effects';

// In app.config.ts (standalone):
export const appConfig = {
  providers: [
    provideStore(),
    provideState(productFeature),
    provideEffects(ProductEffects),
    provideStoreDevtools({ maxAge: 25 }),
  ],
};
```

### Entity Management

#### EntityState and EntityAdapter

```typescript
// store/products-entity.reducer.ts
import { createEntityAdapter, EntityState } from '@ngrx/entity';
import { createReducer, on } from '@ngrx/store';
import { Product } from '../models/product.model';
import { ProductActions } from './products.actions';

export interface ProductEntityState extends EntityState<Product> {
  loading: boolean;
  error: string | null;
}

export const productAdapter = createEntityAdapter<Product>({
  selectId: (product) => product.id,
  sortComparer: (a, b) => a.name.localeCompare(b.name),
});

export const initialEntityState: ProductEntityState = productAdapter.getInitialState({
  loading: false,
  error: null,
});

export const productEntityReducer = createReducer(
  initialEntityState,
  on(ProductActions.loadProductsSuccess, (state, { products }) =>
    productAdapter.setAll(products, { ...state, loading: false })
  ),
  on(ProductActions.addProduct, (state, { product }) =>
    productAdapter.addOne(product, state)
  ),
  on(ProductActions.updateProduct, (state, { product }) =>
    productAdapter.updateOne({ id: product.id, changes: product }, state)
  ),
  on(ProductActions.removeProduct, (state, { id }) =>
    productAdapter.removeOne(id, state)
  )
);
```

#### Entity Selectors

```typescript
// store/products-entity.selectors.ts
import { createFeatureSelector, createSelector } from '@ngrx/store';
import { productAdapter, ProductEntityState } from './products-entity.reducer';

const selectProductEntityState = createFeatureSelector<ProductEntityState>('products');

const { selectAll, selectEntities, selectIds, selectTotal } =
  productAdapter.getSelectors(selectProductEntityState);

export const selectAllProducts = selectAll;
export const selectProductEntities = selectEntities;
export const selectProductIds = selectIds;
export const selectProductTotal = selectTotal;

export const selectProductById = (id: string) =>
  createSelector(selectProductEntities, (entities) => entities[id]);
```

### Effects Patterns

#### API Call Effects

```typescript
loadProducts$ = createEffect(() =>
  this.actions$.pipe(
    ofType(ProductActions.loadProducts),
    switchMap(() =>
      this.productService.getAll().pipe(
        map(products => ProductActions.loadProductsSuccess({ products })),
        catchError(error =>
          of(ProductActions.loadProductsFailure({ error: error.message }))
        )
      )
    )
  )
);
```

#### Navigation Effects

```typescript
import { Router } from '@angular/router';

navigateToProduct$ = createEffect(() =>
  this.actions$.pipe(
    ofType(ProductActions.selectProduct),
    tap(({ id }) => this.router.navigate(['/products', id]))
  ),
  { dispatch: false }
);
```

#### Error Handling in Effects

```typescript
import { MatSnackBar } from '@angular/material/snack-bar';

showError$ = createEffect(() =>
  this.actions$.pipe(
    ofType(ProductActions.loadProductsFailure),
    tap(({ error }) => {
      this.snackBar.open(error, 'Dismiss', { duration: 5000 });
    })
  ),
  { dispatch: false }
);
```

#### Conditional Effects

```typescript
loadProductsIfNeeded$ = createEffect(() =>
  this.actions$.pipe(
    ofType(ProductActions.loadProducts),
    withLatestFrom(this.store.select(productFeature.selectProducts)),
    filter(([_action, products]) => products.length === 0),
    mergeMap(() =>
      this.productService.getAll().pipe(
        map(products => ProductActions.loadProductsSuccess({ products })),
        catchError(error =>
          of(ProductActions.loadProductsFailure({ error: error.message }))
        )
      )
    )
  )
);
```

### NgRx DevTools

#### Store DevTools Setup

```typescript
// app.config.ts
import { provideStoreDevtools } from '@ngrx/store-devtools';
import { isDevMode } from '@angular/core';

export const appConfig = {
  providers: [
    provideStore(),
    provideStoreDevtools({
      maxAge: 25,             // Retains last 25 states
      logOnly: !isDevMode(),  // Restrict in production
      autoPause: true,        // Pause when extension window is not open
      connectInZone: true,    // Connect inside Angular zone for change detection
    }),
  ],
};
```

#### Time-Travel Debugging

- Install the Redux DevTools browser extension.
- Open DevTools, select the "Redux" tab.
- Use the slider or click on actions to travel between states.
- Use "Jump" to set the store to a previous state.
- Use "Skip" to remove an action from the sequence.

#### Action Log Filtering

In Redux DevTools, filter actions by typing in the filter box. Useful patterns:
- `[Products]` shows only product-related actions.
- `Success` shows only success actions.
- `Failure` shows only failure actions.

---

## NgRx Component Store (Angular)

### When to Use Component Store vs Store

| Criteria | NgRx Store | Component Store |
|---|---|---|
| Scope | Application-wide | Component-scoped |
| Boilerplate | Higher (actions, reducers, effects) | Lower (single class) |
| DevTools | Full Redux DevTools | Limited |
| Best for | Complex, shared state | Local, encapsulated state |
| Lifecycle | Singleton (app-wide) | Tied to component lifecycle |

### Full Example: Component with Local State Management

```typescript
// product-list.store.ts
import { Injectable } from '@angular/core';
import { ComponentStore } from '@ngrx/component-store';
import { tapResponse } from '@ngrx/operators';
import { switchMap, tap } from 'rxjs';
import { inject } from '@angular/core';
import { ProductService } from '../services/product.service';

export interface ProductListState {
  products: Product[];
  loading: boolean;
  error: string | null;
  searchTerm: string;
  page: number;
}

const initialState: ProductListState = {
  products: [],
  loading: false,
  error: null,
  searchTerm: '',
  page: 1,
};

@Injectable()
export class ProductListStore extends ComponentStore<ProductListState> {
  private productService = inject(ProductService);

  constructor() {
    super(initialState);
  }

  // Selectors
  readonly products$ = this.select(state => state.products);
  readonly loading$ = this.select(state => state.loading);
  readonly error$ = this.select(state => state.error);
  readonly searchTerm$ = this.select(state => state.searchTerm);

  readonly filteredProducts$ = this.select(
    this.products$,
    this.searchTerm$,
    (products, term) =>
      term
        ? products.filter(p => p.name.toLowerCase().includes(term.toLowerCase()))
        : products
  );

  readonly vm$ = this.select(
    this.filteredProducts$,
    this.loading$,
    this.error$,
    (products, loading, error) => ({ products, loading, error })
  );

  // Updaters (synchronous state changes)
  readonly setSearchTerm = this.updater<string>((state, searchTerm) => ({
    ...state,
    searchTerm,
  }));

  readonly setPage = this.updater<number>((state, page) => ({
    ...state,
    page,
  }));

  readonly addProduct = this.updater<Product>((state, product) => ({
    ...state,
    products: [...state.products, product],
  }));

  // Effects (async operations)
  readonly loadProducts = this.effect<void>(trigger$ =>
    trigger$.pipe(
      tap(() => this.patchState({ loading: true, error: null })),
      switchMap(() =>
        this.productService.getAll().pipe(
          tapResponse(
            (products) => this.patchState({ products, loading: false }),
            (error: Error) => this.patchState({ error: error.message, loading: false })
          )
        )
      )
    )
  );
}
```

```typescript
// product-list.component.ts
@Component({
  selector: 'app-product-list',
  standalone: true,
  providers: [ProductListStore],
  template: `
    <ng-container *ngIf="store.vm$ | async as vm">
      <input (input)="store.setSearchTerm($event.target.value)" placeholder="Search..." />
      <div *ngIf="vm.loading">Loading...</div>
      <div *ngIf="vm.error" class="error">{{ vm.error }}</div>
      <ul>
        <li *ngFor="let product of vm.products">{{ product.name }}</li>
      </ul>
    </ng-container>
  `,
})
export class ProductListComponent implements OnInit {
  store = inject(ProductListStore);

  ngOnInit() {
    this.store.loadProducts();
  }
}
```

---

## NgRx SignalStore (Angular)

### withState, withComputed, withMethods

```typescript
// product.store.ts
import {
  signalStore,
  withState,
  withComputed,
  withMethods,
  patchState,
} from '@ngrx/signals';
import { computed, inject } from '@angular/core';
import { ProductService } from '../services/product.service';

type ProductState = {
  products: Product[];
  loading: boolean;
  error: string | null;
  filter: string;
};

const initialState: ProductState = {
  products: [],
  loading: false,
  error: null,
  filter: '',
};

export const ProductStore = signalStore(
  { providedIn: 'root' },
  withState(initialState),
  withComputed(({ products, filter }) => ({
    filteredProducts: computed(() => {
      const term = filter().toLowerCase();
      return term
        ? products().filter(p => p.name.toLowerCase().includes(term))
        : products();
    }),
    productCount: computed(() => products().length),
    totalValue: computed(() =>
      products().reduce((sum, p) => sum + p.price, 0)
    ),
  })),
  withMethods((store, productService = inject(ProductService)) => ({
    setFilter(filter: string) {
      patchState(store, { filter });
    },
    async loadProducts() {
      patchState(store, { loading: true, error: null });
      try {
        const products = await productService.getAll();
        patchState(store, { products, loading: false });
      } catch (e) {
        patchState(store, {
          error: e instanceof Error ? e.message : 'Unknown error',
          loading: false,
        });
      }
    },
    addProduct(product: Product) {
      patchState(store, { products: [...store.products(), product] });
    },
    removeProduct(id: string) {
      patchState(store, {
        products: store.products().filter(p => p.id !== id),
      });
    },
  }))
);
```

### withEntities for Collection Management

```typescript
import { signalStore, withMethods } from '@ngrx/signals';
import {
  withEntities,
  addEntity,
  updateEntity,
  removeEntity,
  setAllEntities,
} from '@ngrx/signals/entities';

type Product = { id: string; name: string; price: number };

export const ProductEntityStore = signalStore(
  { providedIn: 'root' },
  withEntities<Product>(),
  withMethods((store) => ({
    setProducts(products: Product[]) {
      setAllEntities(products)(store);
    },
    addProduct(product: Product) {
      addEntity(product)(store);
    },
    updateProduct(id: string, changes: Partial<Product>) {
      updateEntity({ id, changes })(store);
    },
    removeProduct(id: string) {
      removeEntity(id)(store);
    },
  }))
);
```

### Custom Store Features with signalStoreFeature

```typescript
import { signalStoreFeature, withState, withComputed, withMethods, patchState } from '@ngrx/signals';
import { computed } from '@angular/core';

// Reusable loading feature
export function withLoadingState() {
  return signalStoreFeature(
    withState({ loading: false, error: null as string | null }),
    withComputed(({ loading, error }) => ({
      hasError: computed(() => error() !== null),
      isReady: computed(() => !loading() && error() === null),
    })),
    withMethods((store) => ({
      setLoading() {
        patchState(store, { loading: true, error: null });
      },
      setLoaded() {
        patchState(store, { loading: false });
      },
      setError(error: string) {
        patchState(store, { loading: false, error });
      },
    }))
  );
}

// Usage
export const OrderStore = signalStore(
  { providedIn: 'root' },
  withLoadingState(),
  withState({ orders: [] as Order[] }),
  withMethods((store, orderService = inject(OrderService)) => ({
    async loadOrders() {
      store.setLoading();
      try {
        const orders = await orderService.getAll();
        patchState(store, { orders });
        store.setLoaded();
      } catch (e) {
        store.setError(e instanceof Error ? e.message : 'Failed to load orders');
      }
    },
  }))
);
```

### Full Example: Feature Store with CRUD Operations

```typescript
// task.store.ts
import {
  signalStore,
  withState,
  withComputed,
  withMethods,
  withHooks,
  patchState,
} from '@ngrx/signals';
import { withEntities, addEntity, updateEntity, removeEntity, setAllEntities } from '@ngrx/signals/entities';
import { computed, inject } from '@angular/core';
import { TaskService } from '../services/task.service';

type Task = {
  id: string;
  title: string;
  completed: boolean;
  priority: 'low' | 'medium' | 'high';
};

export const TaskStore = signalStore(
  { providedIn: 'root' },
  withEntities<Task>(),
  withState({ loading: false, error: null as string | null, filterPriority: null as string | null }),
  withComputed(({ entities, filterPriority }) => ({
    completedTasks: computed(() => entities().filter(t => t.completed)),
    pendingTasks: computed(() => entities().filter(t => !t.completed)),
    filteredTasks: computed(() => {
      const priority = filterPriority();
      return priority ? entities().filter(t => t.priority === priority) : entities();
    }),
    taskStats: computed(() => ({
      total: entities().length,
      completed: entities().filter(t => t.completed).length,
      pending: entities().filter(t => !t.completed).length,
    })),
  })),
  withMethods((store, taskService = inject(TaskService)) => ({
    async loadTasks() {
      patchState(store, { loading: true, error: null });
      try {
        const tasks = await taskService.getAll();
        setAllEntities(tasks)(store);
        patchState(store, { loading: false });
      } catch (e) {
        patchState(store, { loading: false, error: (e as Error).message });
      }
    },
    async createTask(title: string, priority: Task['priority']) {
      const task = await taskService.create({ title, priority, completed: false });
      addEntity(task)(store);
    },
    async toggleTask(id: string) {
      const task = store.entityMap()[id];
      if (task) {
        updateEntity({ id, changes: { completed: !task.completed } })(store);
        await taskService.update(id, { completed: !task.completed });
      }
    },
    async deleteTask(id: string) {
      removeEntity(id)(store);
      await taskService.delete(id);
    },
    setFilterPriority(priority: string | null) {
      patchState(store, { filterPriority: priority });
    },
  })),
  withHooks({
    onInit(store) {
      store.loadTasks();
    },
  })
);
```

```typescript
// task-list.component.ts
@Component({
  selector: 'app-task-list',
  standalone: true,
  template: `
    <h2>Tasks ({{ taskStore.taskStats().total }})</h2>
    <p>Completed: {{ taskStore.taskStats().completed }} | Pending: {{ taskStore.taskStats().pending }}</p>

    <select (change)="taskStore.setFilterPriority($event.target.value || null)">
      <option value="">All</option>
      <option value="low">Low</option>
      <option value="medium">Medium</option>
      <option value="high">High</option>
    </select>

    @if (taskStore.loading()) {
      <p>Loading tasks...</p>
    }

    @for (task of taskStore.filteredTasks(); track task.id) {
      <div class="task" [class.completed]="task.completed">
        <input type="checkbox" [checked]="task.completed" (change)="taskStore.toggleTask(task.id)" />
        <span>{{ task.title }} [{{ task.priority }}]</span>
        <button (click)="taskStore.deleteTask(task.id)">Delete</button>
      </div>
    }
  `,
})
export class TaskListComponent {
  taskStore = inject(TaskStore);
}
```

---

## Comparison Tables

### Pinia vs NgRx vs Vuex vs Component Store vs SignalStore

| Feature | Pinia | NgRx Store | NgRx Component Store | NgRx SignalStore | Vuex |
|---|---|---|---|---|---|
| Setup complexity | Low | High | Medium | Medium | Medium |
| Bundle size | ~1.5 KB | ~15 KB (full) | ~3 KB | ~4 KB | ~5 KB |
| DevTools | Vue DevTools | Redux DevTools | Limited | Limited | Vue DevTools |
| TypeScript | Excellent | Excellent | Excellent | Excellent | Fair |
| SSR support | Yes (Nuxt) | Yes (Angular Universal) | Manual | Manual | Yes (Nuxt 2) |
| Learning curve | Low | High | Medium | Medium | Low-Medium |
| Best for | Vue 3 apps | Large Angular apps | Component-scoped state | Modern Angular apps | Legacy Vue 2 apps |

### Reactive Patterns Comparison

| Pattern | Vue (Pinia) | Angular (NgRx Store) | Angular (NgRx SignalStore) |
|---|---|---|---|
| State definition | `ref()` / `reactive()` | `createReducer()` + `on()` | `withState()` |
| Computed / derived | `computed()` | `createSelector()` | `withComputed()` + `computed()` |
| Async operations | `async` functions in store | `createEffect()` with RxJS | `async` methods via `withMethods()` |
| Side effects | Direct in actions | Dedicated Effects class | Direct in methods or `rxMethod()` |
| Subscriptions | `$subscribe`, `$onAction` | `store.select().subscribe()` | Signals are auto-tracked |
| State mutation | Direct assignment | Immutable via reducers | `patchState()` |
| Composition | Import other stores | Inject Store, select state | `signalStoreFeature()` |

---

## Reactive Patterns

### Vue Reactivity (ref, reactive, computed)

```typescript
import { ref, reactive, computed, watch, watchEffect } from 'vue';

// Primitive reactive value
const count = ref(0);
count.value++; // Access via .value in script

// Object reactive value
const user = reactive({
  name: 'Alice',
  age: 30,
});
user.name = 'Bob'; // Direct mutation, no .value needed

// Derived / computed values
const doubleCount = computed(() => count.value * 2);
const isAdult = computed(() => user.age >= 18);

// Watchers
watch(count, (newVal, oldVal) => {
  console.log(`count changed from ${oldVal} to ${newVal}`);
});

watchEffect(() => {
  // Runs immediately, re-runs when dependencies change
  console.log(`User is ${user.name}, count is ${count.value}`);
});
```

### Angular Signals (signal, computed, effect)

```typescript
import { signal, computed, effect } from '@angular/core';

// Primitive signal
const count = signal(0);
count.set(1);        // Set new value
count.update(v => v + 1); // Update based on current

// Object signal
const user = signal({ name: 'Alice', age: 30 });
user.update(u => ({ ...u, name: 'Bob' })); // Immutable update

// Derived / computed signals
const doubleCount = computed(() => count() * 2);
const isAdult = computed(() => user().age >= 18);

// Effects (side effects when signals change)
effect(() => {
  console.log(`User is ${user().name}, count is ${count()}`);
});
```

### RxJS Observables vs Signals

| Aspect | RxJS Observables | Angular Signals |
|---|---|---|
| Push/Pull | Push-based | Pull-based (lazy) |
| Subscription | Manual subscribe/unsubscribe | Auto-tracked in templates |
| Async | Built-in (async pipe, operators) | Wrap with `toSignal()` |
| Operators | Hundreds of operators | None (use computed) |
| Glitch-free | Not guaranteed | Yes |
| Memory leaks | Possible if not unsubscribed | No |
| Best for | Streams, events, HTTP | Synchronous derived state, UI |

### When to Use Each Approach

- **Signals**: Synchronous state, UI bindings, derived/computed values, component state.
- **RxJS**: HTTP requests, WebSocket streams, complex async pipelines, event debouncing, combining multiple async sources.
- **Both together**: Use RxJS for async data fetching, convert to signals for template consumption.

### Interop Patterns

```typescript
import { toSignal, toObservable } from '@angular/core/rxjs-interop';
import { signal } from '@angular/core';
import { interval } from 'rxjs';

// Observable to Signal
const counter$ = interval(1000);
const counterSignal = toSignal(counter$, { initialValue: 0 });
// Use in template: {{ counterSignal() }}

// Signal to Observable
const searchTerm = signal('');
const searchTerm$ = toObservable(searchTerm);
// Use with RxJS operators:
searchTerm$.pipe(
  debounceTime(300),
  distinctUntilChanged(),
  switchMap(term => this.searchService.search(term))
);
```

---

## DevTools Integration

### Vue DevTools for Pinia

- Install Vue DevTools browser extension (v6+ for Vue 3).
- Pinia stores appear automatically under the "Pinia" tab.
- Inspect current state, edit values live, and track mutations.
- Timeline view shows action calls and state changes.

```typescript
// Enable detailed timeline events (optional)
const pinia = createPinia();
// DevTools integration is automatic; no extra config needed.
```

### Redux DevTools for NgRx

```typescript
// app.config.ts
import { provideStoreDevtools } from '@ngrx/store-devtools';
import { isDevMode } from '@angular/core';

export const appConfig = {
  providers: [
    provideStore(),
    provideStoreDevtools({
      maxAge: 50,
      logOnly: !isDevMode(),
      autoPause: true,
      trace: true,         // Include stack traces
      traceLimit: 25,
    }),
  ],
};
```

### Custom DevTools Plugins (Vue)

```typescript
// plugins/devtools.ts
import { setupDevtoolsPlugin } from '@vue/devtools-api';

export function registerCustomDevtools(app: App) {
  setupDevtoolsPlugin(
    {
      id: 'my-custom-plugin',
      label: 'My Custom Plugin',
      packageName: 'my-custom-plugin',
      app,
    },
    (api) => {
      api.addTimelineLayer({
        id: 'my-events',
        label: 'My Events',
        color: 0x7fff00, // chartreuse
      });

      api.addTimelineEvent({
        layerId: 'my-events',
        event: {
          time: Date.now(),
          data: { message: 'Plugin initialized' },
        },
      });
    }
  );
}
```

### Debugging State Issues

1. **Unexpected state**: Check the action log to see which action caused the change.
2. **Missing updates**: Verify selectors are using the correct state slice. In Vue, check that `ref` values are accessed via `.value` in script.
3. **Stale state in effects**: Use `withLatestFrom` (NgRx) or `computed` (Vue/Angular) instead of closures.
4. **Hydration mismatch**: Ensure SSR state is serializable. Avoid `Date`, `Map`, `Set` in store state that gets serialized.

### Time-Travel Debugging

- **Vue DevTools**: Use the timeline view to step through state changes. Click on any mutation to see state before/after.
- **Redux DevTools (NgRx)**: Use the slider to move between dispatched actions. "Jump" sets the store to that state; "Skip" removes an action from history.

---

## Advanced Patterns

### Optimistic Updates

```typescript
// Vue (Pinia) - Optimistic delete with rollback
export const useTodoStore = defineStore('todos', () => {
  const todos = ref<Todo[]>([]);

  async function deleteTodo(id: string) {
    const index = todos.value.findIndex(t => t.id === id);
    if (index === -1) return;

    const removed = todos.value[index];
    todos.value.splice(index, 1); // Optimistic removal

    try {
      await todoApi.delete(id);
    } catch (e) {
      todos.value.splice(index, 0, removed); // Rollback on failure
      throw e;
    }
  }

  return { todos, deleteTodo };
});
```

```typescript
// Angular (NgRx) - Optimistic update with rollback effect
export const TodoActions = createActionGroup({
  source: 'Todos',
  events: {
    'Delete Todo': props<{ id: string }>(),
    'Delete Todo Success': props<{ id: string }>(),
    'Delete Todo Failure': props<{ id: string; todo: Todo }>(),
  },
});

// Reducer applies optimistic removal immediately
on(TodoActions.deleteTodo, (state, { id }) => ({
  ...state,
  todos: state.todos.filter(t => t.id !== id),
})),
// Rollback on failure
on(TodoActions.deleteTodoFailure, (state, { todo }) => ({
  ...state,
  todos: [...state.todos, todo],
})),

// Effect handles the API call
deleteTodo$ = createEffect(() =>
  this.actions$.pipe(
    ofType(TodoActions.deleteTodo),
    withLatestFrom(this.store.select(selectAllTodos)),
    mergeMap(([{ id }, todos]) => {
      const todo = todos.find(t => t.id === id);
      return this.todoService.delete(id).pipe(
        map(() => TodoActions.deleteTodoSuccess({ id })),
        catchError(() => of(TodoActions.deleteTodoFailure({ id, todo: todo! })))
      );
    })
  )
);
```

### Undo/Redo Pattern

```typescript
// Vue (Pinia) - Generic undo/redo
export const useUndoStore = defineStore('undo', () => {
  const past = ref<string[]>([]);
  const future = ref<string[]>([]);
  const present = ref<string>('');

  function pushState(state: string) {
    past.value.push(present.value);
    present.value = state;
    future.value = [];
  }

  function undo() {
    if (past.value.length === 0) return;
    future.value.push(present.value);
    present.value = past.value.pop()!;
  }

  function redo() {
    if (future.value.length === 0) return;
    past.value.push(present.value);
    present.value = future.value.pop()!;
  }

  const canUndo = computed(() => past.value.length > 0);
  const canRedo = computed(() => future.value.length > 0);

  return { present, pushState, undo, redo, canUndo, canRedo };
});
```

### State Persistence (localStorage, sessionStorage)

```typescript
// Vue (Pinia) - Selective persistence
import { watch } from 'vue';

export const useSettingsStore = defineStore('settings', () => {
  const theme = ref<'light' | 'dark'>('light');
  const locale = ref('en');
  const sidebarCollapsed = ref(false);

  // Restore from localStorage on init
  const stored = localStorage.getItem('app-settings');
  if (stored) {
    const parsed = JSON.parse(stored);
    theme.value = parsed.theme ?? 'light';
    locale.value = parsed.locale ?? 'en';
    sidebarCollapsed.value = parsed.sidebarCollapsed ?? false;
  }

  // Persist on changes
  watch(
    [theme, locale, sidebarCollapsed],
    ([t, l, s]) => {
      localStorage.setItem(
        'app-settings',
        JSON.stringify({ theme: t, locale: l, sidebarCollapsed: s })
      );
    },
    { deep: true }
  );

  return { theme, locale, sidebarCollapsed };
});
```

```typescript
// Angular (SignalStore) - Persistence feature
import { signalStoreFeature, withHooks, withMethods, getState } from '@ngrx/signals';

export function withPersistence(key: string) {
  return signalStoreFeature(
    withMethods((store) => ({
      _saveState() {
        const state = getState(store);
        localStorage.setItem(key, JSON.stringify(state));
      },
      _loadState() {
        const stored = localStorage.getItem(key);
        if (stored) {
          patchState(store, JSON.parse(stored));
        }
      },
    })),
    withHooks({
      onInit(store: any) {
        store._loadState();
      },
    })
  );
}
```

### State Hydration for SSR

```typescript
// Vue (Nuxt) - Pinia SSR hydration is automatic
// The Pinia Nuxt module serializes state during SSR and rehydrates on the client.
// Key rules for SSR state:
// 1. Only store serializable values (no functions, Dates, Maps)
// 2. Use useAsyncData or useFetch in stores for data fetching
// 3. Avoid side effects in store constructors

// Angular Universal - NgRx hydration
// app.config.server.ts
import { provideStoreHydration } from '@ngrx/store';

export const serverConfig = {
  providers: [
    provideStoreHydration(), // Transfers state from server to client
  ],
};
```

### State Normalization

```typescript
// Normalized state shape (both frameworks)
interface NormalizedState {
  entities: Record<string, Product>;
  ids: string[];
}

// Instead of nested data:
// BAD: { orders: [{ id: '1', items: [{ product: { id: 'a', ... } }] }] }

// GOOD: Normalized
// {
//   orders: { entities: { '1': { id: '1', itemIds: ['i1'] } }, ids: ['1'] },
//   orderItems: { entities: { 'i1': { id: 'i1', productId: 'a', qty: 2 } }, ids: ['i1'] },
//   products: { entities: { 'a': { id: 'a', name: 'Widget' } }, ids: ['a'] },
// }

// Angular: Use @ngrx/entity (EntityAdapter) for automatic normalization
// Vue: Normalize manually or use a helper

function normalizeArray<T extends { id: string }>(items: T[]): NormalizedState {
  const entities: Record<string, T> = {};
  const ids: string[] = [];
  for (const item of items) {
    entities[item.id] = item;
    ids.push(item.id);
  }
  return { entities, ids } as unknown as NormalizedState;
}
```

### Pagination State

```typescript
// Vue (Pinia) - Pagination store
export const usePaginatedStore = defineStore('paginated', () => {
  const items = ref<Item[]>([]);
  const currentPage = ref(1);
  const pageSize = ref(20);
  const totalItems = ref(0);
  const loading = ref(false);

  const totalPages = computed(() => Math.ceil(totalItems.value / pageSize.value));
  const hasNextPage = computed(() => currentPage.value < totalPages.value);
  const hasPrevPage = computed(() => currentPage.value > 1);

  async function fetchPage(page: number) {
    loading.value = true;
    currentPage.value = page;
    try {
      const response = await api.getItems({ page, limit: pageSize.value });
      items.value = response.data;
      totalItems.value = response.total;
    } finally {
      loading.value = false;
    }
  }

  async function nextPage() {
    if (hasNextPage.value) await fetchPage(currentPage.value + 1);
  }

  async function prevPage() {
    if (hasPrevPage.value) await fetchPage(currentPage.value - 1);
  }

  return {
    items, currentPage, pageSize, totalItems, totalPages,
    hasNextPage, hasPrevPage, fetchPage, nextPage, prevPage, loading,
  };
});
```

```typescript
// Angular (SignalStore) - Pagination feature
export function withPagination<T>() {
  return signalStoreFeature(
    withState({
      currentPage: 1,
      pageSize: 20,
      totalItems: 0,
    }),
    withComputed(({ currentPage, pageSize, totalItems }) => ({
      totalPages: computed(() => Math.ceil(totalItems() / pageSize())),
      hasNextPage: computed(() => currentPage() < Math.ceil(totalItems() / pageSize())),
      hasPrevPage: computed(() => currentPage() > 1),
    })),
    withMethods((store) => ({
      setPage(page: number) {
        patchState(store, { currentPage: page });
      },
      setTotalItems(total: number) {
        patchState(store, { totalItems: total });
      },
      nextPage() {
        if (store.hasNextPage()) {
          patchState(store, { currentPage: store.currentPage() + 1 });
        }
      },
      prevPage() {
        if (store.hasPrevPage()) {
          patchState(store, { currentPage: store.currentPage() - 1 });
        }
      },
    }))
  );
}
```

### WebSocket State Synchronization

```typescript
// Vue (Pinia) - WebSocket sync
export const useRealtimeStore = defineStore('realtime', () => {
  const messages = ref<Message[]>([]);
  const connected = ref(false);
  let ws: WebSocket | null = null;

  function connect(url: string) {
    ws = new WebSocket(url);

    ws.onopen = () => {
      connected.value = true;
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      switch (data.type) {
        case 'MESSAGE_ADDED':
          messages.value.push(data.payload);
          break;
        case 'MESSAGE_DELETED':
          messages.value = messages.value.filter(m => m.id !== data.payload.id);
          break;
        case 'STATE_SYNC':
          messages.value = data.payload.messages;
          break;
      }
    };

    ws.onclose = () => {
      connected.value = false;
      setTimeout(() => connect(url), 3000); // Auto-reconnect
    };
  }

  function sendMessage(content: string) {
    ws?.send(JSON.stringify({ type: 'SEND_MESSAGE', payload: { content } }));
  }

  function disconnect() {
    ws?.close();
    ws = null;
  }

  return { messages, connected, connect, sendMessage, disconnect };
});
```

```typescript
// Angular (NgRx) - WebSocket effect
@Injectable()
export class WebSocketEffects {
  private actions$ = inject(Actions);
  private wsService = inject(WebSocketService);

  connect$ = createEffect(() =>
    this.actions$.pipe(
      ofType(WsActions.connect),
      switchMap(({ url }) =>
        this.wsService.connect(url).pipe(
          map(message => {
            switch (message.type) {
              case 'MESSAGE_ADDED':
                return WsActions.messageReceived({ message: message.payload });
              case 'MESSAGE_DELETED':
                return WsActions.messageDeleted({ id: message.payload.id });
              default:
                return WsActions.unknownMessage({ message });
            }
          }),
          catchError(error => of(WsActions.connectionError({ error: error.message })))
        )
      )
    )
  );
}
```

---

## Anti-Patterns

| Anti-Pattern | Problem | Solution |
|---|---|---|
| Over-using global state | Everything in one store; hard to reason about, poor performance | Only put truly shared state in stores. Use local state for component-specific data. |
| Storing derived data | Duplicating data that can be computed from other state (e.g., storing `filteredList` alongside `list` and `filter`) | Use `computed()` (Vue) or `createSelector()` / `computed()` signals (Angular) to derive values on the fly. |
| Not normalizing nested state | Deeply nested objects lead to complex update logic and potential bugs from missed references | Normalize state: store entities by ID in flat maps. Use `@ngrx/entity` or manual normalization. |
| Mutating state directly (NgRx) | Bypasses the reducer, breaks time-travel debugging and DevTools | Always return new state objects from reducers. Use spread operators or `immer` for complex updates. |
| Putting UI state in global store | Storing modal visibility, form field focus, scroll position globally | Keep UI state local to the component that owns it. Only lift to a store if multiple components depend on it. |
| Dispatching actions in reducers | Creates side effects in pure functions, unpredictable state transitions | Use NgRx Effects for side effects. Reducers must be pure functions. |
| Not unsubscribing (Angular) | Memory leaks from RxJS subscriptions | Use `async` pipe, `takeUntilDestroyed()`, or `DestroyRef` for cleanup. Signals avoid this entirely. |
| Using store for server cache | Manually managing loading, caching, invalidation, and refetching | Use dedicated tools like TanStack Query, Apollo Client, or SWR for server state. Reserve stores for client state. |
| Circular store dependencies | Store A imports Store B which imports Store A; causes runtime errors | Restructure into a third shared store, or pass data via actions/events instead of direct imports. |
| Giant monolithic store | One store holding all application state; poor code organization | Split into feature stores. Each domain gets its own store (auth, products, cart, etc.). |
| Synchronous API calls in reducers | Blocks the reducer, makes state updates unpredictable | Perform all async work in effects (NgRx) or actions (Pinia). Reducers/state updates are synchronous. |
| Storing non-serializable values | Functions, class instances, Symbols in state break DevTools, SSR hydration, and persistence | Only store plain objects, arrays, strings, numbers, booleans, and null. Transform complex types before storing. |
