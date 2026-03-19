# ML Algorithms Guide — Algorithm Selection and When to Use What

Quick reference for selecting the right machine learning algorithm based on your problem, data characteristics, and requirements. Use this when advising on model selection or building ML pipelines.

---

## Algorithm Selection Flowchart

```
Start
  ├── Is it supervised?
  │     ├── Classification
  │     │     ├── Binary → Logistic Regression → Random Forest → XGBoost/LightGBM
  │     │     ├── Multi-class → Logistic (OvR) → Random Forest → XGBoost
  │     │     ├── Multi-label → Binary Relevance → Classifier Chains
  │     │     └── Imbalanced → Class weights → SMOTE → Threshold tuning
  │     ├── Regression
  │     │     ├── Linear relationship → Ridge/Lasso → ElasticNet
  │     │     ├── Non-linear → Random Forest → XGBoost/LightGBM
  │     │     └── Time series → ARIMA → Prophet → LightGBM with lag features
  │     └── Ranking
  │           └── LambdaMART (XGBoost rank) → LightGBM rank
  └── Is it unsupervised?
        ├── Clustering
        │     ├── Known K → K-Means / K-Medoids
        │     ├── Unknown K → DBSCAN / HDBSCAN
        │     └── Hierarchical → Agglomerative Clustering
        ├── Dimensionality Reduction
        │     ├── Linear → PCA
        │     ├── Non-linear → t-SNE (visualization) / UMAP
        │     └── Feature selection → Mutual Information / L1 regularization
        └── Anomaly Detection
              ├── Isolation Forest (general purpose)
              ├── Local Outlier Factor (density-based)
              └── One-Class SVM (boundary-based)
```

---

## Classification Algorithms

### Logistic Regression

**When to use:**
- Baseline for any classification task
- When interpretability is important
- When you need well-calibrated probabilities
- Linear decision boundary is sufficient
- Small to medium datasets

**When NOT to use:**
- Complex non-linear relationships
- Many feature interactions needed
- Very high-dimensional sparse data (prefer SGDClassifier)

**Key parameters:**
```python
from sklearn.linear_model import LogisticRegression

model = LogisticRegression(
    C=1.0,                    # Regularization strength (smaller = stronger)
    penalty="l2",             # 'l1', 'l2', 'elasticnet', None
    class_weight="balanced",  # Handle imbalanced classes
    max_iter=1000,            # Increase if not converging
    solver="lbfgs",           # 'lbfgs', 'saga' (for l1/elasticnet), 'liblinear'
    random_state=42,
)
```

**Tuning guide:**
- `C`: Start at 1.0, try [0.001, 0.01, 0.1, 1, 10, 100]
- `penalty`: Try l1 for feature selection, l2 for general use
- If not converging: increase `max_iter`, try different `solver`

---

### Random Forest

**When to use:**
- Strong baseline that works well out-of-the-box
- Need feature importance
- Robust to outliers and noisy features
- Can handle mixed feature types
- Don't need extensive tuning

**When NOT to use:**
- Need fastest possible inference
- Very high-dimensional sparse data
- Extrapolation required (forests can't extrapolate)
- Very large datasets (slower than boosting)

**Key parameters:**
```python
from sklearn.ensemble import RandomForestClassifier

model = RandomForestClassifier(
    n_estimators=200,         # More trees = better, diminishing returns after ~200
    max_depth=10,             # None for full depth (risk of overfitting)
    min_samples_leaf=5,       # Prevents overfitting
    min_samples_split=10,     # Minimum samples to split a node
    max_features="sqrt",      # 'sqrt', 'log2', float fraction
    class_weight="balanced",  # Handle imbalanced classes
    n_jobs=-1,                # Parallel training
    random_state=42,
)
```

**Tuning guide:**
- `n_estimators`: 100-500, more is usually better
- `max_depth`: Start None, reduce if overfitting (try 5-20)
- `min_samples_leaf`: 1-20, increase to reduce overfitting
- `max_features`: sqrt is default for classification, try 0.3-0.8

---

### XGBoost

**When to use:**
- State-of-the-art for tabular data
- Kaggle competitions (historically dominant)
- When you need the best possible performance
- Handles missing values natively
- Can handle large datasets efficiently

**When NOT to use:**
- Small datasets (<1000 rows) — may overfit
- Need simple interpretable model
- Need to extrapolate beyond training range
- Real-time inference with strict latency (<1ms)

**Key parameters:**
```python
import xgboost as xgb

params = {
    "objective": "binary:logistic",  # 'multi:softmax', 'reg:squarederror'
    "eval_metric": "auc",            # 'logloss', 'rmse', 'mae'
    "max_depth": 6,                  # 3-10, deeper = more complex
    "learning_rate": 0.1,            # 0.01-0.3, lower with more trees
    "n_estimators": 200,             # With early stopping, set high
    "subsample": 0.8,               # Row sampling, prevents overfitting
    "colsample_bytree": 0.8,        # Feature sampling per tree
    "min_child_weight": 5,           # Minimum sum of instance weight in child
    "reg_alpha": 0.1,               # L1 regularization
    "reg_lambda": 1.0,              # L2 regularization
    "scale_pos_weight": 1.0,        # Set to neg/pos ratio for imbalanced
    "tree_method": "hist",           # Fast histogram-based method
    "early_stopping_rounds": 50,     # Stop if no improvement
}
```

**Tuning priority (in order):**
1. `max_depth` + `min_child_weight` (complexity control)
2. `subsample` + `colsample_bytree` (randomization)
3. `learning_rate` + `n_estimators` (lower LR, more trees)
4. `reg_alpha` + `reg_lambda` (regularization)
5. `gamma` (minimum loss reduction for split)

---

### LightGBM

**When to use:**
- Very large datasets (fastest gradient boosting)
- Need native categorical feature support
- Memory constrained (more memory efficient than XGBoost)
- Distributed training needed
- Similar or better accuracy to XGBoost with less tuning

**When NOT to use:**
- Very small datasets (<1000 rows) — more prone to overfitting
- Need model interpretability
- Need deterministic results (slight non-determinism)

**Key parameters:**
```python
import lightgbm as lgb

params = {
    "objective": "binary",           # 'multiclass', 'regression'
    "metric": "auc",                 # 'binary_logloss', 'rmse'
    "max_depth": 7,                  # -1 for no limit
    "num_leaves": 63,                # 2^max_depth - 1
    "learning_rate": 0.05,           # 0.01-0.3
    "subsample": 0.8,               # 'bagging_fraction'
    "colsample_bytree": 0.8,        # 'feature_fraction'
    "min_child_samples": 20,         # Minimum data in leaf
    "reg_alpha": 0.1,               # L1 regularization
    "reg_lambda": 1.0,              # L2 regularization
    "is_unbalance": True,           # For imbalanced data
    "verbose": -1,
}
```

**LightGBM vs XGBoost:**
- LightGBM is generally faster (leaf-wise vs level-wise growth)
- LightGBM handles categoricals natively (no one-hot needed)
- XGBoost is more stable, better for small datasets
- Both achieve similar accuracy on most tasks

---

### CatBoost

**When to use:**
- Many categorical features (best native handling)
- Don't want to tune hyperparameters much
- Need good out-of-the-box performance
- Ordered boosting to reduce overfitting

**Key parameters:**
```python
from catboost import CatBoostClassifier

model = CatBoostClassifier(
    iterations=500,
    depth=6,
    learning_rate=0.1,
    l2_leaf_reg=3,
    cat_features=categorical_indices,  # Specify categorical columns
    auto_class_weights="Balanced",
    random_seed=42,
    verbose=100,
)
```

---

### Support Vector Machines (SVM)

**When to use:**
- Small to medium datasets
- High-dimensional data
- Clear margin of separation between classes
- When kernel trick is needed (non-linear boundaries)

**When NOT to use:**
- Large datasets (>100K rows) — doesn't scale well
- Many features with noise
- Need probability estimates (slow calibration)
- Need fast training

```python
from sklearn.svm import SVC

model = SVC(
    C=1.0,                  # Regularization
    kernel="rbf",           # 'linear', 'poly', 'rbf', 'sigmoid'
    gamma="scale",          # Kernel coefficient
    class_weight="balanced",
    probability=True,       # Enable probability estimates (slower)
)
```

---

### K-Nearest Neighbors (KNN)

**When to use:**
- Small datasets
- Non-parametric baseline
- Recommendation systems (item similarity)
- Anomaly detection

**When NOT to use:**
- Large datasets (slow inference)
- High-dimensional data (curse of dimensionality)
- Need interpretable model

```python
from sklearn.neighbors import KNeighborsClassifier

model = KNeighborsClassifier(
    n_neighbors=5,
    weights="distance",     # 'uniform' or 'distance'
    metric="euclidean",     # 'manhattan', 'cosine'
    n_jobs=-1,
)
```

---

## Regression Algorithms

### Linear Regression Variants

| Algorithm | Regularization | When to Use |
|-----------|---------------|-------------|
| LinearRegression | None | Baseline, small datasets |
| Ridge | L2 | Multicollinearity, many features |
| Lasso | L1 | Feature selection, sparse solutions |
| ElasticNet | L1 + L2 | Best of both, correlated features |

```python
from sklearn.linear_model import Ridge, Lasso, ElasticNet

# Ridge: keeps all features, shrinks coefficients
ridge = Ridge(alpha=1.0)

# Lasso: sets some coefficients to zero (feature selection)
lasso = Lasso(alpha=1.0, max_iter=5000)

# ElasticNet: mix of L1 and L2
elasticnet = ElasticNet(alpha=1.0, l1_ratio=0.5, max_iter=5000)
```

**Tuning guide:**
- `alpha`: [0.001, 0.01, 0.1, 1.0, 10, 100] — higher = more regularization
- `l1_ratio` (ElasticNet): 0 = Ridge, 1 = Lasso, try [0.1, 0.3, 0.5, 0.7, 0.9]

---

### Gradient Boosting Regression

Same algorithms as classification but with regression objectives:

```python
# XGBoost regression
xgb_params = {"objective": "reg:squarederror", "eval_metric": "rmse"}

# LightGBM regression
lgb_params = {"objective": "regression", "metric": "rmse"}

# Huber loss for robustness to outliers
xgb_params = {"objective": "reg:pseudohubererror"}
lgb_params = {"objective": "huber"}
```

---

## Clustering Algorithms

### K-Means

**When to use:** Known number of clusters, spherical clusters, large datasets
**When NOT:** Non-spherical clusters, unknown K, many outliers

```python
from sklearn.cluster import KMeans

model = KMeans(
    n_clusters=5,
    init="k-means++",      # Smart initialization
    n_init=10,              # Number of restarts
    max_iter=300,
    random_state=42,
)
```

### DBSCAN

**When to use:** Unknown number of clusters, arbitrary shapes, noise detection
**When NOT:** Varying density clusters, high-dimensional data, need to predict new points

```python
from sklearn.cluster import DBSCAN

model = DBSCAN(
    eps=0.5,              # Maximum distance between neighbors
    min_samples=5,        # Minimum points to form cluster
    metric="euclidean",
)
```

### HDBSCAN

**When to use:** Unknown K, varying density, better than DBSCAN for most cases
**When NOT:** Very large datasets, need deterministic results

```python
import hdbscan

model = hdbscan.HDBSCAN(
    min_cluster_size=15,   # Minimum cluster size
    min_samples=5,         # Core point threshold
    metric="euclidean",
)
```

---

## Anomaly Detection

| Algorithm | Approach | Best For |
|-----------|----------|----------|
| Isolation Forest | Tree-based isolation | General purpose, high-dimensional |
| Local Outlier Factor | Density-based | Local anomalies in dense regions |
| One-Class SVM | Boundary-based | When normal data is well-defined |
| Elliptic Envelope | Gaussian fit | Normally distributed features |

```python
from sklearn.ensemble import IsolationForest
from sklearn.neighbors import LocalOutlierFactor

# Isolation Forest (most commonly used)
iforest = IsolationForest(
    contamination=0.05,     # Expected fraction of anomalies
    n_estimators=200,
    random_state=42,
)
labels = iforest.fit_predict(X)  # -1 = anomaly, 1 = normal

# LOF (good for local anomalies)
lof = LocalOutlierFactor(
    n_neighbors=20,
    contamination=0.05,
)
labels = lof.fit_predict(X)
```

---

## Dimensionality Reduction

| Algorithm | Type | Best For |
|-----------|------|----------|
| PCA | Linear | Feature reduction, noise removal |
| t-SNE | Non-linear | 2D/3D visualization |
| UMAP | Non-linear | Visualization + preserves global structure |
| Truncated SVD | Linear | Sparse data (TF-IDF) |
| LDA | Supervised | Classification with dim reduction |

```python
from sklearn.decomposition import PCA, TruncatedSVD
from sklearn.manifold import TSNE

# PCA: retain 95% variance
pca = PCA(n_components=0.95, random_state=42)
X_pca = pca.fit_transform(X_scaled)
print(f"Components: {pca.n_components_}, Variance: {pca.explained_variance_ratio_.sum():.2%}")

# t-SNE: for visualization only (2D)
tsne = TSNE(n_components=2, perplexity=30, random_state=42)
X_tsne = tsne.fit_transform(X_scaled)

# UMAP: faster, preserves more global structure
import umap
reducer = umap.UMAP(n_components=2, n_neighbors=15, min_dist=0.1, random_state=42)
X_umap = reducer.fit_transform(X_scaled)

# Truncated SVD: for sparse data
svd = TruncatedSVD(n_components=100, random_state=42)
X_svd = svd.fit_transform(X_sparse)
```

---

## Feature Selection Methods

| Method | Type | When to Use |
|--------|------|-------------|
| Mutual Information | Filter | Any data, model-agnostic |
| Chi-squared | Filter | Categorical features |
| F-test (ANOVA) | Filter | Continuous features |
| L1 (Lasso) | Embedded | Linear models, sparse selection |
| Tree importance | Embedded | Tree models |
| Permutation importance | Wrapper | Any model, post-training |
| Recursive Feature Elimination | Wrapper | When accuracy matters more than speed |
| Boruta | Wrapper | Statistical rigor, all relevant features |

```python
from sklearn.feature_selection import (
    SelectKBest, mutual_info_classif, f_classif,
    RFE,
)
from sklearn.inspection import permutation_importance

# Mutual Information (works for any relationship)
selector = SelectKBest(mutual_info_classif, k=20)
X_selected = selector.fit_transform(X, y)
selected_features = X.columns[selector.get_support()].tolist()

# Permutation Importance (model-agnostic, after training)
result = permutation_importance(model, X_test, y_test, n_repeats=10, random_state=42)
importance = pd.Series(result.importances_mean, index=X.columns).sort_values(ascending=False)

# RFE (Recursive Feature Elimination)
from sklearn.ensemble import RandomForestClassifier
rfe = RFE(RandomForestClassifier(n_estimators=100), n_features_to_select=20)
rfe.fit(X, y)
selected = X.columns[rfe.support_].tolist()
```

---

## Cross-Validation Strategies

| Strategy | When to Use |
|----------|-------------|
| StratifiedKFold | Classification (preserves class distribution) |
| KFold | Regression |
| TimeSeriesSplit | Time series data (respects temporal ordering) |
| GroupKFold | Grouped data (all samples from same group in same fold) |
| RepeatedStratifiedKFold | Need more robust estimates |
| LeaveOneOut | Very small datasets |

```python
from sklearn.model_selection import (
    StratifiedKFold, KFold, TimeSeriesSplit,
    GroupKFold, RepeatedStratifiedKFold,
)

# Classification
cv = StratifiedKFold(n_splits=5, shuffle=True, random_state=42)

# Regression
cv = KFold(n_splits=5, shuffle=True, random_state=42)

# Time series
cv = TimeSeriesSplit(n_splits=5, gap=7)  # 7-day gap between train/test

# Grouped (e.g., all data from same user in same fold)
cv = GroupKFold(n_splits=5)
# Usage: cv.split(X, y, groups=df["user_id"])

# More robust estimates
cv = RepeatedStratifiedKFold(n_splits=5, n_repeats=3, random_state=42)
```

---

## Metric Selection

### Classification Metrics

| Metric | When to Use | Formula |
|--------|-------------|---------|
| Accuracy | Balanced classes only | (TP+TN) / Total |
| Precision | Minimize false positives (spam, fraud) | TP / (TP+FP) |
| Recall | Minimize false negatives (disease, defects) | TP / (TP+FN) |
| F1 | Balance precision and recall | 2 * (P*R) / (P+R) |
| ROC-AUC | Ranking quality, threshold-independent | Area under ROC curve |
| PR-AUC | Imbalanced data, focus on positive class | Area under PR curve |
| MCC | Best single metric for imbalanced | Matthews Correlation Coefficient |
| Log Loss | Probability calibration matters | Cross-entropy loss |

**Decision tree:**
- Balanced classes → Accuracy or F1
- Imbalanced classes → PR-AUC, F1, or MCC
- Need to rank → ROC-AUC
- Need calibrated probabilities → Log Loss
- Cost of FP ≠ cost of FN → F-beta with appropriate beta

### Regression Metrics

| Metric | When to Use | Sensitivity to Outliers |
|--------|-------------|------------------------|
| RMSE | Default, penalizes large errors | High |
| MAE | Robust to outliers | Low |
| MAPE | Relative errors matter | Low |
| R² | Explain variance proportion | Medium |
| Adjusted R² | Compare models with different # features | Medium |
| Huber Loss | Want something between MAE and MSE | Configurable |

**Decision tree:**
- Standard regression → RMSE
- Many outliers → MAE
- Need percentage interpretation → MAPE
- Comparing models → R²
- Business has specific loss function → Custom metric

---

## Algorithm Complexity Reference

| Algorithm | Training | Prediction | Memory |
|-----------|----------|-----------|--------|
| Logistic Regression | O(n * d) | O(d) | O(d) |
| Decision Tree | O(n * d * log n) | O(log n) | O(nodes) |
| Random Forest | O(k * n * d * log n) | O(k * log n) | O(k * nodes) |
| XGBoost/LightGBM | O(k * n * d) | O(k * depth) | O(k * leaves) |
| KNN | O(1) | O(n * d) | O(n * d) |
| SVM (RBF) | O(n² * d) | O(sv * d) | O(sv * d) |
| K-Means | O(n * k * d * i) | O(k * d) | O(k * d) |

Where: n = samples, d = features, k = trees/clusters, i = iterations, sv = support vectors

---

## Quick Start Templates

### Binary Classification Pipeline

```python
from sklearn.model_selection import train_test_split
from sklearn.ensemble import HistGradientBoostingClassifier
from sklearn.metrics import classification_report, roc_auc_score

X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, stratify=y, random_state=42)

model = HistGradientBoostingClassifier(max_iter=200, random_state=42)
model.fit(X_train, y_train)

y_pred = model.predict(X_test)
y_proba = model.predict_proba(X_test)[:, 1]

print(classification_report(y_test, y_pred))
print(f"ROC-AUC: {roc_auc_score(y_test, y_proba):.4f}")
```

### Regression Pipeline

```python
from sklearn.model_selection import train_test_split
from sklearn.ensemble import HistGradientBoostingRegressor
from sklearn.metrics import mean_squared_error, r2_score
import numpy as np

X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

model = HistGradientBoostingRegressor(max_iter=200, random_state=42)
model.fit(X_train, y_train)

y_pred = model.predict(X_test)

print(f"RMSE: {np.sqrt(mean_squared_error(y_test, y_pred)):.4f}")
print(f"R²: {r2_score(y_test, y_pred):.4f}")
```

---

## Ensemble Strategies

### When to Ensemble

| Strategy | When to Use | Typical Gain |
|----------|-------------|-------------|
| Bagging (Random Forest) | Reduce variance, unstable models | 5-15% |
| Boosting (XGBoost) | Reduce bias, sequential improvement | 10-30% |
| Stacking | Combine diverse model types | 2-5% over best single |
| Blending | Simple average of strong models | 1-3% |
| Voting | Quick combination, no retraining | 1-5% |

### Stacking Rules of Thumb

```python
# Good base models for stacking (diverse algorithms):
# Level 0: Different algorithm families
base_models = [
    ("lr", LogisticRegression(C=1.0, max_iter=1000)),
    ("rf", RandomForestClassifier(n_estimators=200)),
    ("xgb", XGBClassifier(n_estimators=200, max_depth=6)),
    ("lgb", LGBMClassifier(n_estimators=200, num_leaves=63)),
    ("knn", KNeighborsClassifier(n_neighbors=20)),
]

# Level 1: Simple meta-learner
meta_model = LogisticRegression(C=1.0)

# Rules:
# 1. Use diverse base models (not 5 random forests with different params)
# 2. Use simple meta-learner (logistic regression, ridge)
# 3. ALWAYS use out-of-fold predictions for meta-features
# 4. Don't include too many base models (3-7 is typical)
# 5. Stacking helps most when base models are strong but different
```

### Blending Weights

```python
# Simple average (surprisingly strong baseline)
pred = (pred_xgb + pred_lgb + pred_rf) / 3

# Weighted average (optimize weights on validation set)
from scipy.optimize import minimize

def neg_auc(weights, predictions, y_true):
    blended = np.average(predictions, axis=0, weights=weights)
    return -roc_auc_score(y_true, blended)

predictions = np.array([pred_xgb, pred_lgb, pred_rf])
result = minimize(
    neg_auc, x0=[1/3, 1/3, 1/3],
    args=(predictions, y_val),
    method="Nelder-Mead",
    constraints={"type": "eq", "fun": lambda w: np.sum(w) - 1},
)
optimal_weights = result.x
```

---

## Handling Data Challenges

### Missing Values Strategy

| Strategy | When to Use |
|----------|-------------|
| Drop rows | < 5% missing, random missingness |
| Mean/Median imputation | Numeric, not informative |
| Mode imputation | Categorical |
| Forward/backward fill | Time series |
| KNN imputation | Complex patterns, related features |
| Indicator column | Missingness itself is informative |
| Model can handle (XGBoost/LightGBM) | Tree models handle natively |

```python
from sklearn.impute import SimpleImputer, KNNImputer

# Simple strategies
num_imputer = SimpleImputer(strategy="median")
cat_imputer = SimpleImputer(strategy="most_frequent")

# KNN (uses similarity to fill)
knn_imputer = KNNImputer(n_neighbors=5)

# Add missingness indicator
df["col_is_missing"] = df["col"].isnull().astype(int)
```

### Categorical Encoding Strategy

| Encoding | When to Use | Cardinality |
|----------|-------------|-------------|
| One-hot | Nominal, few categories | < 20 |
| Ordinal | Ordered categories | Any |
| Target encoding | High cardinality, classification | > 20 |
| Frequency encoding | Simple, no target leakage | Any |
| Binary encoding | Medium cardinality | 20-100 |
| Hash encoding | Very high cardinality | > 100 |
| Leave as-is | LightGBM/CatBoost (native support) | Any |

```python
from sklearn.preprocessing import OneHotEncoder, OrdinalEncoder, TargetEncoder

# One-hot (nominal, low cardinality)
ohe = OneHotEncoder(handle_unknown="ignore", sparse_output=False)

# Ordinal (ordered)
oe = OrdinalEncoder(categories=[["low", "medium", "high"]])

# Target encoding (high cardinality) — built into sklearn 1.3+
te = TargetEncoder(smooth="auto")
```

### Feature Scaling Strategy

| Scaler | When to Use |
|--------|-------------|
| StandardScaler | Default for most algorithms, assumes ~normal |
| MinMaxScaler | Need bounded [0,1] range, neural networks |
| RobustScaler | Data has outliers |
| MaxAbsScaler | Sparse data |
| None | Tree-based models (don't need scaling) |

```python
from sklearn.preprocessing import StandardScaler, RobustScaler

# Standard (mean=0, std=1)
scaler = StandardScaler()

# Robust (uses median and IQR, resistant to outliers)
scaler = RobustScaler()
```

**Which algorithms need scaling?**
- Need scaling: Logistic Regression, SVM, KNN, Neural Networks, PCA
- Don't need scaling: Decision Trees, Random Forest, XGBoost, LightGBM

---

## Debugging Models

### Diagnosis Checklist

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| Train ≈ 100%, test poor | Overfitting | Regularize, more data, simpler model |
| Train poor, test poor | Underfitting | More features, complex model |
| Train good, test good, prod bad | Data shift | Check feature distributions |
| Metrics very unstable across folds | High variance | More data, regularize, bag |
| Single feature dominates importance | Possible leakage | Check if feature is available at inference |
| Perfect score on one feature | Target leakage | Remove the leaking feature |
| Performance varies by subgroup | Bias | Evaluate per subgroup, adjust |

### Learning Curve Interpretation

```
High bias (underfitting):
  - Train and test scores both low
  - Train and test scores converge
  - Adding more data won't help
  → Fix: More features, less regularization, more complex model

High variance (overfitting):
  - Train score high, test score low
  - Large gap between train and test
  - Adding more data may help
  → Fix: More data, regularization, simpler model, dropout

Good fit:
  - Train and test scores both reasonable
  - Small gap between them
  - Both improve with more data
```

---

## Hardware and Scaling Guidelines

| Dataset Size | Recommended Approach |
|-------------|---------------------|
| < 10K rows | Any algorithm, sklearn |
| 10K - 1M rows | XGBoost/LightGBM, sklearn |
| 1M - 100M rows | LightGBM (hist), Dask-ML |
| > 100M rows | PySpark MLlib, distributed |
| GPU available | XGBoost gpu_hist, RAPIDS cuML |

```python
# GPU-accelerated XGBoost
params = {"tree_method": "gpu_hist", "device": "cuda"}

# Distributed LightGBM
# Use Dask-LightGBM or LightGBM's built-in distributed mode
```
