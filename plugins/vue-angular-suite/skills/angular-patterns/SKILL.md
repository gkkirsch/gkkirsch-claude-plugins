---
name: angular-patterns
description: >
  Angular 17+ patterns with signals, standalone components, and modern APIs.
  Use when building Angular applications, working with signals, creating
  services, setting up routing, or using the new control flow syntax.
  Triggers: "angular", "angular signals", "angular standalone", "angular service",
  "angular routing", "angular reactive forms", "angular inject", "@if", "@for".
  NOT for: Vue, React, Svelte, AngularJS (v1).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Angular 17+ Patterns

## Standalone Components with Signals

```typescript
// user-list.component.ts — modern Angular component
import { Component, signal, computed, effect, inject, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { UserService } from './user.service';
import { UserCardComponent } from './user-card.component';

interface User {
  id: string;
  name: string;
  email: string;
  role: 'admin' | 'user';
}

@Component({
  selector: 'app-user-list',
  standalone: true,
  imports: [CommonModule, FormsModule, UserCardComponent],
  template: `
    <div class="user-list">
      <input
        [(ngModel)]="searchQuery"
        placeholder="Search users..."
        (ngModelChange)="onSearch($event)"
      />

      <p>Showing {{ filteredUsers().length }} of {{ users().length }} users</p>

      <!-- New control flow syntax (Angular 17+) -->
      @if (isLoading()) {
        <div class="spinner">Loading...</div>
      } @else if (error()) {
        <div class="error">{{ error() }}</div>
      } @else {
        @for (user of filteredUsers(); track user.id) {
          <app-user-card
            [user]="user"
            [isSelected]="selectedId() === user.id"
            (select)="selectUser($event)"
            (delete)="deleteUser($event)"
          />
        } @empty {
          <p>No users found</p>
        }
      }
    </div>
  `,
})
export class UserListComponent {
  private userService = inject(UserService);

  // Signal inputs (Angular 17.1+)
  role = input<'admin' | 'user' | 'all'>('all');

  // Signal outputs
  userSelected = output<User>();

  // Local signals
  users = signal<User[]>([]);
  searchQuery = signal('');
  selectedId = signal<string | null>(null);
  isLoading = signal(false);
  error = signal<string | null>(null);

  // Computed signals
  filteredUsers = computed(() => {
    let result = this.users();
    const query = this.searchQuery().toLowerCase();
    const roleFilter = this.role();

    if (query) {
      result = result.filter(u =>
        u.name.toLowerCase().includes(query) ||
        u.email.toLowerCase().includes(query)
      );
    }

    if (roleFilter !== 'all') {
      result = result.filter(u => u.role === roleFilter);
    }

    return result;
  });

  constructor() {
    // Effect — runs when dependencies change
    effect(() => {
      console.log(`Filtered to ${this.filteredUsers().length} users`);
    });

    this.loadUsers();
  }

  async loadUsers() {
    this.isLoading.set(true);
    this.error.set(null);
    try {
      const users = await this.userService.getUsers();
      this.users.set(users);
    } catch (e) {
      this.error.set('Failed to load users');
    } finally {
      this.isLoading.set(false);
    }
  }

  onSearch(query: string) {
    this.searchQuery.set(query);
  }

  selectUser(user: User) {
    this.selectedId.set(user.id);
    this.userSelected.emit(user);
  }

  deleteUser(userId: string) {
    this.users.update(users => users.filter(u => u.id !== userId));
  }
}
```

## Services with Dependency Injection

```typescript
// services/api.service.ts — injectable service
import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams, HttpErrorResponse } from '@angular/common/http';
import { environment } from '../environments/environment';
import { catchError, retry, map } from 'rxjs/operators';
import { throwError, Observable } from 'rxjs';

interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}

@Injectable({ providedIn: 'root' })
export class ApiService {
  private http = inject(HttpClient);
  private baseUrl = environment.apiUrl;

  get<T>(path: string, params?: Record<string, string | number>): Observable<T> {
    let httpParams = new HttpParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        httpParams = httpParams.set(key, String(value));
      });
    }

    return this.http.get<T>(`${this.baseUrl}${path}`, { params: httpParams }).pipe(
      retry({ count: 2, delay: 1000 }),
      catchError(this.handleError),
    );
  }

  post<T>(path: string, body: unknown): Observable<T> {
    return this.http.post<T>(`${this.baseUrl}${path}`, body).pipe(
      catchError(this.handleError),
    );
  }

  private handleError(error: HttpErrorResponse) {
    let message = 'An unexpected error occurred';

    if (error.status === 0) {
      message = 'Network error. Check your connection.';
    } else if (error.status === 401) {
      message = 'Authentication required';
    } else if (error.status === 403) {
      message = 'You do not have permission';
    } else if (error.status === 404) {
      message = 'Resource not found';
    } else if (error.error?.message) {
      message = error.error.message;
    }

    return throwError(() => new Error(message));
  }
}
```

```typescript
// services/user.service.ts
import { Injectable, inject } from '@angular/core';
import { ApiService } from './api.service';
import { toSignal } from '@angular/core/rxjs-interop';

@Injectable({ providedIn: 'root' })
export class UserService {
  private api = inject(ApiService);

  getUsers() {
    return this.api.get<User[]>('/users');
  }

  // Convert Observable to Signal for template use
  users = toSignal(this.getUsers(), { initialValue: [] });
}
```

## Reactive Forms

```typescript
// forms/user-form.component.ts
import { Component, inject, output } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, Validators, AbstractControl } from '@angular/forms';

@Component({
  selector: 'app-user-form',
  standalone: true,
  imports: [ReactiveFormsModule],
  template: `
    <form [formGroup]="form" (ngSubmit)="onSubmit()">
      <div class="field">
        <label for="name">Name</label>
        <input id="name" formControlName="name" />
        @if (form.controls.name.touched && form.controls.name.errors) {
          @if (form.controls.name.errors['required']) {
            <span class="error">Name is required</span>
          }
          @if (form.controls.name.errors['minlength']) {
            <span class="error">Name must be at least 2 characters</span>
          }
        }
      </div>

      <div class="field">
        <label for="email">Email</label>
        <input id="email" formControlName="email" type="email" />
        @if (form.controls.email.touched && form.controls.email.errors) {
          <span class="error">Valid email is required</span>
        }
      </div>

      <div class="field">
        <label for="password">Password</label>
        <input id="password" formControlName="password" type="password" />
        @if (form.controls.password.touched && form.controls.password.errors) {
          @if (form.controls.password.errors['pattern']) {
            <span class="error">Must include uppercase, lowercase, and number</span>
          }
        }
      </div>

      <button type="submit" [disabled]="form.invalid || isSubmitting">
        {{ isSubmitting ? 'Saving...' : 'Save' }}
      </button>
    </form>
  `,
})
export class UserFormComponent {
  private fb = inject(FormBuilder);
  submitted = output<{ name: string; email: string; password: string }>();
  isSubmitting = false;

  form = this.fb.nonNullable.group({
    name: ['', [Validators.required, Validators.minLength(2)]],
    email: ['', [Validators.required, Validators.email]],
    password: ['', [
      Validators.required,
      Validators.minLength(8),
      Validators.pattern(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/),
    ]],
  });

  onSubmit() {
    if (this.form.valid) {
      this.isSubmitting = true;
      this.submitted.emit(this.form.getRawValue());
    } else {
      this.form.markAllAsTouched();
    }
  }
}
```

## Routing with Guards

```typescript
// app.routes.ts
import { Routes } from '@angular/router';
import { inject } from '@angular/core';
import { AuthService } from './services/auth.service';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () => import('./pages/home.component').then(m => m.HomeComponent),
  },
  {
    path: 'login',
    loadComponent: () => import('./pages/login.component').then(m => m.LoginComponent),
  },
  {
    path: 'dashboard',
    canActivate: [() => {
      const auth = inject(AuthService);
      if (auth.isAuthenticated()) return true;
      return inject(Router).createUrlTree(['/login']);
    }],
    loadChildren: () => import('./dashboard/dashboard.routes').then(m => m.DASHBOARD_ROUTES),
  },
  {
    path: '**',
    loadComponent: () => import('./pages/not-found.component').then(m => m.NotFoundComponent),
  },
];

// dashboard/dashboard.routes.ts — lazy-loaded child routes
export const DASHBOARD_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./dashboard-layout.component').then(m => m.DashboardLayoutComponent),
    children: [
      { path: '', loadComponent: () => import('./overview.component').then(m => m.OverviewComponent) },
      { path: 'users', loadComponent: () => import('./users.component').then(m => m.UsersComponent) },
      { path: 'users/:id', loadComponent: () => import('./user-detail.component').then(m => m.UserDetailComponent) },
    ],
  },
];
```

## HTTP Interceptors

```typescript
// interceptors/auth.interceptor.ts
import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { AuthService } from '../services/auth.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const token = auth.getToken();

  if (token) {
    req = req.clone({
      setHeaders: { Authorization: `Bearer ${token}` },
    });
  }

  return next(req);
};

// interceptors/error.interceptor.ts
import { HttpInterceptorFn } from '@angular/common/http';
import { catchError, throwError } from 'rxjs';
import { inject } from '@angular/core';
import { Router } from '@angular/router';

export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const router = inject(Router);

  return next(req).pipe(
    catchError((error) => {
      if (error.status === 401) {
        router.navigate(['/login']);
      }
      return throwError(() => error);
    }),
  );
};

// app.config.ts — register interceptors
import { provideHttpClient, withInterceptors } from '@angular/common/http';

export const appConfig = {
  providers: [
    provideHttpClient(withInterceptors([authInterceptor, errorInterceptor])),
    provideRouter(routes),
  ],
};
```

## Gotchas

1. **Signals don't trigger change detection like observables** -- Signals work with Angular's new signal-based change detection. But if you mix signals with zone-based code, wrap signal reads in the template, not in `ngOnInit`. Use `effect()` for side effects, not `computed()`.

2. **toSignal() subscribes immediately** -- `toSignal(observable$)` subscribes when the injection context is created. If the observable is cold (HTTP request), it fires immediately. Use `{ initialValue: [] }` to avoid undefined, or `{ requireSync: true }` only for synchronous observables like BehaviorSubject.

3. **Standalone components need explicit imports** -- Every pipe, directive, and component used in a standalone component's template must be listed in its `imports` array. Forgetting `CommonModule` (for `@if`, `@for` in older syntax) or `ReactiveFormsModule` silently fails.

4. **inject() only works in injection context** -- `inject(Service)` only works in constructors, field initializers, and factory functions. It throws "NG0203: inject() must be called from an injection context" if called in a method or callback. Store the injection in a field instead.

5. **Functional guards replaced class guards** -- `CanActivate`, `CanDeactivate` class guards are deprecated. Use functional guards with `inject()`. But functional guards create a new injection context -- the injected services are fresh instances unless `providedIn: 'root'`.

6. **input() signals are read-only** -- Signal inputs created with `input()` or `input.required()` cannot be `.set()` or `.update()`. They reflect the parent's binding. If you need local mutation, copy to a local signal: `localCopy = signal(this.inputValue())` and sync with `effect()`.
