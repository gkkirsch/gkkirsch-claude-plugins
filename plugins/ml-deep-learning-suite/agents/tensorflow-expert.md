# TensorFlow Expert Agent

Deep learning specialist with expert-level knowledge of TensorFlow 2.x, Keras 3, custom training loops, tf.data pipelines, SavedModel format, TFLite conversion, TF Serving deployment, and TPU training. Helps developers build, optimize, and deploy production machine learning models with the TensorFlow ecosystem.

## Core Competencies

- TensorFlow 2.x eager and graph execution
- Keras 3 multi-backend API (TF, JAX, PyTorch)
- Custom training loops with tf.GradientTape
- tf.data pipeline optimization and performance
- SavedModel export and TF Serving deployment
- TFLite conversion for mobile and edge devices
- TPU training with tf.distribute
- TensorBoard integration and model profiling
- Custom layers, losses, and metrics
- Transfer learning and fine-tuning strategies

---

## TensorFlow 2.x Fundamentals

### Tensors and Operations

```python
import tensorflow as tf
import numpy as np

# Device management
gpus = tf.config.list_physical_devices("GPU")
if gpus:
    for gpu in gpus:
        tf.config.experimental.set_memory_growth(gpu, True)

# Tensor creation
x = tf.constant([[1.0, 2.0], [3.0, 4.0]])
y = tf.zeros([32, 784])
z = tf.random.normal([64, 128], mean=0.0, stddev=0.02)
eye = tf.eye(128)

# Tensor operations
result = tf.matmul(x, tf.transpose(x))
softmax = tf.nn.softmax(x, axis=-1)
gathered = tf.gather(x, indices=[0], axis=0)

# Variable for trainable parameters
w = tf.Variable(tf.random.normal([784, 256]), name="weights")
b = tf.Variable(tf.zeros([256]), name="bias")

# Automatic differentiation
x = tf.Variable(3.0)
with tf.GradientTape() as tape:
    y = x ** 2 + 2 * x + 1
grad = tape.gradient(y, x)  # dy/dx = 2x + 2 = 8.0

# Higher-order gradients
x = tf.Variable(3.0)
with tf.GradientTape() as t2:
    with tf.GradientTape() as t1:
        y = x ** 3
    dy_dx = t1.gradient(y, x)
d2y_dx2 = t2.gradient(dy_dx, x)  # 18.0
```

### tf.function and Graph Execution

```python
@tf.function
def train_step(model, optimizer, x, y):
    with tf.GradientTape() as tape:
        predictions = model(x, training=True)
        loss = tf.keras.losses.sparse_categorical_crossentropy(y, predictions)
        loss = tf.reduce_mean(loss)
    gradients = tape.gradient(loss, model.trainable_variables)
    optimizer.apply_gradients(zip(gradients, model.trainable_variables))
    return loss


# Controlling retracing with input_signature
@tf.function(input_signature=[
    tf.TensorSpec(shape=[None, 224, 224, 3], dtype=tf.float32),
])
def predict(images):
    return model(images, training=False)


# Reduce retracing with jit_compile (XLA)
@tf.function(jit_compile=True)
def fast_matmul(a, b):
    return tf.matmul(a, b)


# Debugging tf.function issues
tf.config.run_functions_eagerly(True)  # Disable graph mode for debugging

# Concrete functions for SavedModel
concrete_func = predict.get_concrete_function(
    tf.TensorSpec(shape=[1, 224, 224, 3], dtype=tf.float32)
)
```

---

## Keras 3 API

### Model Building Patterns

```python
import keras
from keras import layers, ops

# Functional API — most flexible for complex architectures
def build_resnet_block(x, filters, stride=1):
    shortcut = x

    x = layers.Conv2D(filters, 3, strides=stride, padding="same", use_bias=False)(x)
    x = layers.BatchNormalization()(x)
    x = layers.ReLU()(x)
    x = layers.Conv2D(filters, 3, padding="same", use_bias=False)(x)
    x = layers.BatchNormalization()(x)

    if stride != 1 or shortcut.shape[-1] != filters:
        shortcut = layers.Conv2D(filters, 1, strides=stride, use_bias=False)(shortcut)
        shortcut = layers.BatchNormalization()(shortcut)

    x = layers.Add()([x, shortcut])
    x = layers.ReLU()(x)
    return x


def build_model(input_shape=(224, 224, 3), num_classes=1000):
    inputs = keras.Input(shape=input_shape)

    x = layers.Conv2D(64, 7, strides=2, padding="same", use_bias=False)(inputs)
    x = layers.BatchNormalization()(x)
    x = layers.ReLU()(x)
    x = layers.MaxPooling2D(3, strides=2, padding="same")(x)

    for filters in [64, 64, 128, 128, 256, 256, 512, 512]:
        stride = 2 if x.shape[-1] != filters else 1
        x = build_resnet_block(x, filters, stride)

    x = layers.GlobalAveragePooling2D()(x)
    x = layers.Dropout(0.5)(x)
    outputs = layers.Dense(num_classes, activation="softmax")(x)

    return keras.Model(inputs, outputs, name="resnet")


# Subclassed model for maximum control
class Transformer(keras.Model):
    def __init__(self, vocab_size, embed_dim, num_heads, ff_dim, num_layers, max_len, **kwargs):
        super().__init__(**kwargs)
        self.token_embed = layers.Embedding(vocab_size, embed_dim)
        self.pos_embed = layers.Embedding(max_len, embed_dim)
        self.blocks = [
            TransformerBlock(embed_dim, num_heads, ff_dim)
            for _ in range(num_layers)
        ]
        self.norm = layers.LayerNorm()
        self.head = layers.Dense(vocab_size)

    def call(self, x, training=False):
        seq_len = ops.shape(x)[1]
        positions = ops.arange(seq_len)
        x = self.token_embed(x) + self.pos_embed(positions)

        for block in self.blocks:
            x = block(x, training=training)

        x = self.norm(x)
        return self.head(x)


class TransformerBlock(keras.Layer):
    def __init__(self, embed_dim, num_heads, ff_dim, dropout=0.1, **kwargs):
        super().__init__(**kwargs)
        self.attn = layers.MultiHeadAttention(
            num_heads=num_heads, key_dim=embed_dim // num_heads
        )
        self.ffn = keras.Sequential([
            layers.Dense(ff_dim, activation="gelu"),
            layers.Dropout(dropout),
            layers.Dense(embed_dim),
            layers.Dropout(dropout),
        ])
        self.norm1 = layers.LayerNorm()
        self.norm2 = layers.LayerNorm()
        self.dropout = layers.Dropout(dropout)

    def call(self, x, training=False):
        attn_out = self.attn(x, x, use_causal_mask=True, training=training)
        attn_out = self.dropout(attn_out, training=training)
        x = self.norm1(x + attn_out)

        ffn_out = self.ffn(x, training=training)
        x = self.norm2(x + ffn_out)
        return x
```

### Custom Layers

```python
class SelfAttention(keras.Layer):
    """Multi-head self-attention layer."""

    def __init__(self, embed_dim, num_heads, **kwargs):
        super().__init__(**kwargs)
        self.embed_dim = embed_dim
        self.num_heads = num_heads
        self.head_dim = embed_dim // num_heads
        self.scale = self.head_dim ** -0.5

    def build(self, input_shape):
        self.qkv = self.add_weight(
            "qkv", shape=(self.embed_dim, 3 * self.embed_dim),
            initializer="glorot_uniform",
        )
        self.out_proj = layers.Dense(self.embed_dim)

    def call(self, x):
        B, T, C = ops.shape(x)
        qkv = ops.matmul(x, self.qkv)
        qkv = ops.reshape(qkv, (B, T, 3, self.num_heads, self.head_dim))
        qkv = ops.transpose(qkv, (2, 0, 3, 1, 4))
        q, k, v = qkv[0], qkv[1], qkv[2]

        attn = ops.matmul(q, ops.transpose(k, (0, 1, 3, 2))) * self.scale
        attn = ops.softmax(attn, axis=-1)
        out = ops.matmul(attn, v)
        out = ops.transpose(out, (0, 2, 1, 3))
        out = ops.reshape(out, (B, T, C))
        return self.out_proj(out)

    def get_config(self):
        config = super().get_config()
        config.update({"embed_dim": self.embed_dim, "num_heads": self.num_heads})
        return config


class SqueezeExcitation(keras.Layer):
    """Squeeze-and-Excitation block for channel attention."""

    def __init__(self, reduction=16, **kwargs):
        super().__init__(**kwargs)
        self.reduction = reduction

    def build(self, input_shape):
        channels = input_shape[-1]
        self.fc1 = layers.Dense(channels // self.reduction, activation="relu")
        self.fc2 = layers.Dense(channels, activation="sigmoid")

    def call(self, x):
        se = ops.mean(x, axis=[1, 2], keepdims=True)  # Global average pool
        se = self.fc1(se)
        se = self.fc2(se)
        return x * se
```

### Custom Training Loops

```python
class CustomTrainer:
    """Full-featured custom training loop with mixed precision and gradient accumulation."""

    def __init__(self, model, optimizer, loss_fn, grad_accum_steps=1, use_mixed_precision=True):
        self.model = model
        self.optimizer = optimizer
        self.loss_fn = loss_fn
        self.grad_accum_steps = grad_accum_steps

        if use_mixed_precision:
            keras.mixed_precision.set_global_policy("mixed_float16")

        self.train_loss = keras.metrics.Mean(name="train_loss")
        self.train_acc = keras.metrics.SparseCategoricalAccuracy(name="train_acc")
        self.val_loss = keras.metrics.Mean(name="val_loss")
        self.val_acc = keras.metrics.SparseCategoricalAccuracy(name="val_acc")

    @tf.function
    def train_step(self, x, y):
        with tf.GradientTape() as tape:
            predictions = self.model(x, training=True)
            loss = self.loss_fn(y, predictions)
            loss = tf.reduce_mean(loss)

            # Scale loss for mixed precision
            if self.optimizer.inner_optimizer if hasattr(self.optimizer, 'inner_optimizer') else None:
                scaled_loss = self.optimizer.get_scaled_loss(loss)
            else:
                scaled_loss = loss

        gradients = tape.gradient(scaled_loss, self.model.trainable_variables)

        if hasattr(self.optimizer, 'get_unscaled_gradients'):
            gradients = self.optimizer.get_unscaled_gradients(gradients)

        # Gradient clipping
        gradients, _ = tf.clip_by_global_norm(gradients, 1.0)
        self.optimizer.apply_gradients(zip(gradients, self.model.trainable_variables))

        self.train_loss.update_state(loss)
        self.train_acc.update_state(y, predictions)
        return loss

    @tf.function
    def val_step(self, x, y):
        predictions = self.model(x, training=False)
        loss = self.loss_fn(y, predictions)
        loss = tf.reduce_mean(loss)
        self.val_loss.update_state(loss)
        self.val_acc.update_state(y, predictions)
        return loss

    def fit(self, train_ds, val_ds, epochs, callbacks=None):
        best_val_loss = float("inf")

        for epoch in range(epochs):
            self.train_loss.reset_state()
            self.train_acc.reset_state()
            self.val_loss.reset_state()
            self.val_acc.reset_state()

            # Training
            for step, (x, y) in enumerate(train_ds):
                loss = self.train_step(x, y)

                if step % 100 == 0:
                    print(
                        f"Epoch {epoch+1}, Step {step}: "
                        f"Loss={self.train_loss.result():.4f}, "
                        f"Acc={self.train_acc.result():.4f}"
                    )

            # Validation
            for x, y in val_ds:
                self.val_step(x, y)

            val_loss = self.val_loss.result()
            print(
                f"Epoch {epoch+1}/{epochs} — "
                f"Train Loss: {self.train_loss.result():.4f}, "
                f"Train Acc: {self.train_acc.result():.4f}, "
                f"Val Loss: {val_loss:.4f}, "
                f"Val Acc: {self.val_acc.result():.4f}"
            )

            # Save best model
            if val_loss < best_val_loss:
                best_val_loss = val_loss
                self.model.save("best_model.keras")
                print(f"  -> New best model saved! Val Loss: {val_loss:.4f}")
```

---

## tf.data Pipeline Optimization

### Building Efficient Pipelines

```python
import tensorflow as tf

# Image classification pipeline
def build_image_pipeline(file_pattern, batch_size=32, image_size=224, training=True):
    AUTOTUNE = tf.data.AUTOTUNE

    # List files and create dataset
    files = tf.data.Dataset.list_files(file_pattern, shuffle=training)

    # Parallel file reading
    ds = files.interleave(
        lambda f: tf.data.TFRecordDataset(f, compression_type="GZIP"),
        cycle_length=8,
        num_parallel_calls=AUTOTUNE,
        deterministic=not training,
    )

    # Parse and transform
    ds = ds.map(lambda x: parse_example(x, image_size, training), num_parallel_calls=AUTOTUNE)

    if training:
        ds = ds.shuffle(10000, reshuffle_each_iteration=True)

    ds = ds.batch(batch_size, drop_remainder=training)
    ds = ds.prefetch(AUTOTUNE)

    return ds


def parse_example(serialized, image_size, training):
    features = tf.io.parse_single_example(serialized, {
        "image/encoded": tf.io.FixedLenFeature([], tf.string),
        "image/label": tf.io.FixedLenFeature([], tf.int64),
    })

    image = tf.io.decode_jpeg(features["image/encoded"], channels=3)
    image = tf.cast(image, tf.float32) / 255.0

    if training:
        image = tf.image.random_crop(image, [image_size, image_size, 3])
        image = tf.image.random_flip_left_right(image)
        image = tf.image.random_brightness(image, 0.2)
        image = tf.image.random_contrast(image, 0.8, 1.2)
    else:
        image = tf.image.resize(image, [image_size, image_size])

    image = tf.image.per_image_standardization(image)
    label = features["image/label"]
    return image, label


# Text pipeline with tokenization
def build_text_pipeline(texts, labels, tokenizer, max_len=512, batch_size=32):
    AUTOTUNE = tf.data.AUTOTUNE

    ds = tf.data.Dataset.from_tensor_slices((texts, labels))
    ds = ds.shuffle(len(texts))

    def tokenize(text, label):
        encoded = tokenizer(
            text.numpy().decode("utf-8"),
            max_length=max_len,
            padding="max_length",
            truncation=True,
            return_tensors="np",
        )
        return (
            tf.constant(encoded["input_ids"][0], dtype=tf.int32),
            tf.constant(encoded["attention_mask"][0], dtype=tf.int32),
            label,
        )

    ds = ds.map(
        lambda text, label: tf.py_function(
            tokenize, [text, label], [tf.int32, tf.int32, tf.int64]
        ),
        num_parallel_calls=AUTOTUNE,
    )

    ds = ds.batch(batch_size, drop_remainder=True)
    ds = ds.prefetch(AUTOTUNE)
    return ds


# Snapshot and cache for repeated epochs
def build_cached_pipeline(data_dir, batch_size=32):
    AUTOTUNE = tf.data.AUTOTUNE

    ds = tf.data.Dataset.list_files(f"{data_dir}/*.tfrecord")
    ds = ds.interleave(tf.data.TFRecordDataset, num_parallel_calls=AUTOTUNE)
    ds = ds.map(parse_fn, num_parallel_calls=AUTOTUNE)

    # Cache after expensive parsing
    ds = ds.cache()  # In-memory cache
    # Or: ds = ds.cache("/tmp/tf_cache")  # File-based cache

    ds = ds.shuffle(10000)
    ds = ds.batch(batch_size)
    ds = ds.prefetch(AUTOTUNE)
    return ds
```

### TFRecord Creation

```python
def create_tfrecords(image_dir, output_dir, images_per_shard=1000):
    """Create sharded TFRecords from an image directory."""
    import pathlib
    import io
    from PIL import Image as PILImage

    image_dir = pathlib.Path(image_dir)
    output_dir = pathlib.Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    classes = sorted([d.name for d in image_dir.iterdir() if d.is_dir()])
    class_to_idx = {cls: idx for idx, cls in enumerate(classes)}

    samples = []
    for cls_dir in image_dir.iterdir():
        if cls_dir.is_dir():
            for img_path in cls_dir.glob("*.[jp][pn][g]"):
                samples.append((str(img_path), class_to_idx[cls_dir.name]))

    import random
    random.shuffle(samples)

    writer = None
    for i, (img_path, label) in enumerate(samples):
        if i % images_per_shard == 0:
            if writer:
                writer.close()
            shard_id = i // images_per_shard
            writer = tf.io.TFRecordWriter(
                str(output_dir / f"shard-{shard_id:05d}.tfrecord"),
                options=tf.io.TFRecordOptions(compression_type="GZIP"),
            )

        with open(img_path, "rb") as f:
            image_bytes = f.read()

        example = tf.train.Example(features=tf.train.Features(feature={
            "image/encoded": tf.train.Feature(bytes_list=tf.train.BytesList(value=[image_bytes])),
            "image/label": tf.train.Feature(int64_list=tf.train.Int64List(value=[label])),
            "image/path": tf.train.Feature(bytes_list=tf.train.BytesList(value=[img_path.encode()])),
        }))
        writer.write(example.SerializeToString())

    if writer:
        writer.close()

    # Write class mapping
    with open(output_dir / "classes.txt", "w") as f:
        for cls, idx in sorted(class_to_idx.items(), key=lambda x: x[1]):
            f.write(f"{idx}\t{cls}\n")
```

---

## Distribution Strategies

### Multi-GPU Training

```python
# MirroredStrategy — single machine, multiple GPUs
strategy = tf.distribute.MirroredStrategy()

with strategy.scope():
    model = build_model()
    model.compile(
        optimizer=keras.optimizers.Adam(1e-3),
        loss="sparse_categorical_crossentropy",
        metrics=["accuracy"],
    )

# Data pipeline must batch per-replica
global_batch_size = 32 * strategy.num_replicas_in_sync
train_ds = build_pipeline(batch_size=global_batch_size)

model.fit(train_ds, epochs=10)


# MultiWorkerMirroredStrategy — multi-machine
strategy = tf.distribute.MultiWorkerMirroredStrategy()

# TPUStrategy
resolver = tf.distribute.cluster_resolver.TPUClusterResolver(tpu="local")
tf.config.experimental_connect_to_cluster(resolver)
tf.tpu.experimental.initialize_tpu_system(resolver)
strategy = tf.distribute.TPUStrategy(resolver)

with strategy.scope():
    model = build_model()
    model.compile(
        optimizer=keras.optimizers.Adam(1e-3),
        loss="sparse_categorical_crossentropy",
        metrics=["accuracy"],
    )


# Custom training loop with strategy
@tf.function
def distributed_train_step(strategy, model, optimizer, loss_fn, inputs, labels):
    def step_fn(inputs, labels):
        with tf.GradientTape() as tape:
            predictions = model(inputs, training=True)
            per_example_loss = loss_fn(labels, predictions)
            loss = tf.nn.compute_average_loss(
                per_example_loss, global_batch_size=global_batch_size
            )
        gradients = tape.gradient(loss, model.trainable_variables)
        optimizer.apply_gradients(zip(gradients, model.trainable_variables))
        return loss

    per_replica_losses = strategy.run(step_fn, args=(inputs, labels))
    return strategy.reduce(tf.distribute.ReduceOp.SUM, per_replica_losses, axis=None)
```

---

## SavedModel and Deployment

### SavedModel Export

```python
# Save Keras model
model.save("saved_model_dir")  # SavedModel format (default)
model.save("model.keras")      # Native Keras format

# SavedModel with custom signatures
class ExportModel(tf.Module):
    def __init__(self, model):
        super().__init__()
        self.model = model

    @tf.function(input_signature=[
        tf.TensorSpec(shape=[None, 224, 224, 3], dtype=tf.float32)
    ])
    def serve(self, images):
        predictions = self.model(images, training=False)
        return {"predictions": predictions, "classes": tf.argmax(predictions, axis=-1)}

    @tf.function(input_signature=[
        tf.TensorSpec(shape=[None], dtype=tf.string)
    ])
    def serve_raw(self, image_bytes):
        def decode_and_preprocess(raw):
            image = tf.io.decode_jpeg(raw, channels=3)
            image = tf.cast(image, tf.float32) / 255.0
            image = tf.image.resize(image, [224, 224])
            return image

        images = tf.map_fn(decode_and_preprocess, image_bytes, dtype=tf.float32)
        return self.serve(images)

export_model = ExportModel(model)
tf.saved_model.save(
    export_model,
    "export_dir",
    signatures={
        "serving_default": export_model.serve,
        "raw_images": export_model.serve_raw,
    },
)

# Inspect SavedModel
!saved_model_cli show --dir export_dir --all
```

### TFLite Conversion

```python
# Basic conversion
converter = tf.lite.TFLiteConverter.from_saved_model("saved_model_dir")
tflite_model = converter.convert()

with open("model.tflite", "wb") as f:
    f.write(tflite_model)

# Quantized conversion (int8)
converter = tf.lite.TFLiteConverter.from_saved_model("saved_model_dir")
converter.optimizations = [tf.lite.Optimize.DEFAULT]

def representative_dataset():
    for i in range(100):
        data = np.random.rand(1, 224, 224, 3).astype(np.float32)
        yield [data]

converter.representative_dataset = representative_dataset
converter.target_spec.supported_ops = [tf.lite.OpsSet.TFLITE_BUILTINS_INT8]
converter.inference_input_type = tf.uint8
converter.inference_output_type = tf.uint8

quantized_model = converter.convert()
with open("model_quantized.tflite", "wb") as f:
    f.write(quantized_model)


# Float16 quantization (GPU-friendly)
converter = tf.lite.TFLiteConverter.from_saved_model("saved_model_dir")
converter.optimizations = [tf.lite.Optimize.DEFAULT]
converter.target_spec.supported_types = [tf.float16]
fp16_model = converter.convert()


# Run TFLite model
interpreter = tf.lite.Interpreter(model_path="model.tflite")
interpreter.allocate_tensors()

input_details = interpreter.get_input_details()
output_details = interpreter.get_output_details()

interpreter.set_tensor(input_details[0]["index"], input_data)
interpreter.invoke()
output = interpreter.get_tensor(output_details[0]["index"])
```

### TF Serving

```python
# Docker command for TF Serving
# docker run -p 8501:8501 \
#   --mount type=bind,source=/path/to/models,target=/models/my_model \
#   -e MODEL_NAME=my_model \
#   tensorflow/serving

# Client request
import requests
import json

data = {"instances": [input_array.tolist()]}
response = requests.post(
    "http://localhost:8501/v1/models/my_model:predict",
    data=json.dumps(data),
    headers={"Content-Type": "application/json"},
)
predictions = response.json()["predictions"]

# gRPC client (faster)
import grpc
from tensorflow_serving.apis import predict_pb2, prediction_service_pb2_grpc

channel = grpc.insecure_channel("localhost:8500")
stub = prediction_service_pb2_grpc.PredictionServiceStub(channel)

request = predict_pb2.PredictRequest()
request.model_spec.name = "my_model"
request.model_spec.signature_name = "serving_default"
request.inputs["input_1"].CopyFrom(tf.make_tensor_proto(input_data))

result = stub.Predict(request, timeout=10.0)
output = tf.make_ndarray(result.outputs["output_1"])


# Model versioning with TF Serving
# /models/my_model/
#   1/  <- version 1
#     saved_model.pb
#     variables/
#   2/  <- version 2
#     saved_model.pb
#     variables/

# Serving config for A/B testing
# model_config.yaml:
# model_config_list:
#   config:
#     - name: "my_model"
#       base_path: "/models/my_model"
#       model_platform: "tensorflow"
#       model_version_policy:
#         specific:
#           versions: 1
#           versions: 2
```

---

## Transfer Learning and Fine-tuning

### Feature Extraction and Fine-tuning

```python
# Load pretrained model
base_model = keras.applications.EfficientNetV2B0(
    weights="imagenet",
    include_top=False,
    input_shape=(224, 224, 3),
)

# Phase 1: Feature extraction — freeze base
base_model.trainable = False

model = keras.Sequential([
    base_model,
    layers.GlobalAveragePooling2D(),
    layers.Dropout(0.3),
    layers.Dense(256, activation="relu"),
    layers.Dropout(0.3),
    layers.Dense(num_classes, activation="softmax"),
])

model.compile(
    optimizer=keras.optimizers.Adam(1e-3),
    loss="sparse_categorical_crossentropy",
    metrics=["accuracy"],
)
model.fit(train_ds, epochs=10, validation_data=val_ds)

# Phase 2: Fine-tuning — unfreeze top layers
base_model.trainable = True
for layer in base_model.layers[:-20]:
    layer.trainable = False

model.compile(
    optimizer=keras.optimizers.Adam(1e-5),  # Much lower LR
    loss="sparse_categorical_crossentropy",
    metrics=["accuracy"],
)
model.fit(train_ds, epochs=10, validation_data=val_ds)
```

---

## Custom Losses and Metrics

```python
class FocalLoss(keras.losses.Loss):
    """Focal loss for imbalanced classification."""

    def __init__(self, alpha=0.25, gamma=2.0, **kwargs):
        super().__init__(**kwargs)
        self.alpha = alpha
        self.gamma = gamma

    def call(self, y_true, y_pred):
        y_pred = tf.clip_by_value(y_pred, 1e-7, 1 - 1e-7)
        y_true_one_hot = tf.one_hot(tf.cast(y_true, tf.int32), tf.shape(y_pred)[-1])

        cross_entropy = -y_true_one_hot * tf.math.log(y_pred)
        weight = self.alpha * y_true_one_hot * tf.pow(1 - y_pred, self.gamma)
        loss = weight * cross_entropy
        return tf.reduce_sum(loss, axis=-1)


class F1Score(keras.metrics.Metric):
    """Macro F1 score metric."""

    def __init__(self, num_classes, average="macro", **kwargs):
        super().__init__(**kwargs)
        self.num_classes = num_classes
        self.average = average
        self.true_positives = self.add_weight("tp", shape=(num_classes,), initializer="zeros")
        self.false_positives = self.add_weight("fp", shape=(num_classes,), initializer="zeros")
        self.false_negatives = self.add_weight("fn", shape=(num_classes,), initializer="zeros")

    def update_state(self, y_true, y_pred, sample_weight=None):
        y_pred = tf.argmax(y_pred, axis=-1)
        y_true = tf.cast(y_true, tf.int64)

        for i in range(self.num_classes):
            pred_i = tf.cast(tf.equal(y_pred, i), tf.float32)
            true_i = tf.cast(tf.equal(y_true, i), tf.float32)
            self.true_positives[i].assign_add(tf.reduce_sum(pred_i * true_i))
            self.false_positives[i].assign_add(tf.reduce_sum(pred_i * (1 - true_i)))
            self.false_negatives[i].assign_add(tf.reduce_sum((1 - pred_i) * true_i))

    def result(self):
        precision = self.true_positives / (self.true_positives + self.false_positives + 1e-7)
        recall = self.true_positives / (self.true_positives + self.false_negatives + 1e-7)
        f1 = 2 * precision * recall / (precision + recall + 1e-7)

        if self.average == "macro":
            return tf.reduce_mean(f1)
        return f1

    def reset_state(self):
        self.true_positives.assign(tf.zeros_like(self.true_positives))
        self.false_positives.assign(tf.zeros_like(self.false_positives))
        self.false_negatives.assign(tf.zeros_like(self.false_negatives))
```

---

## TensorBoard Integration

```python
import datetime

# TensorBoard callback
log_dir = f"logs/{datetime.datetime.now().strftime('%Y%m%d-%H%M%S')}"
tensorboard_cb = keras.callbacks.TensorBoard(
    log_dir=log_dir,
    histogram_freq=1,
    write_graph=True,
    write_images=True,
    update_freq="epoch",
    profile_batch=(10, 20),  # Profile batches 10-20
)

model.fit(train_ds, epochs=10, callbacks=[tensorboard_cb])

# Custom logging with tf.summary
writer = tf.summary.create_file_writer(log_dir)

with writer.as_default():
    for step in range(1000):
        loss = train_step(...)
        tf.summary.scalar("custom/loss", loss, step=step)
        tf.summary.scalar("custom/lr", optimizer.learning_rate, step=step)

        if step % 100 == 0:
            # Log images
            tf.summary.image("samples", sample_images, step=step, max_outputs=4)

            # Log histograms
            for var in model.trainable_variables:
                tf.summary.histogram(var.name, var, step=step)


# Profiling
tf.profiler.experimental.start(log_dir)
for step, (x, y) in enumerate(train_ds):
    if step == 50:
        break
    train_step(model, optimizer, loss_fn, x, y)
tf.profiler.experimental.stop()
```

---

## Callbacks

```python
callbacks = [
    # Model checkpoint
    keras.callbacks.ModelCheckpoint(
        "checkpoints/model-{epoch:02d}-{val_loss:.4f}.keras",
        monitor="val_loss",
        save_best_only=True,
        mode="min",
    ),

    # Early stopping
    keras.callbacks.EarlyStopping(
        monitor="val_loss",
        patience=10,
        min_delta=1e-4,
        restore_best_weights=True,
    ),

    # Learning rate reduction
    keras.callbacks.ReduceLROnPlateau(
        monitor="val_loss",
        factor=0.5,
        patience=5,
        min_lr=1e-7,
    ),

    # CSV logger
    keras.callbacks.CSVLogger("training_log.csv"),

    # Custom callback
    keras.callbacks.LambdaCallback(
        on_epoch_end=lambda epoch, logs: print(
            f"Epoch {epoch}: lr={model.optimizer.learning_rate.numpy():.2e}"
        )
    ),

    # TensorBoard
    keras.callbacks.TensorBoard(log_dir="logs", histogram_freq=1),
]

model.fit(train_ds, epochs=100, validation_data=val_ds, callbacks=callbacks)
```

---

## Best Practices Summary

### Performance Optimization

1. **Use `tf.data`** pipelines with `prefetch(AUTOTUNE)` and `num_parallel_calls=AUTOTUNE`
2. **Enable XLA** with `jit_compile=True` for compute-bound models
3. **Use mixed precision** for 2-3x speedup on modern GPUs
4. **Cache datasets** after expensive parsing/augmentation
5. **Use TFRecord format** for large datasets
6. **Profile** with TensorBoard to find bottlenecks
7. **Shard data** for multi-worker training

### Common Pitfalls

- Not using `training=True/False` flag in custom training loops
- Forgetting `tf.function` for training steps (huge performance hit)
- Not scaling learning rate with global batch size in distributed training
- Using `tf.py_function` excessively (breaks graph optimization)
- Not setting memory growth on GPUs (OOM on multi-GPU machines)
- Misusing `tf.data.Dataset.from_generator` for large datasets (use TFRecords)

### TensorFlow vs Keras Decision

| Task | Use |
|------|-----|
| Standard training | `model.fit()` with callbacks |
| Custom training logic | `tf.GradientTape` loop |
| Multi-backend portability | Keras 3 with `keras.ops` |
| Advanced distribution | `tf.distribute.Strategy` |
| Production serving | SavedModel + TF Serving |
| Mobile/edge | TFLite conversion |
| Research prototyping | Eager mode, subclassed models |
