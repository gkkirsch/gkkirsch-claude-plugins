# Angular 17+ Patterns Quick Reference

> Claude Code plugin reference for modern Angular development patterns.
> Covers Signals, Standalone Components, DI, RxJS, Material, Forms, Router, HTTP, Control Flow, SSR, and Performance.

---

## Signals Quick Reference

### Core Signal Functions

| Function | Import | Description | Returns |
|---|---|---|---|
| `signal(initialValue)` | `@angular/core` | Creates a writable signal | `WritableSignal<T>` |
| `computed(fn)` | `@angular/core` | Derives a read-only signal from other signals | `Signal<T>` |
| `effect(fn)` | `@angular/core` | Runs side effects when dependencies change | `EffectRef` |
| `untracked(fn)` | `@angular/core` | Reads signals without tracking dependencies | `T` |

### Writable Signal API

| Method | Description |
|---|---|
| `signal.set(value)` | Replace the current value |
| `signal.update(fn)` | Derive new value from old: `count.update(c => c + 1)` |
| `signal()` | Read the current value (getter) |

```typescript
import { signal, computed, effect, untracked } from '@angular/core';

// Writable signal
const count = signal(0);
count.set(5);
count.update(c => c + 1);
console.log(count()); // 6

// Computed signal (read-only, auto-tracks dependencies)
const doubled = computed(() => count() * 2);
console.log(doubled()); // 12

// Effect (runs when tracked signals change)
effect(() => {
  console.log(`Count is now: ${count()}`);
});

// Untracked read (does not create dependency)
effect(() => {
  const tracked = count();
  const ignored = untracked(() => someOtherSignal());
});
```

### Signal Inputs

| Function | Description | Required |
|---|---|---|
| `input<T>()` | Optional signal input | No |
| `input<T>(default)` | Signal input with default | No |
| `input.required<T>()` | Required signal input | Yes |

```typescript
import { Component, input } from '@angular/core';

@Component({
  selector: 'app-user-card',
  template: `
    <div class="card">
      <h2>{{ name() }}</h2>
      <p>Age: {{ age() }}</p>
      <p *ngIf="email()">{{ email() }}</p>
    </div>
  `,
})
export class UserCardComponent {
  name = input.required<string>();
  age = input<number>(0);
  email = input<string | undefined>();
}
```

### Signal Outputs

```typescript
import { Component, output } from '@angular/core';

@Component({
  selector: 'app-search-box',
  template: `
    <input #searchInput (keyup.enter)="onSearch(searchInput.value)" />
    <button (click)="onSearch(searchInput.value)">Search</button>
  `,
})
export class SearchBoxComponent {
  searched = output<string>();
  cleared = output<void>();

  onSearch(term: string) {
    this.searched.emit(term);
  }

  onClear() {
    this.cleared.emit();
  }
}
```

Usage in parent:
```html
<app-search-box (searched)="handleSearch($event)" (cleared)="handleClear()" />
```

### Signal Queries

| Function | Description | Returns |
|---|---|---|
| `viewChild(ref)` | Single child in view | `Signal<T \| undefined>` |
| `viewChild.required(ref)` | Required single child in view | `Signal<T>` |
| `viewChildren(ref)` | All matching children in view | `Signal<readonly T[]>` |
| `contentChild(ref)` | Single projected child | `Signal<T \| undefined>` |
| `contentChildren(ref)` | All matching projected children | `Signal<readonly T[]>` |

```typescript
import { Component, viewChild, viewChildren, contentChild, contentChildren, ElementRef } from '@angular/core';

@Component({
  selector: 'app-dashboard',
  template: `
    <input #searchInput />
    <app-widget *ngFor="let w of widgets" />
  `,
})
export class DashboardComponent {
  searchInput = viewChild.required<ElementRef>('searchInput');
  widgets = viewChildren(WidgetComponent);

  focusSearch() {
    this.searchInput().nativeElement.focus();
  }
}

@Component({
  selector: 'app-tab-group',
  template: `<ng-content />`,
})
export class TabGroupComponent {
  activeTab = contentChild(TabComponent);
  allTabs = contentChildren(TabComponent);
}
```

### Model Inputs (Two-Way Binding)

```typescript
import { Component, model } from '@angular/core';

@Component({
  selector: 'app-rating',
  template: `
    @for (star of stars(); track star) {
      <button (click)="value.set(star)" [class.active]="star <= value()">
        ★
      </button>
    }
  `,
})
export class RatingComponent {
  value = model(0);
  max = input(5);
  stars = computed(() => Array.from({ length: this.max() }, (_, i) => i + 1));
}
```

Usage with two-way binding:
```html
<app-rating [(value)]="userRating" [max]="10" />
```

### Signal Interop with RxJS

| Function | Import | Description |
|---|---|---|
| `toSignal(obs$)` | `@angular/core/rxjs-interop` | Converts Observable to Signal |
| `toObservable(sig)` | `@angular/core/rxjs-interop` | Converts Signal to Observable |

```typescript
import { Component, signal, inject } from '@angular/core';
import { toSignal, toObservable } from '@angular/core/rxjs-interop';
import { HttpClient } from '@angular/common/http';
import { switchMap } from 'rxjs/operators';

@Component({ /* ... */ })
export class ProductListComponent {
  private http = inject(HttpClient);

  // Observable -> Signal
  products = toSignal(this.http.get<Product[]>('/api/products'), {
    initialValue: [],
  });

  // Signal -> Observable (useful for triggering HTTP calls)
  searchTerm = signal('');
  private searchTerm$ = toObservable(this.searchTerm);

  results = toSignal(
    this.searchTerm$.pipe(
      switchMap(term => this.http.get<Product[]>(`/api/search?q=${term}`))
    ),
    { initialValue: [] }
  );
}
```

### Signal Equality Functions

```typescript
import { signal } from '@angular/core';

// Default: Object.is
const count = signal(0);

// Custom equality function
const user = signal(
  { id: 1, name: 'Alice' },
  { equal: (a, b) => a.id === b.id }
);

// Setting to same id won't trigger updates
user.set({ id: 1, name: 'Alice Updated' }); // No notification
user.set({ id: 2, name: 'Bob' }); // Triggers notification
```

---

## Standalone Components

### Component (Standalone by Default in Angular 17+)

```typescript
import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-header',
  imports: [CommonModule, RouterLink],
  template: `
    <nav>
      <a routerLink="/">Home</a>
      <a routerLink="/about">About</a>
    </nav>
  `,
  styles: [`
    nav { display: flex; gap: 1rem; }
  `],
})
export class HeaderComponent {}
```

### Standalone Directives

```typescript
import { Directive, ElementRef, inject, input, effect } from '@angular/core';

@Directive({
  selector: '[appHighlight]',
})
export class HighlightDirective {
  private el = inject(ElementRef);
  color = input('yellow', { alias: 'appHighlight' });

  constructor() {
    effect(() => {
      this.el.nativeElement.style.backgroundColor = this.color();
    });
  }
}
```

### Standalone Pipes

```typescript
import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'truncate',
})
export class TruncatePipe implements PipeTransform {
  transform(value: string, limit = 50, trail = '...'): string {
    return value.length > limit ? value.substring(0, limit) + trail : value;
  }
}
```

### Importing in Standalone Components

```typescript
import { Component } from '@angular/core';
import { HeaderComponent } from './header.component';
import { HighlightDirective } from './highlight.directive';
import { TruncatePipe } from './truncate.pipe';

@Component({
  selector: 'app-page',
  imports: [HeaderComponent, HighlightDirective, TruncatePipe],
  template: `
    <app-header />
    <p [appHighlight]="'lightblue'">{{ longText | truncate:80 }}</p>
  `,
})
export class PageComponent {
  longText = 'This is a very long text that should be truncated...';
}
```

### Provider Patterns

| Provider Function | Package | Purpose |
|---|---|---|
| `provideRouter(routes)` | `@angular/router` | Configure router |
| `provideHttpClient()` | `@angular/common/http` | Configure HttpClient |
| `provideAnimationsAsync()` | `@angular/platform-browser/animations/async` | Enable animations (lazy) |
| `provideAnimations()` | `@angular/platform-browser/animations` | Enable animations (eager) |
| `provideClientHydration()` | `@angular/platform-browser` | Enable SSR hydration |
| `provideZoneChangeDetection()` | `@angular/core` | Configure zone.js |

### bootstrapApplication Setup

```typescript
// main.ts
import { bootstrapApplication } from '@angular/platform-browser';
import { provideRouter, withComponentInputBinding } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { AppComponent } from './app/app.component';
import { routes } from './app/app.routes';
import { authInterceptor } from './app/auth.interceptor';

bootstrapApplication(AppComponent, {
  providers: [
    provideRouter(routes, withComponentInputBinding()),
    provideHttpClient(withInterceptors([authInterceptor])),
    provideAnimationsAsync(),
  ],
});
```

### Route-Based Lazy Loading with loadComponent

```typescript
// app.routes.ts
import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () =>
      import('./home/home.component').then(m => m.HomeComponent),
  },
  {
    path: 'products',
    loadComponent: () =>
      import('./products/products.component').then(m => m.ProductsComponent),
  },
  {
    path: 'admin',
    loadChildren: () =>
      import('./admin/admin.routes').then(m => m.ADMIN_ROUTES),
  },
];
```

---

## Dependency Injection Patterns

### inject() vs Constructor Injection

```typescript
// Modern: inject() function (preferred in Angular 17+)
import { Component, inject } from '@angular/core';
import { UserService } from './user.service';

@Component({ /* ... */ })
export class UserProfileComponent {
  private userService = inject(UserService);
  private router = inject(Router);
}

// Legacy: Constructor injection
@Component({ /* ... */ })
export class LegacyComponent {
  constructor(
    private userService: UserService,
    private router: Router
  ) {}
}
```

### InjectionToken Creation

```typescript
import { InjectionToken, inject } from '@angular/core';

// Simple value token
export const API_BASE_URL = new InjectionToken<string>('API_BASE_URL');

// Token with factory (no provider needed)
export const WINDOW = new InjectionToken<Window>('Window', {
  providedIn: 'root',
  factory: () => window,
});

// Token with dependencies in factory
export const API_CLIENT = new InjectionToken<ApiClient>('ApiClient', {
  providedIn: 'root',
  factory: () => {
    const baseUrl = inject(API_BASE_URL);
    const http = inject(HttpClient);
    return new ApiClient(http, baseUrl);
  },
});
```

### Provider Types

| Type | Syntax | Description |
|---|---|---|
| `useClass` | `{ provide: Logger, useClass: FileLogger }` | Substitute one class for another |
| `useValue` | `{ provide: API_URL, useValue: 'https://api.example.com' }` | Provide a static value |
| `useFactory` | `{ provide: Cache, useFactory: () => new Cache(100) }` | Create via factory function |
| `useExisting` | `{ provide: AbstractLogger, useExisting: FileLogger }` | Alias one token to another |

```typescript
// Providing in bootstrapApplication
bootstrapApplication(AppComponent, {
  providers: [
    { provide: API_BASE_URL, useValue: 'https://api.example.com' },
    { provide: LoggerService, useClass: ConsoleLoggerService },
    {
      provide: CacheService,
      useFactory: () => {
        const isProduction = inject(IS_PRODUCTION);
        return isProduction ? new RedisCacheService() : new MemoryCacheService();
      },
    },
    { provide: AbstractDataService, useExisting: ConcreteDataService },
  ],
});
```

### providedIn Options

| Option | Scope | Tree-Shakable |
|---|---|---|
| `'root'` | Application-wide singleton | Yes |
| `'platform'` | Shared across multiple apps | Yes |
| `'any'` | Unique instance per lazy module | Yes |

```typescript
@Injectable({ providedIn: 'root' })
export class AuthService {
  // Single instance across the entire application
}

@Injectable({ providedIn: 'any' })
export class FeatureStateService {
  // Each lazy-loaded boundary gets its own instance
}
```

### Hierarchical Injection

```typescript
// Parent provides a service
@Component({
  selector: 'app-dashboard',
  providers: [DashboardStateService],
  template: `<app-widget />`,
})
export class DashboardComponent {
  state = inject(DashboardStateService);
}

// Child inherits parent's instance
@Component({
  selector: 'app-widget',
  template: `<p>{{ state.title() }}</p>`,
})
export class WidgetComponent {
  state = inject(DashboardStateService);
  // Same instance as parent
}

// Optional injection
@Component({ /* ... */ })
export class FlexibleComponent {
  logger = inject(LoggerService, { optional: true });
  // Returns null if not provided

  parentRef = inject(ParentComponent, { skipSelf: true, optional: true });
  // Skips own injector, looks upward
}
```

### Environmental Injectors

```typescript
import { EnvironmentInjector, createEnvironmentInjector, inject } from '@angular/core';

@Component({ /* ... */ })
export class DynamicLoaderComponent {
  private parentInjector = inject(EnvironmentInjector);

  loadFeature() {
    const childInjector = createEnvironmentInjector(
      [
        { provide: FEATURE_CONFIG, useValue: { theme: 'dark' } },
        FeatureService,
      ],
      this.parentInjector
    );
  }
}
```

---

## RxJS Operator Reference

### Transformation Operators

| Operator | Description | Use When |
|---|---|---|
| `map` | Transform each emitted value | Simple value mapping |
| `switchMap` | Map to inner observable, cancel previous | Latest value matters (search, navigation) |
| `mergeMap` | Map to inner observable, run concurrently | All results matter, order doesn't |
| `concatMap` | Map to inner observable, queue sequentially | Order matters (sequential writes) |
| `exhaustMap` | Map to inner observable, ignore until complete | Prevent duplicate submissions |
| `scan` | Accumulate values over time | Running totals, state accumulation |
| `pluck` | Extract nested property (deprecated, use `map`) | N/A |

### Filtering Operators

| Operator | Description |
|---|---|
| `filter` | Emit only values matching a predicate |
| `take(n)` | Emit first n values, then complete |
| `takeUntil(notifier$)` | Emit until notifier emits |
| `takeWhile(predicate)` | Emit while predicate is true |
| `skip(n)` | Skip first n values |
| `distinctUntilChanged()` | Suppress duplicate consecutive values |
| `debounceTime(ms)` | Wait for silence before emitting |
| `throttleTime(ms)` | Emit at most once per interval |
| `first()` | Emit first value (or first matching), then complete |
| `last()` | Emit last value on completion |

### Combination Operators

| Operator | Description | Use When |
|---|---|---|
| `combineLatest` | Emit latest from each source when any emits | Dependent on multiple changing sources |
| `forkJoin` | Emit last value from each source on completion | Parallel HTTP requests |
| `merge` | Interleave emissions from multiple sources | Multiple event streams |
| `zip` | Pair values by index from each source | Lock-step combination |
| `withLatestFrom` | Combine with latest from another source | Primary stream + context data |
| `concat` | Subscribe to observables in sequence | Ordered stream composition |

### Error Handling Operators

| Operator | Description |
|---|---|
| `catchError(fn)` | Handle error and return replacement observable |
| `retry(n)` | Resubscribe on error up to n times |
| `retryWhen(notifier)` | Resubscribe based on notifier logic (deprecated) |
| `retry({ count, delay })` | Retry with configurable delay and count |

### Utility Operators

| Operator | Description |
|---|---|
| `tap(fn)` | Perform side effects without altering the stream |
| `finalize(fn)` | Execute callback on completion or error |
| `shareReplay({ bufferSize, refCount })` | Share and replay emissions to late subscribers |
| `delay(ms)` | Delay each emission by specified time |
| `timeout(ms)` | Error if no emission within specified time |
| `startWith(value)` | Prepend a value before emissions begin |

### Decision Table: switchMap vs mergeMap vs concatMap vs exhaustMap

| Scenario | Operator | Reason |
|---|---|---|
| Typeahead / live search | `switchMap` | Cancel stale requests when new input arrives |
| Auto-save on change | `switchMap` | Only the latest save matters |
| Navigation / route change | `switchMap` | Cancel outgoing requests on re-navigate |
| File upload (multiple files) | `mergeMap` | Upload all concurrently |
| Bulk notifications | `mergeMap` | Send all, order irrelevant |
| Sequential form submissions | `concatMap` | Preserve order of operations |
| Chat messages | `concatMap` | Messages must arrive in order |
| Login button click | `exhaustMap` | Ignore clicks while login is in progress |
| Payment submission | `exhaustMap` | Prevent duplicate charges |

### takeUntilDestroyed Pattern

```typescript
import { Component, inject, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { interval } from 'rxjs';

@Component({ /* ... */ })
export class PollingComponent {
  private destroyRef = inject(DestroyRef);

  constructor() {
    // Automatically unsubscribes when component is destroyed
    interval(5000)
      .pipe(takeUntilDestroyed())
      .subscribe(() => this.refresh());

    // Can also be used outside constructor with explicit DestroyRef
    this.setupLater();
  }

  private setupLater() {
    interval(10000)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(() => this.checkStatus());
  }

  private refresh() { /* ... */ }
  private checkStatus() { /* ... */ }
}
```

### Subject Types Comparison

| Type | Description | Replays | Initial Value |
|---|---|---|---|
| `Subject` | Basic multicast observable | None | No |
| `BehaviorSubject` | Holds current value, emits to new subscribers | Last (1) | Yes (required) |
| `ReplaySubject(n)` | Replays last n values to new subscribers | Last n | No |
| `AsyncSubject` | Emits last value only on completion | Last on complete | No |

```typescript
import { Subject, BehaviorSubject, ReplaySubject, AsyncSubject } from 'rxjs';

const subject = new Subject<number>();
const behavior = new BehaviorSubject<number>(0);     // Must have initial value
const replay = new ReplaySubject<number>(3);          // Replays last 3
const async$ = new AsyncSubject<number>();            // Emits only on complete

behavior.getValue(); // Synchronous access to current value: 0
```

---

## Angular Material Patterns

### Setup

```typescript
// main.ts
import { bootstrapApplication } from '@angular/platform-browser';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { AppComponent } from './app/app.component';

bootstrapApplication(AppComponent, {
  providers: [provideAnimationsAsync()],
});
```

### Theming with CSS Custom Properties

```scss
// styles.scss
@use '@angular/material' as mat;

$theme: mat.define-theme((
  color: (
    theme-type: light,
    primary: mat.$azure-palette,
    tertiary: mat.$blue-palette,
  ),
  typography: (
    brand-family: 'Roboto',
    plain-family: 'Roboto',
  ),
  density: (
    scale: 0,
  ),
));

html {
  @include mat.all-component-themes($theme);
}
```

### Common Component Patterns

#### Data Table

```typescript
import { Component, signal, computed, inject } from '@angular/core';
import { MatTableModule } from '@angular/material/table';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatSortModule, Sort } from '@angular/material/sort';

interface User {
  id: number;
  name: string;
  email: string;
  role: string;
}

@Component({
  selector: 'app-user-table',
  imports: [MatTableModule, MatPaginatorModule, MatSortModule],
  template: `
    <table mat-table [dataSource]="users()" matSort (matSortChange)="onSort($event)">
      <ng-container matColumnDef="name">
        <th mat-header-cell *matHeaderCellDef mat-sort-header>Name</th>
        <td mat-cell *matCellDef="let user">{{ user.name }}</td>
      </ng-container>

      <ng-container matColumnDef="email">
        <th mat-header-cell *matHeaderCellDef mat-sort-header>Email</th>
        <td mat-cell *matCellDef="let user">{{ user.email }}</td>
      </ng-container>

      <ng-container matColumnDef="role">
        <th mat-header-cell *matHeaderCellDef>Role</th>
        <td mat-cell *matCellDef="let user">{{ user.role }}</td>
      </ng-container>

      <tr mat-header-row *matHeaderRowDef="displayedColumns"></tr>
      <tr mat-row *matRowDef="let row; columns: displayedColumns"></tr>
    </table>

    <mat-paginator
      [length]="totalUsers()"
      [pageSize]="10"
      [pageSizeOptions]="[5, 10, 25]"
      (page)="onPage($event)" />
  `,
})
export class UserTableComponent {
  users = signal<User[]>([]);
  totalUsers = signal(0);
  displayedColumns = ['name', 'email', 'role'];

  onSort(sort: Sort) { /* ... */ }
  onPage(event: PageEvent) { /* ... */ }
}
```

#### Dialog

```typescript
import { Component, inject } from '@angular/core';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';

@Component({
  selector: 'app-confirm-dialog',
  imports: [MatDialogModule, MatButtonModule],
  template: `
    <h2 mat-dialog-title>Confirm Action</h2>
    <mat-dialog-content>Are you sure you want to proceed?</mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Cancel</button>
      <button mat-flat-button color="warn" [mat-dialog-close]="true">Confirm</button>
    </mat-dialog-actions>
  `,
})
export class ConfirmDialogComponent {}

// Opening the dialog
@Component({ /* ... */ })
export class ParentComponent {
  private dialog = inject(MatDialog);

  openConfirm() {
    const dialogRef = this.dialog.open(ConfirmDialogComponent, {
      width: '400px',
      disableClose: true,
    });

    dialogRef.afterClosed().subscribe(result => {
      if (result) {
        this.performAction();
      }
    });
  }

  private performAction() { /* ... */ }
}
```

#### Form Fields and Snackbar

```typescript
import { Component, inject } from '@angular/core';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';

@Component({
  selector: 'app-contact-form',
  imports: [MatFormFieldModule, MatInputModule, MatSnackBarModule, ReactiveFormsModule],
  template: `
    <form [formGroup]="form" (ngSubmit)="onSubmit()">
      <mat-form-field appearance="outline">
        <mat-label>Email</mat-label>
        <input matInput formControlName="email" type="email" />
        <mat-error>Valid email required</mat-error>
      </mat-form-field>

      <mat-form-field appearance="outline">
        <mat-label>Message</mat-label>
        <textarea matInput formControlName="message" rows="4"></textarea>
        <mat-hint>{{ form.get('message')?.value?.length || 0 }} / 500</mat-hint>
        <mat-error>Message is required</mat-error>
      </mat-form-field>
    </form>
  `,
})
export class ContactFormComponent {
  private fb = inject(FormBuilder);
  private snackBar = inject(MatSnackBar);

  form = this.fb.group({
    email: ['', [Validators.required, Validators.email]],
    message: ['', [Validators.required, Validators.maxLength(500)]],
  });

  onSubmit() {
    if (this.form.valid) {
      this.snackBar.open('Message sent!', 'Close', { duration: 3000 });
    }
  }
}
```

### CDK Utilities

| CDK Module | Import | Purpose |
|---|---|---|
| Overlay | `@angular/cdk/overlay` | Floating panels, tooltips, dropdowns |
| DragDrop | `@angular/cdk/drag-drop` | Drag and drop lists and items |
| VirtualScroll | `@angular/cdk/scrolling` | Efficiently render large lists |
| A11y | `@angular/cdk/a11y` | Focus management, live announcer |
| Clipboard | `@angular/cdk/clipboard` | Copy text to clipboard |
| Portal | `@angular/cdk/portal` | Dynamic content rendering |

```typescript
// Virtual Scrolling
import { Component, signal } from '@angular/core';
import { ScrollingModule } from '@angular/cdk/scrolling';

@Component({
  selector: 'app-virtual-list',
  imports: [ScrollingModule],
  template: `
    <cdk-virtual-scroll-viewport itemSize="48" class="viewport">
      <div *cdkVirtualFor="let item of items()" class="item">
        {{ item.name }}
      </div>
    </cdk-virtual-scroll-viewport>
  `,
  styles: [`.viewport { height: 400px; }`],
})
export class VirtualListComponent {
  items = signal(Array.from({ length: 10000 }, (_, i) => ({ name: `Item ${i}` })));
}
```

---

## Forms Patterns

### Reactive Forms with FormBuilder

```typescript
import { Component, inject } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';

@Component({
  selector: 'app-registration',
  imports: [ReactiveFormsModule],
  template: `
    <form [formGroup]="form" (ngSubmit)="onSubmit()">
      <input formControlName="name" placeholder="Name" />
      <input formControlName="email" placeholder="Email" />
      <input formControlName="password" type="password" placeholder="Password" />
      <button type="submit" [disabled]="form.invalid">Register</button>
    </form>
  `,
})
export class RegistrationComponent {
  private fb = inject(FormBuilder);

  form = this.fb.group({
    name: ['', [Validators.required, Validators.minLength(2)]],
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, Validators.minLength(8)]],
  });

  onSubmit() {
    if (this.form.valid) {
      console.log(this.form.value);
    }
  }
}
```

### Typed Forms (NonNullableFormBuilder)

```typescript
import { Component, inject } from '@angular/core';
import { ReactiveFormsModule, NonNullableFormBuilder, Validators } from '@angular/forms';

interface ProfileForm {
  firstName: string;
  lastName: string;
  age: number;
  notifications: boolean;
}

@Component({
  selector: 'app-profile',
  imports: [ReactiveFormsModule],
  template: `
    <form [formGroup]="form" (ngSubmit)="onSubmit()">
      <input formControlName="firstName" />
      <input formControlName="lastName" />
      <input formControlName="age" type="number" />
      <label>
        <input formControlName="notifications" type="checkbox" />
        Notifications
      </label>
      <button type="submit">Save</button>
      <button type="button" (click)="form.reset()">Reset</button>
    </form>
  `,
})
export class ProfileComponent {
  private fb = inject(NonNullableFormBuilder);

  form = this.fb.group({
    firstName: ['', Validators.required],
    lastName: ['', Validators.required],
    age: [0, [Validators.required, Validators.min(0)]],
    notifications: [true],
  });

  onSubmit() {
    // Type-safe: form.getRawValue() returns ProfileForm-like shape
    const values = this.form.getRawValue();
    // values.firstName is string (never null due to NonNullable)
  }
}
```

### Custom Validators

```typescript
import { AbstractControl, ValidationErrors, ValidatorFn, AsyncValidatorFn } from '@angular/forms';
import { Observable, map, catchError, of, delay } from 'rxjs';
import { inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';

// Sync validator
export function forbiddenNameValidator(forbidden: RegExp): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null => {
    const isForbidden = forbidden.test(control.value);
    return isForbidden ? { forbiddenName: { value: control.value } } : null;
  };
}

// Async validator (checks server)
export function uniqueEmailValidator(http: HttpClient): AsyncValidatorFn {
  return (control: AbstractControl): Observable<ValidationErrors | null> => {
    return http.get<{ available: boolean }>(`/api/check-email?email=${control.value}`).pipe(
      map(res => (res.available ? null : { emailTaken: true })),
      catchError(() => of(null))
    );
  };
}

// Cross-field validator
export function passwordMatchValidator(): ValidatorFn {
  return (group: AbstractControl): ValidationErrors | null => {
    const password = group.get('password')?.value;
    const confirm = group.get('confirmPassword')?.value;
    return password === confirm ? null : { passwordMismatch: true };
  };
}

// Usage
@Component({ /* ... */ })
export class SignupComponent {
  private fb = inject(FormBuilder);
  private http = inject(HttpClient);

  form = this.fb.group(
    {
      username: ['', [Validators.required, forbiddenNameValidator(/admin/i)]],
      email: ['', [Validators.required, Validators.email], [uniqueEmailValidator(this.http)]],
      password: ['', [Validators.required, Validators.minLength(8)]],
      confirmPassword: ['', Validators.required],
    },
    { validators: passwordMatchValidator() }
  );
}
```

### Dynamic Forms with FormArray

```typescript
import { Component, inject } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, FormArray, Validators } from '@angular/forms';

@Component({
  selector: 'app-invoice',
  imports: [ReactiveFormsModule],
  template: `
    <form [formGroup]="form">
      <div formArrayName="items">
        @for (item of items.controls; track item; let i = $index) {
          <div [formGroupName]="i" class="item-row">
            <input formControlName="description" placeholder="Description" />
            <input formControlName="quantity" type="number" placeholder="Qty" />
            <input formControlName="price" type="number" placeholder="Price" />
            <button type="button" (click)="removeItem(i)">Remove</button>
          </div>
        }
      </div>
      <button type="button" (click)="addItem()">Add Item</button>
      <p>Total: {{ total() }}</p>
    </form>
  `,
})
export class InvoiceComponent {
  private fb = inject(FormBuilder);

  form = this.fb.group({
    items: this.fb.array([this.createItem()]),
  });

  get items(): FormArray {
    return this.form.get('items') as FormArray;
  }

  total = computed(() => {
    // For signal-based reactivity, consider toSignal or manual tracking
    return this.items.controls.reduce((sum, ctrl) => {
      const qty = ctrl.get('quantity')?.value || 0;
      const price = ctrl.get('price')?.value || 0;
      return sum + qty * price;
    }, 0);
  });

  createItem() {
    return this.fb.group({
      description: ['', Validators.required],
      quantity: [1, [Validators.required, Validators.min(1)]],
      price: [0, [Validators.required, Validators.min(0)]],
    });
  }

  addItem() {
    this.items.push(this.createItem());
  }

  removeItem(index: number) {
    this.items.removeAt(index);
  }
}
```

### ControlValueAccessor for Custom Form Controls

```typescript
import { Component, forwardRef, signal } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';

@Component({
  selector: 'app-star-rating',
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => StarRatingComponent),
      multi: true,
    },
  ],
  template: `
    @for (star of [1,2,3,4,5]; track star) {
      <button
        type="button"
        (click)="selectStar(star)"
        [class.filled]="star <= value()"
        [disabled]="disabled()">
        ★
      </button>
    }
  `,
  styles: [`
    button { background: none; border: none; font-size: 1.5rem; cursor: pointer; color: #ccc; }
    button.filled { color: gold; }
    button:disabled { cursor: not-allowed; opacity: 0.5; }
  `],
})
export class StarRatingComponent implements ControlValueAccessor {
  value = signal(0);
  disabled = signal(false);

  private onChange: (value: number) => void = () => {};
  private onTouched: () => void = () => {};

  writeValue(val: number): void {
    this.value.set(val || 0);
  }

  registerOnChange(fn: (value: number) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    this.disabled.set(isDisabled);
  }

  selectStar(star: number): void {
    this.value.set(star);
    this.onChange(star);
    this.onTouched();
  }
}
```

---

## Router Patterns

### Route Configuration

```typescript
// app.routes.ts
import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', redirectTo: '/dashboard', pathMatch: 'full' },
  {
    path: 'dashboard',
    loadComponent: () =>
      import('./dashboard/dashboard.component').then(m => m.DashboardComponent),
    title: 'Dashboard',
  },
  {
    path: 'users/:id',
    loadComponent: () =>
      import('./user-detail/user-detail.component').then(m => m.UserDetailComponent),
    title: 'User Detail',
  },
  {
    path: 'admin',
    loadChildren: () => import('./admin/admin.routes').then(m => m.ADMIN_ROUTES),
    canMatch: [adminGuard],
  },
  { path: '**', loadComponent: () => import('./not-found.component').then(m => m.NotFoundComponent) },
];
```

### Functional Guards

```typescript
import { inject } from '@angular/core';
import { CanActivateFn, CanDeactivateFn, CanMatchFn, Router } from '@angular/router';
import { AuthService } from './auth.service';

// CanActivateFn
export const authGuard: CanActivateFn = (route, state) => {
  const auth = inject(AuthService);
  const router = inject(Router);

  if (auth.isAuthenticated()) {
    return true;
  }

  return router.createUrlTree(['/login'], {
    queryParams: { returnUrl: state.url },
  });
};

// CanDeactivateFn (unsaved changes warning)
export interface HasUnsavedChanges {
  hasUnsavedChanges(): boolean;
}

export const unsavedChangesGuard: CanDeactivateFn<HasUnsavedChanges> = (component) => {
  if (component.hasUnsavedChanges()) {
    return window.confirm('You have unsaved changes. Leave anyway?');
  }
  return true;
};

// CanMatchFn (role-based route matching)
export const adminGuard: CanMatchFn = () => {
  const auth = inject(AuthService);
  return auth.hasRole('admin');
};
```

### Functional Resolvers

```typescript
import { inject } from '@angular/core';
import { ResolveFn } from '@angular/router';
import { UserService } from './user.service';
import { User } from './user.model';

export const userResolver: ResolveFn<User> = (route) => {
  const userService = inject(UserService);
  const userId = route.paramMap.get('id')!;
  return userService.getById(userId);
};

// Route config
const routes: Routes = [
  {
    path: 'users/:id',
    loadComponent: () => import('./user-detail.component').then(m => m.UserDetailComponent),
    resolve: { user: userResolver },
  },
];

// Component reads resolved data
@Component({ /* ... */ })
export class UserDetailComponent {
  user = input.required<User>(); // Via withComponentInputBinding()
}
```

### withComponentInputBinding

```typescript
// main.ts
import { provideRouter, withComponentInputBinding } from '@angular/router';

bootstrapApplication(AppComponent, {
  providers: [
    provideRouter(routes, withComponentInputBinding()),
  ],
});

// Component receives route params, query params, data, and resolve as inputs
@Component({ /* ... */ })
export class ProductComponent {
  id = input<string>();           // From :id route param
  search = input<string>();       // From ?search= query param
  user = input<User>();           // From resolve: { user: ... }
}
```

### Preloading Strategies

```typescript
import { provideRouter, withPreloading, PreloadAllModules } from '@angular/router';

// Preload all lazy routes
bootstrapApplication(AppComponent, {
  providers: [
    provideRouter(routes, withPreloading(PreloadAllModules)),
  ],
});

// Custom preloading strategy
import { PreloadingStrategy, Route } from '@angular/router';
import { Observable, of, EMPTY } from 'rxjs';

export class SelectivePreloadStrategy implements PreloadingStrategy {
  preload(route: Route, load: () => Observable<any>): Observable<any> {
    return route.data?.['preload'] ? load() : EMPTY;
  }
}

// Mark routes for preloading
const routes: Routes = [
  {
    path: 'dashboard',
    loadComponent: () => import('./dashboard.component').then(m => m.DashboardComponent),
    data: { preload: true },
  },
  {
    path: 'settings',
    loadComponent: () => import('./settings.component').then(m => m.SettingsComponent),
    // Not preloaded
  },
];
```

---

## HTTP Patterns

### Typed HTTP Requests

```typescript
import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';

interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}

@Injectable({ providedIn: 'root' })
export class ProductService {
  private http = inject(HttpClient);
  private baseUrl = '/api/products';

  getAll(page = 1, pageSize = 20): Observable<PaginatedResponse<Product>> {
    const params = new HttpParams()
      .set('page', page)
      .set('pageSize', pageSize);

    return this.http.get<PaginatedResponse<Product>>(this.baseUrl, { params });
  }

  getById(id: string): Observable<Product> {
    return this.http.get<Product>(`${this.baseUrl}/${id}`);
  }

  create(product: Omit<Product, 'id'>): Observable<Product> {
    return this.http.post<Product>(this.baseUrl, product);
  }

  update(id: string, product: Partial<Product>): Observable<Product> {
    return this.http.patch<Product>(`${this.baseUrl}/${id}`, product);
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
```

### Functional Interceptors

```typescript
import { HttpInterceptorFn, HttpRequest, HttpHandlerFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { AuthService } from './auth.service';
import { catchError, throwError } from 'rxjs';
import { Router } from '@angular/router';

// Auth token interceptor
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const token = auth.getToken();

  if (token) {
    const cloned = req.clone({
      setHeaders: { Authorization: `Bearer ${token}` },
    });
    return next(cloned);
  }

  return next(req);
};

// Error handling interceptor
export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const router = inject(Router);

  return next(req).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status === 401) {
        router.navigate(['/login']);
      } else if (error.status === 403) {
        router.navigate(['/forbidden']);
      }
      return throwError(() => error);
    })
  );
};

// Logging interceptor
export const loggingInterceptor: HttpInterceptorFn = (req, next) => {
  const started = performance.now();
  return next(req).pipe(
    finalize(() => {
      const elapsed = performance.now() - started;
      console.log(`${req.method} ${req.urlWithParams} - ${elapsed.toFixed(0)}ms`);
    })
  );
};
```

### Registering Interceptors

```typescript
import { provideHttpClient, withInterceptors, withInterceptorsFromDi } from '@angular/common/http';

bootstrapApplication(AppComponent, {
  providers: [
    provideHttpClient(
      withInterceptors([authInterceptor, errorInterceptor, loggingInterceptor])
    ),
    // Or for DI-based (class) interceptors:
    // provideHttpClient(withInterceptorsFromDi()),
  ],
});
```

### Retry Strategies

```typescript
import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { retry, timer } from 'rxjs';

@Injectable({ providedIn: 'root' })
export class ResilientApiService {
  private http = inject(HttpClient);

  fetchData() {
    return this.http.get('/api/data').pipe(
      retry({
        count: 3,
        delay: (error, retryCount) => {
          // Exponential backoff: 1s, 2s, 4s
          const delayMs = Math.pow(2, retryCount - 1) * 1000;
          console.warn(`Retry #${retryCount} in ${delayMs}ms`);
          return timer(delayMs);
        },
      })
    );
  }
}
```

### Request Caching

```typescript
import { HttpInterceptorFn } from '@angular/common/http';
import { of, tap } from 'rxjs';

const cache = new Map<string, { response: any; timestamp: number }>();
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

export const cachingInterceptor: HttpInterceptorFn = (req, next) => {
  if (req.method !== 'GET') {
    return next(req);
  }

  const cached = cache.get(req.urlWithParams);
  if (cached && Date.now() - cached.timestamp < CACHE_DURATION) {
    return of(cached.response.clone());
  }

  return next(req).pipe(
    tap(response => {
      cache.set(req.urlWithParams, {
        response: response.clone(),
        timestamp: Date.now(),
      });
    })
  );
};
```

### File Upload and Download

```typescript
import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpEventType } from '@angular/common/http';
import { Observable, filter, map } from 'rxjs';

@Injectable({ providedIn: 'root' })
export class FileService {
  private http = inject(HttpClient);

  upload(file: File): Observable<number | string> {
    const formData = new FormData();
    formData.append('file', file, file.name);

    return this.http
      .post('/api/upload', formData, {
        reportProgress: true,
        observe: 'events',
      })
      .pipe(
        map(event => {
          switch (event.type) {
            case HttpEventType.UploadProgress:
              return Math.round(((event.loaded || 0) / (event.total || 1)) * 100);
            case HttpEventType.Response:
              return event.body as string;
            default:
              return 0;
          }
        })
      );
  }

  download(fileId: string): void {
    this.http.get(`/api/files/${fileId}`, { responseType: 'blob' }).subscribe(blob => {
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `file-${fileId}`;
      a.click();
      URL.revokeObjectURL(url);
    });
  }
}
```

---

## New Control Flow Syntax

### @if / @else if / @else

```typescript
@Component({
  template: `
    @if (user(); as user) {
      <h1>Welcome, {{ user.name }}</h1>
      @if (user.role === 'admin') {
        <app-admin-panel />
      } @else if (user.role === 'editor') {
        <app-editor-tools />
      } @else {
        <app-viewer />
      }
    } @else {
      <app-login-prompt />
    }
  `,
})
export class HeaderComponent {
  user = signal<User | null>(null);
}
```

### @for with track

```typescript
@Component({
  template: `
    <ul>
      @for (item of items(); track item.id) {
        <li>
          {{ item.name }} - {{ item.price | currency }}
          <button (click)="remove(item.id)">Remove</button>
        </li>
      } @empty {
        <li class="empty">No items found.</li>
      }
    </ul>
  `,
})
export class ItemListComponent {
  items = signal<Item[]>([]);

  remove(id: number) {
    this.items.update(items => items.filter(i => i.id !== id));
  }
}
```

The `track` expression is required and tells Angular how to identify each item for efficient DOM updates.

| Track Expression | Use When |
|---|---|
| `track item.id` | Items have a unique identifier |
| `track item` | Items are primitives or immutable objects |
| `track $index` | Last resort when no unique key exists |

### @switch / @case / @default

```typescript
@Component({
  template: `
    @switch (status()) {
      @case ('loading') {
        <app-spinner />
      }
      @case ('error') {
        <app-error-message [message]="errorMessage()" />
      }
      @case ('empty') {
        <p>No results found.</p>
      }
      @case ('success') {
        <app-results [data]="data()" />
      }
      @default {
        <p>Unknown state</p>
      }
    }
  `,
})
export class DataViewComponent {
  status = signal<'loading' | 'error' | 'empty' | 'success'>('loading');
  data = signal<any[]>([]);
  errorMessage = signal('');
}
```

### @defer / @loading / @placeholder / @error

```typescript
@Component({
  template: `
    <!-- Basic deferral -->
    @defer {
      <app-heavy-chart [data]="chartData()" />
    } @placeholder {
      <div class="placeholder">Chart will appear here</div>
    } @loading (minimum 500ms) {
      <app-spinner />
    } @error {
      <p>Failed to load chart component.</p>
    }

    <!-- Trigger on viewport -->
    @defer (on viewport) {
      <app-comments [postId]="postId()" />
    } @placeholder {
      <div style="height: 200px">Comments load when scrolled into view</div>
    }

    <!-- Trigger on interaction -->
    @defer (on interaction(loadBtn)) {
      <app-details [itemId]="itemId()" />
    } @placeholder {
      <button #loadBtn>Load Details</button>
    }

    <!-- Trigger on hover -->
    @defer (on hover(card)) {
      <app-preview [id]="id()" />
    } @placeholder {
      <div #card class="card">Hover to preview</div>
    }

    <!-- Trigger on idle -->
    @defer (on idle) {
      <app-analytics />
    }

    <!-- Trigger on timer -->
    @defer (on timer(3s)) {
      <app-promotion-banner />
    } @placeholder {
      <div style="height: 80px"></div>
    }

    <!-- Prefetching -->
    @defer (on interaction; prefetch on idle) {
      <app-dashboard-widgets />
    } @placeholder {
      <button>Show Widgets</button>
    }

    <!-- Conditional with when -->
    @defer (when showDetails()) {
      <app-full-details />
    }
  `,
})
export class DeferExamplesComponent {
  chartData = signal<number[]>([]);
  postId = signal('');
  itemId = signal('');
  id = signal('');
  showDetails = signal(false);
}
```

### Migration from Structural Directives

| Old Syntax | New Syntax |
|---|---|
| `*ngIf="cond"` | `@if (cond) { ... }` |
| `*ngIf="cond; else elseBlock"` | `@if (cond) { ... } @else { ... }` |
| `*ngIf="obs$ \| async as val"` | `@if (signal(); as val) { ... }` |
| `*ngFor="let item of items; trackBy: trackFn"` | `@for (item of items(); track item.id) { ... }` |
| `*ngSwitch / *ngSwitchCase` | `@switch (val) { @case (x) { ... } }` |
| None | `@defer { ... }` (new capability) |

---

## SSR Patterns

### Angular SSR Setup

```typescript
// app.config.server.ts
import { ApplicationConfig, mergeApplicationConfig } from '@angular/core';
import { provideServerRendering } from '@angular/platform-server';
import { provideServerRoutesConfig } from '@angular/ssr';
import { appConfig } from './app.config';
import { serverRoutes } from './app.routes.server';

const serverConfig: ApplicationConfig = {
  providers: [
    provideServerRendering(),
    provideServerRoutesConfig(serverRoutes),
  ],
};

export const config = mergeApplicationConfig(appConfig, serverConfig);
```

### Client Hydration

```typescript
// app.config.ts
import { provideClientHydration, withEventReplay } from '@angular/platform-browser';

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    provideHttpClient(withFetch()),
    provideClientHydration(withEventReplay()),
  ],
};
```

### Transfer State

```typescript
import { Injectable, inject } from '@angular/core';
import { TransferState, makeStateKey } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of, tap } from 'rxjs';

const PRODUCTS_KEY = makeStateKey<Product[]>('products');

@Injectable({ providedIn: 'root' })
export class ProductService {
  private http = inject(HttpClient);
  private transferState = inject(TransferState);

  getProducts(): Observable<Product[]> {
    // Check if data was transferred from server
    const stored = this.transferState.get(PRODUCTS_KEY, null);
    if (stored) {
      this.transferState.remove(PRODUCTS_KEY);
      return of(stored);
    }

    // Fetch and store for transfer
    return this.http.get<Product[]>('/api/products').pipe(
      tap(products => this.transferState.set(PRODUCTS_KEY, products))
    );
  }
}
```

### Platform Checks

```typescript
import { Component, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser, isPlatformServer } from '@angular/common';

@Component({ /* ... */ })
export class ChartComponent {
  private platformId = inject(PLATFORM_ID);

  ngOnInit() {
    if (isPlatformBrowser(this.platformId)) {
      // Safe to access window, document, localStorage
      this.initChart();
    }

    if (isPlatformServer(this.platformId)) {
      // Server-only logic
    }
  }

  private initChart() {
    // Browser-only chart library
  }
}
```

### afterNextRender and afterRender

```typescript
import { Component, afterNextRender, afterRender, ElementRef, viewChild } from '@angular/core';

@Component({
  selector: 'app-canvas',
  template: `<canvas #canvas width="800" height="600"></canvas>`,
})
export class CanvasComponent {
  canvas = viewChild.required<ElementRef<HTMLCanvasElement>>('canvas');

  constructor() {
    // Runs once after the NEXT render (SSR-safe)
    afterNextRender(() => {
      const ctx = this.canvas().nativeElement.getContext('2d');
      if (ctx) {
        this.drawInitial(ctx);
      }
    });

    // Runs after EVERY render cycle (SSR-safe)
    afterRender(() => {
      // Sync DOM measurements, update third-party libs, etc.
    });
  }

  private drawInitial(ctx: CanvasRenderingContext2D) {
    ctx.fillStyle = '#333';
    ctx.fillRect(0, 0, 800, 600);
  }
}
```

### SSR-Safe Patterns

| Unsafe | SSR-Safe Alternative |
|---|---|
| `window.innerWidth` | `afterNextRender(() => { ... })` |
| `document.getElementById(...)` | `viewChild()` signal query |
| `localStorage.getItem(...)` | Check `isPlatformBrowser()` first |
| `setTimeout` in constructor | `afterNextRender` or platform check |
| Direct DOM manipulation | Renderer2 or signal-based templates |

---

## Performance Patterns

### OnPush Change Detection

```typescript
import { Component, ChangeDetectionStrategy, signal, computed } from '@angular/core';

@Component({
  selector: 'app-product-card',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="card">
      <h3>{{ product().name }}</h3>
      <p>{{ formattedPrice() }}</p>
      <button (click)="addToCart()">Add to Cart</button>
    </div>
  `,
})
export class ProductCardComponent {
  product = input.required<Product>();

  formattedPrice = computed(() =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' })
      .format(this.product().price)
  );

  addToCart() { /* ... */ }
}
```

### Signal-Based Change Detection

Signals automatically notify Angular when values change, making OnPush even more effective:

```typescript
@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <p>Count: {{ count() }}</p>
    <p>Doubled: {{ doubled() }}</p>
    <button (click)="increment()">+1</button>
  `,
})
export class CounterComponent {
  count = signal(0);
  doubled = computed(() => this.count() * 2);

  increment() {
    this.count.update(c => c + 1);
    // Angular automatically detects the signal change
    // No need for ChangeDetectorRef.markForCheck()
  }
}
```

### track Expression for @for Loops

```typescript
@Component({
  template: `
    <!-- Good: track by unique identifier -->
    @for (user of users(); track user.id) {
      <app-user-card [user]="user" />
    }

    <!-- Acceptable for primitives -->
    @for (tag of tags(); track tag) {
      <span class="tag">{{ tag }}</span>
    }

    <!-- Last resort: index-based tracking -->
    @for (item of items(); track $index) {
      <div>{{ item }}</div>
    }
  `,
})
export class ListComponent {
  users = signal<User[]>([]);
  tags = signal<string[]>([]);
  items = signal<string[]>([]);
}
```

### Lazy Loading Strategies

```typescript
// 1. Route-level lazy loading
const routes: Routes = [
  {
    path: 'reports',
    loadComponent: () => import('./reports/reports.component').then(m => m.ReportsComponent),
  },
  {
    path: 'admin',
    loadChildren: () => import('./admin/admin.routes').then(m => m.ADMIN_ROUTES),
  },
];

// 2. Defer block for component-level lazy loading
@Component({
  template: `
    @defer (on viewport) {
      <app-heavy-widget />
    } @placeholder {
      <div style="height:300px"></div>
    }
  `,
})
export class DashboardComponent {}

// 3. Dynamic import in service
@Injectable({ providedIn: 'root' })
export class ChartService {
  private chartLib: any;

  async loadChart(container: HTMLElement) {
    if (!this.chartLib) {
      this.chartLib = await import('chart.js');
    }
    return new this.chartLib.Chart(container, { /* config */ });
  }
}
```

### Bundle Optimization Checklist

| Technique | Impact | How |
|---|---|---|
| Route lazy loading | High | `loadComponent` / `loadChildren` |
| @defer blocks | High | Component-level code splitting |
| Tree shaking | High | Use `providedIn: 'root'` for services |
| OnPush detection | Medium | `changeDetection: ChangeDetectionStrategy.OnPush` |
| Signals over RxJS | Medium | Reduces zone.js overhead |
| Image optimization | Medium | `NgOptimizedImage` directive |
| Preloading strategies | Low-Med | `withPreloading(PreloadAllModules)` |
| Source map explorer | Diagnostic | `npx source-map-explorer dist/**/*.js` |

### NgOptimizedImage

```typescript
import { Component } from '@angular/core';
import { NgOptimizedImage } from '@angular/common';

@Component({
  selector: 'app-hero',
  imports: [NgOptimizedImage],
  template: `
    <!-- LCP image: add priority attribute -->
    <img ngSrc="/assets/hero.jpg" width="1200" height="600" priority />

    <!-- Responsive image with fill -->
    <div class="banner">
      <img ngSrc="/assets/banner.jpg" fill [loaderParams]="{ quality: 80 }" />
    </div>

    <!-- Sized image with placeholder -->
    <img ngSrc="/assets/product.jpg" width="400" height="400" placeholder />
  `,
})
export class HeroComponent {}
```

---

## Anti-Patterns

| Pattern | Problem | Solution |
|---|---|---|
| Subscribing in templates without async pipe or toSignal | Memory leaks, no auto-cleanup | Use `toSignal()` or `async` pipe |
| Manual `subscribe()` without unsubscribe | Memory leaks on component destruction | Use `takeUntilDestroyed()` or `DestroyRef` |
| Using `any` type for HttpClient responses | No type safety, runtime errors | Always type: `http.get<User[]>(url)` |
| Mutating signal values in-place | Change detection misses, stale UI | Use `signal.update()` with new references |
| Business logic in components | Hard to test, tight coupling | Extract to injectable services |
| Deeply nested subscriptions | Callback hell, difficult to follow | Use RxJS `switchMap`, `mergeMap`, etc. |
| Not using OnPush | Excessive change detection cycles | Set `changeDetection: ChangeDetectionStrategy.OnPush` |
| Importing entire modules | Large bundles, slow initial load | Import standalone components directly |
| Using `ngOnInit` for signal setup | Unnecessary lifecycle hook | Initialize signals in field declarations or constructor |
| Direct DOM manipulation | Breaks SSR, bypasses Angular | Use template bindings, `viewChild()`, or `Renderer2` |
| Large monolithic components | Hard to maintain and test | Split into smart (container) + dumb (presentational) |
| `setTimeout` / `setInterval` without cleanup | Memory leaks, SSR issues | Use `afterNextRender`, RxJS `timer`, or `DestroyRef` |
| Storing component state in services globally | Shared mutable state bugs | Use component-scoped providers or signals |
| Not using `track` expression wisely | Poor @for loop performance | Track by unique ID, not `$index` |
| Class-based guards and resolvers | More boilerplate, harder to tree-shake | Use functional guards: `CanActivateFn`, `ResolveFn` |
| Calling signals in event bindings with `()` | Creates function calls per change detection | Bind to computed or store result in template variable |

---

## File Organization

### Feature-Based Structure

```
src/
  app/
    core/                          # Singleton services, guards, interceptors
      auth/
        auth.service.ts
        auth.guard.ts
        auth.interceptor.ts
      error/
        error-handler.service.ts
        error.interceptor.ts
      layout/
        header.component.ts
        footer.component.ts
        sidebar.component.ts
    shared/                        # Reusable components, directives, pipes
      components/
        button/
          button.component.ts
          button.component.spec.ts
        modal/
          modal.component.ts
      directives/
        highlight.directive.ts
        tooltip.directive.ts
      pipes/
        truncate.pipe.ts
        time-ago.pipe.ts
      models/
        user.model.ts
        product.model.ts
    features/                      # Feature areas
      dashboard/
        dashboard.component.ts
        dashboard.component.spec.ts
        widgets/
          stats-widget.component.ts
          chart-widget.component.ts
        dashboard.routes.ts
      products/
        product-list.component.ts
        product-detail.component.ts
        product.service.ts
        product.model.ts
        products.routes.ts
      users/
        user-list.component.ts
        user-profile.component.ts
        user.service.ts
        users.routes.ts
    app.component.ts
    app.config.ts
    app.routes.ts
  environments/
    environment.ts
    environment.prod.ts
  main.ts
  styles.scss
```

### Naming Conventions

| Type | Convention | Example |
|---|---|---|
| Component | `kebab-case.component.ts` | `user-profile.component.ts` |
| Service | `kebab-case.service.ts` | `auth.service.ts` |
| Directive | `kebab-case.directive.ts` | `highlight.directive.ts` |
| Pipe | `kebab-case.pipe.ts` | `truncate.pipe.ts` |
| Guard | `kebab-case.guard.ts` | `auth.guard.ts` |
| Interceptor | `kebab-case.interceptor.ts` | `error.interceptor.ts` |
| Model/Interface | `kebab-case.model.ts` | `user.model.ts` |
| Routes | `kebab-case.routes.ts` | `admin.routes.ts` |
| Test | `kebab-case.component.spec.ts` | `user-profile.component.spec.ts` |
| Config | `app.config.ts` / `app.config.server.ts` | Top-level app config |

### Module-Less Architecture (Angular 17+)

In Angular 17+, NgModules are no longer required. The recommended approach uses standalone components with route-based organization:

```typescript
// app.config.ts (replaces AppModule)
import { ApplicationConfig } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import { routes } from './app.routes';

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    provideHttpClient(),
  ],
};

// app.routes.ts (replaces RouterModule.forRoot)
import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: 'dashboard',
    loadComponent: () =>
      import('./features/dashboard/dashboard.component').then(m => m.DashboardComponent),
  },
  {
    path: 'products',
    loadChildren: () =>
      import('./features/products/products.routes').then(m => m.PRODUCT_ROUTES),
  },
];

// features/products/products.routes.ts (replaces feature modules)
import { Routes } from '@angular/router';

export const PRODUCT_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () =>
      import('./product-list.component').then(m => m.ProductListComponent),
  },
  {
    path: ':id',
    loadComponent: () =>
      import('./product-detail.component').then(m => m.ProductDetailComponent),
  },
];
```

### Service Organization Patterns

```typescript
// Feature service with typed state
@Injectable({ providedIn: 'root' })
export class CartService {
  private items = signal<CartItem[]>([]);

  readonly cartItems = this.items.asReadonly();
  readonly itemCount = computed(() => this.items().length);
  readonly total = computed(() =>
    this.items().reduce((sum, item) => sum + item.price * item.quantity, 0)
  );

  addItem(product: Product, quantity = 1): void {
    this.items.update(items => {
      const existing = items.find(i => i.productId === product.id);
      if (existing) {
        return items.map(i =>
          i.productId === product.id
            ? { ...i, quantity: i.quantity + quantity }
            : i
        );
      }
      return [...items, { productId: product.id, name: product.name, price: product.price, quantity }];
    });
  }

  removeItem(productId: string): void {
    this.items.update(items => items.filter(i => i.productId !== productId));
  }

  clear(): void {
    this.items.set([]);
  }
}
```

### Shared Component Patterns

```typescript
// Reusable presentational component
@Component({
  selector: 'app-data-table',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <table>
      <thead>
        <tr>
          @for (col of columns(); track col.key) {
            <th (click)="sort.emit(col.key)">{{ col.label }}</th>
          }
        </tr>
      </thead>
      <tbody>
        @for (row of data(); track trackBy()(row)) {
          <tr (click)="rowClick.emit(row)">
            @for (col of columns(); track col.key) {
              <td>{{ row[col.key] }}</td>
            }
          </tr>
        } @empty {
          <tr>
            <td [attr.colspan]="columns().length">{{ emptyMessage() }}</td>
          </tr>
        }
      </tbody>
    </table>
  `,
})
export class DataTableComponent<T extends Record<string, any>> {
  columns = input.required<{ key: string; label: string }[]>();
  data = input.required<T[]>();
  trackBy = input<(item: T) => any>(() => (item: T) => item['id']);
  emptyMessage = input('No data available');

  sort = output<string>();
  rowClick = output<T>();
}
```

---

## Quick Decision Guides

### When to Use Signals vs RxJS

| Use Case | Signals | RxJS |
|---|---|---|
| Component state | Preferred | Overkill |
| Derived / computed values | `computed()` | `combineLatest` + `map` |
| Template binding | `signal()` | `async` pipe or `toSignal()` |
| HTTP responses | `toSignal(http.get(...))` | `http.get(...)` |
| Complex async flows | Convert at boundary | Preferred |
| Time-based operations | Convert at boundary | `timer`, `interval`, `debounceTime` |
| Event streams | Convert at boundary | `fromEvent`, Subjects |
| State management | Signal stores | NgRx with Observables |

### When to Use @defer vs Route Lazy Loading

| Scenario | @defer | Route Lazy Loading |
|---|---|---|
| Below-the-fold content | Preferred | N/A |
| Heavy component on same page | Preferred | N/A |
| Separate page / view | N/A | Preferred |
| Feature area with many routes | N/A | `loadChildren` |
| Conditionally shown component | `when` trigger | N/A |
| User-initiated load | `on interaction` | N/A |

### Injection Scope Selection

| Need | Pattern |
|---|---|
| Application singleton | `@Injectable({ providedIn: 'root' })` |
| Per-component instance | `@Component({ providers: [MyService] })` |
| Per-route instance | Provide in route config `providers` array |
| Per-lazy-boundary | `@Injectable({ providedIn: 'any' })` |
| Conditional/dynamic | `EnvironmentInjector` + `createEnvironmentInjector` |
