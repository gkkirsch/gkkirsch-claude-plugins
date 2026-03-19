# PyTorch Patterns — Model Architectures, Training, and Optimization

Quick reference for common PyTorch patterns, model architectures, training techniques, and performance optimization.

---

## Model Architecture Patterns

### Convolutional Neural Networks

#### Basic CNN Block

```python
import torch
import torch.nn as nn
import torch.nn.functional as F


def conv_block(in_ch, out_ch, kernel_size=3, stride=1, groups=1, act=True):
    """Standard Conv-BN-ReLU block."""
    layers = [
        nn.Conv2d(in_ch, out_ch, kernel_size, stride, kernel_size // 2,
                  groups=groups, bias=False),
        nn.BatchNorm2d(out_ch),
    ]
    if act:
        layers.append(nn.ReLU(inplace=True))
    return nn.Sequential(*layers)
```

#### Inverted Residual Block (MobileNetV2)

```python
class InvertedResidual(nn.Module):
    def __init__(self, in_ch, out_ch, stride, expand_ratio):
        super().__init__()
        hidden = int(in_ch * expand_ratio)
        self.use_residual = stride == 1 and in_ch == out_ch

        layers = []
        if expand_ratio != 1:
            layers.append(conv_block(in_ch, hidden, 1))
        layers.extend([
            conv_block(hidden, hidden, 3, stride, groups=hidden),  # Depthwise
            conv_block(hidden, out_ch, 1, act=False),              # Pointwise
        ])
        self.conv = nn.Sequential(*layers)

    def forward(self, x):
        if self.use_residual:
            return x + self.conv(x)
        return self.conv(x)
```

#### ConvNeXt Block

```python
class ConvNeXtBlock(nn.Module):
    """ConvNeXt block: depthwise conv -> LayerNorm -> 1x1 -> GELU -> 1x1."""
    def __init__(self, dim, expansion=4, drop_path=0.0):
        super().__init__()
        self.dwconv = nn.Conv2d(dim, dim, 7, padding=3, groups=dim)
        self.norm = nn.LayerNorm(dim)
        self.pwconv1 = nn.Linear(dim, dim * expansion)
        self.act = nn.GELU()
        self.pwconv2 = nn.Linear(dim * expansion, dim)
        self.gamma = nn.Parameter(torch.ones(dim) * 1e-6)
        self.drop_path = StochasticDepth(drop_path) if drop_path > 0 else nn.Identity()

    def forward(self, x):
        shortcut = x
        x = self.dwconv(x)
        x = x.permute(0, 2, 3, 1)  # (B, C, H, W) -> (B, H, W, C)
        x = self.norm(x)
        x = self.pwconv1(x)
        x = self.act(x)
        x = self.pwconv2(x)
        x = self.gamma * x
        x = x.permute(0, 3, 1, 2)  # Back to (B, C, H, W)
        return shortcut + self.drop_path(x)
```

### Attention Mechanisms

#### Multi-Head Attention

```python
class MultiHeadAttention(nn.Module):
    def __init__(self, dim, num_heads=8, qkv_bias=False, attn_drop=0.0, proj_drop=0.0):
        super().__init__()
        self.num_heads = num_heads
        self.head_dim = dim // num_heads
        self.scale = self.head_dim ** -0.5

        self.qkv = nn.Linear(dim, dim * 3, bias=qkv_bias)
        self.attn_drop = nn.Dropout(attn_drop)
        self.proj = nn.Linear(dim, dim)
        self.proj_drop = nn.Dropout(proj_drop)

    def forward(self, x):
        B, N, C = x.shape
        qkv = self.qkv(x).reshape(B, N, 3, self.num_heads, self.head_dim).permute(2, 0, 3, 1, 4)
        q, k, v = qkv.unbind(0)

        # Use flash attention (PyTorch 2.x)
        x = F.scaled_dot_product_attention(q, k, v, dropout_p=self.attn_drop.p if self.training else 0)

        x = x.transpose(1, 2).reshape(B, N, C)
        x = self.proj(x)
        x = self.proj_drop(x)
        return x
```

#### Cross-Attention

```python
class CrossAttention(nn.Module):
    def __init__(self, dim, context_dim, num_heads=8):
        super().__init__()
        self.num_heads = num_heads
        self.head_dim = dim // num_heads
        self.scale = self.head_dim ** -0.5

        self.q = nn.Linear(dim, dim)
        self.kv = nn.Linear(context_dim, dim * 2)
        self.proj = nn.Linear(dim, dim)

    def forward(self, x, context):
        B, N, C = x.shape
        _, S, _ = context.shape
        H = self.num_heads
        D = self.head_dim

        q = self.q(x).reshape(B, N, H, D).permute(0, 2, 1, 3)
        kv = self.kv(context).reshape(B, S, 2, H, D).permute(2, 0, 3, 1, 4)
        k, v = kv.unbind(0)

        x = F.scaled_dot_product_attention(q, k, v)
        x = x.transpose(1, 2).reshape(B, N, C)
        return self.proj(x)
```

#### Rotary Position Embedding (RoPE)

```python
class RotaryEmbedding(nn.Module):
    def __init__(self, dim, max_seq_len=8192, base=10000):
        super().__init__()
        inv_freq = 1.0 / (base ** (torch.arange(0, dim, 2).float() / dim))
        self.register_buffer("inv_freq", inv_freq)

        t = torch.arange(max_seq_len)
        freqs = torch.einsum("i,j->ij", t, self.inv_freq)
        emb = torch.cat([freqs, freqs], dim=-1)
        self.register_buffer("cos_cached", emb.cos())
        self.register_buffer("sin_cached", emb.sin())

    def forward(self, x, seq_len):
        return self.cos_cached[:seq_len], self.sin_cached[:seq_len]


def rotate_half(x):
    x1, x2 = x.chunk(2, dim=-1)
    return torch.cat([-x2, x1], dim=-1)


def apply_rotary_emb(q, k, cos, sin):
    q_embed = (q * cos) + (rotate_half(q) * sin)
    k_embed = (k * cos) + (rotate_half(k) * sin)
    return q_embed, k_embed
```

### Normalization Variants

```python
# RMSNorm (used in LLaMA, Gemma)
class RMSNorm(nn.Module):
    def __init__(self, dim, eps=1e-6):
        super().__init__()
        self.weight = nn.Parameter(torch.ones(dim))
        self.eps = eps

    def forward(self, x):
        norm = torch.rsqrt(x.pow(2).mean(-1, keepdim=True) + self.eps)
        return x * norm * self.weight


# Group Normalization
group_norm = nn.GroupNorm(num_groups=32, num_channels=256)

# Instance Normalization (for style transfer)
instance_norm = nn.InstanceNorm2d(256, affine=True)
```

---

## Training Loop Patterns

### Standard Training Step

```python
def train_one_epoch(model, dataloader, optimizer, criterion, device, scaler=None):
    model.train()
    total_loss = 0
    correct = 0
    total = 0

    for inputs, targets in dataloader:
        inputs = inputs.to(device, non_blocking=True)
        targets = targets.to(device, non_blocking=True)

        optimizer.zero_grad(set_to_none=True)

        if scaler:  # Mixed precision
            with torch.autocast("cuda"):
                outputs = model(inputs)
                loss = criterion(outputs, targets)
            scaler.scale(loss).backward()
            scaler.unscale_(optimizer)
            torch.nn.utils.clip_grad_norm_(model.parameters(), 1.0)
            scaler.step(optimizer)
            scaler.update()
        else:
            outputs = model(inputs)
            loss = criterion(outputs, targets)
            loss.backward()
            torch.nn.utils.clip_grad_norm_(model.parameters(), 1.0)
            optimizer.step()

        total_loss += loss.item() * inputs.size(0)
        correct += (outputs.argmax(1) == targets).sum().item()
        total += targets.size(0)

    return total_loss / total, correct / total
```

### Gradient Accumulation

```python
def train_with_accumulation(model, dataloader, optimizer, criterion, device,
                            accum_steps=4, scaler=None):
    model.train()
    optimizer.zero_grad(set_to_none=True)

    for step, (inputs, targets) in enumerate(dataloader):
        inputs = inputs.to(device, non_blocking=True)
        targets = targets.to(device, non_blocking=True)

        with torch.autocast("cuda", enabled=scaler is not None):
            outputs = model(inputs)
            loss = criterion(outputs, targets) / accum_steps

        if scaler:
            scaler.scale(loss).backward()
        else:
            loss.backward()

        if (step + 1) % accum_steps == 0:
            if scaler:
                scaler.unscale_(optimizer)
                torch.nn.utils.clip_grad_norm_(model.parameters(), 1.0)
                scaler.step(optimizer)
                scaler.update()
            else:
                torch.nn.utils.clip_grad_norm_(model.parameters(), 1.0)
                optimizer.step()
            optimizer.zero_grad(set_to_none=True)
```

### Knowledge Distillation

```python
def distillation_loss(student_logits, teacher_logits, targets,
                      temperature=4.0, alpha=0.5):
    """Combined hard label + soft label loss for distillation."""
    soft_targets = F.softmax(teacher_logits / temperature, dim=-1)
    soft_loss = F.kl_div(
        F.log_softmax(student_logits / temperature, dim=-1),
        soft_targets,
        reduction="batchmean",
    ) * (temperature ** 2)

    hard_loss = F.cross_entropy(student_logits, targets)

    return alpha * soft_loss + (1 - alpha) * hard_loss
```

### Contrastive Learning (SimCLR)

```python
def nt_xent_loss(z1, z2, temperature=0.5):
    """Normalized temperature-scaled cross-entropy loss."""
    z1 = F.normalize(z1, dim=1)
    z2 = F.normalize(z2, dim=1)

    batch_size = z1.size(0)
    z = torch.cat([z1, z2], dim=0)  # (2B, D)

    sim = torch.mm(z, z.t()) / temperature  # (2B, 2B)

    # Mask out self-similarities
    mask = ~torch.eye(2 * batch_size, dtype=torch.bool, device=z.device)
    sim = sim.masked_select(mask).reshape(2 * batch_size, -1)

    # Positive pairs
    labels = torch.cat([
        torch.arange(batch_size, 2 * batch_size),
        torch.arange(0, batch_size),
    ]).to(z.device)

    # Adjust labels for removed diagonal
    adjusted_labels = labels - (labels > torch.arange(2 * batch_size, device=z.device)).long()

    return F.cross_entropy(sim, adjusted_labels)
```

---

## Mixed Precision Patterns

### BFloat16 Training (Recommended for Ampere+)

```python
# BFloat16 — no scaler needed
model = model.to(device)
optimizer = torch.optim.AdamW(model.parameters(), lr=1e-4)

for inputs, targets in dataloader:
    inputs = inputs.to(device)
    targets = targets.to(device)

    with torch.autocast("cuda", dtype=torch.bfloat16):
        outputs = model(inputs)
        loss = criterion(outputs, targets)

    loss.backward()
    optimizer.step()
    optimizer.zero_grad(set_to_none=True)
```

### Float16 Training (Legacy GPUs)

```python
scaler = torch.amp.GradScaler("cuda")

for inputs, targets in dataloader:
    inputs = inputs.to(device)
    targets = targets.to(device)

    with torch.autocast("cuda", dtype=torch.float16):
        outputs = model(inputs)
        loss = criterion(outputs, targets)

    scaler.scale(loss).backward()
    scaler.unscale_(optimizer)
    torch.nn.utils.clip_grad_norm_(model.parameters(), 1.0)
    scaler.step(optimizer)
    scaler.update()
    optimizer.zero_grad(set_to_none=True)
```

---

## Custom Dataset Patterns

### Image Dataset with Caching

```python
class CachedImageDataset(torch.utils.data.Dataset):
    def __init__(self, paths, labels, transform=None, cache_size=10000):
        self.paths = paths
        self.labels = labels
        self.transform = transform
        self.cache = {}
        self.cache_size = cache_size

    def __len__(self):
        return len(self.paths)

    def __getitem__(self, idx):
        if idx in self.cache:
            img = self.cache[idx]
        else:
            from PIL import Image
            img = Image.open(self.paths[idx]).convert("RGB")
            if len(self.cache) < self.cache_size:
                self.cache[idx] = img

        if self.transform:
            img = self.transform(img)
        return img, self.labels[idx]
```

### Multi-Modal Dataset

```python
class MultiModalDataset(torch.utils.data.Dataset):
    def __init__(self, images, texts, labels, image_transform=None, tokenizer=None, max_len=128):
        self.images = images
        self.texts = texts
        self.labels = labels
        self.image_transform = image_transform
        self.tokenizer = tokenizer
        self.max_len = max_len

    def __len__(self):
        return len(self.labels)

    def __getitem__(self, idx):
        from PIL import Image
        img = Image.open(self.images[idx]).convert("RGB")
        if self.image_transform:
            img = self.image_transform(img)

        text = self.texts[idx]
        if self.tokenizer:
            tokens = self.tokenizer(
                text, max_length=self.max_len,
                padding="max_length", truncation=True,
                return_tensors="pt"
            )
            tokens = {k: v.squeeze(0) for k, v in tokens.items()}
        else:
            tokens = text

        return {"image": img, "text": tokens, "label": self.labels[idx]}
```

### Streaming / Iterable Dataset

```python
class StreamingDataset(torch.utils.data.IterableDataset):
    def __init__(self, file_paths, transform=None):
        self.file_paths = file_paths
        self.transform = transform

    def __iter__(self):
        worker_info = torch.utils.data.get_worker_info()
        if worker_info is None:
            files = self.file_paths
        else:
            # Split files across workers
            per_worker = len(self.file_paths) // worker_info.num_workers
            start = worker_info.id * per_worker
            end = start + per_worker if worker_info.id < worker_info.num_workers - 1 else len(self.file_paths)
            files = self.file_paths[start:end]

        for path in files:
            for sample in self._read_file(path):
                if self.transform:
                    sample = self.transform(sample)
                yield sample

    def _read_file(self, path):
        import json
        with open(path) as f:
            for line in f:
                yield json.loads(line)
```

---

## Gradient Accumulation Patterns

```python
# Effective batch size = batch_size * accum_steps * num_gpus
effective_batch = 32 * 4 * 2  # = 256

# Scale learning rate with effective batch size
base_lr = 1e-4
lr = base_lr * (effective_batch / 256)
```

---

## Model Quantization

### Post-Training Quantization

```python
# Dynamic quantization (CPU inference)
quantized = torch.quantization.quantize_dynamic(
    model.cpu(), {nn.Linear}, dtype=torch.qint8
)

# Static quantization
model.eval()
model.qconfig = torch.quantization.get_default_qconfig("x86")
model_prepared = torch.quantization.prepare(model)

# Calibrate with representative data
with torch.no_grad():
    for inputs, _ in calibration_loader:
        model_prepared(inputs)

model_quantized = torch.quantization.convert(model_prepared)
```

### Quantization-Aware Training

```python
model.train()
model.qconfig = torch.quantization.get_default_qat_qconfig("x86")
model_qat = torch.quantization.prepare_qat(model)

# Train as normal
for epoch in range(num_epochs):
    for inputs, targets in train_loader:
        outputs = model_qat(inputs)
        loss = criterion(outputs, targets)
        loss.backward()
        optimizer.step()
        optimizer.zero_grad()

# Convert to quantized
model_qat.eval()
model_quantized = torch.quantization.convert(model_qat)
```

---

## Loss Function Recipes

### Label Smoothing Cross-Entropy

```python
class LabelSmoothingCrossEntropy(nn.Module):
    def __init__(self, smoothing=0.1):
        super().__init__()
        self.smoothing = smoothing

    def forward(self, pred, target):
        n_classes = pred.size(-1)
        log_preds = F.log_softmax(pred, dim=-1)
        nll = -log_preds.gather(dim=-1, index=target.unsqueeze(-1)).squeeze(-1)
        smooth = -log_preds.mean(dim=-1)
        loss = (1 - self.smoothing) * nll + self.smoothing * smooth
        return loss.mean()
```

### Focal Loss

```python
class FocalLoss(nn.Module):
    def __init__(self, alpha=0.25, gamma=2.0):
        super().__init__()
        self.alpha = alpha
        self.gamma = gamma

    def forward(self, pred, target):
        ce_loss = F.cross_entropy(pred, target, reduction="none")
        pt = torch.exp(-ce_loss)
        focal_loss = self.alpha * (1 - pt) ** self.gamma * ce_loss
        return focal_loss.mean()
```

### Dice Loss (Segmentation)

```python
class DiceLoss(nn.Module):
    def __init__(self, smooth=1.0):
        super().__init__()
        self.smooth = smooth

    def forward(self, pred, target):
        pred = torch.sigmoid(pred)
        pred_flat = pred.view(-1)
        target_flat = target.view(-1)
        intersection = (pred_flat * target_flat).sum()
        return 1 - (2 * intersection + self.smooth) / (
            pred_flat.sum() + target_flat.sum() + self.smooth
        )
```

### Triplet Loss

```python
class TripletLoss(nn.Module):
    def __init__(self, margin=1.0):
        super().__init__()
        self.margin = margin

    def forward(self, anchor, positive, negative):
        pos_dist = F.pairwise_distance(anchor, positive)
        neg_dist = F.pairwise_distance(anchor, negative)
        loss = F.relu(pos_dist - neg_dist + self.margin)
        return loss.mean()
```

---

## Performance Optimization

### Compile + CUDA Graphs

```python
# torch.compile for general speedup
model = torch.compile(model, mode="reduce-overhead")

# Manual CUDA graphs for maximum throughput
static_input = torch.randn(batch_size, *input_shape, device="cuda")
static_target = torch.randint(0, num_classes, (batch_size,), device="cuda")

# Warmup
for _ in range(3):
    output = model(static_input)
    loss = criterion(output, static_target)
    loss.backward()
    optimizer.step()
    optimizer.zero_grad()

# Capture graph
g = torch.cuda.CUDAGraph()
with torch.cuda.graph(g):
    output = model(static_input)
    loss = criterion(output, static_target)
    loss.backward()
    optimizer.step()
    optimizer.zero_grad()

# Replay (very fast)
for inputs, targets in dataloader:
    static_input.copy_(inputs)
    static_target.copy_(targets)
    g.replay()
```

### Memory Optimization

```python
# Gradient checkpointing
from torch.utils.checkpoint import checkpoint_sequential

class EfficientModel(nn.Module):
    def __init__(self, layers):
        super().__init__()
        self.layers = nn.ModuleList(layers)

    def forward(self, x):
        # Checkpoint every 2 layers
        return checkpoint_sequential(self.layers, 2, x, use_reentrant=False)


# Activation offloading to CPU
class CPUOffloadBlock(nn.Module):
    def __init__(self, block):
        super().__init__()
        self.block = block

    def forward(self, x):
        x_cpu = x.cpu()  # Offload to CPU during forward
        torch.cuda.empty_cache()
        x_gpu = x_cpu.to(x.device)
        return self.block(x_gpu)
```

---

## Inference Optimization

### Torch Export + Optimization

```python
model.eval()

# Optimize for inference
with torch.no_grad():
    traced = torch.jit.trace(model, example_input)
    optimized = torch.jit.optimize_for_inference(traced)

    # Freeze parameters
    frozen = torch.jit.freeze(optimized)

# Benchmark
import time
with torch.no_grad():
    # Warmup
    for _ in range(10):
        frozen(example_input)
    torch.cuda.synchronize()

    start = time.time()
    for _ in range(100):
        frozen(example_input)
    torch.cuda.synchronize()
    elapsed = time.time() - start
    print(f"Avg inference: {elapsed / 100 * 1000:.2f} ms")
```

### Batched Inference

```python
@torch.no_grad()
def batch_predict(model, dataset, batch_size=256, device="cuda"):
    model.eval()
    loader = torch.utils.data.DataLoader(dataset, batch_size=batch_size, shuffle=False,
                                          num_workers=4, pin_memory=True)
    all_preds = []
    for inputs, _ in loader:
        inputs = inputs.to(device, non_blocking=True)
        with torch.autocast("cuda"):
            outputs = model(inputs)
        all_preds.append(outputs.cpu())
    return torch.cat(all_preds)
```

---

## Common Architecture Recipes

### Quick Reference Table

| Task | Architecture | Key Modules |
|------|-------------|-------------|
| Image Classification | ViT, ConvNeXt, EfficientNet | Patch Embed, MHA, MLP |
| Object Detection | DETR, YOLOv8 | Backbone + Neck + Head |
| Segmentation | U-Net, SegFormer | Encoder-Decoder, Skip Connections |
| Language Modeling | GPT, LLaMA | Causal Attention, RoPE, RMSNorm |
| Sequence Classification | BERT, RoBERTa | Bidirectional Attention, CLS token |
| Image Generation | Diffusion, GAN | U-Net, Time Embedding, Discriminator |
| Speech Recognition | Whisper, Wav2Vec | Conv Feature Extractor, Transformer |
| Multimodal | CLIP, LLaVA | Dual Encoder, Cross-Attention |

### Layer Selection Guide

| Component | Options | When to Use |
|-----------|---------|-------------|
| Normalization | LayerNorm, RMSNorm, BatchNorm | LN for transformers, BN for CNNs, RMS for LLMs |
| Activation | GELU, SiLU/Swish, ReLU | GELU for transformers, SiLU for modern CNNs |
| Attention | MHA, GQA, MQA | MHA general, GQA/MQA for efficient LLMs |
| Position Encoding | Sinusoidal, Learned, RoPE, ALiBi | RoPE for modern LLMs, learned for ViT |
| Dropout | Standard, DropPath, DropBlock | DropPath for residual nets, DropBlock for CNNs |
