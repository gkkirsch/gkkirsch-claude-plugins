---
name: pytorch-training
description: >
  PyTorch model training patterns for production ML workflows.
  Use when building training loops, creating custom datasets, implementing
  learning rate schedules, distributed training, or model evaluation.
  Triggers: "pytorch", "torch", "training loop", "dataloader", "model training",
  "loss function", "optimizer", "learning rate", "mixed precision", "distributed training".
  NOT for: TensorFlow/Keras, scikit-learn, data preprocessing without models.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# PyTorch Training Patterns

## Production Training Loop

```python
import torch
import torch.nn as nn
from torch.utils.data import DataLoader, Dataset
from torch.cuda.amp import GradScaler, autocast
from pathlib import Path
import json
import time
from dataclasses import dataclass, asdict

@dataclass
class TrainConfig:
    """Single source of truth for training hyperparameters."""
    model_name: str = "classifier-v1"
    epochs: int = 50
    batch_size: int = 32
    learning_rate: float = 3e-4
    weight_decay: float = 1e-2
    warmup_steps: int = 500
    max_grad_norm: float = 1.0
    mixed_precision: bool = True
    early_stopping_patience: int = 5
    checkpoint_dir: str = "checkpoints"
    seed: int = 42

    def save(self, path: str):
        with open(path, "w") as f:
            json.dump(asdict(self), f, indent=2)

    @classmethod
    def load(cls, path: str) -> "TrainConfig":
        with open(path) as f:
            return cls(**json.load(f))


class Trainer:
    def __init__(self, model: nn.Module, config: TrainConfig):
        self.config = config
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        self.model = model.to(self.device)

        # Optimizer with weight decay decoupled from lr
        self.optimizer = torch.optim.AdamW(
            self.model.parameters(),
            lr=config.learning_rate,
            weight_decay=config.weight_decay,
        )

        # Mixed precision
        self.scaler = GradScaler(enabled=config.mixed_precision)

        # Tracking
        self.best_val_loss = float("inf")
        self.patience_counter = 0
        self.global_step = 0
        self.history: list[dict] = []

        # Checkpoint directory
        Path(config.checkpoint_dir).mkdir(parents=True, exist_ok=True)

    def train(self, train_loader: DataLoader, val_loader: DataLoader):
        """Full training loop with validation, checkpointing, and early stopping."""
        scheduler = torch.optim.lr_scheduler.OneCycleLR(
            self.optimizer,
            max_lr=self.config.learning_rate,
            epochs=self.config.epochs,
            steps_per_epoch=len(train_loader),
        )

        for epoch in range(self.config.epochs):
            # Train
            train_metrics = self._train_epoch(train_loader, scheduler)

            # Validate
            val_metrics = self._validate(val_loader)

            # Record
            record = {
                "epoch": epoch + 1,
                "train_loss": train_metrics["loss"],
                "val_loss": val_metrics["loss"],
                "val_accuracy": val_metrics.get("accuracy", 0),
                "lr": scheduler.get_last_lr()[0],
                "time": train_metrics["time"],
            }
            self.history.append(record)

            print(
                f"Epoch {epoch+1}/{self.config.epochs} | "
                f"Train Loss: {record['train_loss']:.4f} | "
                f"Val Loss: {record['val_loss']:.4f} | "
                f"Val Acc: {record['val_accuracy']:.4f} | "
                f"LR: {record['lr']:.2e}"
            )

            # Checkpoint best model
            if val_metrics["loss"] < self.best_val_loss:
                self.best_val_loss = val_metrics["loss"]
                self.patience_counter = 0
                self._save_checkpoint(epoch, is_best=True)
            else:
                self.patience_counter += 1

            # Early stopping
            if self.patience_counter >= self.config.early_stopping_patience:
                print(f"Early stopping at epoch {epoch+1}")
                break

        return self.history

    def _train_epoch(self, loader: DataLoader, scheduler) -> dict:
        self.model.train()
        total_loss = 0.0
        start = time.time()

        for batch in loader:
            inputs = batch["input"].to(self.device)
            targets = batch["target"].to(self.device)

            self.optimizer.zero_grad(set_to_none=True)

            with autocast(enabled=self.config.mixed_precision):
                outputs = self.model(inputs)
                loss = nn.functional.cross_entropy(outputs, targets)

            self.scaler.scale(loss).backward()

            # Gradient clipping
            self.scaler.unscale_(self.optimizer)
            torch.nn.utils.clip_grad_norm_(
                self.model.parameters(), self.config.max_grad_norm
            )

            self.scaler.step(self.optimizer)
            self.scaler.update()
            scheduler.step()

            total_loss += loss.item()
            self.global_step += 1

        return {
            "loss": total_loss / len(loader),
            "time": time.time() - start,
        }

    @torch.no_grad()
    def _validate(self, loader: DataLoader) -> dict:
        self.model.eval()
        total_loss = 0.0
        correct = 0
        total = 0

        for batch in loader:
            inputs = batch["input"].to(self.device)
            targets = batch["target"].to(self.device)

            with autocast(enabled=self.config.mixed_precision):
                outputs = self.model(inputs)
                loss = nn.functional.cross_entropy(outputs, targets)

            total_loss += loss.item()
            preds = outputs.argmax(dim=-1)
            correct += (preds == targets).sum().item()
            total += targets.size(0)

        return {
            "loss": total_loss / len(loader),
            "accuracy": correct / total if total > 0 else 0,
        }

    def _save_checkpoint(self, epoch: int, is_best: bool = False):
        state = {
            "epoch": epoch,
            "model_state_dict": self.model.state_dict(),
            "optimizer_state_dict": self.optimizer.state_dict(),
            "best_val_loss": self.best_val_loss,
            "config": asdict(self.config),
            "global_step": self.global_step,
        }
        path = Path(self.config.checkpoint_dir)
        torch.save(state, path / "last.pt")
        if is_best:
            torch.save(state, path / "best.pt")
        self.config.save(str(path / "config.json"))

    def load_checkpoint(self, path: str):
        checkpoint = torch.load(path, map_location=self.device, weights_only=False)
        self.model.load_state_dict(checkpoint["model_state_dict"])
        self.optimizer.load_state_dict(checkpoint["optimizer_state_dict"])
        self.best_val_loss = checkpoint["best_val_loss"]
        self.global_step = checkpoint["global_step"]
        return checkpoint["epoch"]
```

## Custom Dataset

```python
from torch.utils.data import Dataset, DataLoader
from torchvision import transforms
from PIL import Image
import pandas as pd

class ImageClassificationDataset(Dataset):
    """Custom dataset with transforms and caching."""

    def __init__(self, csv_path: str, img_dir: str, split: str = "train"):
        self.df = pd.read_csv(csv_path)
        self.df = self.df[self.df["split"] == split].reset_index(drop=True)
        self.img_dir = Path(img_dir)

        self.transform = transforms.Compose([
            transforms.Resize(256),
            transforms.CenterCrop(224),
            transforms.RandomHorizontalFlip() if split == "train" else transforms.Lambda(lambda x: x),
            transforms.ToTensor(),
            transforms.Normalize(mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225]),
        ])

        # Build label mapping
        self.labels = sorted(self.df["label"].unique())
        self.label_to_idx = {label: idx for idx, label in enumerate(self.labels)}

    def __len__(self) -> int:
        return len(self.df)

    def __getitem__(self, idx: int) -> dict:
        row = self.df.iloc[idx]
        img_path = self.img_dir / row["filename"]
        image = Image.open(img_path).convert("RGB")
        image = self.transform(image)

        return {
            "input": image,
            "target": self.label_to_idx[row["label"]],
            "filename": row["filename"],
        }

    @property
    def num_classes(self) -> int:
        return len(self.labels)


def create_dataloaders(config: TrainConfig, csv_path: str, img_dir: str):
    """Create train/val/test dataloaders with proper settings."""
    datasets = {
        split: ImageClassificationDataset(csv_path, img_dir, split=split)
        for split in ["train", "val", "test"]
    }

    loaders = {}
    for split, ds in datasets.items():
        loaders[split] = DataLoader(
            ds,
            batch_size=config.batch_size,
            shuffle=(split == "train"),
            num_workers=4,
            pin_memory=True,         # Faster GPU transfer
            persistent_workers=True, # Keep workers alive between epochs
            drop_last=(split == "train"),  # Avoid small final batch in training
        )

    return loaders
```

## Model Architecture

```python
import torch.nn as nn

class ConvBlock(nn.Module):
    """Conv + BatchNorm + ReLU block."""
    def __init__(self, in_ch: int, out_ch: int, kernel_size: int = 3):
        super().__init__()
        self.block = nn.Sequential(
            nn.Conv2d(in_ch, out_ch, kernel_size, padding=kernel_size // 2, bias=False),
            nn.BatchNorm2d(out_ch),
            nn.ReLU(inplace=True),
        )

    def forward(self, x):
        return self.block(x)


class Classifier(nn.Module):
    """Simple CNN classifier for demonstration."""
    def __init__(self, num_classes: int, in_channels: int = 3):
        super().__init__()
        self.features = nn.Sequential(
            ConvBlock(in_channels, 64),
            nn.MaxPool2d(2),
            ConvBlock(64, 128),
            nn.MaxPool2d(2),
            ConvBlock(128, 256),
            nn.AdaptiveAvgPool2d(1),
        )
        self.classifier = nn.Sequential(
            nn.Dropout(0.5),
            nn.Linear(256, num_classes),
        )

    def forward(self, x):
        x = self.features(x)
        x = x.flatten(1)
        return self.classifier(x)


# Transfer learning with frozen backbone
def create_transfer_model(num_classes: int, backbone: str = "resnet50"):
    from torchvision import models

    weights = models.ResNet50_Weights.DEFAULT
    model = models.resnet50(weights=weights)

    # Freeze backbone
    for param in model.parameters():
        param.requires_grad = False

    # Replace classifier head (unfrozen)
    model.fc = nn.Sequential(
        nn.Dropout(0.5),
        nn.Linear(model.fc.in_features, num_classes),
    )

    return model
```

## Learning Rate Schedules

```python
# Common schedules compared
from torch.optim.lr_scheduler import (
    CosineAnnealingLR,
    OneCycleLR,
    CosineAnnealingWarmRestarts,
    ReduceLROnPlateau,
)

# OneCycleLR — best general-purpose schedule
scheduler = OneCycleLR(
    optimizer,
    max_lr=3e-4,
    epochs=50,
    steps_per_epoch=len(train_loader),
    pct_start=0.1,      # 10% warmup
    anneal_strategy="cos",
)
# Call scheduler.step() EVERY BATCH (not every epoch)

# CosineAnnealing — smooth decay
scheduler = CosineAnnealingLR(optimizer, T_max=50, eta_min=1e-6)
# Call scheduler.step() every EPOCH

# ReduceLROnPlateau — reduce when stuck
scheduler = ReduceLROnPlateau(optimizer, mode="min", factor=0.5, patience=5)
# Call scheduler.step(val_loss) every EPOCH

# Warmup + cosine (manual)
def get_lr(step: int, warmup: int, total: int, base_lr: float) -> float:
    if step < warmup:
        return base_lr * step / warmup
    progress = (step - warmup) / (total - warmup)
    return base_lr * 0.5 * (1 + math.cos(math.pi * progress))
```

## Gotchas

1. **model.eval() and torch.no_grad() are different** -- `model.eval()` changes BatchNorm and Dropout behavior. `torch.no_grad()` disables gradient computation (saves memory). You need BOTH for validation: `model.eval()` + `with torch.no_grad():`. Forgetting `model.eval()` gives wrong predictions; forgetting `no_grad()` wastes memory.

2. **optimizer.zero_grad(set_to_none=True)** -- Default `zero_grad()` sets gradients to zero tensors. `set_to_none=True` sets them to None, which uses less memory and is faster. Always use `set_to_none=True` unless you have gradient accumulation that relies on zero values.

3. **DataLoader num_workers > 0 hangs on error** -- If your dataset's `__getitem__` raises an exception, worker processes crash silently. Set `num_workers=0` for debugging. Also set `persistent_workers=True` with `num_workers > 0` to avoid re-spawning workers every epoch.

4. **Mixed precision with gradient clipping** -- Must call `scaler.unscale_(optimizer)` BEFORE `clip_grad_norm_()`, then `scaler.step(optimizer)`. Calling `clip_grad_norm_` on scaled gradients clips at the wrong threshold. The order matters: scale backward -> unscale -> clip -> step -> update.

5. **Checkpoint loading with map_location** -- `torch.load("model.pt")` fails if saved on GPU and loaded on CPU. Always use `torch.load(path, map_location=device)`. Also, `weights_only=True` (PyTorch 2.6+ default) blocks loading optimizers and custom objects. Use `weights_only=False` for full checkpoints.

6. **Learning rate scheduler step timing** -- `OneCycleLR` and `CosineAnnealingWarmRestarts` step PER BATCH. `CosineAnnealingLR` and `ReduceLROnPlateau` step PER EPOCH. Using the wrong frequency silently gives wrong learning rates. Check the docs for each scheduler.
