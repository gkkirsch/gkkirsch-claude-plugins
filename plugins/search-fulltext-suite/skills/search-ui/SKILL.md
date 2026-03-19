---
name: search-ui
description: >
  Build search user interfaces — search bars, autocomplete, faceted filtering,
  result highlighting, pagination, and search analytics.
  Triggers: "search UI", "search component", "search bar", "autocomplete",
  "faceted search", "search results page", "search box", "typeahead".
  NOT for: backend search setup (use meilisearch-setup, elasticsearch-setup, full-text-postgres).
version: 1.0.0
argument-hint: "[component-type]"
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Search UI Components

Build polished search interfaces with React — search bars, autocomplete, faceted filtering, and result displays.

## Search Bar with Debounce

```tsx
// src/components/SearchBar.tsx
import { useState, useCallback, useRef, useEffect } from 'react';
import { Search, X, Loader2 } from 'lucide-react';

interface SearchBarProps {
  onSearch: (query: string) => void;
  placeholder?: string;
  debounceMs?: number;
  isLoading?: boolean;
}

export function SearchBar({
  onSearch,
  placeholder = 'Search...',
  debounceMs = 300,
  isLoading = false,
}: SearchBarProps) {
  const [query, setQuery] = useState('');
  const timerRef = useRef<ReturnType<typeof setTimeout>>();
  const inputRef = useRef<HTMLInputElement>(null);

  const debouncedSearch = useCallback(
    (value: string) => {
      if (timerRef.current) clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => onSearch(value), debounceMs);
    },
    [onSearch, debounceMs]
  );

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setQuery(value);
    debouncedSearch(value);
  };

  const handleClear = () => {
    setQuery('');
    onSearch('');
    inputRef.current?.focus();
  };

  // Keyboard shortcut: Ctrl/Cmd + K to focus
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        inputRef.current?.focus();
      }
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, []);

  return (
    <div className="relative w-full max-w-xl">
      <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
      <input
        ref={inputRef}
        type="text"
        value={query}
        onChange={handleChange}
        placeholder={placeholder}
        className="w-full pl-10 pr-10 py-2.5 rounded-lg border border-input bg-background
                   text-sm placeholder:text-muted-foreground
                   focus:outline-none focus:ring-2 focus:ring-ring"
      />
      {isLoading ? (
        <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 animate-spin text-muted-foreground" />
      ) : query ? (
        <button
          onClick={handleClear}
          className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
        >
          <X className="h-4 w-4" />
        </button>
      ) : (
        <kbd className="absolute right-3 top-1/2 -translate-y-1/2 hidden sm:inline-flex
                        h-5 items-center gap-1 rounded border bg-muted px-1.5
                        font-mono text-[10px] text-muted-foreground">
          <span className="text-xs">⌘</span>K
        </kbd>
      )}
    </div>
  );
}
```

## Autocomplete / Typeahead

```tsx
// src/components/Autocomplete.tsx
import { useState, useRef, useEffect, useCallback } from 'react';
import { Search, Loader2 } from 'lucide-react';

interface Suggestion {
  text: string;
  category?: string;
  type?: string;
}

interface AutocompleteProps {
  onSearch: (query: string) => void;
  getSuggestions: (query: string) => Promise<Suggestion[]>;
  placeholder?: string;
}

export function Autocomplete({ onSearch, getSuggestions, placeholder = 'Search...' }: AutocompleteProps) {
  const [query, setQuery] = useState('');
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const inputRef = useRef<HTMLInputElement>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout>>();

  const fetchSuggestions = useCallback(async (value: string) => {
    if (value.length < 2) {
      setSuggestions([]);
      setIsOpen(false);
      return;
    }
    setIsLoading(true);
    try {
      const results = await getSuggestions(value);
      setSuggestions(results);
      setIsOpen(results.length > 0);
      setSelectedIndex(-1);
    } finally {
      setIsLoading(false);
    }
  }, [getSuggestions]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setQuery(value);
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => fetchSuggestions(value), 200);
  };

  const handleSelect = (suggestion: Suggestion) => {
    setQuery(suggestion.text);
    setIsOpen(false);
    onSearch(suggestion.text);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!isOpen) {
      if (e.key === 'Enter') onSearch(query);
      return;
    }

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setSelectedIndex((prev) => Math.min(prev + 1, suggestions.length - 1));
        break;
      case 'ArrowUp':
        e.preventDefault();
        setSelectedIndex((prev) => Math.max(prev - 1, -1));
        break;
      case 'Enter':
        e.preventDefault();
        if (selectedIndex >= 0) handleSelect(suggestions[selectedIndex]);
        else onSearch(query);
        setIsOpen(false);
        break;
      case 'Escape':
        setIsOpen(false);
        break;
    }
  };

  return (
    <div className="relative w-full max-w-xl">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          onFocus={() => suggestions.length > 0 && setIsOpen(true)}
          onBlur={() => setTimeout(() => setIsOpen(false), 200)}
          placeholder={placeholder}
          role="combobox"
          aria-expanded={isOpen}
          aria-autocomplete="list"
          className="w-full pl-10 pr-10 py-2.5 rounded-lg border border-input bg-background text-sm
                     focus:outline-none focus:ring-2 focus:ring-ring"
        />
        {isLoading && (
          <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 animate-spin text-muted-foreground" />
        )}
      </div>

      {isOpen && (
        <ul
          role="listbox"
          className="absolute z-50 w-full mt-1 py-1 bg-popover border rounded-lg shadow-lg
                     max-h-60 overflow-auto"
        >
          {suggestions.map((suggestion, i) => (
            <li
              key={i}
              role="option"
              aria-selected={i === selectedIndex}
              onMouseDown={() => handleSelect(suggestion)}
              className={`px-3 py-2 cursor-pointer text-sm flex items-center justify-between
                ${i === selectedIndex ? 'bg-accent text-accent-foreground' : 'hover:bg-accent/50'}`}
            >
              <span>{suggestion.text}</span>
              {suggestion.category && (
                <span className="text-xs text-muted-foreground">{suggestion.category}</span>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
```

## Faceted Filters

```tsx
// src/components/SearchFilters.tsx
interface FacetGroup {
  name: string;
  field: string;
  buckets: Array<{ key: string; count: number }>;
}

interface SearchFiltersProps {
  facets: FacetGroup[];
  activeFilters: Record<string, string[]>;
  onFilterChange: (field: string, values: string[]) => void;
  onClearAll: () => void;
}

export function SearchFilters({ facets, activeFilters, onFilterChange, onClearAll }: SearchFiltersProps) {
  const hasActiveFilters = Object.values(activeFilters).some((v) => v.length > 0);

  const toggleFilter = (field: string, value: string) => {
    const current = activeFilters[field] || [];
    const updated = current.includes(value)
      ? current.filter((v) => v !== value)
      : [...current, value];
    onFilterChange(field, updated);
  };

  return (
    <div className="space-y-6">
      {hasActiveFilters && (
        <button
          onClick={onClearAll}
          className="text-sm text-primary hover:underline"
        >
          Clear all filters
        </button>
      )}

      {facets.map((facet) => (
        <div key={facet.field}>
          <h3 className="font-medium text-sm mb-2">{facet.name}</h3>
          <div className="space-y-1">
            {facet.buckets.map((bucket) => {
              const isActive = (activeFilters[facet.field] || []).includes(bucket.key);
              return (
                <label
                  key={bucket.key}
                  className="flex items-center gap-2 cursor-pointer text-sm py-0.5"
                >
                  <input
                    type="checkbox"
                    checked={isActive}
                    onChange={() => toggleFilter(facet.field, bucket.key)}
                    className="rounded border-input"
                  />
                  <span className={isActive ? 'font-medium' : ''}>{bucket.key}</span>
                  <span className="text-muted-foreground ml-auto text-xs">({bucket.count})</span>
                </label>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
}
```

## Search Results with Highlighting

```tsx
// src/components/SearchResults.tsx
interface SearchHit {
  id: string;
  title: string;
  description: string;
  imageUrl?: string;
  price?: number;
  rating?: number;
  highlight?: {
    title?: string[];
    description?: string[];
  };
}

interface SearchResultsProps {
  hits: SearchHit[];
  total: number;
  query: string;
  isLoading: boolean;
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export function SearchResults({
  hits, total, query, isLoading, page, totalPages, onPageChange,
}: SearchResultsProps) {
  if (isLoading) {
    return (
      <div className="space-y-4">
        {[...Array(5)].map((_, i) => (
          <div key={i} className="animate-pulse flex gap-4">
            <div className="h-24 w-24 bg-muted rounded" />
            <div className="flex-1 space-y-2">
              <div className="h-4 bg-muted rounded w-3/4" />
              <div className="h-3 bg-muted rounded w-1/2" />
              <div className="h-3 bg-muted rounded w-full" />
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (query && hits.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-lg font-medium">No results found</p>
        <p className="text-muted-foreground mt-1">
          Try different keywords or remove some filters
        </p>
      </div>
    );
  }

  return (
    <div>
      {query && (
        <p className="text-sm text-muted-foreground mb-4">
          {total.toLocaleString()} results for "{query}"
        </p>
      )}

      <div className="space-y-4">
        {hits.map((hit) => (
          <article key={hit.id} className="flex gap-4 p-4 rounded-lg border hover:bg-accent/30 transition-colors">
            {hit.imageUrl && (
              <img
                src={hit.imageUrl}
                alt={hit.title}
                className="h-24 w-24 object-cover rounded"
                loading="lazy"
              />
            )}
            <div className="flex-1 min-w-0">
              <h3
                className="font-medium text-base"
                dangerouslySetInnerHTML={{
                  __html: hit.highlight?.title?.[0] || hit.title,
                }}
              />
              <p
                className="text-sm text-muted-foreground mt-1 line-clamp-2"
                dangerouslySetInnerHTML={{
                  __html: hit.highlight?.description?.[0] || hit.description,
                }}
              />
              {hit.price !== undefined && (
                <p className="text-sm font-medium mt-2">
                  ${(hit.price / 100).toFixed(2)}
                </p>
              )}
            </div>
          </article>
        ))}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <nav className="flex justify-center gap-1 mt-8">
          <button
            onClick={() => onPageChange(page - 1)}
            disabled={page <= 1}
            className="px-3 py-1.5 text-sm rounded border disabled:opacity-50 hover:bg-accent"
          >
            Previous
          </button>
          {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
            let pageNum: number;
            if (totalPages <= 7) pageNum = i + 1;
            else if (page <= 4) pageNum = i + 1;
            else if (page >= totalPages - 3) pageNum = totalPages - 6 + i;
            else pageNum = page - 3 + i;

            return (
              <button
                key={pageNum}
                onClick={() => onPageChange(pageNum)}
                className={`px-3 py-1.5 text-sm rounded border
                  ${pageNum === page ? 'bg-primary text-primary-foreground' : 'hover:bg-accent'}`}
              >
                {pageNum}
              </button>
            );
          })}
          <button
            onClick={() => onPageChange(page + 1)}
            disabled={page >= totalPages}
            className="px-3 py-1.5 text-sm rounded border disabled:opacity-50 hover:bg-accent"
          >
            Next
          </button>
        </nav>
      )}
    </div>
  );
}
```

## Complete Search Page

```tsx
// src/pages/SearchPage.tsx
import { useState, useCallback } from 'react';
import { SearchBar } from '../components/SearchBar';
import { SearchFilters } from '../components/SearchFilters';
import { SearchResults } from '../components/SearchResults';

export function SearchPage() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [filters, setFilters] = useState<Record<string, string[]>>({});
  const [page, setPage] = useState(1);
  const [sort, setSort] = useState('');

  const performSearch = useCallback(async (q: string, f = filters, p = 1, s = sort) => {
    setIsLoading(true);
    try {
      const params = new URLSearchParams();
      if (q) params.set('q', q);
      if (p > 1) params.set('page', String(p));
      if (s) params.set('sort', s);

      // Add active filters
      for (const [field, values] of Object.entries(f)) {
        for (const value of values) {
          params.append(field, value);
        }
      }

      const response = await fetch(`/api/search?${params}`);
      const data = await response.json();
      setResults(data);
    } finally {
      setIsLoading(false);
    }
  }, [filters, sort]);

  const handleSearch = (q: string) => {
    setQuery(q);
    setPage(1);
    performSearch(q, filters, 1);
  };

  const handleFilterChange = (field: string, values: string[]) => {
    const newFilters = { ...filters, [field]: values };
    setFilters(newFilters);
    setPage(1);
    performSearch(query, newFilters, 1);
  };

  const handlePageChange = (p: number) => {
    setPage(p);
    performSearch(query, filters, p);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="mb-8">
        <SearchBar onSearch={handleSearch} isLoading={isLoading} />
      </div>

      <div className="flex gap-8">
        {/* Sidebar filters */}
        {results?.facets && (
          <aside className="w-64 shrink-0 hidden lg:block">
            <SearchFilters
              facets={[
                { name: 'Category', field: 'category', buckets: results.facets.categories?.buckets || [] },
                { name: 'Brand', field: 'brand', buckets: results.facets.brands?.buckets || [] },
              ]}
              activeFilters={filters}
              onFilterChange={handleFilterChange}
              onClearAll={() => { setFilters({}); performSearch(query, {}, 1); }}
            />
          </aside>
        )}

        {/* Results */}
        <main className="flex-1 min-w-0">
          {/* Sort */}
          <div className="flex justify-between items-center mb-4">
            <div />
            <select
              value={sort}
              onChange={(e) => { setSort(e.target.value); performSearch(query, filters, 1, e.target.value); }}
              className="text-sm border rounded-md px-2 py-1.5"
            >
              <option value="">Relevance</option>
              <option value="price_asc">Price: Low to High</option>
              <option value="price_desc">Price: High to Low</option>
              <option value="rating">Highest Rated</option>
              <option value="newest">Newest</option>
            </select>
          </div>

          <SearchResults
            hits={results?.hits || []}
            total={results?.total || 0}
            query={query}
            isLoading={isLoading}
            page={page}
            totalPages={Math.ceil((results?.total || 0) / 20)}
            onPageChange={handlePageChange}
          />
        </main>
      </div>
    </div>
  );
}
```

## Search Analytics Hook

```typescript
// src/hooks/useSearchAnalytics.ts
export function useSearchAnalytics() {
  const trackSearch = useCallback((query: string, resultCount: number) => {
    // Track what users search for
    fetch('/api/analytics/search', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        query,
        resultCount,
        timestamp: new Date().toISOString(),
      }),
    }).catch(() => {}); // Fire and forget
  }, []);

  const trackClick = useCallback((query: string, resultId: string, position: number) => {
    // Track which results get clicked
    fetch('/api/analytics/click', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ query, resultId, position }),
    }).catch(() => {});
  }, []);

  return { trackSearch, trackClick };
}
```

## Gotchas

- Always debounce search input (300ms minimum) to avoid hammering the API
- Set `loading="lazy"` on result images below the fold
- Use `dangerouslySetInnerHTML` carefully with highlights — sanitize HTML from the search API if user-generated content is indexed
- Keyboard navigation (arrow keys, Enter, Escape) is essential for autocomplete accessibility
- Show loading skeletons, not spinners — they feel faster and prevent layout shifts
- URL state: sync search query and filters with URL params so search results are shareable and back-button works
- Empty state matters — show helpful suggestions when no results match ("Try different keywords" or popular searches)
- Mobile: replace sidebar facets with a bottom sheet or expandable section
