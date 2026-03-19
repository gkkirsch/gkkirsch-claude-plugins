# scikit-learn Recipes — Classification, Regression, Clustering, and Pipeline Patterns

Quick reference for scikit-learn patterns, algorithm recipes, pipeline construction, and feature engineering.

---

## Classification Recipes

### Binary Classification Pipeline

```python
from sklearn.pipeline import Pipeline
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import StandardScaler, OneHotEncoder
from sklearn.impute import SimpleImputer
from sklearn.model_selection import train_test_split, cross_val_score
import xgboost as xgb

# Data split
X_train, X_test, y_train, y_test = train_test_split(
    X, y, test_size=0.2, stratify=y, random_state=42
)

# Preprocessing
preprocessor = ColumnTransformer([
    ("num", Pipeline([
        ("imputer", SimpleImputer(strategy="median")),
        ("scaler", StandardScaler()),
    ]), numeric_cols),
    ("cat", Pipeline([
        ("imputer", SimpleImputer(strategy="most_frequent")),
        ("encoder", OneHotEncoder(handle_unknown="ignore", sparse_output=False)),
    ]), categorical_cols),
])

# Pipeline
pipeline = Pipeline([
    ("prep", preprocessor),
    ("model", xgb.XGBClassifier(
        n_estimators=200, max_depth=6, learning_rate=0.1,
        tree_method="hist", eval_metric="logloss", random_state=42,
    )),
])

# Cross-validation
scores = cross_val_score(pipeline, X_train, y_train, cv=5, scoring="roc_auc")
print(f"CV ROC AUC: {scores.mean():.4f} (+/- {scores.std():.4f})")

# Fit and evaluate
pipeline.fit(X_train, y_train)
y_proba = pipeline.predict_proba(X_test)[:, 1]
```

### Multiclass Classification

```python
from sklearn.metrics import classification_report, top_k_accuracy_score
from sklearn.preprocessing import LabelEncoder

# Encode labels
le = LabelEncoder()
y_encoded = le.fit_transform(y)

# Train
pipeline.fit(X_train, y_encoded_train)

# Evaluate
y_pred = pipeline.predict(X_test)
print(classification_report(y_encoded_test, y_pred, target_names=le.classes_))

# Top-k accuracy
y_proba = pipeline.predict_proba(X_test)
top3_acc = top_k_accuracy_score(y_encoded_test, y_proba, k=3)
```

### Imbalanced Classification

```python
from imblearn.pipeline import Pipeline as ImbPipeline
from imblearn.over_sampling import SMOTE
from sklearn.metrics import f1_score, precision_recall_curve

# SMOTE pipeline
pipeline = ImbPipeline([
    ("prep", preprocessor),
    ("smote", SMOTE(sampling_strategy=0.5, random_state=42)),
    ("model", xgb.XGBClassifier(
        scale_pos_weight=len(y[y==0]) / len(y[y==1]),
        tree_method="hist", random_state=42,
    )),
])

# Threshold optimization
y_proba = pipeline.predict_proba(X_test)[:, 1]
precisions, recalls, thresholds = precision_recall_curve(y_test, y_proba)
f1_scores = 2 * precisions * recalls / (precisions + recalls + 1e-8)
best_threshold = thresholds[f1_scores.argmax()]
y_pred_optimized = (y_proba >= best_threshold).astype(int)
```

---

## Regression Recipes

### Standard Regression Pipeline

```python
from sklearn.linear_model import Ridge, Lasso, ElasticNet
from sklearn.ensemble import HistGradientBoostingRegressor
from sklearn.metrics import mean_squared_error, r2_score
import numpy as np

pipeline = Pipeline([
    ("prep", preprocessor),
    ("model", HistGradientBoostingRegressor(
        max_iter=500, max_depth=6, learning_rate=0.05,
        min_samples_leaf=20, random_state=42,
    )),
])

pipeline.fit(X_train, y_train)
y_pred = pipeline.predict(X_test)

rmse = np.sqrt(mean_squared_error(y_test, y_pred))
r2 = r2_score(y_test, y_pred)
print(f"RMSE: {rmse:.4f}, R2: {r2:.4f}")
```

### Log-Transformed Target

```python
from sklearn.compose import TransformedTargetRegressor

model = TransformedTargetRegressor(
    regressor=HistGradientBoostingRegressor(max_iter=500, random_state=42),
    func=np.log1p,
    inverse_func=np.expm1,
)

pipeline = Pipeline([("prep", preprocessor), ("model", model)])
pipeline.fit(X_train, y_train)
```

### Quantile Regression

```python
from sklearn.ensemble import GradientBoostingRegressor

# Predict median and prediction intervals
quantiles = [0.1, 0.5, 0.9]
models = {}

for q in quantiles:
    models[q] = Pipeline([
        ("prep", preprocessor),
        ("model", GradientBoostingRegressor(
            loss="quantile", alpha=q, n_estimators=200, random_state=42,
        )),
    ])
    models[q].fit(X_train, y_train)

# Prediction interval
lower = models[0.1].predict(X_test)
median = models[0.5].predict(X_test)
upper = models[0.9].predict(X_test)
```

---

## Clustering Recipes

### K-Means Pipeline

```python
from sklearn.cluster import KMeans
from sklearn.preprocessing import StandardScaler
from sklearn.metrics import silhouette_score

# Scale features before clustering
scaler = StandardScaler()
X_scaled = scaler.fit_transform(X)

# Elbow method
inertias = []
silhouettes = []
K_range = range(2, 15)

for k in K_range:
    km = KMeans(n_clusters=k, n_init=10, random_state=42)
    labels = km.fit_predict(X_scaled)
    inertias.append(km.inertia_)
    silhouettes.append(silhouette_score(X_scaled, labels))

# Best k by silhouette
best_k = K_range[np.argmax(silhouettes)]
print(f"Best k: {best_k}, Silhouette: {max(silhouettes):.4f}")

# Final model
km = KMeans(n_clusters=best_k, n_init=10, random_state=42)
labels = km.fit_predict(X_scaled)
```

### DBSCAN

```python
from sklearn.cluster import DBSCAN
from sklearn.neighbors import NearestNeighbors

# Find eps with k-distance graph
nn = NearestNeighbors(n_neighbors=5)
nn.fit(X_scaled)
distances, _ = nn.kneighbors(X_scaled)
k_distances = np.sort(distances[:, -1])
# Plot k_distances to find the elbow → that's your eps

dbscan = DBSCAN(eps=0.5, min_samples=5)
labels = dbscan.fit_predict(X_scaled)

n_clusters = len(set(labels)) - (1 if -1 in labels else 0)
n_noise = (labels == -1).sum()
print(f"Clusters: {n_clusters}, Noise points: {n_noise}")
```

### Gaussian Mixture Models

```python
from sklearn.mixture import GaussianMixture

# BIC-based model selection
bics = []
for n in range(2, 15):
    gmm = GaussianMixture(n_components=n, covariance_type="full", random_state=42)
    gmm.fit(X_scaled)
    bics.append(gmm.bic(X_scaled))

best_n = range(2, 15)[np.argmin(bics)]
gmm = GaussianMixture(n_components=best_n, covariance_type="full", random_state=42)
labels = gmm.fit_predict(X_scaled)
probs = gmm.predict_proba(X_scaled)  # Soft assignments
```

---

## Ensemble Recipes

### Stacking Ensemble

```python
from sklearn.ensemble import StackingClassifier

stacking = StackingClassifier(
    estimators=[
        ("rf", RandomForestClassifier(n_estimators=200, random_state=42)),
        ("xgb", xgb.XGBClassifier(n_estimators=200, tree_method="hist", random_state=42)),
        ("lgb", lgb.LGBMClassifier(n_estimators=200, verbose=-1, random_state=42)),
        ("lr", LogisticRegression(max_iter=1000, C=0.1)),
    ],
    final_estimator=LogisticRegression(max_iter=1000),
    cv=5,
    stack_method="predict_proba",
    n_jobs=-1,
)

pipeline = Pipeline([("prep", preprocessor), ("stack", stacking)])
pipeline.fit(X_train, y_train)
```

### Voting Ensemble

```python
from sklearn.ensemble import VotingClassifier

voting = VotingClassifier(
    estimators=[
        ("rf", RandomForestClassifier(n_estimators=200, random_state=42)),
        ("xgb", xgb.XGBClassifier(n_estimators=200, tree_method="hist", random_state=42)),
        ("lgb", lgb.LGBMClassifier(n_estimators=200, verbose=-1, random_state=42)),
    ],
    voting="soft",
    weights=[1, 2, 2],
    n_jobs=-1,
)
```

---

## Feature Selection Recipes

### Filter Methods

```python
from sklearn.feature_selection import SelectKBest, mutual_info_classif, f_classif

# Mutual Information
selector = SelectKBest(mutual_info_classif, k=30)
X_selected = selector.fit_transform(X_train, y_train)
selected_features = X_train.columns[selector.get_support()]

# F-test (ANOVA)
selector = SelectKBest(f_classif, k=30)
X_selected = selector.fit_transform(X_train, y_train)

# Variance threshold
from sklearn.feature_selection import VarianceThreshold
selector = VarianceThreshold(threshold=0.01)
X_selected = selector.fit_transform(X_train)
```

### Wrapper Methods

```python
from sklearn.feature_selection import SequentialFeatureSelector, RFECV

# Forward selection
sfs = SequentialFeatureSelector(
    xgb.XGBClassifier(n_estimators=100, tree_method="hist", random_state=42),
    n_features_to_select=20,
    direction="forward",
    cv=5,
    scoring="roc_auc",
    n_jobs=-1,
)
sfs.fit(X_train, y_train)
selected = X_train.columns[sfs.get_support()]

# RFE with cross-validation
rfecv = RFECV(
    estimator=RandomForestClassifier(n_estimators=100, random_state=42),
    step=5,
    cv=5,
    scoring="roc_auc",
    min_features_to_select=10,
    n_jobs=-1,
)
rfecv.fit(X_train, y_train)
print(f"Optimal features: {rfecv.n_features_}")
```

### Embedded Methods (L1 Regularization)

```python
from sklearn.linear_model import LogisticRegression
from sklearn.feature_selection import SelectFromModel

# L1 regularization feature selection
selector = SelectFromModel(
    LogisticRegression(penalty="l1", C=0.1, solver="saga", max_iter=5000, random_state=42),
    threshold="median",
)
selector.fit(X_train_scaled, y_train)
selected = X_train.columns[selector.get_support()]
```

---

## Cross-Validation Recipes

### Stratified K-Fold

```python
from sklearn.model_selection import StratifiedKFold, cross_validate

skf = StratifiedKFold(n_splits=5, shuffle=True, random_state=42)

results = cross_validate(
    pipeline, X, y, cv=skf,
    scoring=["roc_auc", "f1", "precision", "recall"],
    return_train_score=True,
    n_jobs=-1,
)

for metric in ["roc_auc", "f1", "precision", "recall"]:
    mean = results[f"test_{metric}"].mean()
    std = results[f"test_{metric}"].std()
    print(f"{metric}: {mean:.4f} (+/- {std:.4f})")
```

### Time Series Split

```python
from sklearn.model_selection import TimeSeriesSplit

tscv = TimeSeriesSplit(n_splits=5, gap=7)  # 7-day gap

for fold, (train_idx, test_idx) in enumerate(tscv.split(X)):
    X_tr, X_te = X.iloc[train_idx], X.iloc[test_idx]
    y_tr, y_te = y.iloc[train_idx], y.iloc[test_idx]

    pipeline.fit(X_tr, y_tr)
    score = pipeline.score(X_te, y_te)
    print(f"Fold {fold}: {score:.4f} | Train: {X_tr.index.min()} to {X_tr.index.max()} | Test: {X_te.index.min()} to {X_te.index.max()}")
```

### Nested Cross-Validation

```python
from sklearn.model_selection import cross_val_score, GridSearchCV, StratifiedKFold

# Inner loop: hyperparameter tuning
inner_cv = StratifiedKFold(n_splits=3, shuffle=True, random_state=42)

param_grid = {
    "model__n_estimators": [100, 200],
    "model__max_depth": [4, 6, 8],
    "model__learning_rate": [0.05, 0.1],
}

grid_search = GridSearchCV(
    pipeline, param_grid, cv=inner_cv, scoring="roc_auc", n_jobs=-1
)

# Outer loop: unbiased performance estimate
outer_cv = StratifiedKFold(n_splits=5, shuffle=True, random_state=42)
nested_scores = cross_val_score(grid_search, X, y, cv=outer_cv, scoring="roc_auc")
print(f"Nested CV ROC AUC: {nested_scores.mean():.4f} (+/- {nested_scores.std():.4f})")
```

### Group K-Fold

```python
from sklearn.model_selection import GroupKFold, StratifiedGroupKFold

# Ensure same group (e.g., customer) doesn't appear in both train and test
gkf = GroupKFold(n_splits=5)

scores = cross_val_score(
    pipeline, X, y, cv=gkf, groups=customer_ids, scoring="roc_auc"
)

# Stratified + Grouped
sgkf = StratifiedGroupKFold(n_splits=5, shuffle=True, random_state=42)
scores = cross_val_score(
    pipeline, X, y, cv=sgkf, groups=customer_ids, scoring="roc_auc"
)
```

---

## Hyperparameter Tuning Recipes

### Optuna with sklearn

```python
import optuna

def objective(trial):
    model_type = trial.suggest_categorical("model", ["xgboost", "lightgbm", "rf"])

    if model_type == "xgboost":
        params = {
            "model": xgb.XGBClassifier(
                n_estimators=trial.suggest_int("n_estimators", 100, 500),
                max_depth=trial.suggest_int("max_depth", 3, 10),
                learning_rate=trial.suggest_float("lr", 0.01, 0.3, log=True),
                subsample=trial.suggest_float("subsample", 0.5, 1.0),
                colsample_bytree=trial.suggest_float("colsample", 0.5, 1.0),
                tree_method="hist", random_state=42,
            ),
        }
    elif model_type == "lightgbm":
        params = {
            "model": lgb.LGBMClassifier(
                n_estimators=trial.suggest_int("n_estimators", 100, 500),
                max_depth=trial.suggest_int("max_depth", 3, 12),
                learning_rate=trial.suggest_float("lr", 0.01, 0.3, log=True),
                num_leaves=trial.suggest_int("num_leaves", 20, 200),
                verbose=-1, random_state=42,
            ),
        }
    else:
        params = {
            "model": RandomForestClassifier(
                n_estimators=trial.suggest_int("n_estimators", 100, 500),
                max_depth=trial.suggest_int("max_depth", 5, 20),
                min_samples_leaf=trial.suggest_int("min_leaf", 1, 20),
                random_state=42, n_jobs=-1,
            ),
        }

    pipe = Pipeline([("prep", preprocessor), ("model", params["model"])])
    scores = cross_val_score(pipe, X_train, y_train, cv=5, scoring="roc_auc")
    return scores.mean()

study = optuna.create_study(direction="maximize")
study.optimize(objective, n_trials=100, show_progress_bar=True)
```

### Halving Search (Efficient Grid Search)

```python
from sklearn.experimental import enable_halving_search_cv
from sklearn.model_selection import HalvingGridSearchCV

param_grid = {
    "model__n_estimators": [100, 200, 300, 500],
    "model__max_depth": [3, 5, 7, 9],
    "model__learning_rate": [0.01, 0.05, 0.1, 0.2],
    "model__subsample": [0.6, 0.8, 1.0],
}

halving = HalvingGridSearchCV(
    pipeline,
    param_grid,
    scoring="roc_auc",
    cv=5,
    factor=3,  # Eliminate 2/3 of candidates each round
    resource="n_samples",
    min_resources=100,
    n_jobs=-1,
    verbose=1,
)
halving.fit(X_train, y_train)
```

---

## Pipeline Patterns

### Column-Specific Preprocessing

```python
from sklearn.compose import make_column_selector

preprocessor = ColumnTransformer([
    ("num", Pipeline([
        ("imputer", SimpleImputer(strategy="median")),
        ("scaler", StandardScaler()),
    ]), make_column_selector(dtype_include=np.number)),
    ("cat", Pipeline([
        ("imputer", SimpleImputer(strategy="constant", fill_value="missing")),
        ("encoder", OneHotEncoder(handle_unknown="ignore", sparse_output=False, min_frequency=5)),
    ]), make_column_selector(dtype_include="object")),
])
```

### Pipeline with Feature Engineering

```python
from sklearn.preprocessing import FunctionTransformer, PolynomialFeatures

pipeline = Pipeline([
    ("date_features", FunctionTransformer(extract_date_features)),
    ("prep", preprocessor),
    ("poly", PolynomialFeatures(degree=2, interaction_only=True, include_bias=False)),
    ("select", SelectKBest(mutual_info_classif, k=50)),
    ("model", xgb.XGBClassifier(tree_method="hist", random_state=42)),
])
```

### Save and Load Pipeline

```python
import joblib

# Save entire pipeline (preprocessor + model)
joblib.dump(pipeline, "pipeline.joblib", compress=3)

# Load
pipeline = joblib.load("pipeline.joblib")
predictions = pipeline.predict(new_data)
```

---

## Metric Selection Guide

| Problem | Primary Metric | Secondary Metrics |
|---------|---------------|-------------------|
| Balanced binary classification | ROC AUC | F1, Accuracy |
| Imbalanced binary classification | PR AUC (Average Precision) | F1, Recall |
| Cost-sensitive classification | Custom cost function | Precision, Recall |
| Multiclass classification | Macro F1 | Weighted F1, Accuracy |
| Ranking | NDCG, MAP | MRR |
| Regression | RMSE | MAE, R2, MAPE |
| Regression with outliers | MAE, MedAE | Huber loss |
| Probabilistic regression | CRPS | Pinball loss |

---

## Algorithm Quick Reference

### When to Use What

| Algorithm | Best For | Avoid When |
|-----------|----------|------------|
| Logistic Regression | Baseline, interpretability, sparse data | Nonlinear relationships |
| Random Forest | Robust baseline, feature importance | Very high-dimensional sparse |
| XGBoost | Tabular data, competitions | Small datasets (< 1K) |
| LightGBM | Large datasets, speed | Need strict reproducibility |
| SVM | Small-medium data, high-dim | > 100K samples (slow) |
| KNN | Small data, local patterns | High dimensions, large data |
| Naive Bayes | Text, very fast training | Feature dependencies |
| Ridge/Lasso | Linear relationships, regularization | Complex nonlinear patterns |
| HistGradientBoosting | Large data, missing values | Very small datasets |

### Hyperparameter Defaults That Work

| Model | Parameter | Good Default | Range to Search |
|-------|-----------|-------------|-----------------|
| XGBoost | n_estimators | 200 | 100-1000 |
| XGBoost | max_depth | 6 | 3-10 |
| XGBoost | learning_rate | 0.1 | 0.01-0.3 |
| XGBoost | subsample | 0.8 | 0.5-1.0 |
| LightGBM | num_leaves | 31 | 20-300 |
| Random Forest | n_estimators | 200 | 100-500 |
| Random Forest | max_depth | 10 | 5-20 |
| Logistic Regression | C | 1.0 | 0.001-100 |
| SVM | C | 1.0 | 0.01-100 |
| KNN | n_neighbors | 5 | 3-15 |
