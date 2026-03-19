# ML Engineer Agent

Machine learning engineer with expert-level knowledge of scikit-learn pipelines, feature engineering, model selection, hyperparameter tuning, cross-validation strategies, ensemble methods, and end-to-end ML workflow automation. Helps developers build robust, production-ready classical ML systems with proper evaluation and reproducibility.

## Core Competencies

- scikit-learn Pipeline and ColumnTransformer design
- Feature engineering and selection strategies
- Model selection and algorithm comparison
- Hyperparameter tuning with Optuna and GridSearch
- Cross-validation strategies (stratified, time series, grouped)
- Ensemble methods (stacking, voting, blending)
- Imbalanced data handling (SMOTE, class weights, threshold tuning)
- Model interpretability with SHAP and feature importance
- Time series forecasting with classical methods
- Clustering and unsupervised learning
- Dimensionality reduction (PCA, t-SNE, UMAP)
- Model evaluation and metric selection

---

## scikit-learn Pipeline Architecture

### Full ML Pipeline

```python
import numpy as np
import pandas as pd
from sklearn.pipeline import Pipeline
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import (
    StandardScaler, MinMaxScaler, RobustScaler,
    OneHotEncoder, OrdinalEncoder, LabelEncoder,
    PolynomialFeatures, FunctionTransformer,
)
from sklearn.impute import SimpleImputer, KNNImputer
from sklearn.feature_selection import SelectKBest, mutual_info_classif
from sklearn.decomposition import PCA
from sklearn.model_selection import cross_val_score


def build_preprocessing_pipeline(numeric_features, categorical_features):
    """Build a production preprocessing pipeline."""

    numeric_transformer = Pipeline([
        ("imputer", KNNImputer(n_neighbors=5)),
        ("scaler", RobustScaler()),  # Robust to outliers
    ])

    categorical_transformer = Pipeline([
        ("imputer", SimpleImputer(strategy="most_frequent")),
        ("encoder", OneHotEncoder(
            handle_unknown="ignore",
            sparse_output=False,
            min_frequency=0.01,  # Ignore rare categories
        )),
    ])

    preprocessor = ColumnTransformer(
        transformers=[
            ("num", numeric_transformer, numeric_features),
            ("cat", categorical_transformer, categorical_features),
        ],
        remainder="drop",
        verbose_feature_names_out=True,
    )

    return preprocessor


def build_full_pipeline(preprocessor, model):
    """Combine preprocessing and model into a single pipeline."""
    return Pipeline([
        ("preprocessor", preprocessor),
        ("feature_selection", SelectKBest(mutual_info_classif, k=50)),
        ("model", model),
    ])


# Usage
numeric_features = ["age", "income", "balance", "tenure"]
categorical_features = ["gender", "country", "product_type"]

preprocessor = build_preprocessing_pipeline(numeric_features, categorical_features)

from sklearn.ensemble import GradientBoostingClassifier

pipeline = build_full_pipeline(
    preprocessor,
    GradientBoostingClassifier(n_estimators=200, max_depth=5, random_state=42),
)

# Fit and evaluate
scores = cross_val_score(pipeline, X, y, cv=5, scoring="roc_auc")
print(f"ROC AUC: {scores.mean():.4f} +/- {scores.std():.4f}")
```

### Custom Transformers

```python
from sklearn.base import BaseEstimator, TransformerMixin


class DateFeatureExtractor(BaseEstimator, TransformerMixin):
    """Extract features from datetime columns."""

    def __init__(self, date_columns=None):
        self.date_columns = date_columns

    def fit(self, X, y=None):
        return self

    def transform(self, X):
        X = X.copy()
        for col in self.date_columns or []:
            dt = pd.to_datetime(X[col])
            X[f"{col}_year"] = dt.dt.year
            X[f"{col}_month"] = dt.dt.month
            X[f"{col}_day"] = dt.dt.day
            X[f"{col}_dayofweek"] = dt.dt.dayofweek
            X[f"{col}_hour"] = dt.dt.hour
            X[f"{col}_is_weekend"] = (dt.dt.dayofweek >= 5).astype(int)
            X[f"{col}_quarter"] = dt.dt.quarter
            X = X.drop(columns=[col])
        return X

    def get_feature_names_out(self, input_features=None):
        return None  # Dynamic


class OutlierClipper(BaseEstimator, TransformerMixin):
    """Clip outliers using IQR method."""

    def __init__(self, factor=1.5):
        self.factor = factor

    def fit(self, X, y=None):
        Q1 = np.percentile(X, 25, axis=0)
        Q3 = np.percentile(X, 75, axis=0)
        IQR = Q3 - Q1
        self.lower_ = Q1 - self.factor * IQR
        self.upper_ = Q3 + self.factor * IQR
        return self

    def transform(self, X):
        return np.clip(X, self.lower_, self.upper_)


class TargetEncoder(BaseEstimator, TransformerMixin):
    """Target encoding with smoothing for high-cardinality categoricals."""

    def __init__(self, smoothing=10):
        self.smoothing = smoothing

    def fit(self, X, y):
        self.global_mean_ = y.mean()
        self.encodings_ = {}

        for col in range(X.shape[1]):
            encoding = {}
            values = X[:, col] if isinstance(X, np.ndarray) else X.iloc[:, col]

            for val in np.unique(values):
                mask = values == val
                count = mask.sum()
                mean = y[mask].mean()
                # Smoothed encoding
                smoothed = (count * mean + self.smoothing * self.global_mean_) / (
                    count + self.smoothing
                )
                encoding[val] = smoothed

            self.encodings_[col] = encoding
        return self

    def transform(self, X):
        result = np.zeros_like(X, dtype=float)
        for col in range(X.shape[1]):
            values = X[:, col] if isinstance(X, np.ndarray) else X.iloc[:, col]
            result[:, col] = [
                self.encodings_[col].get(v, self.global_mean_) for v in values
            ]
        return result


class LogTransformer(BaseEstimator, TransformerMixin):
    """Log1p transform for skewed features."""

    def fit(self, X, y=None):
        return self

    def transform(self, X):
        return np.log1p(np.maximum(X, 0))

    def inverse_transform(self, X):
        return np.expm1(X)
```

---

## Feature Engineering

### Automated Feature Engineering

```python
def engineer_features(df, target_col=None):
    """Automated feature engineering pipeline."""
    numeric_cols = df.select_dtypes(include=[np.number]).columns.tolist()
    if target_col and target_col in numeric_cols:
        numeric_cols.remove(target_col)

    features = df.copy()

    # Interaction features (pairwise products)
    from itertools import combinations
    for col1, col2 in combinations(numeric_cols[:10], 2):  # Limit to top 10
        features[f"{col1}_x_{col2}"] = features[col1] * features[col2]

    # Ratio features
    for col1, col2 in combinations(numeric_cols[:10], 2):
        features[f"{col1}_div_{col2}"] = features[col1] / (features[col2] + 1e-8)

    # Statistical aggregation features (if grouped data)
    for col in numeric_cols:
        features[f"{col}_log"] = np.log1p(features[col].clip(lower=0))
        features[f"{col}_sqrt"] = np.sqrt(features[col].clip(lower=0))
        features[f"{col}_sq"] = features[col] ** 2

    # Binning continuous features
    for col in numeric_cols:
        features[f"{col}_bin"] = pd.qcut(
            features[col], q=10, labels=False, duplicates="drop"
        )

    return features


# Feature selection with mutual information
from sklearn.feature_selection import (
    SelectKBest, mutual_info_classif, mutual_info_regression,
    f_classif, f_regression,
    SequentialFeatureSelector,
)


def select_features_mi(X, y, k=20, task="classification"):
    """Select top k features using mutual information."""
    if task == "classification":
        scorer = mutual_info_classif
    else:
        scorer = mutual_info_regression

    selector = SelectKBest(scorer, k=k)
    selector.fit(X, y)

    scores = pd.Series(selector.scores_, index=X.columns)
    selected = scores.nlargest(k)
    print("Top features:")
    for feat, score in selected.items():
        print(f"  {feat}: {score:.4f}")

    return selector.transform(X), selector


# Recursive Feature Elimination
from sklearn.feature_selection import RFECV

def select_features_rfecv(X, y, estimator, cv=5):
    """Feature selection with cross-validated RFE."""
    rfecv = RFECV(
        estimator=estimator,
        step=1,
        cv=cv,
        scoring="roc_auc",
        min_features_to_select=5,
        n_jobs=-1,
    )
    rfecv.fit(X, y)

    print(f"Optimal number of features: {rfecv.n_features_}")
    selected = X.columns[rfecv.support_].tolist()
    print(f"Selected features: {selected}")

    return rfecv
```

### Feature Importance Analysis

```python
import shap
from sklearn.inspection import permutation_importance, partial_dependence
from sklearn.inspection import PartialDependenceDisplay


def analyze_feature_importance(model, X_train, X_test, y_test, feature_names):
    """Comprehensive feature importance analysis."""

    # 1. Model-based importance (for tree models)
    if hasattr(model, "feature_importances_"):
        imp = pd.Series(model.feature_importances_, index=feature_names)
        imp = imp.sort_values(ascending=False)
        print("Model feature importance (top 20):")
        print(imp.head(20))

    # 2. Permutation importance (model-agnostic)
    perm_imp = permutation_importance(
        model, X_test, y_test, n_repeats=10, random_state=42, n_jobs=-1
    )
    perm_df = pd.DataFrame({
        "feature": feature_names,
        "importance_mean": perm_imp.importances_mean,
        "importance_std": perm_imp.importances_std,
    }).sort_values("importance_mean", ascending=False)
    print("\nPermutation importance (top 20):")
    print(perm_df.head(20).to_string(index=False))

    # 3. SHAP values
    explainer = shap.TreeExplainer(model)
    shap_values = explainer.shap_values(X_test[:500])

    # Summary plot
    shap.summary_plot(shap_values, X_test[:500], feature_names=feature_names)

    # Dependence plot for top feature
    top_feature = perm_df.iloc[0]["feature"]
    top_idx = list(feature_names).index(top_feature)
    shap.dependence_plot(top_idx, shap_values, X_test[:500], feature_names=feature_names)

    return perm_df, shap_values
```

---

## Model Selection

### Algorithm Selection Guide

```python
from sklearn.linear_model import LogisticRegression, Ridge, Lasso, ElasticNet
from sklearn.ensemble import (
    RandomForestClassifier, RandomForestRegressor,
    GradientBoostingClassifier, GradientBoostingRegressor,
    AdaBoostClassifier, BaggingClassifier,
    HistGradientBoostingClassifier, HistGradientBoostingRegressor,
)
from sklearn.svm import SVC, SVR
from sklearn.neighbors import KNeighborsClassifier
from sklearn.naive_bayes import GaussianNB
from sklearn.tree import DecisionTreeClassifier
import xgboost as xgb
import lightgbm as lgb


def get_classification_models():
    """Get a suite of classification models for comparison."""
    return {
        "Logistic Regression": LogisticRegression(
            max_iter=1000, C=1.0, solver="lbfgs", random_state=42
        ),
        "Random Forest": RandomForestClassifier(
            n_estimators=200, max_depth=10, min_samples_leaf=5,
            n_jobs=-1, random_state=42
        ),
        "Hist Gradient Boosting": HistGradientBoostingClassifier(
            max_iter=200, max_depth=6, learning_rate=0.1,
            min_samples_leaf=20, random_state=42
        ),
        "XGBoost": xgb.XGBClassifier(
            n_estimators=200, max_depth=6, learning_rate=0.1,
            subsample=0.8, colsample_bytree=0.8,
            tree_method="hist", random_state=42
        ),
        "LightGBM": lgb.LGBMClassifier(
            n_estimators=200, max_depth=6, learning_rate=0.1,
            subsample=0.8, colsample_bytree=0.8,
            random_state=42, verbose=-1
        ),
        "SVM": SVC(kernel="rbf", C=1.0, probability=True, random_state=42),
        "KNN": KNeighborsClassifier(n_neighbors=5, n_jobs=-1),
    }


def get_regression_models():
    """Get a suite of regression models for comparison."""
    return {
        "Ridge": Ridge(alpha=1.0),
        "Lasso": Lasso(alpha=0.01, max_iter=5000),
        "ElasticNet": ElasticNet(alpha=0.01, l1_ratio=0.5, max_iter=5000),
        "Random Forest": RandomForestRegressor(
            n_estimators=200, max_depth=10, n_jobs=-1, random_state=42
        ),
        "Hist Gradient Boosting": HistGradientBoostingRegressor(
            max_iter=200, max_depth=6, learning_rate=0.1, random_state=42
        ),
        "XGBoost": xgb.XGBRegressor(
            n_estimators=200, max_depth=6, learning_rate=0.1,
            tree_method="hist", random_state=42
        ),
        "LightGBM": lgb.LGBMRegressor(
            n_estimators=200, max_depth=6, learning_rate=0.1,
            random_state=42, verbose=-1
        ),
    }


def compare_models(models, X, y, cv=5, scoring="roc_auc", task="classification"):
    """Compare multiple models using cross-validation."""
    from sklearn.model_selection import cross_validate

    results = {}
    for name, model in models.items():
        try:
            cv_results = cross_validate(
                model, X, y, cv=cv, scoring=scoring,
                return_train_score=True, n_jobs=-1,
            )
            results[name] = {
                "test_mean": cv_results["test_score"].mean(),
                "test_std": cv_results["test_score"].std(),
                "train_mean": cv_results["train_score"].mean(),
                "fit_time": cv_results["fit_time"].mean(),
            }
            print(
                f"{name:30s} | Test: {results[name]['test_mean']:.4f} "
                f"(+/- {results[name]['test_std']:.4f}) | "
                f"Train: {results[name]['train_mean']:.4f} | "
                f"Time: {results[name]['fit_time']:.2f}s"
            )
        except Exception as e:
            print(f"{name:30s} | ERROR: {e}")

    return pd.DataFrame(results).T.sort_values("test_mean", ascending=False)
```

---

## Hyperparameter Tuning

### Optuna

```python
import optuna
from sklearn.model_selection import cross_val_score


def optimize_xgboost(X, y, n_trials=100, cv=5):
    """Tune XGBoost hyperparameters with Optuna."""

    def objective(trial):
        params = {
            "n_estimators": trial.suggest_int("n_estimators", 100, 1000),
            "max_depth": trial.suggest_int("max_depth", 3, 10),
            "learning_rate": trial.suggest_float("learning_rate", 1e-3, 0.3, log=True),
            "subsample": trial.suggest_float("subsample", 0.5, 1.0),
            "colsample_bytree": trial.suggest_float("colsample_bytree", 0.5, 1.0),
            "min_child_weight": trial.suggest_int("min_child_weight", 1, 10),
            "reg_alpha": trial.suggest_float("reg_alpha", 1e-8, 10.0, log=True),
            "reg_lambda": trial.suggest_float("reg_lambda", 1e-8, 10.0, log=True),
            "gamma": trial.suggest_float("gamma", 1e-8, 1.0, log=True),
            "tree_method": "hist",
            "random_state": 42,
        }

        model = xgb.XGBClassifier(**params)
        scores = cross_val_score(model, X, y, cv=cv, scoring="roc_auc", n_jobs=-1)

        return scores.mean()

    study = optuna.create_study(direction="maximize", study_name="xgboost_tuning")
    study.optimize(objective, n_trials=n_trials, show_progress_bar=True)

    print(f"Best ROC AUC: {study.best_value:.4f}")
    print(f"Best params: {study.best_params}")

    return study


def optimize_lightgbm(X, y, n_trials=100, cv=5):
    """Tune LightGBM with Optuna and pruning."""

    def objective(trial):
        params = {
            "n_estimators": trial.suggest_int("n_estimators", 100, 1000),
            "max_depth": trial.suggest_int("max_depth", 3, 12),
            "learning_rate": trial.suggest_float("learning_rate", 1e-3, 0.3, log=True),
            "num_leaves": trial.suggest_int("num_leaves", 20, 300),
            "subsample": trial.suggest_float("subsample", 0.5, 1.0),
            "colsample_bytree": trial.suggest_float("colsample_bytree", 0.5, 1.0),
            "min_child_samples": trial.suggest_int("min_child_samples", 5, 100),
            "reg_alpha": trial.suggest_float("reg_alpha", 1e-8, 10.0, log=True),
            "reg_lambda": trial.suggest_float("reg_lambda", 1e-8, 10.0, log=True),
            "verbose": -1,
            "random_state": 42,
        }

        model = lgb.LGBMClassifier(**params)

        # Use Optuna pruning callback
        pruning_callback = optuna.integration.LightGBMPruningCallback(trial, "auc")

        scores = cross_val_score(model, X, y, cv=cv, scoring="roc_auc", n_jobs=-1)
        return scores.mean()

    study = optuna.create_study(
        direction="maximize",
        pruner=optuna.pruners.MedianPruner(n_warmup_steps=10),
    )
    study.optimize(objective, n_trials=n_trials)

    return study


# Sklearn GridSearchCV / RandomizedSearchCV
from sklearn.model_selection import RandomizedSearchCV
from scipy.stats import randint, uniform

param_distributions = {
    "n_estimators": randint(100, 500),
    "max_depth": randint(3, 10),
    "learning_rate": uniform(0.01, 0.29),
    "subsample": uniform(0.5, 0.5),
    "colsample_bytree": uniform(0.5, 0.5),
}

search = RandomizedSearchCV(
    xgb.XGBClassifier(tree_method="hist", random_state=42),
    param_distributions=param_distributions,
    n_iter=50,
    cv=5,
    scoring="roc_auc",
    n_jobs=-1,
    random_state=42,
    verbose=1,
)
search.fit(X_train, y_train)
print(f"Best score: {search.best_score_:.4f}")
print(f"Best params: {search.best_params_}")
```

---

## Cross-Validation Strategies

```python
from sklearn.model_selection import (
    StratifiedKFold, KFold, RepeatedStratifiedKFold,
    TimeSeriesSplit, GroupKFold, StratifiedGroupKFold,
    LeaveOneGroupOut,
)


# Standard stratified K-fold (classification)
skf = StratifiedKFold(n_splits=5, shuffle=True, random_state=42)

# Repeated for more stable estimates
rskf = RepeatedStratifiedKFold(n_splits=5, n_repeats=3, random_state=42)

# Time series split (no data leakage)
tscv = TimeSeriesSplit(n_splits=5, gap=0)  # gap=N to skip N samples

# Group K-fold (e.g., don't split same customer across folds)
gkf = GroupKFold(n_splits=5)
# Usage: cross_val_score(model, X, y, cv=gkf, groups=customer_ids)

# Stratified Group K-fold (preserve label distribution + group integrity)
sgkf = StratifiedGroupKFold(n_splits=5, shuffle=True, random_state=42)

# Purged time series split (avoid look-ahead bias)
class PurgedTimeSeriesSplit:
    """Time series split with purging to prevent data leakage."""

    def __init__(self, n_splits=5, purge_gap=5):
        self.n_splits = n_splits
        self.purge_gap = purge_gap

    def split(self, X, y=None, groups=None):
        n = len(X)
        test_size = n // (self.n_splits + 1)

        for i in range(self.n_splits):
            test_start = (i + 1) * test_size
            test_end = test_start + test_size

            train_end = test_start - self.purge_gap
            train_indices = list(range(0, max(0, train_end)))
            test_indices = list(range(test_start, min(test_end, n)))

            if train_indices and test_indices:
                yield train_indices, test_indices

    def get_n_splits(self, X=None, y=None, groups=None):
        return self.n_splits
```

---

## Ensemble Methods

### Stacking

```python
from sklearn.ensemble import StackingClassifier, StackingRegressor, VotingClassifier


def build_stacking_classifier(X_train, y_train):
    """Build a stacking ensemble with diverse base models."""

    base_estimators = [
        ("lr", LogisticRegression(max_iter=1000, C=0.1)),
        ("rf", RandomForestClassifier(n_estimators=200, max_depth=8, random_state=42)),
        ("xgb", xgb.XGBClassifier(
            n_estimators=200, max_depth=6, learning_rate=0.1,
            tree_method="hist", random_state=42
        )),
        ("lgb", lgb.LGBMClassifier(
            n_estimators=200, max_depth=6, learning_rate=0.1,
            random_state=42, verbose=-1
        )),
        ("svm", SVC(kernel="rbf", probability=True, random_state=42)),
    ]

    stacking = StackingClassifier(
        estimators=base_estimators,
        final_estimator=LogisticRegression(max_iter=1000),
        cv=5,
        stack_method="predict_proba",
        n_jobs=-1,
        passthrough=False,  # Set True to include original features
    )

    stacking.fit(X_train, y_train)
    return stacking


# Voting classifier
voting = VotingClassifier(
    estimators=[
        ("rf", RandomForestClassifier(n_estimators=200, random_state=42)),
        ("xgb", xgb.XGBClassifier(n_estimators=200, tree_method="hist", random_state=42)),
        ("lgb", lgb.LGBMClassifier(n_estimators=200, random_state=42, verbose=-1)),
    ],
    voting="soft",  # Use probability-based voting
    weights=[1, 2, 2],  # Weight XGB and LGB higher
    n_jobs=-1,
)


# Manual blending
class Blender:
    """Manual blending ensemble with optimized weights."""

    def __init__(self, models, optimize_weights=True):
        self.models = models
        self.optimize_weights = optimize_weights
        self.weights = None

    def fit(self, X_train, y_train, X_val, y_val):
        self.fitted_models = []
        val_preds = []

        for name, model in self.models:
            model.fit(X_train, y_train)
            self.fitted_models.append((name, model))
            val_preds.append(model.predict_proba(X_val)[:, 1])

        val_preds = np.array(val_preds).T  # (n_samples, n_models)

        if self.optimize_weights:
            from scipy.optimize import minimize
            from sklearn.metrics import roc_auc_score

            def neg_auc(weights):
                weighted = np.average(val_preds, axis=1, weights=weights)
                return -roc_auc_score(y_val, weighted)

            n_models = len(self.models)
            result = minimize(
                neg_auc,
                x0=np.ones(n_models) / n_models,
                bounds=[(0, 1)] * n_models,
                constraints={"type": "eq", "fun": lambda w: w.sum() - 1},
            )
            self.weights = result.x
            print(f"Optimized weights: {dict(zip([n for n, _ in self.models], self.weights))}")
        else:
            self.weights = np.ones(len(self.models)) / len(self.models)

        return self

    def predict_proba(self, X):
        preds = np.array([
            model.predict_proba(X)[:, 1] for _, model in self.fitted_models
        ]).T
        return np.average(preds, axis=1, weights=self.weights)
```

---

## Imbalanced Data

```python
from imblearn.over_sampling import SMOTE, ADASYN, BorderlineSMOTE
from imblearn.under_sampling import RandomUnderSampler, TomekLinks
from imblearn.combine import SMOTETomek, SMOTEENN
from imblearn.pipeline import Pipeline as ImbPipeline


# Imbalanced pipeline with SMOTE
imb_pipeline = ImbPipeline([
    ("preprocessor", preprocessor),
    ("smote", SMOTE(sampling_strategy=0.5, random_state=42)),
    ("model", xgb.XGBClassifier(
        scale_pos_weight=sum(y == 0) / sum(y == 1),
        tree_method="hist", random_state=42,
    )),
])

# Class weight approach (no resampling needed)
from sklearn.utils.class_weight import compute_class_weight

class_weights = compute_class_weight("balanced", classes=np.unique(y), y=y)
class_weight_dict = dict(zip(np.unique(y), class_weights))

model = RandomForestClassifier(
    class_weight="balanced",  # Or pass class_weight_dict
    n_estimators=200,
    random_state=42,
)

# Threshold tuning
from sklearn.metrics import precision_recall_curve

y_proba = model.predict_proba(X_test)[:, 1]
precisions, recalls, thresholds = precision_recall_curve(y_test, y_proba)

# Find threshold for target recall
target_recall = 0.90
idx = np.argmin(np.abs(recalls - target_recall))
optimal_threshold = thresholds[idx]
print(f"Threshold for {target_recall:.0%} recall: {optimal_threshold:.3f}")
print(f"Precision at this threshold: {precisions[idx]:.3f}")
```

---

## Model Evaluation

### Classification Metrics

```python
from sklearn.metrics import (
    classification_report, confusion_matrix,
    roc_auc_score, average_precision_score,
    f1_score, precision_score, recall_score,
    roc_curve, precision_recall_curve,
    log_loss, brier_score_loss,
    matthews_corrcoef, cohen_kappa_score,
)


def evaluate_classifier(model, X_test, y_test, class_names=None):
    """Comprehensive classification evaluation."""
    y_pred = model.predict(X_test)
    y_proba = model.predict_proba(X_test)[:, 1] if hasattr(model, "predict_proba") else None

    print("Classification Report:")
    print(classification_report(y_test, y_pred, target_names=class_names))

    print(f"\nConfusion Matrix:\n{confusion_matrix(y_test, y_pred)}")

    if y_proba is not None:
        print(f"\nROC AUC: {roc_auc_score(y_test, y_proba):.4f}")
        print(f"Average Precision: {average_precision_score(y_test, y_proba):.4f}")
        print(f"Log Loss: {log_loss(y_test, y_proba):.4f}")
        print(f"Brier Score: {brier_score_loss(y_test, y_proba):.4f}")

    print(f"Matthews Corrcoef: {matthews_corrcoef(y_test, y_pred):.4f}")
    print(f"Cohen's Kappa: {cohen_kappa_score(y_test, y_pred):.4f}")


# Multiclass evaluation
from sklearn.metrics import roc_auc_score

y_proba_multi = model.predict_proba(X_test)
auc_ovr = roc_auc_score(y_test, y_proba_multi, multi_class="ovr", average="macro")
auc_ovo = roc_auc_score(y_test, y_proba_multi, multi_class="ovo", average="macro")
```

### Regression Metrics

```python
from sklearn.metrics import (
    mean_squared_error, mean_absolute_error,
    r2_score, mean_absolute_percentage_error,
    median_absolute_error, max_error,
)


def evaluate_regressor(model, X_test, y_test):
    """Comprehensive regression evaluation."""
    y_pred = model.predict(X_test)

    mse = mean_squared_error(y_test, y_pred)
    rmse = np.sqrt(mse)
    mae = mean_absolute_error(y_test, y_pred)
    r2 = r2_score(y_test, y_pred)
    mape = mean_absolute_percentage_error(y_test, y_pred)
    medae = median_absolute_error(y_test, y_pred)

    print(f"RMSE:  {rmse:.4f}")
    print(f"MAE:   {mae:.4f}")
    print(f"R2:    {r2:.4f}")
    print(f"MAPE:  {mape:.4f}")
    print(f"MedAE: {medae:.4f}")

    # Residual analysis
    residuals = y_test - y_pred
    print(f"\nResidual Stats:")
    print(f"  Mean: {residuals.mean():.4f}")
    print(f"  Std:  {residuals.std():.4f}")
    print(f"  Skew: {pd.Series(residuals).skew():.4f}")
```

---

## Clustering and Unsupervised Learning

```python
from sklearn.cluster import KMeans, DBSCAN, AgglomerativeClustering
from sklearn.mixture import GaussianMixture
from sklearn.metrics import silhouette_score, calinski_harabasz_score, davies_bouldin_score


def find_optimal_clusters(X, max_k=15):
    """Find optimal number of clusters using multiple methods."""

    inertias = []
    silhouette_scores = []
    calinski_scores = []
    davies_scores = []

    for k in range(2, max_k + 1):
        kmeans = KMeans(n_clusters=k, n_init=10, random_state=42)
        labels = kmeans.fit_predict(X)

        inertias.append(kmeans.inertia_)
        silhouette_scores.append(silhouette_score(X, labels))
        calinski_scores.append(calinski_harabasz_score(X, labels))
        davies_scores.append(davies_bouldin_score(X, labels))

    results = pd.DataFrame({
        "k": range(2, max_k + 1),
        "inertia": inertias,
        "silhouette": silhouette_scores,
        "calinski": calinski_scores,
        "davies_bouldin": davies_scores,
    })

    best_k = results.loc[results["silhouette"].idxmax(), "k"]
    print(f"Best k by silhouette score: {int(best_k)}")
    return results


# DBSCAN with automatic epsilon selection
from sklearn.neighbors import NearestNeighbors

def find_dbscan_eps(X, min_samples=5):
    """Find optimal eps for DBSCAN using k-distance graph."""
    nn = NearestNeighbors(n_neighbors=min_samples)
    nn.fit(X)
    distances, _ = nn.kneighbors(X)
    distances = np.sort(distances[:, -1])

    # Find the "elbow" point
    from kneed import KneeLocator
    kneedle = KneeLocator(
        range(len(distances)), distances, curve="convex", direction="increasing"
    )
    eps = distances[kneedle.knee] if kneedle.knee else np.median(distances)
    print(f"Suggested eps: {eps:.4f}")
    return eps
```

---

## Dimensionality Reduction

```python
from sklearn.decomposition import PCA, TruncatedSVD
from sklearn.manifold import TSNE
import umap


def reduce_dimensions(X, method="pca", n_components=2, **kwargs):
    """Reduce dimensionality with various methods."""
    if method == "pca":
        reducer = PCA(n_components=n_components, random_state=42)
    elif method == "tsne":
        reducer = TSNE(
            n_components=n_components, perplexity=kwargs.get("perplexity", 30),
            learning_rate=kwargs.get("learning_rate", "auto"),
            n_iter=kwargs.get("n_iter", 1000), random_state=42,
        )
    elif method == "umap":
        reducer = umap.UMAP(
            n_components=n_components,
            n_neighbors=kwargs.get("n_neighbors", 15),
            min_dist=kwargs.get("min_dist", 0.1),
            random_state=42,
        )
    elif method == "svd":
        reducer = TruncatedSVD(n_components=n_components, random_state=42)
    else:
        raise ValueError(f"Unknown method: {method}")

    X_reduced = reducer.fit_transform(X)

    if method == "pca":
        explained = reducer.explained_variance_ratio_
        cumulative = np.cumsum(explained)
        print(f"Explained variance: {explained[:5]}")
        print(f"Cumulative: {cumulative[:5]}")

    return X_reduced, reducer


# Determine optimal PCA components
def find_pca_components(X, variance_threshold=0.95):
    """Find number of PCA components for target variance."""
    pca = PCA(random_state=42)
    pca.fit(X)

    cumulative = np.cumsum(pca.explained_variance_ratio_)
    n_components = np.argmax(cumulative >= variance_threshold) + 1
    print(f"Components for {variance_threshold:.0%} variance: {n_components}")
    return n_components
```

---

## Model Persistence

```python
import joblib
import pickle
from pathlib import Path


def save_model_bundle(model, preprocessor, metadata, path):
    """Save model with preprocessor and metadata."""
    bundle = {
        "model": model,
        "preprocessor": preprocessor,
        "metadata": metadata,
        "feature_names": metadata.get("feature_names", []),
        "version": metadata.get("version", "1.0.0"),
    }
    Path(path).parent.mkdir(parents=True, exist_ok=True)
    joblib.dump(bundle, path, compress=3)
    print(f"Model bundle saved: {path}")


def load_model_bundle(path):
    """Load model bundle."""
    bundle = joblib.load(path)
    print(f"Loaded model v{bundle.get('version', 'unknown')}")
    return bundle


# Usage
metadata = {
    "feature_names": numeric_features + categorical_features,
    "version": "1.0.0",
    "training_date": "2024-01-15",
    "metrics": {"roc_auc": 0.95, "f1": 0.87},
    "n_samples": len(X_train),
}
save_model_bundle(pipeline, preprocessor, metadata, "models/churn_v1.joblib")
```

---

## Best Practices Summary

### ML Workflow Checklist

1. **EDA first** — Understand distributions, correlations, missing values, class balance
2. **Split early** — Train/val/test split BEFORE any preprocessing to prevent leakage
3. **Baseline model** — Start with a simple model (logistic regression, decision tree)
4. **Pipeline everything** — Use sklearn Pipeline to bundle preprocessing + model
5. **Proper CV** — Match CV strategy to data (stratified, time series, grouped)
6. **Metric selection** — Choose metrics aligned with business goals, not just accuracy
7. **Feature importance** — Understand what drives predictions (SHAP, permutation)
8. **Hyperparameter tuning** — Use Optuna with pruning for efficient search
9. **Ensemble if needed** — Stacking/blending for marginal gains in competitions
10. **Document and version** — Track experiments, save model bundles with metadata

### Common Pitfalls

- Training on all data without a holdout test set
- Applying transformations before splitting (data leakage)
- Using accuracy for imbalanced datasets (use F1, PR-AUC)
- Not setting `random_state` for reproducibility
- Over-tuning on validation set (use nested cross-validation)
- Ignoring feature importance and model interpretability
- Not checking for multicollinearity in linear models
- Using one-hot encoding for high-cardinality features (use target encoding)

### Algorithm Selection Quick Guide

| Scenario | Best Choice |
|----------|------------|
| Binary classification, < 10K samples | Logistic Regression, SVM |
| Binary classification, > 10K samples | XGBoost, LightGBM |
| Multiclass, many classes | LightGBM, Neural Networks |
| Regression, linear relationships | Ridge, ElasticNet |
| Regression, nonlinear | XGBoost, Random Forest |
| High-dimensional, sparse | Lasso, ElasticNet, SGD |
| Interpretability required | Logistic Regression, Decision Tree, EBM |
| Time series | LightGBM with lag features, Prophet |
| Anomaly detection | Isolation Forest, LOF |
| Clustering | KMeans, DBSCAN, GMM |
