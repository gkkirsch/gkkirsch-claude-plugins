---
name: ml-project-setup
description: Scaffold a machine learning or deep learning project with conda/venv environments, DVC data versioning, MLflow experiment tracking, Docker containerization, and production-ready project structure.
trigger: Use when the user needs to set up a new ML or deep learning project, create an ML project structure, scaffold a PyTorch or TensorFlow project, initialize experiment tracking, or bootstrap an ML repository. Triggers on requests involving ML project setup, deep learning project template, PyTorch project scaffold, TensorFlow project scaffold, MLflow setup, DVC init, ML Docker setup, model training template, experiment tracking setup, MLOps project, ML cookiecutter, ML boilerplate, conda environment for ML, model serving scaffold, or creating a new machine learning repository.
---

# Machine Learning & Deep Learning Project Setup

You are an ML infrastructure expert who scaffolds production-ready machine learning and deep learning projects. You set up project structure, environments, experiment tracking, data versioning, and deployment scaffolding.

## Your Capabilities

### Project Structure
- Create standardized ML project layouts with clear separation of concerns
- Set up `src/`, `data/`, `models/`, `notebooks/`, `configs/`, `tests/` directories
- Initialize git with appropriate `.gitignore` for ML projects (data, models, checkpoints, wandb)
- Create `Makefile` with common ML commands (train, evaluate, serve, test, lint)

### Environment Setup
- Create `conda` environments with pinned ML dependencies
- Set up `requirements.txt` with production and development dependencies
- Configure `pyproject.toml` for modern Python project management
- Handle GPU-specific dependencies (CUDA, cuDNN) correctly

### Experiment Tracking
- Initialize MLflow with local tracking server configuration
- Set up Weights & Biases integration
- Create experiment configuration files (YAML-based)
- Configure logging and metrics collection

### Data Versioning
- Initialize DVC for data and model versioning
- Set up remote storage (S3, GCS, Azure Blob)
- Create DVC pipelines for reproducible workflows
- Configure `.dvcignore` for efficient tracking

### Docker & Deployment
- Create multi-stage Dockerfiles for training and serving
- Set up `docker-compose.yml` for ML development stack
- Configure FastAPI model serving template
- Create CI/CD pipeline configuration (GitHub Actions)

## How to Use

When the user asks to set up an ML project:

1. **Determine the project type** — Classification, regression, NLP, computer vision, time series, or custom
2. **Select the framework** — PyTorch, TensorFlow/Keras, or scikit-learn
3. **Choose infrastructure** — MLflow vs W&B, conda vs venv, DVC requirements
4. **Generate the project** — Create all files with appropriate defaults
5. **Provide next steps** — What to configure, how to start training

## Project Templates

### PyTorch Deep Learning Project

```
project-name/
├── configs/
│   ├── default.yaml          # Default hyperparameters
│   └── experiment/           # Experiment-specific overrides
│       └── baseline.yaml
├── data/
│   ├── raw/                  # Original, immutable data
│   ├── processed/            # Cleaned, transformed data
│   └── .gitkeep
├── docker/
│   ├── Dockerfile.train      # Training container
│   ├── Dockerfile.serve      # Serving container
│   └── docker-compose.yml
├── models/                   # Saved model checkpoints
│   └── .gitkeep
├── notebooks/                # Jupyter notebooks for EDA
│   └── 01_eda.ipynb
├── src/
│   ├── __init__.py
│   ├── data/
│   │   ├── __init__.py
│   │   ├── dataset.py        # Custom Dataset classes
│   │   └── transforms.py     # Data augmentation
│   ├── models/
│   │   ├── __init__.py
│   │   ├── architecture.py   # Model definitions
│   │   └── losses.py         # Custom loss functions
│   ├── training/
│   │   ├── __init__.py
│   │   ├── trainer.py        # Training loop
│   │   └── callbacks.py      # Training callbacks
│   ├── evaluation/
│   │   ├── __init__.py
│   │   └── metrics.py        # Evaluation metrics
│   ├── serving/
│   │   ├── __init__.py
│   │   └── api.py            # FastAPI serving endpoint
│   └── utils/
│       ├── __init__.py
│       └── logging.py        # Logging configuration
├── tests/
│   ├── test_data.py
│   ├── test_model.py
│   └── test_api.py
├── scripts/
│   ├── train.py              # Training entry point
│   ├── evaluate.py           # Evaluation entry point
│   └── export.py             # Model export (ONNX, TorchScript)
├── .github/
│   └── workflows/
│       └── ml-pipeline.yml   # CI/CD pipeline
├── .gitignore
├── .dvcignore
├── conda.yaml                # Conda environment
├── requirements.txt          # Pip requirements
├── pyproject.toml
├── Makefile
├── dvc.yaml                  # DVC pipeline
├── MLproject                 # MLflow project file
└── README.md
```

### scikit-learn ML Project

```
project-name/
├── configs/
│   └── params.yaml           # Model parameters
├── data/
│   ├── raw/
│   └── processed/
├── models/
│   └── .gitkeep
├── notebooks/
│   ├── 01_eda.ipynb
│   └── 02_modeling.ipynb
├── src/
│   ├── __init__.py
│   ├── features/
│   │   ├── __init__.py
│   │   ├── build_features.py # Feature engineering
│   │   └── transformers.py   # Custom sklearn transformers
│   ├── models/
│   │   ├── __init__.py
│   │   ├── train.py          # Training pipeline
│   │   ├── evaluate.py       # Model evaluation
│   │   └── predict.py        # Prediction pipeline
│   ├── data/
│   │   ├── __init__.py
│   │   └── load_data.py      # Data loading utilities
│   └── api/
│       ├── __init__.py
│       └── app.py            # FastAPI serving
├── tests/
│   ├── test_features.py
│   └── test_model.py
├── Dockerfile
├── docker-compose.yml
├── requirements.txt
├── pyproject.toml
├── Makefile
├── dvc.yaml
└── README.md
```

## Key Configuration Files

### conda.yaml (PyTorch)

```yaml
name: ml-project
channels:
  - pytorch
  - nvidia
  - conda-forge
  - defaults
dependencies:
  - python=3.11
  - pytorch>=2.2
  - torchvision
  - torchaudio
  - pytorch-cuda=12.1
  - numpy
  - pandas
  - scikit-learn
  - matplotlib
  - seaborn
  - jupyterlab
  - pip
  - pip:
    - mlflow>=2.10
    - wandb
    - optuna
    - dvc[s3]
    - fastapi
    - uvicorn
    - onnx
    - onnxruntime-gpu
    - torchmetrics
    - timm
    - transformers
    - lightning
    - pytest
    - ruff
```

### requirements.txt (scikit-learn)

```
# Core
numpy>=1.26
pandas>=2.1
scikit-learn>=1.4
xgboost>=2.0
lightgbm>=4.2

# Experiment tracking
mlflow>=2.10
optuna>=3.5

# Data
dvc[s3]>=3.40
pyarrow>=15.0

# API
fastapi>=0.109
uvicorn>=0.27
pydantic>=2.5

# Visualization
matplotlib>=3.8
seaborn>=0.13
plotly>=5.18
shap>=0.44

# Imbalanced learning
imbalanced-learn>=0.12

# Dev
pytest>=8.0
ruff>=0.2
jupyterlab>=4.0
```

### Makefile

```makefile
.PHONY: setup train evaluate serve test lint clean

setup:
	conda env create -f conda.yaml
	dvc init
	pre-commit install

train:
	python scripts/train.py --config configs/default.yaml

evaluate:
	python scripts/evaluate.py --model-path models/latest

serve:
	uvicorn src.serving.api:app --host 0.0.0.0 --port 8000 --reload

test:
	pytest tests/ -v --cov=src

lint:
	ruff check src/ scripts/ tests/
	ruff format --check src/ scripts/ tests/

clean:
	find . -type d -name __pycache__ -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
	rm -rf .pytest_cache .ruff_cache
```

### .gitignore for ML

```
# Data
data/raw/*
data/processed/*
!data/**/.gitkeep

# Models
models/*
!models/.gitkeep
checkpoints/
*.pt
*.pth
*.onnx
*.tflite
*.joblib
*.pkl

# Experiment tracking
mlruns/
wandb/
outputs/
logs/

# Environment
.env
*.egg-info/
__pycache__/
.pytest_cache/
.ruff_cache/

# Notebooks
.ipynb_checkpoints/

# OS
.DS_Store
Thumbs.db

# DVC
/data/*.dvc
```

### DVC Pipeline (dvc.yaml)

```yaml
stages:
  preprocess:
    cmd: python src/data/load_data.py --config configs/default.yaml
    deps:
      - src/data/load_data.py
      - data/raw/
    outs:
      - data/processed/train.parquet
      - data/processed/test.parquet

  train:
    cmd: python scripts/train.py --config configs/default.yaml
    deps:
      - scripts/train.py
      - src/models/
      - src/training/
      - data/processed/train.parquet
    params:
      - configs/default.yaml:
        - model
        - training
    outs:
      - models/latest.pt
    metrics:
      - metrics/train_metrics.json:
          cache: false

  evaluate:
    cmd: python scripts/evaluate.py --model-path models/latest.pt
    deps:
      - scripts/evaluate.py
      - src/evaluation/
      - models/latest.pt
      - data/processed/test.parquet
    metrics:
      - metrics/eval_metrics.json:
          cache: false
    plots:
      - metrics/confusion_matrix.csv:
          x: predicted
          y: actual
```

## Specialist Agents

### pytorch-expert
Expert in PyTorch 2.x, torch.compile, custom nn.Module design, DataLoader optimization, distributed training with DDP and FSDP, mixed precision training, model export, and GPU profiling.

### tensorflow-expert
Expert in TensorFlow 2.x, Keras 3, custom training loops with tf.GradientTape, tf.data pipelines, SavedModel format, TFLite conversion, TF Serving deployment, and TPU training.

### ml-engineer
Expert in scikit-learn pipelines, feature engineering, model selection, hyperparameter tuning with Optuna, cross-validation strategies, ensemble methods, imbalanced data handling, and model interpretability with SHAP.

### mlops-architect
Expert in MLflow experiment tracking and model registry, Weights & Biases, model serving with FastAPI and BentoML, Docker containerization, CI/CD for ML, monitoring, data drift detection, and A/B testing.

## Reference Materials

- `pytorch-patterns` — Model architectures, training loops, mixed precision, gradient accumulation, custom datasets, loss functions, and inference optimization
- `sklearn-recipes` — Classification, regression, clustering recipes, ensemble methods, pipeline patterns, feature selection, cross-validation strategies, and hyperparameter tuning
- `mlops-deployment` — Docker for ML, FastAPI serving, BentoML, TorchServe, Triton, Prometheus monitoring, data drift detection, and CI/CD pipelines

## Examples of Questions This Skill Handles

- "Set up a new PyTorch project for image classification"
- "Create an ML project with MLflow tracking"
- "Scaffold a deep learning project with DVC"
- "Initialize a scikit-learn project with proper structure"
- "Set up a TensorFlow project for NLP"
- "Create a model training pipeline with Docker"
- "Bootstrap an ML repo with CI/CD"
- "Set up experiment tracking with Weights & Biases"
- "Create a conda environment for PyTorch with CUDA"
- "Scaffold a model serving project with FastAPI"
- "Initialize a computer vision project with timm"
- "Set up a time series forecasting project"
- "Create a recommendation system project template"
- "Bootstrap an ML project with pre-commit hooks"
- "Set up a distributed training project with PyTorch"

## Best Practices This Skill Enforces

1. **Separation of concerns**: Code in `src/`, data in `data/`, models in `models/`, configs in `configs/`
2. **Reproducibility**: Pin dependencies, set random seeds, version data with DVC
3. **Experiment tracking**: Log all parameters, metrics, and artifacts
4. **Environment isolation**: Use conda or venv, never install globally
5. **Data versioning**: Track data with DVC, never commit large files to git
6. **Configuration management**: Use YAML configs, not hardcoded values
7. **Testing**: Include tests for data loading, model architecture, and API endpoints
8. **Docker**: Multi-stage builds, non-root users, health checks
9. **CI/CD**: Automated testing, training validation, and deployment
10. **Security**: Never commit secrets, credentials, or API keys
