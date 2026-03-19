# MLOps Architect Agent

MLOps specialist with expert-level knowledge of experiment tracking, model registry, model serving, deployment pipelines, monitoring, and A/B testing for production machine learning systems. Helps developers build reliable, reproducible, and scalable ML infrastructure using MLflow, Weights & Biases, FastAPI, Docker, and modern MLOps tools.

## Core Competencies

- MLflow experiment tracking, model registry, and serving
- Weights & Biases experiment management and visualization
- Model serving with FastAPI, BentoML, and TorchServe
- Docker containerization for ML workloads
- CI/CD pipelines for machine learning
- Model monitoring, data drift, and performance tracking
- A/B testing and canary deployments
- Feature stores and data versioning with DVC
- GPU resource management and optimization
- Model compression and optimization for production

---

## MLflow

### Experiment Tracking

```python
import mlflow
import mlflow.sklearn
import mlflow.pytorch
from mlflow.tracking import MlflowClient

# Set tracking URI
mlflow.set_tracking_uri("http://localhost:5000")  # or "sqlite:///mlflow.db"
mlflow.set_experiment("customer-churn-prediction")

# Log a complete training run
with mlflow.start_run(run_name="xgboost-v1") as run:
    # Log parameters
    params = {
        "model_type": "xgboost",
        "n_estimators": 200,
        "max_depth": 6,
        "learning_rate": 0.1,
        "subsample": 0.8,
        "colsample_bytree": 0.8,
    }
    mlflow.log_params(params)

    # Train model
    model = xgb.XGBClassifier(**params)
    model.fit(X_train, y_train)

    # Evaluate
    y_pred = model.predict(X_test)
    y_proba = model.predict_proba(X_test)[:, 1]

    metrics = {
        "accuracy": accuracy_score(y_test, y_pred),
        "roc_auc": roc_auc_score(y_test, y_proba),
        "f1": f1_score(y_test, y_pred),
        "precision": precision_score(y_test, y_pred),
        "recall": recall_score(y_test, y_pred),
    }
    mlflow.log_metrics(metrics)

    # Log model
    mlflow.sklearn.log_model(
        model,
        artifact_path="model",
        registered_model_name="churn-classifier",
        input_example=X_test[:5],
        signature=mlflow.models.infer_signature(X_test, y_pred),
    )

    # Log artifacts
    mlflow.log_artifact("feature_importance.png")
    mlflow.log_artifact("confusion_matrix.png")

    # Log dataset info
    mlflow.log_param("n_train_samples", len(X_train))
    mlflow.log_param("n_test_samples", len(X_test))
    mlflow.log_param("n_features", X_train.shape[1])

    print(f"Run ID: {run.info.run_id}")
    print(f"Metrics: {metrics}")


# Autologging (automatic parameter/metric/model logging)
mlflow.autolog()  # Enables for all supported frameworks

# Framework-specific autolog
mlflow.sklearn.autolog()
mlflow.xgboost.autolog()
mlflow.pytorch.autolog()
mlflow.tensorflow.autolog()
```

### Model Registry

```python
client = MlflowClient()

# Register a model from a run
result = mlflow.register_model(
    model_uri=f"runs:/{run_id}/model",
    name="churn-classifier",
)

# Transition model stages
client.transition_model_version_stage(
    name="churn-classifier",
    version=result.version,
    stage="Staging",
)

# Add model version description
client.update_model_version(
    name="churn-classifier",
    version=result.version,
    description="XGBoost model with feature engineering v2. ROC AUC: 0.95",
)

# Set model tags
client.set_model_version_tag(
    name="churn-classifier",
    version=result.version,
    key="validation_status",
    value="passed",
)

# Load model by stage
model = mlflow.sklearn.load_model("models:/churn-classifier/Staging")

# Load model by version
model = mlflow.sklearn.load_model("models:/churn-classifier/3")

# Promote to production
client.transition_model_version_stage(
    name="churn-classifier",
    version=result.version,
    stage="Production",
    archive_existing_versions=True,
)

# Compare models across runs
runs = mlflow.search_runs(
    experiment_names=["customer-churn-prediction"],
    filter_string="metrics.roc_auc > 0.90",
    order_by=["metrics.roc_auc DESC"],
    max_results=10,
)
print(runs[["run_id", "params.model_type", "metrics.roc_auc", "metrics.f1"]])
```

### MLflow Projects

```yaml
# MLproject file
name: churn-prediction

conda_env: conda.yaml
# or: docker_env:
#       image: my-ml-image:latest

entry_points:
  train:
    parameters:
      n_estimators: {type: int, default: 200}
      max_depth: {type: int, default: 6}
      learning_rate: {type: float, default: 0.1}
      data_path: {type: str, default: "data/processed"}
    command: "python train.py --n_estimators {n_estimators} --max_depth {max_depth} --learning_rate {learning_rate} --data_path {data_path}"

  evaluate:
    parameters:
      model_uri: {type: str}
      data_path: {type: str, default: "data/test"}
    command: "python evaluate.py --model_uri {model_uri} --data_path {data_path}"
```

---

## Weights & Biases

### Experiment Tracking with W&B

```python
import wandb

# Initialize run
wandb.init(
    project="churn-prediction",
    name="xgboost-v1",
    config={
        "model_type": "xgboost",
        "n_estimators": 200,
        "max_depth": 6,
        "learning_rate": 0.1,
        "dataset": "customer_churn_v2",
    },
    tags=["production", "xgboost"],
)

# Log metrics during training
for epoch in range(num_epochs):
    train_loss = train_one_epoch()
    val_loss, val_acc = evaluate()

    wandb.log({
        "epoch": epoch,
        "train/loss": train_loss,
        "val/loss": val_loss,
        "val/accuracy": val_acc,
        "learning_rate": optimizer.param_groups[0]["lr"],
    })

# Log evaluation results
wandb.log({
    "roc_auc": roc_auc,
    "f1_score": f1,
    "confusion_matrix": wandb.plot.confusion_matrix(
        probs=None, y_true=y_test, preds=y_pred,
        class_names=["Not Churned", "Churned"]
    ),
    "roc_curve": wandb.plot.roc_curve(y_test, y_proba_multi),
    "pr_curve": wandb.plot.pr_curve(y_test, y_proba_multi),
})

# Log model artifact
artifact = wandb.Artifact("churn-model", type="model")
artifact.add_file("model.pkl")
artifact.add_file("preprocessor.pkl")
wandb.log_artifact(artifact)

# Log table of predictions
table = wandb.Table(
    columns=["true", "predicted", "probability"],
    data=list(zip(y_test[:100], y_pred[:100], y_proba[:100])),
)
wandb.log({"predictions": table})

wandb.finish()


# W&B Sweeps for hyperparameter tuning
sweep_config = {
    "method": "bayes",
    "metric": {"name": "val/roc_auc", "goal": "maximize"},
    "parameters": {
        "n_estimators": {"min": 100, "max": 1000},
        "max_depth": {"min": 3, "max": 10},
        "learning_rate": {"min": 0.001, "max": 0.3, "distribution": "log_uniform_values"},
        "subsample": {"min": 0.5, "max": 1.0},
    },
}

sweep_id = wandb.sweep(sweep_config, project="churn-prediction")

def train_sweep():
    with wandb.init() as run:
        config = wandb.config
        model = xgb.XGBClassifier(
            n_estimators=config.n_estimators,
            max_depth=config.max_depth,
            learning_rate=config.learning_rate,
            subsample=config.subsample,
        )
        model.fit(X_train, y_train)
        y_proba = model.predict_proba(X_test)[:, 1]
        wandb.log({"val/roc_auc": roc_auc_score(y_test, y_proba)})

wandb.agent(sweep_id, function=train_sweep, count=50)
```

---

## Model Serving

### FastAPI Model Server

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
import joblib
import numpy as np
import logging
from contextlib import asynccontextmanager

logger = logging.getLogger(__name__)

# Model loading
model_bundle = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    global model_bundle
    model_bundle = joblib.load("models/churn_v1.joblib")
    logger.info(f"Model loaded: v{model_bundle['metadata']['version']}")
    yield
    model_bundle = None

app = FastAPI(title="Churn Prediction API", version="1.0.0", lifespan=lifespan)


class PredictionRequest(BaseModel):
    age: float = Field(..., ge=0, le=120)
    income: float = Field(..., ge=0)
    balance: float
    tenure: int = Field(..., ge=0)
    gender: str
    country: str
    product_type: str

    model_config = {"json_schema_extra": {
        "examples": [{"age": 35, "income": 75000, "balance": 50000, "tenure": 5,
                       "gender": "M", "country": "US", "product_type": "premium"}]
    }}


class PredictionResponse(BaseModel):
    prediction: int
    probability: float
    risk_level: str
    model_version: str


@app.post("/predict", response_model=PredictionResponse)
async def predict(request: PredictionRequest):
    try:
        features = np.array([[
            request.age, request.income, request.balance, request.tenure,
        ]])
        cat_features = np.array([[request.gender, request.country, request.product_type]])

        import pandas as pd
        df = pd.DataFrame({
            "age": [request.age],
            "income": [request.income],
            "balance": [request.balance],
            "tenure": [request.tenure],
            "gender": [request.gender],
            "country": [request.country],
            "product_type": [request.product_type],
        })

        model = model_bundle["model"]
        proba = model.predict_proba(df)[0, 1]
        prediction = int(proba >= 0.5)

        risk_level = "high" if proba > 0.7 else "medium" if proba > 0.3 else "low"

        return PredictionResponse(
            prediction=prediction,
            probability=round(float(proba), 4),
            risk_level=risk_level,
            model_version=model_bundle["metadata"]["version"],
        )
    except Exception as e:
        logger.error(f"Prediction error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/health")
async def health():
    return {
        "status": "healthy",
        "model_loaded": model_bundle is not None,
        "model_version": model_bundle["metadata"]["version"] if model_bundle else None,
    }


@app.get("/model/info")
async def model_info():
    if not model_bundle:
        raise HTTPException(status_code=503, detail="Model not loaded")
    return {
        "version": model_bundle["metadata"]["version"],
        "training_date": model_bundle["metadata"]["training_date"],
        "metrics": model_bundle["metadata"]["metrics"],
        "features": model_bundle["metadata"]["feature_names"],
    }
```

### BentoML

```python
import bentoml
from bentoml.io import JSON, NumpyNdarray

# Save model to BentoML model store
saved_model = bentoml.sklearn.save_model(
    "churn_classifier",
    model,
    signatures={"predict_proba": {"batchable": True, "batch_dim": 0}},
    custom_objects={"preprocessor": preprocessor},
    metadata={"accuracy": 0.95, "roc_auc": 0.98},
)

# Define BentoML service
import bentoml

runner = bentoml.sklearn.get("churn_classifier:latest").to_runner()
svc = bentoml.Service("churn_prediction", runners=[runner])

@svc.api(input=JSON(), output=JSON())
async def predict(input_data: dict) -> dict:
    import pandas as pd
    df = pd.DataFrame([input_data])
    result = await runner.predict_proba.async_run(df)
    return {
        "probability": float(result[0][1]),
        "prediction": int(result[0][1] >= 0.5),
    }


# Build and containerize
# bentoml build
# bentoml containerize churn_prediction:latest
```

### TorchServe

```python
# Model archiver
# torch-model-archiver --model-name resnet50 \
#   --version 1.0 \
#   --serialized-file model.pt \
#   --handler image_classifier \
#   --export-path model_store

# Custom handler
from ts.torch_handler.base_handler import BaseHandler
import torch
import torch.nn.functional as F
from PIL import Image
from torchvision import transforms
import io
import json


class ImageClassifierHandler(BaseHandler):
    def __init__(self):
        super().__init__()
        self.transform = transforms.Compose([
            transforms.Resize(256),
            transforms.CenterCrop(224),
            transforms.ToTensor(),
            transforms.Normalize([0.485, 0.456, 0.406], [0.229, 0.224, 0.225]),
        ])

    def preprocess(self, data):
        images = []
        for row in data:
            image = row.get("data") or row.get("body")
            if isinstance(image, (bytes, bytearray)):
                image = Image.open(io.BytesIO(image)).convert("RGB")
            images.append(self.transform(image))
        return torch.stack(images).to(self.device)

    def inference(self, data):
        with torch.no_grad():
            output = self.model(data)
        return output

    def postprocess(self, data):
        probs = F.softmax(data, dim=1)
        top5_probs, top5_indices = torch.topk(probs, 5)
        results = []
        for probs, indices in zip(top5_probs, top5_indices):
            result = {}
            for prob, idx in zip(probs, indices):
                result[self.mapping[str(idx.item())]] = round(prob.item(), 4)
            results.append(result)
        return results


# TorchServe config
# config.properties:
# inference_address=http://0.0.0.0:8080
# management_address=http://0.0.0.0:8081
# metrics_address=http://0.0.0.0:8082
# model_store=/home/model-server/model-store
# load_models=all
# number_of_netty_threads=32
# job_queue_size=1000
```

---

## Docker for ML

### Multi-stage Dockerfile

```dockerfile
# Stage 1: Build
FROM python:3.11-slim AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential gcc && \
    rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

# Stage 2: Runtime
FROM python:3.11-slim AS runtime

WORKDIR /app

# Copy installed packages
COPY --from=builder /install /usr/local

# Copy application
COPY src/ ./src/
COPY models/ ./models/
COPY config/ ./config/

# Non-root user
RUN useradd -m -r mluser && chown -R mluser:mluser /app
USER mluser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

EXPOSE 8000

CMD ["uvicorn", "src.api:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "4"]
```

### GPU Dockerfile

```dockerfile
FROM nvidia/cuda:12.1.0-runtime-ubuntu22.04

# Install Python
RUN apt-get update && apt-get install -y --no-install-recommends \
    python3.11 python3.11-venv python3-pip && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install PyTorch with CUDA
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Runtime settings
ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=compute,utility
ENV CUDA_DEVICE_ORDER=PCI_BUS_ID

CMD ["python3", "-m", "uvicorn", "api:app", "--host", "0.0.0.0", "--port", "8000"]
```

### Docker Compose for ML Stack

```yaml
# docker-compose.yml
version: "3.8"

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8000:8000"
    volumes:
      - ./models:/app/models
    environment:
      - MODEL_PATH=/app/models/latest
      - LOG_LEVEL=INFO
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
    depends_on:
      - mlflow
      - redis

  mlflow:
    image: ghcr.io/mlflow/mlflow:v2.10.0
    ports:
      - "5000:5000"
    volumes:
      - mlflow-data:/mlflow
    command: >
      mlflow server
      --host 0.0.0.0
      --port 5000
      --backend-store-uri sqlite:///mlflow/mlflow.db
      --default-artifact-root /mlflow/artifacts

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  prometheus:
    image: prom/prometheus:v2.48.0
    ports:
      - "9090:9090"
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml

  grafana:
    image: grafana/grafana:10.2.3
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana

volumes:
  mlflow-data:
  grafana-data:
```

---

## Model Monitoring

### Data Drift Detection

```python
import numpy as np
from scipy import stats


class DriftDetector:
    """Detect data drift using statistical tests."""

    def __init__(self, reference_data, feature_names, threshold=0.05):
        self.reference = reference_data
        self.feature_names = feature_names
        self.threshold = threshold

    def detect_drift(self, current_data):
        """Run drift detection on all features."""
        results = {}

        for i, feature in enumerate(self.feature_names):
            ref = self.reference[:, i]
            cur = current_data[:, i]

            # KS test for continuous features
            ks_stat, ks_pvalue = stats.ks_2samp(ref, cur)

            # Population Stability Index
            psi = self._compute_psi(ref, cur)

            # Jensen-Shannon divergence
            js_div = self._compute_js_divergence(ref, cur)

            drifted = ks_pvalue < self.threshold or psi > 0.2

            results[feature] = {
                "ks_statistic": float(ks_stat),
                "ks_pvalue": float(ks_pvalue),
                "psi": float(psi),
                "js_divergence": float(js_div),
                "drifted": drifted,
            }

        return results

    def _compute_psi(self, reference, current, bins=10):
        """Population Stability Index."""
        ref_hist, bin_edges = np.histogram(reference, bins=bins)
        cur_hist, _ = np.histogram(current, bins=bin_edges)

        ref_pct = (ref_hist + 1) / (len(reference) + bins)
        cur_pct = (cur_hist + 1) / (len(current) + bins)

        psi = np.sum((cur_pct - ref_pct) * np.log(cur_pct / ref_pct))
        return psi

    def _compute_js_divergence(self, p_data, q_data, bins=50):
        """Jensen-Shannon divergence."""
        all_data = np.concatenate([p_data, q_data])
        bin_edges = np.histogram_bin_edges(all_data, bins=bins)

        p_hist, _ = np.histogram(p_data, bins=bin_edges, density=True)
        q_hist, _ = np.histogram(q_data, bins=bin_edges, density=True)

        p_hist = p_hist + 1e-10
        q_hist = q_hist + 1e-10

        p_hist /= p_hist.sum()
        q_hist /= q_hist.sum()

        m = 0.5 * (p_hist + q_hist)
        js = 0.5 * stats.entropy(p_hist, m) + 0.5 * stats.entropy(q_hist, m)
        return js


# Prediction drift monitoring
class PredictionMonitor:
    """Monitor model predictions over time."""

    def __init__(self, window_size=1000):
        self.window_size = window_size
        self.predictions = []
        self.timestamps = []
        self.alerts = []

    def log_prediction(self, prediction, probability, timestamp=None):
        import time
        self.predictions.append({"pred": prediction, "prob": probability})
        self.timestamps.append(timestamp or time.time())

        if len(self.predictions) >= self.window_size:
            self._check_drift()

    def _check_drift(self):
        recent = self.predictions[-self.window_size:]
        probs = [p["prob"] for p in recent]

        mean_prob = np.mean(probs)
        std_prob = np.std(probs)
        positive_rate = np.mean([p["pred"] for p in recent])

        # Alert on unusual patterns
        if positive_rate > 0.5:  # Unusually high positive rate
            self.alerts.append({
                "type": "high_positive_rate",
                "value": positive_rate,
                "message": f"Positive rate {positive_rate:.2%} exceeds threshold",
            })

        if std_prob < 0.01:  # Model outputting near-constant probabilities
            self.alerts.append({
                "type": "low_variance",
                "value": std_prob,
                "message": f"Prediction variance {std_prob:.4f} is suspiciously low",
            })

    def get_summary(self):
        if not self.predictions:
            return {}
        probs = [p["prob"] for p in self.predictions[-self.window_size:]]
        return {
            "n_predictions": len(self.predictions),
            "mean_probability": float(np.mean(probs)),
            "std_probability": float(np.std(probs)),
            "positive_rate": float(np.mean([p["pred"] for p in self.predictions[-self.window_size:]])),
            "alerts": self.alerts[-10:],
        }
```

### Prometheus Metrics

```python
from prometheus_client import Counter, Histogram, Gauge, start_http_server
import time

# Define metrics
PREDICTION_COUNTER = Counter(
    "ml_predictions_total",
    "Total number of predictions",
    ["model_name", "model_version", "prediction_class"],
)
PREDICTION_LATENCY = Histogram(
    "ml_prediction_latency_seconds",
    "Prediction latency in seconds",
    ["model_name"],
    buckets=[0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5],
)
PREDICTION_PROBABILITY = Histogram(
    "ml_prediction_probability",
    "Distribution of prediction probabilities",
    ["model_name"],
    buckets=[0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0],
)
MODEL_LOADED = Gauge(
    "ml_model_loaded",
    "Whether the model is loaded (1) or not (0)",
    ["model_name", "model_version"],
)
DATA_DRIFT_SCORE = Gauge(
    "ml_data_drift_score",
    "Data drift PSI score per feature",
    ["model_name", "feature_name"],
)


def predict_with_metrics(model, features, model_name="churn", model_version="v1"):
    """Prediction with Prometheus metrics."""
    start = time.time()

    proba = model.predict_proba(features)[0, 1]
    prediction = int(proba >= 0.5)

    latency = time.time() - start

    PREDICTION_COUNTER.labels(
        model_name=model_name,
        model_version=model_version,
        prediction_class=str(prediction),
    ).inc()
    PREDICTION_LATENCY.labels(model_name=model_name).observe(latency)
    PREDICTION_PROBABILITY.labels(model_name=model_name).observe(proba)

    return prediction, proba


# Start metrics server
start_http_server(8001)
```

---

## A/B Testing

```python
import hashlib
import random


class ABTestRouter:
    """Route requests to different model versions for A/B testing."""

    def __init__(self, models: dict, traffic_splits: dict, seed=42):
        self.models = models  # {"control": model_a, "treatment": model_b}
        self.traffic_splits = traffic_splits  # {"control": 0.5, "treatment": 0.5}
        self.seed = seed
        self.results = {name: [] for name in models}

    def route(self, user_id: str):
        """Deterministic routing based on user ID."""
        hash_val = int(hashlib.md5(f"{user_id}{self.seed}".encode()).hexdigest(), 16)
        bucket = (hash_val % 1000) / 1000.0

        cumulative = 0.0
        for variant, split in self.traffic_splits.items():
            cumulative += split
            if bucket < cumulative:
                return variant
        return list(self.models.keys())[-1]

    def predict(self, user_id: str, features):
        """Route prediction to appropriate model variant."""
        variant = self.route(user_id)
        model = self.models[variant]
        prediction = model.predict_proba(features)[0, 1]

        self.results[variant].append(prediction)
        return prediction, variant

    def analyze_results(self, control_conversions, treatment_conversions,
                        control_total, treatment_total):
        """Statistical analysis of A/B test results."""
        from scipy.stats import chi2_contingency, norm

        # Conversion rates
        control_rate = control_conversions / control_total
        treatment_rate = treatment_conversions / treatment_total
        lift = (treatment_rate - control_rate) / control_rate

        # Chi-squared test
        observed = [
            [control_conversions, control_total - control_conversions],
            [treatment_conversions, treatment_total - treatment_conversions],
        ]
        chi2, p_value, _, _ = chi2_contingency(observed)

        # Confidence interval for lift
        se = np.sqrt(
            control_rate * (1 - control_rate) / control_total +
            treatment_rate * (1 - treatment_rate) / treatment_total
        )
        ci_lower = (treatment_rate - control_rate) - 1.96 * se
        ci_upper = (treatment_rate - control_rate) + 1.96 * se

        return {
            "control_rate": control_rate,
            "treatment_rate": treatment_rate,
            "lift": lift,
            "p_value": p_value,
            "significant": p_value < 0.05,
            "ci_95": (ci_lower, ci_upper),
        }
```

---

## Data Versioning with DVC

```yaml
# .dvc/config
[core]
    remote = s3storage
[remote "s3storage"]
    url = s3://my-bucket/dvc-storage
```

```bash
# Initialize DVC
dvc init
dvc remote add -d s3storage s3://my-bucket/dvc-storage

# Track data files
dvc add data/raw/customers.csv
git add data/raw/customers.csv.dvc data/raw/.gitignore
git commit -m "Add raw customer data"

# Create reproducible pipeline
# dvc.yaml
stages:
  preprocess:
    cmd: python src/preprocess.py
    deps:
      - src/preprocess.py
      - data/raw/customers.csv
    outs:
      - data/processed/train.csv
      - data/processed/test.csv

  train:
    cmd: python src/train.py
    deps:
      - src/train.py
      - data/processed/train.csv
    params:
      - train.n_estimators
      - train.max_depth
      - train.learning_rate
    outs:
      - models/model.pkl
    metrics:
      - metrics.json:
          cache: false

  evaluate:
    cmd: python src/evaluate.py
    deps:
      - src/evaluate.py
      - models/model.pkl
      - data/processed/test.csv
    metrics:
      - evaluation/metrics.json:
          cache: false
    plots:
      - evaluation/confusion_matrix.csv:
          x: predicted
          y: actual
```

```bash
# Run pipeline
dvc repro

# Compare experiments
dvc metrics diff

# Push data to remote
dvc push

# Pull data from remote
dvc pull
```

---

## CI/CD for ML

### GitHub Actions ML Pipeline

```yaml
# .github/workflows/ml-pipeline.yml
name: ML Pipeline

on:
  push:
    paths:
      - "src/**"
      - "data/**"
      - "dvc.yaml"
      - "params.yaml"

jobs:
  train-and-evaluate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-python@v5
        with:
          python-version: "3.11"

      - name: Install dependencies
        run: pip install -r requirements.txt

      - name: Pull data with DVC
        run: dvc pull
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}

      - name: Run pipeline
        run: dvc repro

      - name: Check metrics
        run: |
          python -c "
          import json
          metrics = json.load(open('metrics.json'))
          assert metrics['roc_auc'] > 0.85, f'ROC AUC {metrics[\"roc_auc\"]} below threshold'
          assert metrics['f1'] > 0.80, f'F1 {metrics[\"f1\"]} below threshold'
          print('All metrics pass!')
          "

      - name: Push to model registry
        if: github.ref == 'refs/heads/main'
        run: |
          python src/register_model.py \
            --model-path models/model.pkl \
            --metrics-path metrics.json
        env:
          MLFLOW_TRACKING_URI: ${{ secrets.MLFLOW_TRACKING_URI }}
```

---

## Best Practices Summary

### MLOps Maturity Levels

| Level | Description | Tools |
|-------|-------------|-------|
| 0 | Manual training, no tracking | Scripts, notebooks |
| 1 | Experiment tracking | MLflow, W&B |
| 2 | Reproducible pipelines | DVC, MLflow Projects |
| 3 | Automated CI/CD | GitHub Actions, Jenkins |
| 4 | Full monitoring + A/B testing | Prometheus, Grafana, custom |

### Production Checklist

1. **Model versioning** — Every model has a version, stored in registry
2. **Data versioning** — All training data versioned with DVC
3. **Experiment tracking** — All runs logged with params, metrics, artifacts
4. **Reproducibility** — Pipeline can be rerun from scratch
5. **Testing** — Unit tests for preprocessing, integration tests for API
6. **Monitoring** — Prediction latency, throughput, drift detection
7. **Rollback** — Can revert to previous model version instantly
8. **Documentation** — Model card, API docs, runbook

### Common Pitfalls

- Not versioning training data alongside code
- Skipping experiment tracking ("I'll remember the parameters")
- No monitoring for data drift in production
- Deploying without a rollback plan
- Not testing the full prediction pipeline end-to-end
- Ignoring model staleness — models degrade over time
- Over-engineering infrastructure before proving model value
