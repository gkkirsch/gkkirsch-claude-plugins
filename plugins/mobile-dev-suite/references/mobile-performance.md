# Mobile Performance Reference

Production patterns for React Native performance optimization. This reference covers JS thread optimization, FlatList/FlashList tuning, image optimization, bundle size reduction, startup time improvement, and memory management.

---

## 1. JS Thread Optimization

### Understanding the Threading Model

React Native uses multiple threads:
- **JS Thread**: Runs your JavaScript/TypeScript code, React reconciliation
- **UI Thread (Main)**: Handles native rendering, touch events, layout
- **Shadow Thread**: Calculates Yoga layouts (being merged with UI in New Architecture)
- **Native Modules Thread**: Runs native module operations

The JS thread is the most common bottleneck. Keep it under 16ms per frame (60fps).

### Offloading Heavy Computation

```typescript
// BAD: Heavy computation on JS thread blocks UI
function processLargeDataset(data: Item[]) {
  // This blocks the JS thread
  return data
    .filter(item => complexFilter(item))
    .sort((a, b) => complexSort(a, b))
    .map(item => transformItem(item));
}

// GOOD: Use InteractionManager to defer non-urgent work
import { InteractionManager } from 'react-native';

function DeferredComponent() {
  const [data, setData] = useState<Item[]>([]);
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    // Wait for animations/transitions to complete
    const task = InteractionManager.runAfterInteractions(() => {
      const processed = processLargeDataset(rawData);
      setData(processed);
      setIsReady(true);
    });

    return () => task.cancel();
  }, [rawData]);

  if (!isReady) return <LoadingSkeleton />;
  return <DataList data={data} />;
}
```

### Worklets with Reanimated (Off JS Thread)

```typescript
// Run animations on the UI thread — zero JS thread overhead
import Animated, {
  useSharedValue,
  useAnimatedStyle,
  withSpring,
  runOnUI,
} from 'react-native-reanimated';

function AnimatedCard() {
  const scale = useSharedValue(1);

  // This runs entirely on the UI thread
  const animatedStyle = useAnimatedStyle(() => ({
    transform: [{ scale: withSpring(scale.value) }],
  }));

  const onPress = () => {
    // Direct UI thread update — no bridge crossing
    scale.value = 0.95;
    setTimeout(() => { scale.value = 1; }, 100);
  };

  return (
    <Animated.View style={animatedStyle}>
      <Pressable onPress={onPress}>
        <Text>Animated Card</Text>
      </Pressable>
    </Animated.View>
  );
}
```

### Avoiding Re-renders

```typescript
// BAD: Inline objects/functions cause re-renders
function ParentComponent() {
  return (
    <ChildComponent
      style={{ marginTop: 10 }}              // New object every render
      onPress={() => handlePress()}           // New function every render
      config={{ theme: 'dark', size: 'lg' }} // New object every render
    />
  );
}

// GOOD: Stable references
const styles = StyleSheet.create({
  child: { marginTop: 10 },
});

const CONFIG = { theme: 'dark', size: 'lg' } as const;

function ParentComponent() {
  const handlePress = useCallback(() => {
    // handler logic
  }, []);

  return (
    <ChildComponent
      style={styles.child}
      onPress={handlePress}
      config={CONFIG}
    />
  );
}

// GOOD: Use React.memo for expensive child components
const ChildComponent = React.memo(function ChildComponent({ data, onPress }: Props) {
  return (
    <Pressable onPress={onPress}>
      <ExpensiveContent data={data} />
    </Pressable>
  );
});
```

### useMemo for Derived Data

```typescript
function FilteredList({ items, searchQuery, sortBy }: Props) {
  // Recompute only when dependencies change
  const filteredItems = useMemo(() => {
    let result = items;

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      result = result.filter(
        item =>
          item.title.toLowerCase().includes(query) ||
          item.description.toLowerCase().includes(query)
      );
    }

    result.sort((a, b) => {
      switch (sortBy) {
        case 'date': return b.createdAt - a.createdAt;
        case 'name': return a.title.localeCompare(b.title);
        case 'priority': return b.priority - a.priority;
        default: return 0;
      }
    });

    return result;
  }, [items, searchQuery, sortBy]);

  return <FlatList data={filteredItems} renderItem={renderItem} />;
}
```

---

## 2. FlatList / FlashList Optimization

### FlatList Best Practices

```typescript
function OptimizedFlatList({ data }: { data: Item[] }) {
  // Extract key outside render
  const keyExtractor = useCallback((item: Item) => item.id, []);

  // Memoize renderItem
  const renderItem = useCallback(({ item }: { item: Item }) => (
    <MemoizedListItem item={item} />
  ), []);

  // Memoize separators
  const ItemSeparator = useCallback(() => <View style={styles.separator} />, []);

  // Estimate item height for better scroll performance
  const getItemLayout = useCallback((_: any, index: number) => ({
    length: ITEM_HEIGHT,
    offset: ITEM_HEIGHT * index,
    index,
  }), []);

  return (
    <FlatList
      data={data}
      renderItem={renderItem}
      keyExtractor={keyExtractor}
      ItemSeparatorComponent={ItemSeparator}
      getItemLayout={getItemLayout}
      // Performance props
      removeClippedSubviews={true}          // Unmount off-screen items (Android)
      maxToRenderPerBatch={10}               // Items per batch render
      updateCellsBatchingPeriod={50}         // Ms between batch renders
      windowSize={5}                         // Render 5 screens worth of items
      initialNumToRender={10}                // Items to render initially
      // Avoid unnecessary re-renders
      extraData={undefined}                  // Only pass if list needs external state
    />
  );
}

const MemoizedListItem = React.memo(function ListItem({ item }: { item: Item }) {
  return (
    <View style={styles.item}>
      <FastImage source={{ uri: item.imageUrl }} style={styles.image} />
      <Text style={styles.title}>{item.title}</Text>
    </View>
  );
});
```

### FlashList (Shopify) — Superior Performance

```typescript
// FlashList is a drop-in replacement with significantly better performance
import { FlashList } from '@shopify/flash-list';

function HighPerformanceList({ data }: { data: Item[] }) {
  const renderItem = useCallback(({ item }: { item: Item }) => (
    <ListItem item={item} />
  ), []);

  return (
    <FlashList
      data={data}
      renderItem={renderItem}
      estimatedItemSize={80}           // REQUIRED: estimated height of items
      keyExtractor={(item) => item.id}
      // FlashList-specific optimizations
      drawDistance={250}                // How far ahead to render (pixels)
      overrideItemLayout={(layout, item, index, maxColumns, extraData) => {
        // For variable-height items, provide exact heights
        layout.size = item.type === 'header' ? 120 : 80;
      }}
    />
  );
}

// FlashList with multiple item types
function MixedList({ data }: { data: (Post | Ad | Header)[] }) {
  return (
    <FlashList
      data={data}
      renderItem={({ item }) => {
        switch (item.type) {
          case 'post': return <PostCard post={item} />;
          case 'ad': return <AdBanner ad={item} />;
          case 'header': return <SectionHeader header={item} />;
        }
      }}
      getItemType={(item) => item.type}  // Critical for mixed lists
      estimatedItemSize={100}
    />
  );
}
```

### SectionList Optimization

```typescript
function OptimizedSectionList({ sections }: Props) {
  return (
    <SectionList
      sections={sections}
      renderItem={({ item }) => <MemoizedItem item={item} />}
      renderSectionHeader={({ section }) => (
        <View style={styles.sectionHeader}>
          <Text style={styles.sectionTitle}>{section.title}</Text>
        </View>
      )}
      stickySectionHeadersEnabled={true}
      // Same performance props as FlatList
      removeClippedSubviews={true}
      maxToRenderPerBatch={10}
      windowSize={5}
      initialNumToRender={10}
    />
  );
}
```

---

## 3. Image Optimization

### expo-image (Recommended)

```typescript
import { Image } from 'expo-image';

// High-performance image with caching, blurhash, transitions
function OptimizedImage({ uri, blurhash }: { uri: string; blurhash?: string }) {
  return (
    <Image
      source={{ uri }}
      placeholder={blurhash ? { blurhash } : undefined}
      contentFit="cover"
      transition={200}
      style={styles.image}
      // Caching
      cachePolicy="memory-disk"
      recyclingKey={uri}
      // Priority
      priority="normal" // "low" | "normal" | "high"
    />
  );
}

// Prefetch images
import { Image } from 'expo-image';

async function prefetchImages(urls: string[]) {
  await Promise.all(urls.map(url => Image.prefetch(url)));
}
```

### react-native-fast-image Alternative

```typescript
import FastImage from 'react-native-fast-image';

function CachedImage({ uri }: { uri: string }) {
  return (
    <FastImage
      source={{
        uri,
        priority: FastImage.priority.normal,
        cache: FastImage.cacheControl.immutable,
      }}
      resizeMode={FastImage.resizeMode.cover}
      style={styles.image}
    />
  );
}

// Preload critical images
FastImage.preload([
  { uri: 'https://example.com/hero.jpg' },
  { uri: 'https://example.com/profile.jpg' },
]);
```

### Image Size Guidelines

| Context | Resolution | Format | Max Size |
|---------|-----------|--------|----------|
| Thumbnail (list) | 150x150 | WebP | 15KB |
| Card image | 400x300 | WebP | 50KB |
| Full-width | 750xAuto | WebP | 100KB |
| Hero/splash | 1125xAuto | WebP/PNG | 200KB |
| Avatar | 100x100 | WebP | 10KB |

### Responsive Images

```typescript
import { PixelRatio, Dimensions } from 'react-native';

function getImageUrl(baseUrl: string, width: number): string {
  const pixelWidth = PixelRatio.getPixelSizeForLayoutSize(width);
  // Round up to nearest supported size
  const sizes = [150, 300, 600, 900, 1200];
  const targetSize = sizes.find(s => s >= pixelWidth) ?? sizes[sizes.length - 1];
  return `${baseUrl}?w=${targetSize}&format=webp&quality=80`;
}
```

---

## 4. Bundle Size Optimization

### Analyzing Bundle Size

```bash
# React Native bundle analysis
npx react-native-bundle-visualizer

# Expo bundle analysis
npx expo export --dump-sourcemap
npx source-map-explorer dist/_expo/static/js/web/*.js

# Metro bundle output
npx react-native bundle \
  --platform ios \
  --dev false \
  --entry-file index.js \
  --bundle-output bundle.js \
  --sourcemap-output bundle.js.map
```

### Common Bundle Bloat Fixes

```typescript
// BAD: Importing entire libraries
import _ from 'lodash';              // ~72KB
import moment from 'moment';         // ~67KB
import { format } from 'date-fns';   // Tree-shakeable but still large

// GOOD: Cherry-pick imports
import groupBy from 'lodash/groupBy';  // ~2KB
import dayjs from 'dayjs';             // ~2KB

// BAD: Large icon libraries
import Icon from 'react-native-vector-icons/MaterialIcons'; // All icons bundled

// GOOD: Use specific icon sets or lucide
import { Home, Settings, User } from 'lucide-react-native'; // Tree-shaken
```

### Lazy Loading Screens

```typescript
// React.lazy for screens loaded on demand
import { lazy, Suspense } from 'react';

const SettingsScreen = lazy(() => import('./screens/SettingsScreen'));
const ProfileEditScreen = lazy(() => import('./screens/ProfileEditScreen'));

// Wrap in Suspense
function AppNavigator() {
  return (
    <Stack.Navigator>
      <Stack.Screen name="Home" component={HomeScreen} />
      <Stack.Screen name="Settings">
        {() => (
          <Suspense fallback={<LoadingScreen />}>
            <SettingsScreen />
          </Suspense>
        )}
      </Stack.Screen>
    </Stack.Navigator>
  );
}
```

### Hermes Engine Optimization

```json
// app.json — Hermes is default in Expo SDK 47+
{
  "expo": {
    "jsEngine": "hermes"
  }
}

// Hermes benefits:
// - Bytecode precompilation → faster startup
// - Reduced memory usage
// - Improved garbage collection
// - Better CPU performance
```

### ProGuard (Android) & Code Stripping (iOS)

```gradle
// android/app/build.gradle
android {
  buildTypes {
    release {
      minifyEnabled true
      shrinkResources true
      proguardFiles getDefaultProguardFile('proguard-android-optimize.txt'), 'proguard-rules.pro'
    }
  }
}
```

---

## 5. Startup Time Optimization

### Measuring Startup Time

```typescript
// utils/startup-timer.ts
import { PerformanceObserver, performance } from 'react-native-performance';

class StartupTimer {
  private marks: Record<string, number> = {};

  mark(name: string) {
    this.marks[name] = performance.now();
  }

  measure(name: string, startMark: string, endMark: string) {
    const start = this.marks[startMark];
    const end = this.marks[endMark];
    if (start && end) {
      console.log(`[Startup] ${name}: ${(end - start).toFixed(1)}ms`);
      analytics.track('startup_timing', { phase: name, durationMs: end - start });
    }
  }

  reportAll() {
    this.measure('JS Bundle Load', 'app_start', 'bundle_loaded');
    this.measure('First Render', 'bundle_loaded', 'first_render');
    this.measure('Navigation Ready', 'first_render', 'navigation_ready');
    this.measure('Data Loaded', 'navigation_ready', 'data_loaded');
    this.measure('Total', 'app_start', 'data_loaded');
  }
}

export const startupTimer = new StartupTimer();

// index.js
startupTimer.mark('app_start');

// App.tsx root
function App() {
  useEffect(() => {
    startupTimer.mark('first_render');
  }, []);
}
```

### Reducing Startup Work

```typescript
// BAD: Everything initializes at startup
import { initAnalytics } from './analytics';
import { initCrashReporting } from './crash-reporting';
import { syncDatabase } from './database';
import { prefetchFeed } from './api';

// All run synchronously before first render
initAnalytics();
initCrashReporting();
syncDatabase();
prefetchFeed();

// GOOD: Prioritize critical path, defer the rest
function App() {
  useEffect(() => {
    // Critical: error reporting (fast)
    initCrashReporting();

    // After first render: analytics
    InteractionManager.runAfterInteractions(() => {
      initAnalytics();
    });

    // After app is interactive: background sync
    setTimeout(() => {
      syncDatabase();
      prefetchFeed();
    }, 2000);
  }, []);
}
```

### Splash Screen Strategy

```typescript
// app/_layout.tsx (Expo Router)
import * as SplashScreen from 'expo-splash-screen';
import { useFonts } from 'expo-font';

// Keep splash visible while loading
SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  const [fontsLoaded] = useFonts({
    'Inter-Regular': require('../assets/fonts/Inter-Regular.ttf'),
    'Inter-Bold': require('../assets/fonts/Inter-Bold.ttf'),
  });

  const { isLoaded: authLoaded } = useAuthStore();

  useEffect(() => {
    if (fontsLoaded && authLoaded) {
      // Hide splash only when critical resources are ready
      SplashScreen.hideAsync();
    }
  }, [fontsLoaded, authLoaded]);

  if (!fontsLoaded || !authLoaded) return null;

  return (
    <Providers>
      <Stack />
    </Providers>
  );
}
```

---

## 6. Memory Management

### Detecting Memory Leaks

```typescript
// Common memory leak patterns in React Native

// LEAK: Unsubscribed listeners
useEffect(() => {
  const subscription = EventEmitter.addListener('event', handler);
  // Missing cleanup!
}, []);

// FIX: Always clean up
useEffect(() => {
  const subscription = EventEmitter.addListener('event', handler);
  return () => subscription.remove();
}, []);

// LEAK: Timers not cleared
useEffect(() => {
  setInterval(pollData, 5000);
  // Missing cleanup!
}, []);

// FIX: Clear timers
useEffect(() => {
  const intervalId = setInterval(pollData, 5000);
  return () => clearInterval(intervalId);
}, []);

// LEAK: Unmounted component state updates
useEffect(() => {
  fetchData().then(data => setData(data)); // May set state after unmount
}, []);

// FIX: Check mounted state or use AbortController
useEffect(() => {
  const controller = new AbortController();

  fetchData({ signal: controller.signal })
    .then(data => setData(data))
    .catch(err => {
      if (err.name !== 'AbortError') throw err;
    });

  return () => controller.abort();
}, []);
```

### Image Memory Management

```typescript
// Clear image cache periodically
import { Image } from 'expo-image';

async function clearImageCache() {
  await Image.clearDiskCache();
  await Image.clearMemoryCache();
}

// Monitor memory usage
import { NativeModules } from 'react-native';

function logMemoryUsage() {
  if (__DEV__) {
    // iOS: use Xcode Instruments
    // Android: use Android Studio Memory Profiler
    console.log('Memory logging — use native tools for accurate measurements');
  }
}
```

### Large List Memory

```typescript
// For very large datasets, use windowed rendering
function VirtualizedLargeList({ data }: { data: Item[] }) {
  return (
    <FlashList
      data={data}
      renderItem={({ item }) => <ListItem item={item} />}
      estimatedItemSize={80}
      // Aggressive memory settings for 10k+ items
      drawDistance={100}          // Render less ahead
      // Unmount items far from viewport
    />
  );
}
```

---

## Quick Reference: Performance Checklist

```
Before release, verify:

JS Thread Performance
  [ ] No synchronous heavy computation in render
  [ ] InteractionManager used for deferred work
  [ ] Animations run on UI thread (Reanimated)
  [ ] React.memo on expensive list items
  [ ] useMemo for derived data, useCallback for handlers

List Performance
  [ ] FlashList or optimized FlatList
  [ ] keyExtractor uses stable IDs
  [ ] renderItem is memoized
  [ ] getItemLayout provided for fixed-height items
  [ ] removeClippedSubviews on Android

Images
  [ ] expo-image or FastImage (not <Image>)
  [ ] Proper caching strategy
  [ ] Responsive image sizes
  [ ] WebP format where possible
  [ ] Blurhash or placeholder

Bundle Size
  [ ] Tree-shaken imports (no lodash/moment full imports)
  [ ] Lazy-loaded screens
  [ ] Hermes engine enabled
  [ ] ProGuard/R8 enabled for Android release

Startup Time
  [ ] Splash screen covers loading
  [ ] Critical path minimal (auth check, fonts)
  [ ] Non-critical init deferred
  [ ] No unnecessary re-renders at startup

Memory
  [ ] All subscriptions cleaned up
  [ ] All timers cleared on unmount
  [ ] AbortController for async operations
  [ ] Image cache management
```
