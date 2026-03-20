---
name: model-deployment
description: >
  ML model serving, export, and deployment patterns.
  Use when deploying PyTorch models to production, exporting to ONNX,
  building inference APIs, setting up model registries, or optimizing
  model serving performance.
  Triggers: "model serving", "model deployment", "ONNX export", "TorchScript",
  "inference API", "model registry", "MLflow", "torch.export", "quantization".
  NOT for: training loops, data preprocessing, experiment tracking during training.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Model Deployment

## ONNX Export

```python
import torch
import torch.onnx

def export_to_onnx(
    model: torch.nn.Module,
    output_path: str,
    input_shape: tuple = (1, 3, 224, 224),
    opset_version: int = 17,
):
    """Export PyTorch model to ONNX format."""
    model.eval()
    dummy_input = torch.randn(*input_shape)

    torch.onnx.export(
        model,
        dummy_input,
        output_path,
        export_params=True,
        opset_version=opset_version,
        do_constant_folding=True,
        input_names=["input"],
        output_names=["output"],
        dynamic_axes={
            "input": {0: "batch_size"},
            "output": {0: "batch_size"},
        },
    )

    # Verify exported model
    import onnx
    onnx_model = onnx.load(output_path)
    onnx.checker.check_model(onnx_model)
    print(f"Exported to {output_path} ({Path(output_path).stat().st_size / 1e6:.1f} MB)")


def verify_onnx(onnx_path: str, input_shape: tuple = (1, 3, 224, 224)):
    """Compare ONNX output with PyTorch output."""
    import onnxruntime as ort
    import numpy as np

    session = ort.InferenceSession(onnx_path)
    dummy = np.random.randn(*input_shape).astype(np.float32)

    onnx_output = session.run(None, {"input": dummy})[0]
    print(f"ONNX output shape: {onnx_output.shape}")
    return onnx_output
```

## FastAPI Inference Server

```python
# server.py — production inference API
from fastapi import FastAPI, HTTPException, UploadFile
from pydantic import BaseModel
import torch
import torch.nn.functional as F
from torchvision import transforms
from PIL import Image
import io
import time
from contextlib import asynccontextmanager

# Global model reference
model = None
device = None
transform = None
labels = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Load model on startup, cleanup on shutdown."""
    global model, device, transform, labels

    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    print(f"Loading model on {device}...")

    # Load model
    checkpoint = torch.load("checkpoints/best.pt", map_location=device, weights_only=False)
    model = create_model(num_classes=checkpoint["config"]["num_classes"])
    model.load_state_dict(checkpoint["model_state_dict"])
    model.eval()

    # Compile for faster inference (PyTorch 2.0+)
    if device.type == "cuda":
        model = torch.compile(model, mode="reduce-overhead")

    # Preprocessing pipeline
    transform = transforms.Compose([
        transforms.Resize(256),
        transforms.CenterCrop(224),
        transforms.ToTensor(),
        transforms.Normalize(mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225]),
    ])

    labels = checkpoint.get("labels", [])
    print(f"Model loaded: {len(labels)} classes")

    yield  # Server runs

    # Cleanup
    del model
    torch.cuda.empty_cache()

app = FastAPI(title="ML Inference API", lifespan=lifespan)


class PredictionResponse(BaseModel):
    label: str
    confidence: float
    top_5: list[dict[str, float]]
    inference_time_ms: float


@app.post("/predict", response_model=PredictionResponse)
async def predict(file: UploadFile):
    if not file.content_type or not file.content_type.startswith("image/"):
        raise HTTPException(400, "File must be an image")

    # Read and preprocess
    contents = await file.read()
    image = Image.open(io.BytesIO(contents)).convert("RGB")
    tensor = transform(image).unsqueeze(0).to(device)

    # Inference
    start = time.perf_counter()
    with torch.no_grad():
        logits = model(tensor)
        probs = F.softmax(logits, dim=-1)
    inference_ms = (time.perf_counter() - start) * 1000

    # Top-5 predictions
    top5_probs, top5_indices = probs[0].topk(5)
    top5 = [
        {"label": labels[idx.item()], "confidence": round(prob.item(), 4)}
        for prob, idx in zip(top5_probs, top5_indices)
    ]

    return PredictionResponse(
        label=top5[0]["label"],
        confidence=top5[0]["confidence"],
        top_5=top5,
        inference_time_ms=round(inference_ms, 2),
    )


@app.get("/health")
async def health():
    return {
        "status": "healthy",
        "model_loaded": model is not None,
        "device": str(device),
        "cuda_available": torch.cuda.is_available(),
    }
```

## Docker for ML Models

```dockerfile
# Dockerfile — multi-stage build for ML serving
FROM python:3.11-slim AS base

WORKDIR /app

# System deps
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgl1-mesa-glx libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

# Python deps (cached layer)
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# App code
COPY server.py .
COPY models/ models/

# Model weights (copy or download)
COPY checkpoints/best.pt checkpoints/best.pt

# Non-root user
RUN useradd -m appuser
USER appuser

EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/health')"

CMD ["uvicorn", "server:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "1"]
```

```yaml
# docker-compose.yml — local development with GPU
services:
  inference:
    build: .
    ports:
      - "8000:8000"
    volumes:
      - ./checkpoints:/app/checkpoints:ro
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
    environment:
      - CUDA_VISIBLE_DEVICES=0
      - TORCH_NUM_THREADS=4
```

## Model Quantization

```python
import torch

def quantize_dynamic(model: torch.nn.Module, output_path: str):
    """Dynamic quantization — easiest, good for NLP models."""
    quantized = torch.quantization.quantize_dynamic(
        model,
        {torch.nn.Linear, torch.nn.LSTM},
        dtype=torch.qint8,
    )
    torch.save(quantized.state_dict(), output_path)

    # Compare sizes
    original_size = sum(p.numel() * p.element_size() for p in model.parameters())
    quant_size = sum(p.numel() * p.element_size() for p in quantized.parameters())
    print(f"Original: {original_size / 1e6:.1f} MB")
    print(f"Quantized: {quant_size / 1e6:.1f} MB")
    print(f"Compression: {original_size / quant_size:.1f}x")

    return quantized


def quantize_static(model, calibration_loader, output_path: str):
    """Static quantization — better accuracy, requires calibration data."""
    model.eval()

    # Specify quantization config
    model.qconfig = torch.quantization.get_default_qconfig("x86")
    torch.quantization.prepare(model, inplace=True)

    # Run calibration data through model
    with torch.no_grad():
        for batch in calibration_loader:
            model(batch["input"])

    torch.quantization.convert(model, inplace=True)
    torch.save(model.state_dict(), output_path)
    return model
```

## MLflow Model Registry

```python
import mlflow
import mlflow.pytorch

def register_model(
    model: torch.nn.Module,
    run_name: str,
    metrics: dict,
    config: dict,
    artifact_path: str = "model",
):
    """Log model and metrics to MLflow, register in model registry."""
    mlflow.set_tracking_uri("http://mlflow-server:5000")
    mlflow.set_experiment("image-classifier")

    with mlflow.start_run(run_name=run_name):
        # Log hyperparameters
        mlflow.log_params(config)

        # Log metrics
        for key, value in metrics.items():
            mlflow.log_metric(key, value)

        # Log model
        mlflow.pytorch.log_model(
            model,
            artifact_path=artifact_path,
            registered_model_name="image-classifier",
            pip_requirements=["torch>=2.0", "torchvision>=0.15"],
        )

        # Log training artifacts
        mlflow.log_artifact("checkpoints/config.json")

    print(f"Model registered: image-classifier/{run_name}")


def load_production_model(model_name: str = "image-classifier"):
    """Load the latest production model from registry."""
    client = mlflow.tracking.MlflowClient()
    latest = client.get_latest_versions(model_name, stages=["Production"])

    if not latest:
        raise ValueError(f"No production model found for {model_name}")

    model_uri = f"models:/{model_name}/Production"
    model = mlflow.pytorch.load_model(model_uri)
    return model
```

## Gotchas

1. **torch.compile breaks on dynamic shapes** -- `torch.compile()` recompiles when input shapes change. For variable-length inputs (NLP), use `dynamic=True` or pad to fixed lengths. Without this, first-request latency can be 10-30 seconds as the model compiles.

2. **CUDA OOM in inference** -- Even with `torch.no_grad()`, large batch inference can OOM. Process in micro-batches: `for i in range(0, len(inputs), batch_size): batch = inputs[i:i+batch_size]`. Also call `torch.cuda.empty_cache()` between large requests.

3. **Model not in eval mode during export** -- Exporting with `model.train()` bakes in dropout and running BatchNorm stats. Always call `model.eval()` before `torch.onnx.export()` or `torch.jit.trace()`. The exported model will behave differently otherwise.

4. **ONNX opset version compatibility** -- Different ONNX runtimes support different opset versions. ONNX Runtime 1.16+ supports opset 17. If targeting edge devices or older runtimes, use opset 13-14. Check operator support before deploying.

5. **FastAPI workers with GPU** -- Multiple Uvicorn workers each load a separate model copy into GPU memory. With a 2GB model and 4 workers, you need 8GB VRAM. Use `--workers 1` with async request handling, or use a model server (Triton, TorchServe) that manages GPU memory.

6. **Quantization accuracy drop** -- Dynamic quantization barely affects Linear-heavy models (transformers) but can significantly hurt Conv-heavy models (vision). Always benchmark quantized vs original on your test set. Static quantization with calibration data is more reliable but requires representative samples.
