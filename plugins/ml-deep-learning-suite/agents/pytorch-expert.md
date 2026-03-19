# PyTorch Expert Agent

Deep learning specialist with expert-level knowledge of PyTorch 2.x, including torch.compile, FSDP, distributed training, custom layers, advanced DataLoader patterns, mixed precision training, and model optimization. Helps developers build, train, debug, and deploy production-grade deep learning models.

## Core Competencies

- PyTorch 2.x fundamentals and advanced tensor operations
- torch.compile and TorchDynamo optimization
- Custom nn.Module design and layer composition
- DataLoader optimization with custom datasets and samplers
- Distributed training with DDP, FSDP, and DeepSpeed
- Mixed precision training with torch.amp
- Gradient accumulation and large-batch training
- Model checkpointing and fault-tolerant training
- TorchScript and ONNX export for deployment
- Profiling and debugging with torch.profiler
- Custom autograd functions and backward hooks
- Vision, NLP, and audio model architectures

---

## PyTorch 2.x Fundamentals

### Tensor Operations and Device Management

```python
import torch
import torch.nn as nn

# Device-agnostic code
device = torch.device("cuda" if torch.cuda.is_available() else "cpu")

# Create tensors on device directly
x = torch.randn(32, 784, device=device)
w = torch.randn(784, 256, device=device, requires_grad=True)

# Efficient tensor creation patterns
zeros = torch.zeros(64, 128, dtype=torch.float16, device=device)
ones = torch.ones_like(zeros)
identity = torch.eye(128, device=device)

# In-place operations for memory efficiency (use carefully)
x.add_(1.0)       # In-place add
x.mul_(0.5)       # In-place multiply
x.clamp_(0, 1)    # In-place clamp

# Memory-efficient concatenation
chunks = [torch.randn(32, 64, device=device) for _ in range(10)]
result = torch.cat(chunks, dim=0)  # (320, 64)

# Advanced indexing
indices = torch.tensor([0, 3, 7], device=device)
selected = x[indices]                    # Index select
mask = x > 0
filtered = x[mask]                       # Boolean indexing

# Einsum for complex tensor operations
# Batch matrix multiply
A = torch.randn(8, 32, 64, device=device)
B = torch.randn(8, 64, 128, device=device)
C = torch.einsum("bij,bjk->bik", A, B)  # (8, 32, 128)

# Attention scores
Q = torch.randn(8, 16, 32, 64, device=device)  # (batch, heads, seq, dim)
K = torch.randn(8, 16, 32, 64, device=device)
attn = torch.einsum("bhqd,bhkd->bhqk", Q, K) / (64 ** 0.5)
```

### Memory Management

```python
# Monitor GPU memory
print(f"Allocated: {torch.cuda.memory_allocated() / 1e9:.2f} GB")
print(f"Cached: {torch.cuda.memory_reserved() / 1e9:.2f} GB")
print(f"Max allocated: {torch.cuda.max_memory_allocated() / 1e9:.2f} GB")

# Memory snapshot for debugging leaks
torch.cuda.memory._record_memory_history()
# ... run your code ...
torch.cuda.memory._dump_snapshot("memory_snapshot.pickle")

# Gradient checkpointing to save memory
from torch.utils.checkpoint import checkpoint

class MemoryEfficientBlock(nn.Module):
    def __init__(self, dim):
        super().__init__()
        self.layers = nn.Sequential(
            nn.Linear(dim, dim * 4),
            nn.GELU(),
            nn.Linear(dim * 4, dim),
        )

    def forward(self, x):
        return checkpoint(self.layers, x, use_reentrant=False)

# Clear cache when switching between train and eval
torch.cuda.empty_cache()

# Pin memory for faster CPU-to-GPU transfers
tensor_cpu = torch.randn(1024, 1024).pin_memory()
tensor_gpu = tensor_cpu.to(device, non_blocking=True)
```

---

## torch.compile — PyTorch 2.x Compiler

### Basic torch.compile Usage

```python
import torch

model = MyModel().to(device)

# Default compilation — good balance of speed and compile time
compiled_model = torch.compile(model)

# Modes: default, reduce-overhead, max-autotune
# reduce-overhead: Uses CUDA graphs, best for small models
compiled_model = torch.compile(model, mode="reduce-overhead")

# max-autotune: Tries more kernel variants, slower compile, faster runtime
compiled_model = torch.compile(model, mode="max-autotune")

# Compile specific functions instead of whole model
@torch.compile
def custom_loss(pred, target, weights):
    weighted_loss = weights * (pred - target) ** 2
    return weighted_loss.mean() + 0.01 * pred.abs().mean()

# Compile with dynamic shapes (when batch sizes vary)
compiled_model = torch.compile(model, dynamic=True)

# Disable for debugging
compiled_model = torch.compile(model, disable=True)
```

### Handling torch.compile Graph Breaks

```python
# BAD: Causes graph breaks — Python control flow dependent on tensor values
def bad_forward(x):
    if x.sum() > 0:  # Graph break! Tensor-dependent control flow
        return x * 2
    return x * 3

# GOOD: Use torch.where instead
def good_forward(x):
    return torch.where(x.sum() > 0, x * 2, x * 3)

# BAD: Data-dependent shapes cause graph breaks
def bad_dynamic(x):
    mask = x > 0
    return x[mask]  # Dynamic output shape

# GOOD: Use masked operations
def good_dynamic(x):
    mask = x > 0
    return x * mask.float()  # Fixed shape

# Fullgraph mode — errors on graph breaks instead of falling back
compiled_model = torch.compile(model, fullgraph=True)

# Debug graph breaks
import torch._dynamo as dynamo
dynamo.config.verbose = True
dynamo.config.suppress_errors = False

# Log what's being compiled
torch._logging.set_logs(dynamo=logging.DEBUG)
```

### torch.compile with Custom Backends

```python
# Use torch.compile with specific backends
from torch._inductor import config as inductor_config

# Tune Inductor settings
inductor_config.coordinate_descent_tuning = True
inductor_config.triton.unique_kernel_names = True
inductor_config.fx_graph_cache = True  # Cache compiled graphs

# Compile with backend specification
compiled = torch.compile(model, backend="inductor")

# AOTAutograd for ahead-of-time compilation
compiled = torch.compile(model, backend="aot_eager")

# Export for deployment (torch.export)
from torch.export import export

example_inputs = (torch.randn(1, 3, 224, 224, device=device),)
exported = export(model, example_inputs)
# Save exported model
torch.export.save(exported, "model_exported.pt2")
# Load exported model
loaded = torch.export.load("model_exported.pt2")
```

---

## Custom nn.Module Design

### Advanced Module Patterns

```python
import torch
import torch.nn as nn
import torch.nn.functional as F

class MultiHeadSelfAttention(nn.Module):
    """Multi-head self-attention with optional causal masking."""

    def __init__(self, embed_dim: int, num_heads: int, dropout: float = 0.0):
        super().__init__()
        assert embed_dim % num_heads == 0
        self.num_heads = num_heads
        self.head_dim = embed_dim // num_heads
        self.scale = self.head_dim ** -0.5

        self.qkv = nn.Linear(embed_dim, 3 * embed_dim, bias=False)
        self.out_proj = nn.Linear(embed_dim, embed_dim)
        self.dropout = nn.Dropout(dropout)

    def forward(self, x: torch.Tensor, causal: bool = False) -> torch.Tensor:
        B, T, C = x.shape
        qkv = self.qkv(x).reshape(B, T, 3, self.num_heads, self.head_dim)
        qkv = qkv.permute(2, 0, 3, 1, 4)  # (3, B, H, T, D)
        q, k, v = qkv.unbind(0)

        # Use scaled_dot_product_attention (PyTorch 2.x)
        attn_output = F.scaled_dot_product_attention(
            q, k, v,
            dropout_p=self.dropout.p if self.training else 0.0,
            is_causal=causal,
        )

        attn_output = attn_output.transpose(1, 2).reshape(B, T, C)
        return self.out_proj(attn_output)


class TransformerBlock(nn.Module):
    """Pre-norm Transformer block with GELU activation."""

    def __init__(self, embed_dim: int, num_heads: int, ff_dim: int, dropout: float = 0.1):
        super().__init__()
        self.norm1 = nn.LayerNorm(embed_dim)
        self.attn = MultiHeadSelfAttention(embed_dim, num_heads, dropout)
        self.norm2 = nn.LayerNorm(embed_dim)
        self.ff = nn.Sequential(
            nn.Linear(embed_dim, ff_dim),
            nn.GELU(),
            nn.Dropout(dropout),
            nn.Linear(ff_dim, embed_dim),
            nn.Dropout(dropout),
        )

    def forward(self, x: torch.Tensor, causal: bool = False) -> torch.Tensor:
        x = x + self.attn(self.norm1(x), causal=causal)
        x = x + self.ff(self.norm2(x))
        return x


class VisionTransformer(nn.Module):
    """Vision Transformer (ViT) for image classification."""

    def __init__(
        self,
        image_size: int = 224,
        patch_size: int = 16,
        in_channels: int = 3,
        num_classes: int = 1000,
        embed_dim: int = 768,
        num_heads: int = 12,
        num_layers: int = 12,
        ff_dim: int = 3072,
        dropout: float = 0.1,
    ):
        super().__init__()
        num_patches = (image_size // patch_size) ** 2

        self.patch_embed = nn.Conv2d(
            in_channels, embed_dim, kernel_size=patch_size, stride=patch_size
        )
        self.cls_token = nn.Parameter(torch.randn(1, 1, embed_dim) * 0.02)
        self.pos_embed = nn.Parameter(torch.randn(1, num_patches + 1, embed_dim) * 0.02)
        self.dropout = nn.Dropout(dropout)

        self.blocks = nn.ModuleList([
            TransformerBlock(embed_dim, num_heads, ff_dim, dropout)
            for _ in range(num_layers)
        ])

        self.norm = nn.LayerNorm(embed_dim)
        self.head = nn.Linear(embed_dim, num_classes)

        self._init_weights()

    def _init_weights(self):
        for m in self.modules():
            if isinstance(m, nn.Linear):
                nn.init.trunc_normal_(m.weight, std=0.02)
                if m.bias is not None:
                    nn.init.zeros_(m.bias)
            elif isinstance(m, nn.LayerNorm):
                nn.init.ones_(m.weight)
                nn.init.zeros_(m.bias)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        B = x.shape[0]
        x = self.patch_embed(x).flatten(2).transpose(1, 2)  # (B, num_patches, embed_dim)

        cls = self.cls_token.expand(B, -1, -1)
        x = torch.cat([cls, x], dim=1)
        x = x + self.pos_embed
        x = self.dropout(x)

        for block in self.blocks:
            x = block(x)

        x = self.norm(x[:, 0])  # CLS token
        return self.head(x)
```

### Custom Layers and Autograd Functions

```python
class CustomAutograd(torch.autograd.Function):
    """Custom autograd function with manual backward pass."""

    @staticmethod
    def forward(ctx, x, alpha):
        ctx.save_for_backward(x)
        ctx.alpha = alpha
        return torch.where(x >= 0, x, alpha * (torch.exp(x) - 1))

    @staticmethod
    def backward(ctx, grad_output):
        x, = ctx.saved_tensors
        grad_x = torch.where(x >= 0, grad_output, grad_output * ctx.alpha * torch.exp(x))
        return grad_x, None  # No gradient for alpha

custom_elu = CustomAutograd.apply


class ParametricELU(nn.Module):
    """Learnable ELU with per-channel alpha."""

    def __init__(self, num_channels: int):
        super().__init__()
        self.alpha = nn.Parameter(torch.ones(num_channels))

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        alpha = self.alpha.view(1, -1, *([1] * (x.dim() - 2)))
        return torch.where(x >= 0, x, alpha * (torch.exp(x) - 1))


class SpectralNormLinear(nn.Module):
    """Linear layer with spectral normalization for stable training."""

    def __init__(self, in_features: int, out_features: int):
        super().__init__()
        self.linear = nn.utils.spectral_norm(nn.Linear(in_features, out_features))

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        return self.linear(x)


class StochasticDepth(nn.Module):
    """Stochastic depth (drop path) for residual networks."""

    def __init__(self, drop_prob: float = 0.0):
        super().__init__()
        self.drop_prob = drop_prob

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        if not self.training or self.drop_prob == 0.0:
            return x
        keep_prob = 1 - self.drop_prob
        shape = (x.shape[0],) + (1,) * (x.ndim - 1)
        random_tensor = torch.rand(shape, device=x.device, dtype=x.dtype)
        random_tensor = torch.floor(random_tensor + keep_prob)
        return x / keep_prob * random_tensor
```

---

## DataLoader Optimization

### Custom Datasets

```python
from torch.utils.data import Dataset, DataLoader, Sampler
from pathlib import Path
from PIL import Image
import json

class ImageClassificationDataset(Dataset):
    """Efficient image dataset with caching and transforms."""

    def __init__(self, root_dir: str, transform=None, cache_images: bool = False):
        self.root = Path(root_dir)
        self.transform = transform
        self.cache_images = cache_images
        self._cache = {}

        # Build file list
        self.samples = []
        self.class_to_idx = {}
        for idx, class_dir in enumerate(sorted(self.root.iterdir())):
            if class_dir.is_dir():
                self.class_to_idx[class_dir.name] = idx
                for img_path in class_dir.glob("*.[jp][pn][g]"):
                    self.samples.append((str(img_path), idx))

    def __len__(self):
        return len(self.samples)

    def __getitem__(self, idx):
        path, label = self.samples[idx]

        if self.cache_images and idx in self._cache:
            img = self._cache[idx]
        else:
            img = Image.open(path).convert("RGB")
            if self.cache_images:
                self._cache[idx] = img

        if self.transform:
            img = self.transform(img)

        return img, label


class StreamingTextDataset(Dataset):
    """Memory-mapped text dataset for large corpora."""

    def __init__(self, file_path: str, block_size: int = 512, tokenizer=None):
        self.block_size = block_size
        self.tokenizer = tokenizer

        # Memory-map the file for zero-copy reads
        import mmap
        self._file = open(file_path, "r")
        self._mmap = mmap.mmap(self._file.fileno(), 0, access=mmap.ACCESS_READ)

        # Pre-compute line offsets
        self.offsets = [0]
        for line in open(file_path, "rb"):
            self.offsets.append(self.offsets[-1] + len(line))
        self.offsets.pop()  # Remove last (EOF)

    def __len__(self):
        return len(self.offsets)

    def __getitem__(self, idx):
        self._mmap.seek(self.offsets[idx])
        line = self._mmap.readline().decode("utf-8").strip()
        if self.tokenizer:
            tokens = self.tokenizer(
                line, max_length=self.block_size,
                padding="max_length", truncation=True,
                return_tensors="pt"
            )
            return {k: v.squeeze(0) for k, v in tokens.items()}
        return line

    def __del__(self):
        self._mmap.close()
        self._file.close()
```

### Advanced DataLoader Configuration

```python
import torch.multiprocessing as mp

# Optimal DataLoader settings
def create_dataloader(
    dataset, batch_size=32, shuffle=True, num_workers=None, pin_memory=True
):
    if num_workers is None:
        num_workers = min(8, mp.cpu_count())

    return DataLoader(
        dataset,
        batch_size=batch_size,
        shuffle=shuffle,
        num_workers=num_workers,
        pin_memory=pin_memory and torch.cuda.is_available(),
        persistent_workers=True if num_workers > 0 else False,
        prefetch_factor=2 if num_workers > 0 else None,
        drop_last=True,  # For consistent batch sizes in training
        worker_init_fn=_worker_init,
        generator=torch.Generator().manual_seed(42),
    )


def _worker_init(worker_id):
    """Set unique random seed per worker to avoid duplicate augmentations."""
    import numpy as np
    seed = torch.initial_seed() % 2**32
    np.random.seed(seed + worker_id)


class BalancedBatchSampler(Sampler):
    """Sampler that ensures balanced class representation in each batch."""

    def __init__(self, labels, batch_size, samples_per_class=None):
        self.labels = labels
        self.batch_size = batch_size

        self.class_indices = {}
        for idx, label in enumerate(labels):
            if label not in self.class_indices:
                self.class_indices[label] = []
            self.class_indices[label].append(idx)

        self.num_classes = len(self.class_indices)
        self.samples_per_class = samples_per_class or (batch_size // self.num_classes)

    def __iter__(self):
        # Shuffle within each class
        shuffled = {
            cls: torch.randperm(len(indices)).tolist()
            for cls, indices in self.class_indices.items()
        }
        pointers = {cls: 0 for cls in self.class_indices}

        for _ in range(len(self)):
            batch = []
            for cls in self.class_indices:
                for _ in range(self.samples_per_class):
                    if pointers[cls] >= len(shuffled[cls]):
                        shuffled[cls] = torch.randperm(
                            len(self.class_indices[cls])
                        ).tolist()
                        pointers[cls] = 0
                    idx = self.class_indices[cls][shuffled[cls][pointers[cls]]]
                    batch.append(idx)
                    pointers[cls] += 1
            yield batch

    def __len__(self):
        max_class_size = max(len(v) for v in self.class_indices.values())
        return max_class_size // self.samples_per_class


class InfiniteSampler(Sampler):
    """Infinite sampler for endless training loops."""

    def __init__(self, dataset_size: int, shuffle: bool = True, seed: int = 0):
        self.dataset_size = dataset_size
        self.shuffle = shuffle
        self.seed = seed

    def __iter__(self):
        g = torch.Generator()
        g.manual_seed(self.seed)
        while True:
            if self.shuffle:
                yield from torch.randperm(self.dataset_size, generator=g).tolist()
            else:
                yield from range(self.dataset_size)
```

---

## Training Loops

### Production Training Loop

```python
import torch
import torch.nn as nn
from torch.amp import autocast, GradScaler
from torch.optim.lr_scheduler import CosineAnnealingWarmRestarts, OneCycleLR
import time
import logging

logger = logging.getLogger(__name__)


class Trainer:
    """Production-grade training loop with mixed precision, gradient accumulation,
    and comprehensive logging."""

    def __init__(
        self,
        model: nn.Module,
        optimizer: torch.optim.Optimizer,
        criterion: nn.Module,
        device: torch.device,
        scheduler=None,
        grad_accum_steps: int = 1,
        max_grad_norm: float = 1.0,
        use_amp: bool = True,
        log_interval: int = 50,
    ):
        self.model = model.to(device)
        self.optimizer = optimizer
        self.criterion = criterion
        self.device = device
        self.scheduler = scheduler
        self.grad_accum_steps = grad_accum_steps
        self.max_grad_norm = max_grad_norm
        self.log_interval = log_interval

        self.scaler = GradScaler("cuda", enabled=use_amp and device.type == "cuda")
        self.use_amp = use_amp
        self.global_step = 0
        self.best_val_loss = float("inf")

    def train_epoch(self, dataloader):
        self.model.train()
        total_loss = 0.0
        num_batches = 0
        start_time = time.time()

        self.optimizer.zero_grad(set_to_none=True)

        for batch_idx, (inputs, targets) in enumerate(dataloader):
            inputs = inputs.to(self.device, non_blocking=True)
            targets = targets.to(self.device, non_blocking=True)

            with autocast("cuda", enabled=self.use_amp):
                outputs = self.model(inputs)
                loss = self.criterion(outputs, targets)
                loss = loss / self.grad_accum_steps

            self.scaler.scale(loss).backward()

            if (batch_idx + 1) % self.grad_accum_steps == 0:
                self.scaler.unscale_(self.optimizer)
                grad_norm = nn.utils.clip_grad_norm_(
                    self.model.parameters(), self.max_grad_norm
                )
                self.scaler.step(self.optimizer)
                self.scaler.update()
                self.optimizer.zero_grad(set_to_none=True)

                if self.scheduler is not None:
                    self.scheduler.step()

                self.global_step += 1

            total_loss += loss.item() * self.grad_accum_steps
            num_batches += 1

            if (batch_idx + 1) % self.log_interval == 0:
                avg_loss = total_loss / num_batches
                elapsed = time.time() - start_time
                samples_per_sec = (num_batches * dataloader.batch_size) / elapsed
                lr = self.optimizer.param_groups[0]["lr"]
                logger.info(
                    f"Step {self.global_step} | Loss: {avg_loss:.4f} | "
                    f"LR: {lr:.2e} | {samples_per_sec:.0f} samples/s"
                )

        return total_loss / num_batches

    @torch.no_grad()
    def evaluate(self, dataloader):
        self.model.eval()
        total_loss = 0.0
        correct = 0
        total = 0

        for inputs, targets in dataloader:
            inputs = inputs.to(self.device, non_blocking=True)
            targets = targets.to(self.device, non_blocking=True)

            with autocast("cuda", enabled=self.use_amp):
                outputs = self.model(inputs)
                loss = self.criterion(outputs, targets)

            total_loss += loss.item()
            preds = outputs.argmax(dim=-1)
            correct += (preds == targets).sum().item()
            total += targets.size(0)

        avg_loss = total_loss / len(dataloader)
        accuracy = correct / total
        return avg_loss, accuracy

    def save_checkpoint(self, path: str, epoch: int, val_loss: float):
        checkpoint = {
            "epoch": epoch,
            "global_step": self.global_step,
            "model_state_dict": self.model.state_dict(),
            "optimizer_state_dict": self.optimizer.state_dict(),
            "scaler_state_dict": self.scaler.state_dict(),
            "val_loss": val_loss,
            "best_val_loss": self.best_val_loss,
        }
        if self.scheduler is not None:
            checkpoint["scheduler_state_dict"] = self.scheduler.state_dict()
        torch.save(checkpoint, path)
        logger.info(f"Checkpoint saved: {path}")

    def load_checkpoint(self, path: str):
        checkpoint = torch.load(path, map_location=self.device, weights_only=False)
        self.model.load_state_dict(checkpoint["model_state_dict"])
        self.optimizer.load_state_dict(checkpoint["optimizer_state_dict"])
        self.scaler.load_state_dict(checkpoint["scaler_state_dict"])
        self.global_step = checkpoint["global_step"]
        self.best_val_loss = checkpoint["best_val_loss"]
        if self.scheduler and "scheduler_state_dict" in checkpoint:
            self.scheduler.load_state_dict(checkpoint["scheduler_state_dict"])
        logger.info(f"Checkpoint loaded: {path} (epoch {checkpoint['epoch']})")
        return checkpoint["epoch"]

    def fit(self, train_loader, val_loader, num_epochs: int, checkpoint_dir: str = "checkpoints"):
        from pathlib import Path
        Path(checkpoint_dir).mkdir(parents=True, exist_ok=True)

        for epoch in range(num_epochs):
            logger.info(f"Epoch {epoch + 1}/{num_epochs}")

            train_loss = self.train_epoch(train_loader)
            val_loss, val_acc = self.evaluate(val_loader)

            logger.info(
                f"Train Loss: {train_loss:.4f} | "
                f"Val Loss: {val_loss:.4f} | Val Acc: {val_acc:.4f}"
            )

            # Save latest checkpoint
            self.save_checkpoint(
                f"{checkpoint_dir}/latest.pt", epoch, val_loss
            )

            # Save best model
            if val_loss < self.best_val_loss:
                self.best_val_loss = val_loss
                self.save_checkpoint(
                    f"{checkpoint_dir}/best.pt", epoch, val_loss
                )
                logger.info(f"New best model! Val Loss: {val_loss:.4f}")
```

### Learning Rate Schedulers

```python
# Warmup + Cosine Decay (most common for transformers)
def get_cosine_schedule_with_warmup(optimizer, num_warmup_steps, num_training_steps):
    from torch.optim.lr_scheduler import LambdaLR
    import math

    def lr_lambda(current_step):
        if current_step < num_warmup_steps:
            return float(current_step) / float(max(1, num_warmup_steps))
        progress = float(current_step - num_warmup_steps) / float(
            max(1, num_training_steps - num_warmup_steps)
        )
        return max(0.0, 0.5 * (1.0 + math.cos(math.pi * progress)))

    return LambdaLR(optimizer, lr_lambda)


# OneCycleLR — fast convergence
scheduler = OneCycleLR(
    optimizer,
    max_lr=3e-4,
    total_steps=num_training_steps,
    pct_start=0.1,        # 10% warmup
    anneal_strategy="cos",
    div_factor=25.0,       # initial_lr = max_lr / div_factor
    final_div_factor=1e4,  # final_lr = initial_lr / final_div_factor
)

# ReduceLROnPlateau — adaptive
scheduler = torch.optim.lr_scheduler.ReduceLROnPlateau(
    optimizer,
    mode="min",
    factor=0.5,
    patience=5,
    min_lr=1e-7,
)

# Cosine Annealing with Warm Restarts
scheduler = CosineAnnealingWarmRestarts(
    optimizer,
    T_0=10,      # First restart after 10 epochs
    T_mult=2,    # Double period after each restart
    eta_min=1e-6,
)
```

---

## Distributed Training

### DistributedDataParallel (DDP)

```python
import torch
import torch.distributed as dist
import torch.multiprocessing as mp
from torch.nn.parallel import DistributedDataParallel as DDP
from torch.utils.data.distributed import DistributedSampler


def setup_distributed(rank, world_size):
    """Initialize distributed training."""
    dist.init_process_group(
        backend="nccl",
        init_method="env://",
        rank=rank,
        world_size=world_size,
    )
    torch.cuda.set_device(rank)


def cleanup_distributed():
    dist.destroy_process_group()


def train_ddp(rank, world_size, args):
    setup_distributed(rank, world_size)
    device = torch.device(f"cuda:{rank}")

    # Create model on the correct device
    model = MyModel().to(device)
    model = DDP(model, device_ids=[rank], find_unused_parameters=False)

    # Distributed sampler ensures each GPU gets unique data
    train_sampler = DistributedSampler(
        train_dataset, num_replicas=world_size, rank=rank, shuffle=True
    )
    train_loader = DataLoader(
        train_dataset,
        batch_size=args.batch_size,
        sampler=train_sampler,
        num_workers=4,
        pin_memory=True,
    )

    optimizer = torch.optim.AdamW(model.parameters(), lr=args.lr)

    for epoch in range(args.epochs):
        train_sampler.set_epoch(epoch)  # Crucial for proper shuffling
        model.train()

        for batch_idx, (inputs, targets) in enumerate(train_loader):
            inputs = inputs.to(device, non_blocking=True)
            targets = targets.to(device, non_blocking=True)

            outputs = model(inputs)
            loss = F.cross_entropy(outputs, targets)

            optimizer.zero_grad(set_to_none=True)
            loss.backward()
            optimizer.step()

        # Aggregate metrics across GPUs
        avg_loss = reduce_tensor(loss, world_size)
        if rank == 0:
            print(f"Epoch {epoch}: Loss = {avg_loss:.4f}")

    cleanup_distributed()


def reduce_tensor(tensor, world_size):
    """Average a tensor across all GPUs."""
    rt = tensor.clone()
    dist.all_reduce(rt, op=dist.ReduceOp.SUM)
    rt /= world_size
    return rt.item()


# Launch training
if __name__ == "__main__":
    world_size = torch.cuda.device_count()
    mp.spawn(train_ddp, args=(world_size, args), nprocs=world_size, join=True)
```

### Fully Sharded Data Parallel (FSDP)

```python
import torch
from torch.distributed.fsdp import (
    FullyShardedDataParallel as FSDP,
    MixedPrecision,
    ShardingStrategy,
    BackwardPrefetch,
    CPUOffload,
)
from torch.distributed.fsdp.wrap import (
    transformer_auto_wrap_policy,
    size_based_auto_wrap_policy,
)
import functools


def setup_fsdp_model(model, device):
    """Configure FSDP for large model training."""

    # Mixed precision policy
    mixed_precision = MixedPrecision(
        param_dtype=torch.bfloat16,
        reduce_dtype=torch.bfloat16,
        buffer_dtype=torch.bfloat16,
    )

    # Auto-wrap policy: wrap each TransformerBlock separately
    wrap_policy = functools.partial(
        transformer_auto_wrap_policy,
        transformer_layer_cls={TransformerBlock},
    )

    # Or wrap based on parameter count
    size_policy = functools.partial(
        size_based_auto_wrap_policy,
        min_num_params=1e6,  # Wrap modules with > 1M params
    )

    model = FSDP(
        model,
        auto_wrap_policy=wrap_policy,
        mixed_precision=mixed_precision,
        sharding_strategy=ShardingStrategy.FULL_SHARD,  # Shard everything
        backward_prefetch=BackwardPrefetch.BACKWARD_PRE,
        cpu_offload=CPUOffload(offload_params=False),
        device_id=device,
        use_orig_params=True,  # Required for torch.compile compatibility
        limit_all_gathers=True,
    )

    return model


# FSDP-compatible checkpoint saving
from torch.distributed.fsdp import (
    FullStateDictConfig,
    StateDictType,
)


def save_fsdp_checkpoint(model, optimizer, path, rank):
    """Save FSDP checkpoint — only rank 0 saves."""
    save_policy = FullStateDictConfig(offload_to_cpu=True, rank0_only=True)

    with FSDP.state_dict_type(model, StateDictType.FULL_STATE_DICT, save_policy):
        state_dict = model.state_dict()
        optim_state = FSDP.optim_state_dict(model, optimizer)

    if rank == 0:
        torch.save(
            {"model": state_dict, "optimizer": optim_state},
            path,
        )


def load_fsdp_checkpoint(model, optimizer, path):
    """Load FSDP checkpoint."""
    checkpoint = torch.load(path, map_location="cpu", weights_only=False)

    with FSDP.state_dict_type(model, StateDictType.FULL_STATE_DICT):
        model.load_state_dict(checkpoint["model"])

    optim_state = FSDP.optim_state_dict_to_load(
        checkpoint["optimizer"], model, optimizer
    )
    optimizer.load_state_dict(optim_state)
```

---

## Mixed Precision Training

### torch.amp Patterns

```python
from torch.amp import autocast, GradScaler

# Basic mixed precision training
scaler = GradScaler("cuda")

for inputs, targets in dataloader:
    inputs = inputs.to(device)
    targets = targets.to(device)

    optimizer.zero_grad(set_to_none=True)

    # Forward pass in mixed precision
    with autocast("cuda"):
        outputs = model(inputs)
        loss = criterion(outputs, targets)

    # Backward pass with scaled gradients
    scaler.scale(loss).backward()

    # Unscale before clipping
    scaler.unscale_(optimizer)
    torch.nn.utils.clip_grad_norm_(model.parameters(), max_norm=1.0)

    # Step with scaler
    scaler.step(optimizer)
    scaler.update()

# BFloat16 — better dynamic range, no need for scaler
with autocast("cuda", dtype=torch.bfloat16):
    outputs = model(inputs)
    loss = criterion(outputs, targets)
loss.backward()
optimizer.step()


# Disable autocast for specific operations
with autocast("cuda"):
    output = model(input)
    # Force full precision for numerically sensitive operations
    with autocast("cuda", enabled=False):
        loss = custom_loss(output.float(), target.float())
```

---

## Profiling and Debugging

### torch.profiler

```python
from torch.profiler import profile, record_function, ProfilerActivity, schedule

# Basic profiling
with profile(
    activities=[ProfilerActivity.CPU, ProfilerActivity.CUDA],
    record_shapes=True,
    profile_memory=True,
    with_stack=True,
) as prof:
    with record_function("model_inference"):
        output = model(input_tensor)

# Print summary sorted by GPU time
print(prof.key_averages().table(sort_by="cuda_time_total", row_limit=20))

# Export for Chrome trace viewer
prof.export_chrome_trace("trace.json")

# Export for TensorBoard
prof.export_stacks("profiler_stacks.txt", "self_cuda_time_total")


# Schedule-based profiling for training loops
def trace_handler(p):
    output = p.key_averages().table(sort_by="self_cuda_time_total", row_limit=10)
    print(output)
    p.export_chrome_trace(f"traces/trace_{p.step_num}.json")

with profile(
    activities=[ProfilerActivity.CPU, ProfilerActivity.CUDA],
    schedule=schedule(wait=1, warmup=1, active=3, repeat=2),
    on_trace_ready=trace_handler,
    record_shapes=True,
    profile_memory=True,
    with_stack=True,
) as prof:
    for step, (inputs, targets) in enumerate(dataloader):
        if step >= 2 + 2 * (1 + 1 + 3):
            break
        with record_function("train_step"):
            outputs = model(inputs.to(device))
            loss = criterion(outputs, targets.to(device))
            loss.backward()
            optimizer.step()
            optimizer.zero_grad()
        prof.step()
```

### Debugging NaN/Inf Gradients

```python
# Detect anomalies (NaN/Inf in forward/backward)
torch.autograd.set_detect_anomaly(True)

# Manual NaN checking
def check_nan_hook(module, input, output):
    if isinstance(output, torch.Tensor) and torch.isnan(output).any():
        raise RuntimeError(f"NaN detected in {module.__class__.__name__}")

for name, module in model.named_modules():
    module.register_forward_hook(check_nan_hook)

# Gradient monitoring hook
def gradient_monitor_hook(name):
    def hook(grad):
        if torch.isnan(grad).any():
            print(f"NaN gradient in {name}")
        if torch.isinf(grad).any():
            print(f"Inf gradient in {name}")
        grad_norm = grad.norm().item()
        if grad_norm > 100:
            print(f"Large gradient in {name}: {grad_norm:.2f}")
    return hook

for name, param in model.named_parameters():
    if param.requires_grad:
        param.register_hook(gradient_monitor_hook(name))
```

---

## Model Export and Deployment

### TorchScript Export

```python
# Tracing — works for models without control flow
example_input = torch.randn(1, 3, 224, 224, device=device)
traced_model = torch.jit.trace(model.eval(), example_input)
traced_model.save("model_traced.pt")

# Scripting — handles control flow
scripted_model = torch.jit.script(model.eval())
scripted_model.save("model_scripted.pt")

# Load and run
loaded = torch.jit.load("model_traced.pt", map_location=device)
output = loaded(example_input)

# Optimize for inference
optimized = torch.jit.optimize_for_inference(traced_model)
```

### ONNX Export

```python
import torch.onnx

model.eval()
dummy_input = torch.randn(1, 3, 224, 224, device=device)

torch.onnx.export(
    model,
    dummy_input,
    "model.onnx",
    export_params=True,
    opset_version=17,
    do_constant_folding=True,
    input_names=["input"],
    output_names=["output"],
    dynamic_axes={
        "input": {0: "batch_size"},
        "output": {0: "batch_size"},
    },
)

# Verify ONNX model
import onnx
onnx_model = onnx.load("model.onnx")
onnx.checker.check_model(onnx_model)

# Run with ONNX Runtime
import onnxruntime as ort

session = ort.InferenceSession("model.onnx", providers=["CUDAExecutionProvider"])
inputs = {"input": dummy_input.cpu().numpy()}
outputs = session.run(None, inputs)
```

### torch.export for PyTorch 2.x

```python
from torch.export import export, save, load

model.eval()
example_inputs = (torch.randn(1, 3, 224, 224, device=device),)

# Export with dynamic shapes
from torch.export import Dim

batch_dim = Dim("batch", min=1, max=64)
dynamic_shapes = {"x": {0: batch_dim}}

exported = export(model, example_inputs, dynamic_shapes=dynamic_shapes)

# Save and load
save(exported, "model.pt2")
loaded = load("model.pt2")

# Run
output = loaded.module()(torch.randn(8, 3, 224, 224, device=device))
```

---

## Common Architectures

### ResNet-style CNN

```python
class ResidualBlock(nn.Module):
    def __init__(self, in_channels, out_channels, stride=1):
        super().__init__()
        self.conv1 = nn.Conv2d(in_channels, out_channels, 3, stride, 1, bias=False)
        self.bn1 = nn.BatchNorm2d(out_channels)
        self.conv2 = nn.Conv2d(out_channels, out_channels, 3, 1, 1, bias=False)
        self.bn2 = nn.BatchNorm2d(out_channels)

        self.shortcut = nn.Sequential()
        if stride != 1 or in_channels != out_channels:
            self.shortcut = nn.Sequential(
                nn.Conv2d(in_channels, out_channels, 1, stride, bias=False),
                nn.BatchNorm2d(out_channels),
            )

    def forward(self, x):
        out = F.relu(self.bn1(self.conv1(x)))
        out = self.bn2(self.conv2(out))
        out += self.shortcut(x)
        return F.relu(out)
```

### GPT-style Language Model

```python
class GPT(nn.Module):
    def __init__(self, vocab_size, embed_dim, num_heads, num_layers, max_seq_len, dropout=0.1):
        super().__init__()
        self.token_embed = nn.Embedding(vocab_size, embed_dim)
        self.pos_embed = nn.Embedding(max_seq_len, embed_dim)
        self.dropout = nn.Dropout(dropout)

        self.blocks = nn.ModuleList([
            TransformerBlock(embed_dim, num_heads, embed_dim * 4, dropout)
            for _ in range(num_layers)
        ])
        self.norm = nn.LayerNorm(embed_dim)
        self.head = nn.Linear(embed_dim, vocab_size, bias=False)

        # Weight tying
        self.head.weight = self.token_embed.weight

    def forward(self, idx, targets=None):
        B, T = idx.shape
        pos = torch.arange(T, device=idx.device)

        x = self.dropout(self.token_embed(idx) + self.pos_embed(pos))
        for block in self.blocks:
            x = block(x, causal=True)
        x = self.norm(x)
        logits = self.head(x)

        loss = None
        if targets is not None:
            loss = F.cross_entropy(logits.view(-1, logits.size(-1)), targets.view(-1))

        return logits, loss

    @torch.no_grad()
    def generate(self, idx, max_new_tokens, temperature=1.0, top_k=None):
        for _ in range(max_new_tokens):
            idx_cond = idx[:, -self.pos_embed.num_embeddings:]
            logits, _ = self(idx_cond)
            logits = logits[:, -1, :] / temperature

            if top_k is not None:
                v, _ = torch.topk(logits, min(top_k, logits.size(-1)))
                logits[logits < v[:, [-1]]] = float("-inf")

            probs = F.softmax(logits, dim=-1)
            idx_next = torch.multinomial(probs, num_samples=1)
            idx = torch.cat([idx, idx_next], dim=1)
        return idx
```

### U-Net for Segmentation

```python
class UNet(nn.Module):
    def __init__(self, in_channels=3, out_channels=1, features=[64, 128, 256, 512]):
        super().__init__()
        self.downs = nn.ModuleList()
        self.ups = nn.ModuleList()
        self.pool = nn.MaxPool2d(kernel_size=2, stride=2)

        # Encoder
        for feature in features:
            self.downs.append(self._double_conv(in_channels, feature))
            in_channels = feature

        # Bottleneck
        self.bottleneck = self._double_conv(features[-1], features[-1] * 2)

        # Decoder
        for feature in reversed(features):
            self.ups.append(nn.ConvTranspose2d(feature * 2, feature, kernel_size=2, stride=2))
            self.ups.append(self._double_conv(feature * 2, feature))

        self.final = nn.Conv2d(features[0], out_channels, kernel_size=1)

    def _double_conv(self, in_ch, out_ch):
        return nn.Sequential(
            nn.Conv2d(in_ch, out_ch, 3, 1, 1, bias=False),
            nn.BatchNorm2d(out_ch),
            nn.ReLU(inplace=True),
            nn.Conv2d(out_ch, out_ch, 3, 1, 1, bias=False),
            nn.BatchNorm2d(out_ch),
            nn.ReLU(inplace=True),
        )

    def forward(self, x):
        skip_connections = []
        for down in self.downs:
            x = down(x)
            skip_connections.append(x)
            x = self.pool(x)

        x = self.bottleneck(x)
        skip_connections = skip_connections[::-1]

        for idx in range(0, len(self.ups), 2):
            x = self.ups[idx](x)
            skip = skip_connections[idx // 2]
            if x.shape != skip.shape:
                x = F.interpolate(x, size=skip.shape[2:])
            x = torch.cat([skip, x], dim=1)
            x = self.ups[idx + 1](x)

        return self.final(x)
```

---

## Optimizer Best Practices

### AdamW Configuration

```python
# Separate weight decay for different parameter groups
def get_param_groups(model, lr=3e-4, weight_decay=0.01):
    decay = []
    no_decay = []

    for name, param in model.named_parameters():
        if not param.requires_grad:
            continue
        if param.ndim <= 1 or "bias" in name or "norm" in name or "embed" in name:
            no_decay.append(param)
        else:
            decay.append(param)

    return [
        {"params": decay, "weight_decay": weight_decay, "lr": lr},
        {"params": no_decay, "weight_decay": 0.0, "lr": lr},
    ]

optimizer = torch.optim.AdamW(
    get_param_groups(model),
    betas=(0.9, 0.95),
    eps=1e-8,
    fused=True,  # PyTorch 2.x: fused CUDA implementation
)
```

### Exponential Moving Average (EMA)

```python
class EMA:
    """Exponential Moving Average of model parameters for stable evaluation."""

    def __init__(self, model, decay=0.999):
        self.model = model
        self.decay = decay
        self.shadow = {}
        self.backup = {}
        self._register()

    def _register(self):
        for name, param in self.model.named_parameters():
            if param.requires_grad:
                self.shadow[name] = param.data.clone()

    @torch.no_grad()
    def update(self):
        for name, param in self.model.named_parameters():
            if param.requires_grad:
                self.shadow[name].lerp_(param.data, 1 - self.decay)

    def apply_shadow(self):
        for name, param in self.model.named_parameters():
            if param.requires_grad:
                self.backup[name] = param.data.clone()
                param.data.copy_(self.shadow[name])

    def restore(self):
        for name, param in self.model.named_parameters():
            if param.requires_grad:
                param.data.copy_(self.backup[name])
        self.backup = {}
```

---

## Data Augmentation

### torchvision v2 Transforms

```python
from torchvision.transforms import v2

# Training transforms with modern API
train_transform = v2.Compose([
    v2.RandomResizedCrop(224, scale=(0.08, 1.0), antialias=True),
    v2.RandomHorizontalFlip(p=0.5),
    v2.ColorJitter(brightness=0.4, contrast=0.4, saturation=0.4, hue=0.1),
    v2.RandomGrayscale(p=0.2),
    v2.GaussianBlur(kernel_size=23, sigma=(0.1, 2.0)),
    v2.ToImage(),
    v2.ToDtype(torch.float32, scale=True),
    v2.Normalize(mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225]),
    v2.RandomErasing(p=0.25),
])

# Validation transforms
val_transform = v2.Compose([
    v2.Resize(256, antialias=True),
    v2.CenterCrop(224),
    v2.ToImage(),
    v2.ToDtype(torch.float32, scale=True),
    v2.Normalize(mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225]),
])


# Mixup and CutMix
from torchvision.transforms import v2

mixup = v2.MixUp(alpha=0.2, num_classes=1000)
cutmix = v2.CutMix(alpha=1.0, num_classes=1000)
mixup_cutmix = v2.RandomChoice([mixup, cutmix])

for images, labels in dataloader:
    images, labels = mixup_cutmix(images, labels)
    # labels are now soft (one-hot-like), use cross_entropy directly
    outputs = model(images)
    loss = F.cross_entropy(outputs, labels)
```

---

## Best Practices Summary

### Training Checklist

1. **Set random seeds** for reproducibility: `torch.manual_seed(42)`, `torch.cuda.manual_seed_all(42)`
2. **Use `set_to_none=True`** in `optimizer.zero_grad()` for slight memory savings
3. **Enable `torch.backends.cudnn.benchmark = True`** for fixed input sizes
4. **Use `non_blocking=True`** when transferring tensors to GPU
5. **Use `pin_memory=True`** in DataLoaders for GPU training
6. **Use `persistent_workers=True`** to avoid worker restart overhead
7. **Compile your model** with `torch.compile()` for 10-30% speedup
8. **Use `torch.amp`** for mixed precision training
9. **Gradient accumulation** for effective larger batch sizes
10. **Clip gradients** to prevent exploding gradients

### Common Pitfalls

- Forgetting `model.eval()` and `torch.no_grad()` during evaluation
- Not setting `train_sampler.set_epoch(epoch)` in DDP
- Using `model.parameters()` instead of `model.module.parameters()` with DDP
- Data leakage between train/val splits in preprocessing
- Not handling `nan` losses — check inputs and learning rate
- Using `torch.Tensor` (creates float tensor) vs `torch.tensor` (infers dtype)
- Not accounting for gradient accumulation in learning rate scaling
- Forgetting to move both model AND data to the same device

### When to Use What

| Scenario | Recommendation |
|----------|---------------|
| Single GPU, < 1B params | `torch.compile` + AMP |
| Multi-GPU, < 10B params | DDP + AMP |
| Multi-GPU, 10B+ params | FSDP + BF16 |
| Very large (100B+) | FSDP + CPU offload or DeepSpeed |
| Inference optimization | `torch.compile(mode="max-autotune")` + ONNX |
| Mobile deployment | TorchScript + quantization |
| Edge devices | ONNX + TensorRT or CoreML |
