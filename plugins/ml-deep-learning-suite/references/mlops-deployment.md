# MLOps Deployment — Docker, Model Serving, Monitoring, and Production Patterns

Quick reference for deploying ML models to production with Docker, FastAPI, BentoML, TorchServe, Triton, monitoring, and data drift detection.

---

## Docker for ML

### Standard ML Dockerfile

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# System dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential libgomp1 && \
    rm -rf /var/lib/apt/lists/*

# Python dependencies (cached layer)
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Application code
COPY src/ ./src/
COPY models/ ./models/

# Non-root user
RUN useradd -m -r appuser && chown -R appuser:appuser /app
USER appuser

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

EXPOSE 8000
CMD ["uvicorn", "src.api:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "4"]
```

### GPU Dockerfile

```dockerfile
FROM nvidia/cuda:12.1.0-runtime-ubuntu22.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    python3.11 python3.11-venv python3-pip curl && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=compute,utility

EXPOSE 8000
CMD ["python3", "-m", "uvicorn", "api:app", "--host", "0.0.0.0", "--port", "8000"]
```

### Multi-stage Build (Smaller Image)

```dockerfile
# Build stage
FROM python:3.11-slim AS builder
WORKDIR /build
COPY requirements.txt .
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

# Runtime stage
FROM python:3.11-slim
WORKDIR /app
COPY --from=builder /install /usr/local
COPY src/ ./src/
COPY models/ ./models/

RUN useradd -m -r appuser && chown -R appuser:appuser /app
USER appuser

EXPOSE 8000
CMD ["uvicorn", "src.api:app", "--host", "0.0.0.0", "--port", "8000"]
```

### Docker Compose ML Stack

```yaml
version: "3.8"

services:
  api:
    build: .
    ports:
      - "8000:8000"
    volumes:
      - ./models:/app/models:ro
    environment:
      - MODEL_PATH=/app/models/latest.joblib
      - LOG_LEVEL=INFO
      - WORKERS=4
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          memory: 4G
          cpus: "2.0"

  mlflow:
    image: ghcr.io/mlflow/mlflow:v2.10.0
    ports:
      - "5000:5000"
    volumes:
      - mlflow-data:/mlflow
    command: >
      mlflow server --host 0.0.0.0 --port 5000
      --backend-store-uri sqlite:///mlflow/mlflow.db
      --default-artifact-root /mlflow/artifacts

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml:ro

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin

volumes:
  mlflow-data:
```

---

## FastAPI Model Serving

### Production API Template

```python
from fastapi import FastAPI, HTTPException, BackgroundTasks
from pydantic import BaseModel, Field
from contextlib import asynccontextmanager
import joblib
import numpy as np
import pandas as pd
import time
import logging

logger = logging.getLogger(__name__)

# Global model state
model_state = {"model": None, "version": None, "loaded_at": None}

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Load model on startup
    model_path = "models/production.joblib"
    bundle = joblib.load(model_path)
    model_state["model"] = bundle["model"]
    model_state["version"] = bundle["metadata"]["version"]
    model_state["loaded_at"] = time.time()
    logger.info(f"Model v{model_state['version']} loaded")
    yield
    model_state["model"] = None

app = FastAPI(title="ML Prediction API", version="1.0.0", lifespan=lifespan)


class PredictRequest(BaseModel):
    features: dict = Field(..., description="Feature dict")

class PredictResponse(BaseModel):
    prediction: float
    probability: float | None = None
    model_version: str
    latency_ms: float


@app.post("/predict", response_model=PredictResponse)
async def predict(request: PredictRequest):
    start = time.time()
    try:
        df = pd.DataFrame([request.features])
        model = model_state["model"]

        if hasattr(model, "predict_proba"):
            proba = float(model.predict_proba(df)[0, 1])
            pred = int(proba >= 0.5)
        else:
            pred = float(model.predict(df)[0])
            proba = None

        latency = (time.time() - start) * 1000
        return PredictResponse(
            prediction=pred,
            probability=proba,
            model_version=model_state["version"],
            latency_ms=round(latency, 2),
        )
    except Exception as e:
        logger.error(f"Prediction error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/predict/batch")
async def predict_batch(requests: list[PredictRequest]):
    df = pd.DataFrame([r.features for r in requests])
    model = model_state["model"]

    if hasattr(model, "predict_proba"):
        probas = model.predict_proba(df)[:, 1].tolist()
        preds = [int(p >= 0.5) for p in probas]
    else:
        preds = model.predict(df).tolist()
        probas = [None] * len(preds)

    return [
        {"prediction": p, "probability": prob, "model_version": model_state["version"]}
        for p, prob in zip(preds, probas)
    ]


@app.get("/health")
async def health():
    return {
        "status": "healthy",
        "model_loaded": model_state["model"] is not None,
        "model_version": model_state["version"],
    }


@app.get("/model/info")
async def model_info():
    return {
        "version": model_state["version"],
        "loaded_at": model_state["loaded_at"],
        "uptime_seconds": time.time() - model_state["loaded_at"],
    }
```

### Request Validation

```python
from pydantic import BaseModel, Field, field_validator

class ChurnPredictRequest(BaseModel):
    age: int = Field(..., ge=18, le=120)
    income: float = Field(..., ge=0)
    tenure_months: int = Field(..., ge=0)
    num_products: int = Field(..., ge=1, le=10)
    has_credit_card: bool
    is_active_member: bool
    country: str = Field(..., pattern="^[A-Z]{2}$")

    @field_validator("country")
    @classmethod
    def validate_country(cls, v):
        allowed = {"US", "UK", "DE", "FR", "ES"}
        if v not in allowed:
            raise ValueError(f"Country must be one of {allowed}")
        return v
```

---

## BentoML Serving

### Define BentoML Service

```python
import bentoml
from bentoml.io import JSON, NumpyNdarray

# Save model
bentoml.sklearn.save_model("churn_model", pipeline, signatures={"predict_proba": {"batchable": True}})

# Service definition
runner = bentoml.sklearn.get("churn_model:latest").to_runner()
svc = bentoml.Service("churn_service", runners=[runner])

@svc.api(input=JSON(), output=JSON())
async def predict(input_data: dict) -> dict:
    import pandas as pd
    df = pd.DataFrame([input_data])
    result = await runner.predict_proba.async_run(df)
    return {"probability": float(result[0][1]), "prediction": int(result[0][1] >= 0.5)}

# Build: bentoml build
# Serve: bentoml serve service:svc
# Containerize: bentoml containerize churn_service:latest
```

### bentofile.yaml

```yaml
service: "service:svc"
include:
  - "*.py"
python:
  requirements_txt: "requirements.txt"
docker:
  system_packages:
    - libgomp1
  env:
    - WORKERS=4
```

---

## TorchServe

### Package and Serve PyTorch Model

```bash
# Archive model
torch-model-archiver \
  --model-name resnet50 \
  --version 1.0 \
  --serialized-file model.pt \
  --handler image_classifier \
  --export-path model_store \
  --extra-files index_to_name.json

# Start TorchServe
torchserve --start --model-store model_store \
  --models resnet50=resnet50.mar \
  --ncs --ts-config config.properties
```

### Custom Handler

```python
from ts.torch_handler.base_handler import BaseHandler
import torch
import io
from PIL import Image
from torchvision import transforms


class CustomHandler(BaseHandler):
    def initialize(self, context):
        super().initialize(context)
        self.transform = transforms.Compose([
            transforms.Resize(256),
            transforms.CenterCrop(224),
            transforms.ToTensor(),
            transforms.Normalize([0.485, 0.456, 0.406], [0.229, 0.224, 0.225]),
        ])

    def preprocess(self, data):
        images = []
        for row in data:
            raw = row.get("data") or row.get("body")
            image = Image.open(io.BytesIO(raw)).convert("RGB")
            images.append(self.transform(image))
        return torch.stack(images).to(self.device)

    def inference(self, data):
        with torch.no_grad():
            return self.model(data)

    def postprocess(self, output):
        probs = torch.softmax(output, dim=1)
        top5 = torch.topk(probs, 5)
        results = []
        for probs_i, indices_i in zip(top5.values, top5.indices):
            results.append({
                self.mapping[str(idx.item())]: round(prob.item(), 4)
                for prob, idx in zip(probs_i, indices_i)
            })
        return results
```

### TorchServe Config

```properties
# config.properties
inference_address=http://0.0.0.0:8080
management_address=http://0.0.0.0:8081
metrics_address=http://0.0.0.0:8082
model_store=/home/model-server/model-store
load_models=all
number_of_netty_threads=32
job_queue_size=1000
default_workers_per_model=4
batch_size=32
max_batch_delay=100
```

---

## NVIDIA Triton Inference Server

### Model Repository Structure

```
model_repository/
├── resnet50/
│   ├── config.pbtxt
│   └── 1/
│       └── model.onnx
├── text_classifier/
│   ├── config.pbtxt
│   └── 1/
│       └── model.pt
```

### Model Config

```protobuf
# config.pbtxt
name: "resnet50"
platform: "onnxruntime_onnx"
max_batch_size: 64

input [
  {
    name: "input"
    data_type: TYPE_FP32
    dims: [3, 224, 224]
  }
]

output [
  {
    name: "output"
    data_type: TYPE_FP32
    dims: [1000]
  }
]

instance_group [
  {
    count: 2
    kind: KIND_GPU
    gpus: [0]
  }
]

dynamic_batching {
  preferred_batch_size: [8, 16, 32]
  max_queue_delay_microseconds: 100
}
```

### Triton Client

```python
import tritonclient.http as httpclient
import numpy as np

client = httpclient.InferenceServerClient(url="localhost:8000")

# Check model
assert client.is_model_ready("resnet50")

# Inference
input_data = np.random.rand(1, 3, 224, 224).astype(np.float32)

inputs = [httpclient.InferInput("input", input_data.shape, "FP32")]
inputs[0].set_data_from_numpy(input_data)

outputs = [httpclient.InferRequestedOutput("output")]

result = client.infer("resnet50", inputs=inputs, outputs=outputs)
output_data = result.as_numpy("output")
```

---

## Monitoring

### Prometheus Metrics for ML

```python
from prometheus_client import Counter, Histogram, Gauge, start_http_server

# Counters
PREDICTIONS_TOTAL = Counter(
    "ml_predictions_total", "Total predictions",
    ["model_name", "model_version", "prediction_class"],
)
ERRORS_TOTAL = Counter(
    "ml_prediction_errors_total", "Prediction errors",
    ["model_name", "error_type"],
)

# Histograms
PREDICTION_LATENCY = Histogram(
    "ml_prediction_latency_seconds", "Prediction latency",
    ["model_name"],
    buckets=[0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0],
)
PREDICTION_SCORE = Histogram(
    "ml_prediction_score", "Distribution of prediction scores",
    ["model_name"],
    buckets=[0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0],
)

# Gauges
MODEL_LOADED = Gauge("ml_model_loaded", "Model load status", ["model_name", "version"])
DATA_DRIFT_PSI = Gauge("ml_data_drift_psi", "PSI score per feature", ["model_name", "feature"])
FEATURE_MEAN = Gauge("ml_feature_mean", "Running mean of features", ["model_name", "feature"])

# Start metrics server on port 8001
start_http_server(8001)
```

### Prometheus Config

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "ml-api"
    static_configs:
      - targets: ["api:8001"]
    metrics_path: /metrics
    scrape_interval: 10s
```

### Grafana Dashboard JSON (Key Panels)

```json
{
  "panels": [
    {
      "title": "Predictions per Second",
      "type": "graph",
      "targets": [{"expr": "rate(ml_predictions_total[5m])"}]
    },
    {
      "title": "P99 Latency",
      "type": "stat",
      "targets": [{"expr": "histogram_quantile(0.99, rate(ml_prediction_latency_seconds_bucket[5m]))"}]
    },
    {
      "title": "Prediction Score Distribution",
      "type": "heatmap",
      "targets": [{"expr": "rate(ml_prediction_score_bucket[5m])"}]
    },
    {
      "title": "Data Drift (PSI)",
      "type": "gauge",
      "targets": [{"expr": "ml_data_drift_psi"}],
      "thresholds": [0.1, 0.2]
    }
  ]
}
```

---

## Data Drift Detection

### Population Stability Index (PSI)

```python
import numpy as np

def compute_psi(reference, current, bins=10):
    """PSI: < 0.1 stable, 0.1-0.2 moderate, > 0.2 significant drift."""
    ref_hist, bin_edges = np.histogram(reference, bins=bins)
    cur_hist, _ = np.histogram(current, bins=bin_edges)

    ref_pct = (ref_hist + 1) / (len(reference) + bins)
    cur_pct = (cur_hist + 1) / (len(current) + bins)

    psi = np.sum((cur_pct - ref_pct) * np.log(cur_pct / ref_pct))
    return psi
```

### KS Test

```python
from scipy.stats import ks_2samp

def check_drift(reference_data, current_data, feature_names, threshold=0.05):
    """Check for drift using Kolmogorov-Smirnov test."""
    results = {}
    for i, feature in enumerate(feature_names):
        stat, p_value = ks_2samp(reference_data[:, i], current_data[:, i])
        psi = compute_psi(reference_data[:, i], current_data[:, i])
        results[feature] = {
            "ks_stat": stat,
            "p_value": p_value,
            "psi": psi,
            "drifted": p_value < threshold or psi > 0.2,
        }
    return results
```

### Evidently AI Integration

```python
from evidently.report import Report
from evidently.metric_preset import DataDriftPreset, DataQualityPreset

report = Report(metrics=[DataDriftPreset(), DataQualityPreset()])
report.run(reference_data=df_reference, current_data=df_current)
report.save_html("drift_report.html")

# Extract drift results programmatically
result = report.as_dict()
dataset_drift = result["metrics"][0]["result"]["dataset_drift"]
drifted_features = [
    col for col, info in result["metrics"][0]["result"]["drift_by_columns"].items()
    if info["drift_detected"]
]
```

---

## Model Versioning and Rollback

### Blue-Green Deployment

```python
import os

class ModelRouter:
    """Route between active and standby models."""

    def __init__(self):
        self.models = {}
        self.active = None

    def load_model(self, version, path):
        import joblib
        self.models[version] = joblib.load(path)

    def set_active(self, version):
        if version not in self.models:
            raise ValueError(f"Model {version} not loaded")
        self.active = version

    def predict(self, features):
        model = self.models[self.active]
        return model.predict(features)

    def rollback(self, version):
        self.set_active(version)

# Usage
router = ModelRouter()
router.load_model("v1", "models/v1.joblib")
router.load_model("v2", "models/v2.joblib")
router.set_active("v2")

# If v2 has issues:
router.rollback("v1")
```

### Canary Deployment

```python
import hashlib

class CanaryRouter:
    """Route traffic between production and canary models."""

    def __init__(self, prod_model, canary_model, canary_pct=0.1):
        self.prod = prod_model
        self.canary = canary_model
        self.canary_pct = canary_pct

    def predict(self, features, request_id=None):
        if request_id:
            bucket = int(hashlib.md5(request_id.encode()).hexdigest(), 16) % 100
        else:
            import random
            bucket = random.randint(0, 99)

        if bucket < self.canary_pct * 100:
            return self.canary.predict(features), "canary"
        return self.prod.predict(features), "production"

    def promote_canary(self):
        self.prod = self.canary
        self.canary_pct = 0.0
```

---

## CI/CD Pipeline

### GitHub Actions for ML

```yaml
name: ML Pipeline

on:
  push:
    paths: ["src/**", "data/**", "dvc.yaml"]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: "3.11"
      - run: pip install -r requirements.txt
      - run: pytest tests/ -v

  train:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: "3.11"
      - run: pip install -r requirements.txt
      - run: dvc pull
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
      - run: dvc repro
      - name: Validate metrics
        run: |
          python -c "
          import json
          m = json.load(open('metrics.json'))
          assert m['roc_auc'] > 0.85
          assert m['f1'] > 0.80
          print('Metrics OK')
          "
      - uses: actions/upload-artifact@v4
        with:
          name: model
          path: models/

  deploy:
    needs: train
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v4
        with:
          name: model
          path: models/
      - run: |
          docker build -t ml-api:${{ github.sha }} .
          docker push ml-api:${{ github.sha }}
```

---

## Production Checklist

| Category | Check | Status |
|----------|-------|--------|
| **Model** | Model versioned in registry | |
| **Model** | Input validation on all features | |
| **Model** | Fallback/default for missing features | |
| **Data** | Training data versioned (DVC/S3) | |
| **Data** | Data validation pipeline | |
| **Serving** | Health check endpoint | |
| **Serving** | Batch + single prediction endpoints | |
| **Serving** | Request/response logging | |
| **Serving** | Rate limiting | |
| **Monitoring** | Prediction latency metrics | |
| **Monitoring** | Throughput metrics | |
| **Monitoring** | Data drift detection | |
| **Monitoring** | Model performance tracking | |
| **Monitoring** | Error rate alerting | |
| **Ops** | Rollback procedure documented | |
| **Ops** | CI/CD pipeline | |
| **Ops** | Load testing completed | |
| **Ops** | Disaster recovery plan | |

---

## Quick Commands

```bash
# Docker
docker build -t ml-api .
docker run -p 8000:8000 ml-api
docker compose up -d

# TorchServe
torchserve --start --model-store model_store --models all
curl http://localhost:8080/predictions/model_name -T input.jpg

# Triton
docker run --gpus all -p 8000:8000 -p 8001:8001 -p 8002:8002 \
  -v $(pwd)/model_repository:/models \
  nvcr.io/nvidia/tritonserver:latest \
  tritonserver --model-repository=/models

# BentoML
bentoml serve service:svc --reload
bentoml build
bentoml containerize my_service:latest

# MLflow
mlflow server --host 0.0.0.0 --port 5000
mlflow models serve -m "models:/model_name/Production" -p 8000

# Load testing
locust -f locustfile.py --host http://localhost:8000
```
