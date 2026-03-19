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

---

## Generative Model Patterns

### VAE (Variational Autoencoder)

```python
class VAE(nn.Module):
    def __init__(self, input_dim, hidden_dim=256, latent_dim=32):
        super().__init__()
        # Encoder
        self.encoder = nn.Sequential(
            nn.Linear(input_dim, hidden_dim),
            nn.ReLU(),
            nn.Linear(hidden_dim, hidden_dim),
            nn.ReLU(),
        )
        self.fc_mu = nn.Linear(hidden_dim, latent_dim)
        self.fc_logvar = nn.Linear(hidden_dim, latent_dim)

        # Decoder
        self.decoder = nn.Sequential(
            nn.Linear(latent_dim, hidden_dim),
            nn.ReLU(),
            nn.Linear(hidden_dim, hidden_dim),
            nn.ReLU(),
            nn.Linear(hidden_dim, input_dim),
            nn.Sigmoid(),
        )

    def encode(self, x):
        h = self.encoder(x)
        return self.fc_mu(h), self.fc_logvar(h)

    def reparameterize(self, mu, logvar):
        std = torch.exp(0.5 * logvar)
        eps = torch.randn_like(std)
        return mu + eps * std

    def forward(self, x):
        mu, logvar = self.encode(x)
        z = self.reparameterize(mu, logvar)
        recon = self.decoder(z)
        return recon, mu, logvar


def vae_loss(recon_x, x, mu, logvar):
    recon_loss = F.binary_cross_entropy(recon_x, x, reduction="sum")
    kl_loss = -0.5 * torch.sum(1 + logvar - mu.pow(2) - logvar.exp())
    return recon_loss + kl_loss
```

### GAN (Generative Adversarial Network)

```python
class Generator(nn.Module):
    def __init__(self, latent_dim=100, img_channels=3, feature_dim=64):
        super().__init__()
        self.net = nn.Sequential(
            self._block(latent_dim, feature_dim * 16, 4, 1, 0),
            self._block(feature_dim * 16, feature_dim * 8, 4, 2, 1),
            self._block(feature_dim * 8, feature_dim * 4, 4, 2, 1),
            self._block(feature_dim * 4, feature_dim * 2, 4, 2, 1),
            nn.ConvTranspose2d(feature_dim * 2, img_channels, 4, 2, 1),
            nn.Tanh(),
        )

    def _block(self, in_ch, out_ch, kernel, stride, padding):
        return nn.Sequential(
            nn.ConvTranspose2d(in_ch, out_ch, kernel, stride, padding, bias=False),
            nn.BatchNorm2d(out_ch),
            nn.ReLU(True),
        )

    def forward(self, z):
        return self.net(z.view(z.size(0), -1, 1, 1))


class Discriminator(nn.Module):
    def __init__(self, img_channels=3, feature_dim=64):
        super().__init__()
        self.net = nn.Sequential(
            nn.Conv2d(img_channels, feature_dim, 4, 2, 1),
            nn.LeakyReLU(0.2, True),
            self._block(feature_dim, feature_dim * 2, 4, 2, 1),
            self._block(feature_dim * 2, feature_dim * 4, 4, 2, 1),
            self._block(feature_dim * 4, feature_dim * 8, 4, 2, 1),
            nn.Conv2d(feature_dim * 8, 1, 4, 1, 0),
        )

    def _block(self, in_ch, out_ch, kernel, stride, padding):
        return nn.Sequential(
            nn.Conv2d(in_ch, out_ch, kernel, stride, padding, bias=False),
            nn.BatchNorm2d(out_ch),
            nn.LeakyReLU(0.2, True),
        )

    def forward(self, x):
        return self.net(x).view(-1)


# WGAN-GP loss
def gradient_penalty(discriminator, real, fake, device):
    alpha = torch.rand(real.size(0), 1, 1, 1, device=device)
    interpolated = (alpha * real + (1 - alpha) * fake).requires_grad_(True)
    mixed_scores = discriminator(interpolated)
    gradient = torch.autograd.grad(
        outputs=mixed_scores,
        inputs=interpolated,
        grad_outputs=torch.ones_like(mixed_scores),
        create_graph=True,
        retain_graph=True,
    )[0]
    gradient = gradient.view(gradient.size(0), -1)
    penalty = ((gradient.norm(2, dim=1) - 1) ** 2).mean()
    return penalty
```

### Diffusion Model (Simplified DDPM)

```python
class SimpleDiffusion(nn.Module):
    """Simplified Denoising Diffusion Probabilistic Model."""

    def __init__(self, model, timesteps=1000, beta_start=1e-4, beta_end=0.02):
        super().__init__()
        self.model = model
        self.timesteps = timesteps

        betas = torch.linspace(beta_start, beta_end, timesteps)
        alphas = 1.0 - betas
        alphas_cumprod = torch.cumprod(alphas, dim=0)

        self.register_buffer("betas", betas)
        self.register_buffer("sqrt_alphas_cumprod", torch.sqrt(alphas_cumprod))
        self.register_buffer("sqrt_one_minus_alphas_cumprod", torch.sqrt(1 - alphas_cumprod))
        self.register_buffer("sqrt_recip_alphas", torch.sqrt(1.0 / alphas))
        self.register_buffer("posterior_variance", betas * (1 - alphas_cumprod.roll(1)) / (1 - alphas_cumprod))

    def forward_diffusion(self, x_0, t, noise=None):
        """Add noise to data at timestep t."""
        if noise is None:
            noise = torch.randn_like(x_0)
        sqrt_alpha = self.sqrt_alphas_cumprod[t].view(-1, 1, 1, 1)
        sqrt_one_minus = self.sqrt_one_minus_alphas_cumprod[t].view(-1, 1, 1, 1)
        return sqrt_alpha * x_0 + sqrt_one_minus * noise, noise

    def loss(self, x_0):
        """Training loss: predict the noise."""
        t = torch.randint(0, self.timesteps, (x_0.size(0),), device=x_0.device)
        x_noisy, noise = self.forward_diffusion(x_0, t)
        noise_pred = self.model(x_noisy, t)
        return F.mse_loss(noise_pred, noise)

    @torch.no_grad()
    def sample(self, shape, device):
        """Generate samples via reverse diffusion."""
        x = torch.randn(shape, device=device)
        for t in reversed(range(self.timesteps)):
            t_batch = torch.full((shape[0],), t, device=device, dtype=torch.long)
            noise_pred = self.model(x, t_batch)
            alpha = 1 - self.betas[t]
            x = self.sqrt_recip_alphas[t] * (
                x - self.betas[t] / self.sqrt_one_minus_alphas_cumprod[t] * noise_pred
            )
            if t > 0:
                noise = torch.randn_like(x)
                x += torch.sqrt(self.posterior_variance[t]) * noise
        return x
```

---

## LoRA (Low-Rank Adaptation) Pattern

```python
class LoRALinear(nn.Module):
    """Low-Rank Adaptation for efficient fine-tuning."""

    def __init__(self, original_linear: nn.Linear, rank: int = 8, alpha: float = 16.0):
        super().__init__()
        self.original = original_linear
        self.original.weight.requires_grad_(False)
        if self.original.bias is not None:
            self.original.bias.requires_grad_(False)

        in_features = original_linear.in_features
        out_features = original_linear.out_features

        self.lora_A = nn.Parameter(torch.randn(in_features, rank) * 0.01)
        self.lora_B = nn.Parameter(torch.zeros(rank, out_features))
        self.scaling = alpha / rank

    def forward(self, x):
        original_out = self.original(x)
        lora_out = (x @ self.lora_A @ self.lora_B) * self.scaling
        return original_out + lora_out


def apply_lora(model, target_modules=["q_proj", "v_proj"], rank=8, alpha=16.0):
    """Apply LoRA to specific modules in a model."""
    for name, module in model.named_modules():
        if any(target in name for target in target_modules):
            if isinstance(module, nn.Linear):
                parent_name = ".".join(name.split(".")[:-1])
                child_name = name.split(".")[-1]
                parent = dict(model.named_modules())[parent_name]
                setattr(parent, child_name, LoRALinear(module, rank=rank, alpha=alpha))

    # Only train LoRA parameters
    for name, param in model.named_parameters():
        if "lora_" not in name:
            param.requires_grad_(False)

    trainable = sum(p.numel() for p in model.parameters() if p.requires_grad)
    total = sum(p.numel() for p in model.parameters())
    print(f"LoRA trainable: {trainable:,} / {total:,} ({trainable/total:.2%})")
```

---

## Distributed Communication Patterns

### All-Reduce and All-Gather

```python
import torch.distributed as dist

# All-reduce: average tensor across all processes
def all_reduce_mean(tensor):
    dist.all_reduce(tensor, op=dist.ReduceOp.SUM)
    tensor /= dist.get_world_size()
    return tensor

# All-gather: collect tensors from all processes
def all_gather_tensors(tensor):
    world_size = dist.get_world_size()
    gathered = [torch.zeros_like(tensor) for _ in range(world_size)]
    dist.all_gather(gathered, tensor)
    return torch.cat(gathered, dim=0)

# Broadcast: send tensor from rank 0 to all
def broadcast_from_rank0(tensor, src=0):
    dist.broadcast(tensor, src=src)
    return tensor
```

### Learning Rate Warmup Implementations

```python
import math

class WarmupCosineScheduler:
    def __init__(self, optimizer, warmup_steps, total_steps, min_lr=1e-6):
        self.optimizer = optimizer
        self.warmup_steps = warmup_steps
        self.total_steps = total_steps
        self.min_lr = min_lr
        self.base_lrs = [pg["lr"] for pg in optimizer.param_groups]
        self.step_count = 0

    def step(self):
        self.step_count += 1
        for pg, base_lr in zip(self.optimizer.param_groups, self.base_lrs):
            if self.step_count < self.warmup_steps:
                lr = base_lr * self.step_count / self.warmup_steps
            else:
                progress = (self.step_count - self.warmup_steps) / (
                    self.total_steps - self.warmup_steps
                )
                lr = self.min_lr + (base_lr - self.min_lr) * 0.5 * (
                    1 + math.cos(math.pi * progress)
                )
            pg["lr"] = lr

    def get_lr(self):
        return [pg["lr"] for pg in self.optimizer.param_groups]
```

---

## Debugging Recipes

### Gradient Flow Check

```python
def check_gradient_flow(named_parameters):
    """Check for vanishing/exploding gradients."""
    for name, param in named_parameters:
        if param.grad is not None:
            grad_norm = param.grad.norm().item()
            if grad_norm == 0:
                print(f"ZERO gradient: {name}")
            elif grad_norm > 100:
                print(f"LARGE gradient: {name} = {grad_norm:.2f}")
            elif math.isnan(grad_norm):
                print(f"NaN gradient: {name}")
```

### Model Summary

```python
def model_summary(model, input_size):
    """Print model summary with parameter counts."""
    total_params = 0
    trainable_params = 0

    print(f"{'Layer':50s} {'Output Shape':20s} {'Params':>10s}")
    print("-" * 82)

    for name, module in model.named_modules():
        if len(list(module.children())) == 0:
            params = sum(p.numel() for p in module.parameters())
            trainable = sum(p.numel() for p in module.parameters() if p.requires_grad)
            total_params += params
            trainable_params += trainable
            if params > 0:
                print(f"{name:50s} {'':20s} {params:>10,d}")

    print("-" * 82)
    print(f"Total params: {total_params:,}")
    print(f"Trainable params: {trainable_params:,}")
    print(f"Non-trainable params: {total_params - trainable_params:,}")
    print(f"Model size: {total_params * 4 / 1024 / 1024:.1f} MB (float32)")
```

### Reproducibility Setup

```python
def set_seed(seed=42):
    """Set all random seeds for reproducibility."""
    import random
    random.seed(seed)
    np.random.seed(seed)
    torch.manual_seed(seed)
    torch.cuda.manual_seed_all(seed)
    torch.backends.cudnn.deterministic = True
    torch.backends.cudnn.benchmark = False

# For DataLoader reproducibility
def seed_worker(worker_id):
    import numpy as np
    worker_seed = torch.initial_seed() % 2**32
    np.random.seed(worker_seed)

g = torch.Generator()
g.manual_seed(42)

dataloader = DataLoader(
    dataset,
    batch_size=32,
    worker_init_fn=seed_worker,
    generator=g,
)
```
