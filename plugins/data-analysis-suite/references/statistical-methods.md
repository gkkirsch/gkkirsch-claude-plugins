# Statistical Methods Reference

> A comprehensive reference manual for working analysts. Every section includes
> theory, interpretation guidelines, and production-ready Python code using
> scipy, numpy, pandas, and statsmodels.

---

## Table of Contents

1. [Descriptive Statistics](#descriptive-statistics)
2. [Probability Distributions](#probability-distributions)
3. [Hypothesis Testing](#hypothesis-testing)
4. [Correlation Analysis](#correlation-analysis)
5. [Regression Analysis](#regression-analysis)
6. [Time Series Analysis](#time-series-analysis)
7. [Sampling Methods](#sampling-methods)
8. [Bayesian Methods](#bayesian-methods)
9. [Effect Size & Power](#effect-size--power)
10. [Practical Guidelines](#practical-guidelines)

---

## Descriptive Statistics

### Measures of Central Tendency

#### Mean (Arithmetic)

The sum of all values divided by the count. Sensitive to outliers.

```python
import numpy as np
import pandas as pd

data = np.array([12, 15, 14, 10, 18, 22, 13, 16, 14, 11])

# numpy
mean_np = np.mean(data)

# pandas
s = pd.Series(data)
mean_pd = s.mean()

# Weighted mean
weights = np.array([1, 1, 2, 2, 1, 1, 3, 1, 2, 1])
weighted_mean = np.average(data, weights=weights)
```

**When to use:** Data is roughly symmetric, no extreme outliers dominate.

**When to avoid:** Highly skewed distributions or data with influential outliers.

#### Median

The middle value when data is sorted. Robust to outliers.

```python
median_val = np.median(data)

# For grouped data or when you need interpolation
s = pd.Series(data)
median_pd = s.median()
```

**When to use:** Skewed distributions, ordinal data, presence of outliers (e.g., income data, housing prices).

#### Mode

The most frequently occurring value. The only measure of central tendency valid for nominal data.

```python
from scipy import stats

mode_result = stats.mode(data, keepdims=True)
mode_val = mode_result.mode[0]
mode_count = mode_result.count[0]

# pandas (handles multimodal)
s = pd.Series(data)
modes = s.mode()  # returns all modes
```

**When to use:** Categorical data, identifying the most common category, bimodal distributions.

#### Trimmed Mean

Mean calculated after removing a percentage of extreme values from both tails.

```python
from scipy import stats

# Remove 10% from each tail
trimmed = stats.trim_mean(data, proportiontocut=0.1)

# Winsorized mean — replaces extreme values instead of removing them
from scipy.stats import mstats
winsorized = mstats.winsorize(data, limits=[0.1, 0.1])
winsorized_mean = np.mean(winsorized)
```

**When to use:** You suspect outliers but don't want to fully switch to the
median. Common in competition scoring (e.g., Olympic judging removes
the highest and lowest scores).

---

### Measures of Dispersion

#### Variance and Standard Deviation

Variance measures the average squared deviation from the mean. Standard
deviation is its square root and shares the unit of the original data.

```python
# Population variance and std dev
var_pop = np.var(data)          # ddof=0 by default
std_pop = np.std(data)

# Sample variance and std dev (use ddof=1 for unbiased estimate)
var_sample = np.var(data, ddof=1)
std_sample = np.std(data, ddof=1)

# pandas uses ddof=1 by default
s = pd.Series(data)
var_pd = s.var()    # ddof=1
std_pd = s.std()    # ddof=1
```

**Interpretation:** A std dev of 3.5 on a dataset with mean 14.5 tells you
roughly 68% of values fall between 11.0 and 18.0 (if approximately normal).

#### Median Absolute Deviation (MAD)

Robust alternative to standard deviation. Less influenced by outliers.

```python
from scipy import stats

mad = stats.median_abs_deviation(data)

# Scale factor for consistency with std dev under normality
mad_scaled = stats.median_abs_deviation(data, scale='normal')
# mad_scaled approximates std dev when data is normal
```

**When to use:** Outlier-contaminated data, robust estimation, anomaly detection thresholds.

#### Interquartile Range (IQR)

Difference between the 75th and 25th percentiles. Defines the box in box plots.

```python
q75, q25 = np.percentile(data, [75, 25])
iqr = q75 - q25

# scipy shortcut
from scipy.stats import iqr as compute_iqr
iqr_val = compute_iqr(data)

# Outlier detection using IQR
lower_fence = q25 - 1.5 * iqr
upper_fence = q75 + 1.5 * iqr
outliers = data[(data < lower_fence) | (data > upper_fence)]
```

#### Range

Simplest measure of spread. Highly sensitive to outliers.

```python
data_range = np.ptp(data)  # peak-to-peak (max - min)
# equivalent to: np.max(data) - np.min(data)
```

#### Coefficient of Variation (CV)

Dimensionless ratio of standard deviation to mean. Enables comparison of
variability across datasets with different scales.

```python
cv = np.std(data, ddof=1) / np.mean(data)
cv_percent = cv * 100  # express as percentage

# scipy
from scipy.stats import variation
cv_scipy = variation(data, ddof=1)
```

**Interpretation:**
- CV < 15%: Low variability
- CV 15-30%: Moderate variability
- CV > 30%: High variability

**When to use:** Comparing variability between measurements on different scales
(e.g., comparing variability of height in cm vs. weight in kg).

---

### Measures of Shape

#### Skewness

Quantifies asymmetry of the distribution.

```python
from scipy import stats

# Fisher's skewness (default, excess skewness)
skew = stats.skew(data)

# pandas
s = pd.Series(data)
skew_pd = s.skew()
```

**Interpretation thresholds:**

| Skewness Value     | Interpretation                                  |
|--------------------|------------------------------------------------|
| -0.5 to +0.5      | Approximately symmetric                        |
| -1.0 to -0.5      | Moderately negatively skewed (left tail longer) |
| +0.5 to +1.0      | Moderately positively skewed (right tail longer)|
| Below -1.0         | Highly negatively skewed                        |
| Above +1.0         | Highly positively skewed                        |

**Practical note:** Income, insurance claims, and website session duration are
typically right-skewed. Test scores and age-at-retirement are often left-skewed.

#### Kurtosis

Quantifies the heaviness of the tails relative to a normal distribution.

```python
# Excess kurtosis (0 = normal distribution)
kurt = stats.kurtosis(data)  # Fisher=True by default

# Pearson kurtosis (3 = normal distribution)
kurt_pearson = stats.kurtosis(data, fisher=False)

# pandas (Fisher/excess by default)
kurt_pd = s.kurtosis()
```

**Interpretation thresholds (excess kurtosis):**

| Excess Kurtosis | Type           | Interpretation                            |
|-----------------|----------------|-------------------------------------------|
| ~ 0             | Mesokurtic     | Tail weight similar to normal             |
| > 0             | Leptokurtic    | Heavier tails, more outliers than normal  |
| < 0             | Platykurtic    | Lighter tails, fewer outliers than normal |
| > 1             | —              | Notably heavy tails; check for outliers   |
| > 3             | —              | Extreme tails; financial data, consider robust methods |

---

### Quantiles and Percentiles

```python
# Percentiles
p25, p50, p75 = np.percentile(data, [25, 50, 75])

# Quantiles (0 to 1 scale)
q1 = np.quantile(data, 0.25)

# Deciles
deciles = np.percentile(data, np.arange(10, 100, 10))

# pandas describe — a quick summary
s = pd.Series(data)
summary = s.describe()
# Includes: count, mean, std, min, 25%, 50%, 75%, max

# Custom percentiles
summary_custom = s.describe(percentiles=[0.05, 0.25, 0.5, 0.75, 0.95])
```

---

### Full Descriptive Statistics Function

```python
import numpy as np
import pandas as pd
from scipy import stats

def descriptive_summary(data, name="Variable"):
    """Generate a comprehensive descriptive statistics report."""
    data = np.asarray(data, dtype=float)
    data = data[~np.isnan(data)]  # remove NaNs
    n = len(data)

    result = {
        "Variable": name,
        "N": n,
        "Mean": np.mean(data),
        "Std Dev": np.std(data, ddof=1),
        "Median": np.median(data),
        "MAD": stats.median_abs_deviation(data, scale='normal'),
        "Min": np.min(data),
        "Max": np.max(data),
        "Range": np.ptp(data),
        "IQR": stats.iqr(data),
        "CV (%)": 100 * np.std(data, ddof=1) / np.mean(data) if np.mean(data) != 0 else np.nan,
        "Skewness": stats.skew(data),
        "Kurtosis (excess)": stats.kurtosis(data),
        "P5": np.percentile(data, 5),
        "P25": np.percentile(data, 25),
        "P75": np.percentile(data, 75),
        "P95": np.percentile(data, 95),
        "SE of Mean": np.std(data, ddof=1) / np.sqrt(n),
    }
    return pd.Series(result)
```

---

## Probability Distributions

### Continuous Distributions

#### Normal (Gaussian) Distribution

**Parameters:** mu (mean, location), sigma (std dev, scale)

**PDF:** f(x) = (1 / (sigma * sqrt(2*pi))) * exp(-(x - mu)^2 / (2 * sigma^2))

**Properties:**
- Symmetric about mu
- 68-95-99.7 rule (within 1, 2, 3 std devs)
- Mean = median = mode
- Sum of independent normals is normal
- Central Limit Theorem guarantees approximate normality of sample means

**When it appears in real data:** Measurement errors, heights, blood pressure,
standardized test scores, manufacturing tolerances.

```python
from scipy import stats
import numpy as np

# Create the distribution
dist = stats.norm(loc=100, scale=15)  # IQ distribution

# PDF, CDF, PPF (inverse CDF)
x = 130
pdf_val = dist.pdf(x)    # density at x
cdf_val = dist.cdf(x)    # P(X <= x)
ppf_val = dist.ppf(0.95) # value at 95th percentile

# Generate random samples
samples = dist.rvs(size=1000, random_state=42)

# Fit to data
mu_hat, sigma_hat = stats.norm.fit(data)

# Test for normality
stat, p_value = stats.shapiro(data)          # Shapiro-Wilk (n < 5000)
stat, p_value = stats.normaltest(data)       # D'Agostino-Pearson
stat, p_value = stats.kstest(data, 'norm', args=(mu_hat, sigma_hat))  # KS test
```

#### Log-Normal Distribution

**Parameters:** mu (mean of log), sigma (std dev of log)

**When it appears:** Income, stock prices, city populations, file sizes,
biological measurements — any quantity that is the product of many
independent positive factors.

```python
# If X ~ LogNormal, then log(X) ~ Normal
dist = stats.lognorm(s=0.5, loc=0, scale=np.exp(2))
# s = sigma (shape), scale = exp(mu)

# Fit to data
shape, loc, scale = stats.lognorm.fit(data, floc=0)
mu_ln = np.log(scale)
sigma_ln = shape

samples = dist.rvs(size=1000, random_state=42)
```

#### Exponential Distribution

**Parameters:** lambda (rate), often parameterized as scale = 1/lambda

**When it appears:** Time between events in a Poisson process — time between
customer arrivals, time between failures, radioactive decay intervals.

**Key property:** Memoryless — P(X > s + t | X > s) = P(X > t)

```python
# Mean time between events = 5 minutes
rate = 1/5
dist = stats.expon(scale=5)  # scale = 1/lambda = mean

# P(wait > 10 minutes)
p_wait = 1 - dist.cdf(10)  # survival function
# equivalently
p_wait = dist.sf(10)

# Fit
loc, scale = stats.expon.fit(data)
```

#### Gamma Distribution

**Parameters:** alpha (shape, k), beta (rate) or theta = 1/beta (scale)

**When it appears:** Waiting time for alpha events in a Poisson process,
insurance claim amounts, rainfall amounts, queuing theory.

```python
dist = stats.gamma(a=2, scale=3)  # a = shape, scale = 1/rate

# Fit to data
shape, loc, scale = stats.gamma.fit(data, floc=0)

# Note: Exponential is Gamma with shape=1
# Chi-squared is Gamma with shape=k/2, scale=2
```

#### Beta Distribution

**Parameters:** alpha, beta (both shape parameters)

**When it appears:** Proportions, probabilities, rates bounded between 0 and 1.
Bayesian modeling of binomial proportions (conjugate prior).

```python
dist = stats.beta(a=2, b=5)  # skewed toward 0

# Common shapes:
# a=1, b=1:   Uniform on [0,1]
# a=0.5, b=0.5: U-shaped (Jeffrey's prior)
# a>1, b>1:   Unimodal
# a<1, b<1:   Bimodal at boundaries

# Fit to data (data must be in [0,1])
a_hat, b_hat, loc, scale = stats.beta.fit(data, floc=0, fscale=1)
```

#### Uniform Distribution

**Parameters:** a (lower), b (upper)

```python
dist = stats.uniform(loc=0, scale=10)  # Uniform on [0, 10]
# loc = a, scale = b - a

# Mean = (a+b)/2, Variance = (b-a)^2 / 12
```

#### Student's t Distribution

**Parameters:** df (degrees of freedom)

**When it appears:** Small-sample inference about the mean when population
variance is unknown. The test statistic in t-tests follows this distribution.

```python
dist = stats.t(df=10)

# Approaches normal as df -> infinity
# df=1 is the Cauchy distribution (very heavy tails)
# df=30 is already close to normal

# Critical value for two-tailed test at alpha=0.05
t_crit = stats.t.ppf(0.975, df=10)  # approximately 2.228
```

#### Chi-Squared Distribution

**Parameters:** df (degrees of freedom)

**When it appears:** Sum of squared standard normals, goodness-of-fit tests,
tests of independence, variance of normal populations.

```python
dist = stats.chi2(df=5)

# Critical value at alpha = 0.05
chi2_crit = stats.chi2.ppf(0.95, df=5)  # one-tailed
```

#### F Distribution

**Parameters:** dfn (numerator df), dfd (denominator df)

**When it appears:** Ratio of two chi-squared variables, ANOVA F-statistic,
comparing two variances.

```python
dist = stats.f(dfn=3, dfd=20)
f_crit = stats.f.ppf(0.95, dfn=3, dfd=20)
```

#### Weibull Distribution

**Parameters:** k (shape), lambda (scale)

**When it appears:** Reliability engineering, failure time analysis, wind speed
modeling, survival analysis.

```python
# scipy uses c for shape parameter
dist = stats.weibull_min(c=1.5, scale=10)

# k < 1: decreasing failure rate (infant mortality)
# k = 1: constant failure rate (same as exponential)
# k > 1: increasing failure rate (aging/wear)

shape, loc, scale = stats.weibull_min.fit(data, floc=0)
```

---

### Discrete Distributions

#### Binomial Distribution

**Parameters:** n (trials), p (probability of success per trial)

**Real-world examples:** Number of defective items in a batch, number of
heads in coin flips, number of patients responding to treatment, click-through
counts from n impressions.

```python
dist = stats.binom(n=20, p=0.3)

# P(X = 5)
pmf_val = dist.pmf(5)

# P(X <= 5)
cdf_val = dist.cdf(5)

# Mean = np, Variance = np(1-p)
print(f"Mean: {dist.mean()}, Variance: {dist.var()}")

# Simulate
samples = dist.rvs(size=1000, random_state=42)
```

#### Poisson Distribution

**Parameters:** lambda (mu) — the average rate

**Real-world examples:** Number of emails per hour, number of accidents per
month, number of mutations per DNA strand, server requests per minute.

**Key property:** If events occur independently at a constant rate, the
count in a fixed interval is Poisson-distributed.

```python
dist = stats.poisson(mu=4.5)

pmf_val = dist.pmf(3)     # P(X = 3)
cdf_val = dist.cdf(6)     # P(X <= 6)

# Mean = Variance = lambda
# If Variance >> Mean, consider Negative Binomial (overdispersion)
```

#### Geometric Distribution

**Parameters:** p (probability of success)

**Real-world examples:** Number of trials until first success — number of
sales calls until a sale, number of login attempts until success.

```python
# scipy counts failures before first success (support starts at 0)
dist = stats.geom(p=0.2)

# P(first success on 5th trial)
pmf_val = dist.pmf(5)

# Mean = 1/p
```

#### Negative Binomial Distribution

**Parameters:** r (number of successes), p (probability of success)

**Real-world examples:** Overdispersed count data (variance > mean), number
of trials to achieve r successes, modeling count data with extra variability.

```python
# scipy parameterization: n = r (number of successes), p = success prob
dist = stats.nbinom(n=5, p=0.4)

# P(X = 10) where X is number of failures before 5th success
pmf_val = dist.pmf(10)

# Alternative parameterization for overdispersed counts:
# Mean = mu, Variance = mu + mu^2/r
```

#### Hypergeometric Distribution

**Parameters:** M (population size), n (number of success states), N (draws)

**Real-world examples:** Drawing cards without replacement, quality control
sampling, lottery probabilities.

```python
# 52-card deck, 13 hearts, draw 5 cards
# P(exactly 3 hearts)
dist = stats.hypergeom(M=52, n=13, N=5)
pmf_val = dist.pmf(3)

# P(at least 1 heart)
p_at_least_1 = 1 - dist.pmf(0)
```

---

### Distribution Fitting

#### Maximum Likelihood Estimation (MLE)

Most scipy distributions support MLE via the `.fit()` method.

```python
from scipy import stats
import numpy as np

data = np.random.lognormal(mean=2, sigma=0.5, size=500)

# Fit several candidate distributions
candidates = {
    'norm': stats.norm,
    'lognorm': stats.lognorm,
    'gamma': stats.gamma,
    'weibull_min': stats.weibull_min,
    'expon': stats.expon,
}

fit_results = {}
for name, dist in candidates.items():
    try:
        params = dist.fit(data)
        # Compute log-likelihood
        ll = np.sum(dist.logpdf(data, *params))
        k = len(params)  # number of parameters
        aic = 2 * k - 2 * ll
        bic = k * np.log(len(data)) - 2 * ll
        fit_results[name] = {
            'params': params,
            'log_likelihood': ll,
            'AIC': aic,
            'BIC': bic,
        }
    except Exception as e:
        fit_results[name] = {'error': str(e)}

# Select best by AIC
best = min(
    ((name, res) for name, res in fit_results.items() if 'AIC' in res),
    key=lambda x: x[1]['AIC']
)
print(f"Best fit: {best[0]} (AIC = {best[1]['AIC']:.2f})")
```

#### Method of Moments

Equate sample moments to theoretical moments and solve.

```python
# Example: fitting Gamma distribution via method of moments
sample_mean = np.mean(data)
sample_var = np.var(data, ddof=1)

# Gamma: E[X] = alpha * beta, Var(X) = alpha * beta^2
beta_hat = sample_var / sample_mean      # scale
alpha_hat = sample_mean / beta_hat       # shape
# equivalently: alpha_hat = sample_mean**2 / sample_var
```

#### Goodness-of-Fit Tests

```python
# Kolmogorov-Smirnov test
# H0: data follows the specified distribution
mu_hat, sigma_hat = stats.norm.fit(data)
ks_stat, ks_p = stats.kstest(data, 'norm', args=(mu_hat, sigma_hat))
# Warning: KS test is conservative when parameters are estimated from data
# Use Lilliefors test (statsmodels) for testing normality with estimated params

from statsmodels.stats.diagnostic import lilliefors
lf_stat, lf_p = lilliefors(data, dist='norm')

# Anderson-Darling test (more sensitive to tail deviations)
result = stats.anderson(data, dist='norm')
print(f"Statistic: {result.statistic:.4f}")
for sl, cv in zip(result.significance_level, result.critical_values):
    print(f"  {sl}%: critical value = {cv:.4f} "
          f"({'reject' if result.statistic > cv else 'fail to reject'})")

# Chi-squared goodness of fit (for any distribution, requires binning)
observed_freq, bin_edges = np.histogram(data, bins=20)
# Expected frequencies under fitted distribution
cdf_vals = stats.norm.cdf(bin_edges, loc=mu_hat, scale=sigma_hat)
expected_freq = np.diff(cdf_vals) * len(data)
chi2_stat, chi2_p = stats.chisquare(observed_freq, f_exp=expected_freq)
```

#### QQ Plots

```python
import matplotlib.pyplot as plt
from scipy import stats

fig, ax = plt.subplots(1, 1, figsize=(6, 6))
stats.probplot(data, dist="norm", plot=ax)
ax.set_title("Normal QQ Plot")
ax.get_lines()[0].set_markerfacecolor('steelblue')
ax.get_lines()[0].set_markersize(4)
plt.tight_layout()
plt.savefig("qq_plot.png", dpi=150)

# Interpreting QQ plots:
# - Points on the diagonal line: good fit
# - S-shaped curve: distribution has heavier tails (leptokurtic)
# - Inverted S: lighter tails (platykurtic)
# - Curve above line at right end: right-skewed
# - Curve below line at left end: left-skewed
```

---

## Hypothesis Testing

### Framework

#### Null and Alternative Hypotheses

- **H0 (null):** No effect, no difference, status quo.
- **H1 (alternative):** There is an effect, a difference exists.
- Tests compute a p-value: the probability of observing data as extreme as
  (or more extreme than) the actual data if H0 were true.

#### Type I and Type II Errors

| Decision          | H0 True          | H0 False            |
|-------------------|------------------|----------------------|
| Reject H0         | Type I (alpha)   | Correct (Power)      |
| Fail to reject H0 | Correct          | Type II (beta)       |

- **alpha (significance level):** P(Type I error), typically 0.05
- **beta:** P(Type II error)
- **Power = 1 - beta:** Probability of detecting a real effect

#### P-Value Interpretation

A p-value is **not** the probability that H0 is true. It is the probability of
seeing data this extreme (or more extreme) assuming H0 is true.

**Guidelines:**
- p < 0.001: Very strong evidence against H0
- p < 0.01: Strong evidence against H0
- p < 0.05: Moderate evidence against H0 (conventional threshold)
- p < 0.10: Weak evidence against H0
- p >= 0.10: Insufficient evidence to reject H0

**Always report exact p-values** (e.g., p = 0.032) rather than just p < 0.05.

#### Effect Sizes

| Measure      | Small  | Medium | Large  | Context                           |
|--------------|--------|--------|--------|-----------------------------------|
| Cohen's d    | 0.2    | 0.5    | 0.8    | Comparing two means               |
| Pearson r    | 0.1    | 0.3    | 0.5    | Linear association                |
| Eta-squared  | 0.01   | 0.06   | 0.14   | ANOVA (proportion of variance)    |
| Odds ratio   | 1.5    | 2.5    | 4.3    | Binary outcomes                   |
| Cohen's w    | 0.1    | 0.3    | 0.5    | Chi-squared tests                 |

```python
def cohens_d(group1, group2):
    """Cohen's d for independent samples."""
    n1, n2 = len(group1), len(group2)
    var1, var2 = np.var(group1, ddof=1), np.var(group2, ddof=1)
    pooled_std = np.sqrt(((n1 - 1) * var1 + (n2 - 1) * var2) / (n1 + n2 - 2))
    return (np.mean(group1) - np.mean(group2)) / pooled_std
```

#### Power Analysis

```python
from statsmodels.stats.power import TTestIndPower, TTestPower
from statsmodels.stats.power import NormalIndPower, GofChisquarePower

analysis = TTestIndPower()

# Find required sample size for independent t-test
n = analysis.solve_power(
    effect_size=0.5,   # Cohen's d (medium)
    alpha=0.05,
    power=0.80,
    ratio=1.0,         # n2/n1 ratio
    alternative='two-sided'
)
print(f"Required sample size per group: {int(np.ceil(n))}")

# Find power given sample size
power = analysis.solve_power(
    effect_size=0.5,
    alpha=0.05,
    nobs1=50,
    ratio=1.0,
    alternative='two-sided'
)
print(f"Power with n=50 per group: {power:.3f}")
```

#### Multiple Testing Corrections

When performing multiple hypothesis tests, the family-wise error rate inflates.

```python
from statsmodels.stats.multitest import multipletests
import numpy as np

# Suppose you have p-values from 10 tests
p_values = np.array([0.001, 0.008, 0.039, 0.041, 0.049,
                     0.062, 0.12, 0.23, 0.44, 0.89])

# Bonferroni correction (most conservative)
reject_bonf, pvals_bonf, _, _ = multipletests(p_values, method='bonferroni')

# Holm-Bonferroni (step-down, uniformly more powerful than Bonferroni)
reject_holm, pvals_holm, _, _ = multipletests(p_values, method='holm')

# Benjamini-Hochberg (controls False Discovery Rate, less conservative)
reject_bh, pvals_bh, _, _ = multipletests(p_values, method='fdr_bh')

# Summary
import pandas as pd
comparison = pd.DataFrame({
    'Original p': p_values,
    'Bonferroni adj': pvals_bonf,
    'Holm adj': pvals_holm,
    'BH adj (FDR)': pvals_bh,
    'Reject (BH)': reject_bh,
})
print(comparison.to_string(index=False))
```

**When to use which:**
- **Bonferroni/Holm:** When you need strict family-wise error control (e.g., clinical trials)
- **Benjamini-Hochberg:** When you can tolerate some false positives among
  discoveries (e.g., genomics, exploratory analysis). Controls FDR at level alpha.

---

### Parametric Tests

#### One-Sample t-Test

**H0:** Population mean equals a specified value mu_0.

**Assumptions:** Data is approximately normal (or n >= 30 by CLT), observations are independent.

```python
from scipy import stats

data = np.array([52, 48, 55, 51, 49, 53, 47, 50, 54, 46])

# Test if mean differs from 50
t_stat, p_value = stats.ttest_1samp(data, popmean=50)
print(f"t = {t_stat:.4f}, p = {p_value:.4f}")

# Confidence interval for the mean
from scipy.stats import t as t_dist
n = len(data)
mean = np.mean(data)
se = stats.sem(data)
ci = t_dist.interval(0.95, df=n-1, loc=mean, scale=se)
print(f"95% CI: ({ci[0]:.2f}, {ci[1]:.2f})")
```

#### Independent Two-Sample t-Test

**H0:** Two population means are equal.

**Assumptions:** Both samples are independent, approximately normal,
equal variances (use Welch's if not).

```python
group_a = np.array([23, 25, 28, 22, 27, 26, 24, 29, 21, 25])
group_b = np.array([30, 33, 29, 35, 31, 34, 28, 32, 36, 30])

# Equal variance assumed
t_stat, p_value = stats.ttest_ind(group_a, group_b, equal_var=True)
print(f"t = {t_stat:.4f}, p = {p_value:.4f}")

# Effect size
d = cohens_d(group_a, group_b)
print(f"Cohen's d = {d:.4f}")

# Check equal variance assumption first
levene_stat, levene_p = stats.levene(group_a, group_b)
print(f"Levene's test: F = {levene_stat:.4f}, p = {levene_p:.4f}")
```

#### Welch's t-Test

Use when variances are unequal. Generally recommended over the equal-variance
t-test as it performs well even when variances are equal.

```python
t_stat, p_value = stats.ttest_ind(group_a, group_b, equal_var=False)
print(f"Welch's t = {t_stat:.4f}, p = {p_value:.4f}")
```

**Recommendation:** Default to Welch's t-test unless you have strong reason to
assume equal variances. It is more robust.

#### Paired t-Test

**H0:** The mean difference between paired observations is zero.

**Assumptions:** Differences are approximately normal, pairs are independent.

```python
before = np.array([200, 210, 190, 220, 205, 215, 195, 225, 198, 208])
after  = np.array([185, 195, 188, 200, 192, 198, 180, 210, 185, 195])

t_stat, p_value = stats.ttest_rel(before, after)
print(f"Paired t = {t_stat:.4f}, p = {p_value:.4f}")

# Effect size for paired data
diffs = before - after
d_paired = np.mean(diffs) / np.std(diffs, ddof=1)
print(f"Cohen's d (paired) = {d_paired:.4f}")
```

#### One-Way ANOVA

**H0:** All group means are equal.

**Assumptions:** Independence, normality within groups, equal variances (homoscedasticity).

```python
group1 = [85, 90, 88, 92, 86]
group2 = [78, 82, 80, 75, 79]
group3 = [92, 95, 91, 98, 94]

# scipy
f_stat, p_value = stats.f_oneway(group1, group2, group3)
print(f"F = {f_stat:.4f}, p = {p_value:.4f}")

# Eta-squared (effect size)
import pandas as pd
all_data = np.concatenate([group1, group2, group3])
groups = np.repeat(['A', 'B', 'C'], [len(group1), len(group2), len(group3)])
grand_mean = np.mean(all_data)
ss_between = sum(len(g) * (np.mean(g) - grand_mean)**2
                 for g in [group1, group2, group3])
ss_total = np.sum((all_data - grand_mean)**2)
eta_sq = ss_between / ss_total
print(f"Eta-squared = {eta_sq:.4f}")

# Post-hoc pairwise comparisons (Tukey HSD)
from statsmodels.stats.multicomp import pairwise_tukeyhsd
tukey = pairwise_tukeyhsd(all_data, groups, alpha=0.05)
print(tukey)
```

#### Two-Way ANOVA

```python
import statsmodels.api as sm
from statsmodels.formula.api import ols

# Example: effect of Drug (A, B) and Gender (M, F) on blood pressure
data = pd.DataFrame({
    'bp': [120, 125, 130, 118, 115, 128, 135, 122, 140, 132, 125, 120],
    'drug': ['A','A','A','A','A','A','B','B','B','B','B','B'],
    'gender': ['M','F','M','F','M','F','M','F','M','F','M','F'],
})

model = ols('bp ~ C(drug) * C(gender)', data=data).fit()
anova_table = sm.stats.anova_lm(model, typ=2)
print(anova_table)

# Type II SS is generally recommended
# Type III SS is used when you want to test each main effect
# after controlling for the other main effect AND the interaction
```

---

### Non-Parametric Tests

#### Mann-Whitney U Test

Non-parametric alternative to the independent two-sample t-test.

**H0:** The distributions of both groups are equal (or: the probability that a
randomly selected value from one group exceeds a randomly selected value from
the other group is 0.5).

**When to use:** Ordinal data, non-normal distributions, small samples where
normality cannot be assumed, presence of outliers.

```python
group_a = [4, 7, 3, 8, 5, 6, 2, 9]
group_b = [9, 12, 10, 8, 11, 14, 7, 13]

u_stat, p_value = stats.mannwhitneyu(group_a, group_b, alternative='two-sided')
print(f"U = {u_stat:.4f}, p = {p_value:.4f}")

# Effect size: rank-biserial correlation
n1, n2 = len(group_a), len(group_b)
r_rb = 1 - (2 * u_stat) / (n1 * n2)
print(f"Rank-biserial r = {r_rb:.4f}")
```

#### Wilcoxon Signed-Rank Test

Non-parametric alternative to the paired t-test.

```python
before = [200, 210, 190, 220, 205, 215, 195, 225, 198, 208]
after  = [185, 195, 188, 200, 192, 198, 180, 210, 185, 195]

w_stat, p_value = stats.wilcoxon(before, after, alternative='two-sided')
print(f"W = {w_stat:.4f}, p = {p_value:.4f}")

# Effect size: r = Z / sqrt(N)
# Z approximation from wilcoxon:
from scipy.stats import norm
n = len(before)
z = norm.ppf(p_value / 2)  # approximate Z
r = abs(z) / np.sqrt(n)
print(f"Effect size r = {r:.4f}")
```

#### Kruskal-Wallis Test

Non-parametric alternative to one-way ANOVA.

```python
g1 = [85, 90, 88, 92, 86]
g2 = [78, 82, 80, 75, 79]
g3 = [92, 95, 91, 98, 94]

h_stat, p_value = stats.kruskal(g1, g2, g3)
print(f"H = {h_stat:.4f}, p = {p_value:.4f}")

# Post-hoc: Dunn's test (requires scikit-posthocs)
# pip install scikit-posthocs
import scikit_posthocs as sp
data_long = pd.DataFrame({
    'value': g1 + g2 + g3,
    'group': ['A']*5 + ['B']*5 + ['C']*5,
})
dunn = sp.posthoc_dunn(data_long, val_col='value', group_col='group', p_adjust='bonferroni')
print(dunn)
```

#### Friedman Test

Non-parametric alternative to repeated measures ANOVA.

```python
# Each row is one subject measured under 3 conditions
condition_1 = [7.1, 6.5, 8.2, 7.8, 6.9]
condition_2 = [8.5, 7.2, 9.1, 8.6, 7.8]
condition_3 = [9.2, 8.1, 9.8, 9.0, 8.5]

chi2_stat, p_value = stats.friedmanchisquare(
    condition_1, condition_2, condition_3
)
print(f"Chi2 = {chi2_stat:.4f}, p = {p_value:.4f}")
```

#### Chi-Squared Test of Independence

Tests whether two categorical variables are independent.

```python
# Contingency table
observed = np.array([
    [50, 30, 20],   # Row 1 (e.g., Male)
    [35, 40, 25],   # Row 2 (e.g., Female)
])

chi2, p_value, dof, expected = stats.chi2_contingency(observed)
print(f"Chi2 = {chi2:.4f}, p = {p_value:.4f}, dof = {dof}")
print(f"Expected frequencies:\n{expected}")

# Effect size: Cramer's V
n = observed.sum()
min_dim = min(observed.shape) - 1
cramers_v = np.sqrt(chi2 / (n * min_dim))
print(f"Cramer's V = {cramers_v:.4f}")
```

**Assumption:** Expected cell frequencies should be >= 5. If not, use Fisher's exact test.

#### Fisher's Exact Test

For 2x2 contingency tables, especially with small expected frequencies.

```python
# 2x2 table
table = np.array([[8, 2], [1, 5]])

odds_ratio, p_value = stats.fisher_exact(table, alternative='two-sided')
print(f"Odds Ratio = {odds_ratio:.4f}, p = {p_value:.4f}")
```

---

### Test Selection Decision Tree

Use this flowchart to select the appropriate statistical test.

```
START: What is your research question?
|
|-- Comparing groups on a measured variable?
|   |
|   |-- How many groups?
|   |   |
|   |   |-- 1 group (vs. known value)
|   |   |   |-- Normal data? --> One-sample t-test
|   |   |   |-- Non-normal?  --> Wilcoxon signed-rank (vs. hypothesized median)
|   |   |
|   |   |-- 2 groups
|   |   |   |-- Paired/matched?
|   |   |   |   |-- Normal differences? --> Paired t-test
|   |   |   |   |-- Non-normal?         --> Wilcoxon signed-rank
|   |   |   |
|   |   |   |-- Independent?
|   |   |       |-- Normal + equal var?     --> Independent t-test
|   |   |       |-- Normal + unequal var?   --> Welch's t-test
|   |   |       |-- Non-normal / ordinal?   --> Mann-Whitney U
|   |   |
|   |   |-- 3+ groups
|   |       |-- Independent?
|   |       |   |-- Normal + equal var? --> One-way ANOVA + Tukey HSD
|   |       |   |-- Non-normal?         --> Kruskal-Wallis + Dunn's test
|   |       |
|   |       |-- Repeated measures?
|   |           |-- Normal?     --> Repeated measures ANOVA
|   |           |-- Non-normal? --> Friedman test
|   |
|-- Testing association between variables?
|   |
|   |-- Both continuous?
|   |   |-- Linear relationship + normal? --> Pearson r
|   |   |-- Monotonic / non-normal?       --> Spearman rho
|   |   |-- Neither?                      --> Kendall tau or distance correlation
|   |
|   |-- Both categorical?
|   |   |-- Expected freq >= 5?  --> Chi-squared test
|   |   |-- Small expected freq? --> Fisher's exact test (2x2)
|   |
|   |-- One continuous, one binary? --> Point-biserial correlation
|   |                                  or independent t-test / Mann-Whitney U
|   |
|   |-- Predicting continuous outcome? --> Regression (see Regression section)
|
|-- Testing distribution fit?
    |-- Normal?      --> Shapiro-Wilk (n<5000) or Anderson-Darling
    |-- Any dist?    --> KS test, Anderson-Darling, chi-squared GOF
```

---

## Correlation Analysis

### Pearson Correlation

Measures linear association between two continuous variables.

**Assumptions:** Both variables are continuous, relationship is linear,
bivariate normality (for inference), no extreme outliers.

```python
from scipy import stats
import numpy as np

x = np.array([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])
y = np.array([2.1, 3.8, 6.2, 7.9, 10.5, 12.1, 14.0, 16.3, 18.1, 20.2])

r, p_value = stats.pearsonr(x, y)
print(f"Pearson r = {r:.4f}, p = {p_value:.6f}")

# Confidence interval for r (Fisher z-transformation)
n = len(x)
z = np.arctanh(r)
se = 1 / np.sqrt(n - 3)
z_lower = z - 1.96 * se
z_upper = z + 1.96 * se
r_lower = np.tanh(z_lower)
r_upper = np.tanh(z_upper)
print(f"95% CI for r: ({r_lower:.4f}, {r_upper:.4f})")
```

**Interpretation:**

| |r|    | Strength      |
|---------|---------------|
| 0 - 0.1   | Negligible    |
| 0.1 - 0.3 | Weak          |
| 0.3 - 0.5 | Moderate      |
| 0.5 - 0.7 | Strong        |
| 0.7 - 1.0 | Very strong   |

### Spearman Rank Correlation

Measures monotonic association. Robust to outliers and works with ordinal data.

```python
rho, p_value = stats.spearmanr(x, y)
print(f"Spearman rho = {rho:.4f}, p = {p_value:.6f}")
```

**When to use:** Non-linear but monotonic relationships, ordinal data,
presence of outliers, non-normal distributions.

### Kendall's Tau

Another rank-based correlation. More robust with small samples and handles
ties better than Spearman.

```python
tau, p_value = stats.kendalltau(x, y)
print(f"Kendall tau = {tau:.4f}, p = {p_value:.6f}")
```

**Note:** Kendall's tau values are typically lower than Spearman's rho for
the same data. This is normal and does not indicate weaker association.

### Point-Biserial Correlation

Pearson correlation between a continuous and a dichotomous (binary) variable.

```python
# continuous variable
scores = np.array([75, 82, 90, 68, 95, 77, 88, 72, 60, 85])
# binary variable (0/1)
passed = np.array([0, 1, 1, 0, 1, 0, 1, 0, 0, 1])

r_pb, p_value = stats.pointbiserialr(passed, scores)
print(f"Point-biserial r = {r_pb:.4f}, p = {p_value:.4f}")
```

### Cramer's V

Measures association between two categorical variables (any table size).

```python
def cramers_v(contingency_table):
    """Compute Cramer's V from a contingency table."""
    chi2 = stats.chi2_contingency(contingency_table)[0]
    n = contingency_table.sum()
    min_dim = min(contingency_table.shape) - 1
    return np.sqrt(chi2 / (n * min_dim))

table = np.array([[50, 30, 20], [35, 40, 25]])
v = cramers_v(table)
print(f"Cramer's V = {v:.4f}")

# Interpretation: same thresholds as Pearson r
```

### Partial Correlation

Correlation between two variables after removing the effect of one or more
control variables.

```python
import pandas as pd
from scipy import stats

df = pd.DataFrame({
    'x': [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
    'y': [2, 4, 5, 4, 5, 7, 8, 9, 10, 12],
    'z': [1, 1, 2, 2, 3, 3, 4, 4, 5, 5],
})

def partial_corr(df, x, y, covariates):
    """Partial correlation between x and y controlling for covariates."""
    from sklearn.linear_model import LinearRegression

    cov_data = df[covariates].values.reshape(-1, len(covariates))

    # Residualize x
    model_x = LinearRegression().fit(cov_data, df[x])
    resid_x = df[x] - model_x.predict(cov_data)

    # Residualize y
    model_y = LinearRegression().fit(cov_data, df[y])
    resid_y = df[y] - model_y.predict(cov_data)

    return stats.pearsonr(resid_x, resid_y)

r_partial, p_val = partial_corr(df, 'x', 'y', ['z'])
print(f"Partial r(x,y | z) = {r_partial:.4f}, p = {p_val:.4f}")
```

### Distance Correlation

Captures non-linear associations where Pearson and Spearman may fail.
A distance correlation of zero implies independence (unlike Pearson).

```python
# requires dcor package: pip install dcor
import dcor

x = np.array([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])
y = x ** 2  # perfect non-linear relationship

pearson_r = stats.pearsonr(x, y)[0]
dist_corr = dcor.distance_correlation(x, y)
print(f"Pearson r = {pearson_r:.4f}")    # ~0.97 (high but not 1)
print(f"Distance corr = {dist_corr:.4f}") # closer to 1.0
```

### Correlation vs. Causation

**A non-zero correlation does NOT imply causation.** Possible explanations for
an observed correlation between X and Y:

1. **X causes Y** (direct causation)
2. **Y causes X** (reverse causation)
3. **Z causes both X and Y** (confounding variable)
4. **Coincidence** (spurious correlation, especially with multiple testing)
5. **Collider bias** (conditioning on a common effect)

To establish causation, you need:
- Randomized controlled experiments
- Instrumental variables
- Natural experiments
- Causal inference frameworks (do-calculus, potential outcomes)

---

## Regression Analysis

### Simple Linear Regression

Model: Y = beta_0 + beta_1 * X + epsilon

```python
import numpy as np
import statsmodels.api as sm
from sklearn.linear_model import LinearRegression

x = np.array([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])
y = np.array([2.3, 4.1, 5.8, 8.2, 9.7, 12.1, 13.8, 16.0, 17.9, 20.5])

# --- statsmodels (detailed output) ---
X_sm = sm.add_constant(x)  # add intercept column
model = sm.OLS(y, X_sm).fit()
print(model.summary())

# Key outputs:
# model.params      -> [intercept, slope]
# model.pvalues     -> p-values for each coefficient
# model.rsquared    -> R-squared
# model.rsquared_adj -> Adjusted R-squared
# model.conf_int()  -> confidence intervals
# model.resid       -> residuals

# --- sklearn (prediction-focused) ---
lr = LinearRegression()
lr.fit(x.reshape(-1, 1), y)
print(f"Intercept: {lr.intercept_:.4f}, Slope: {lr.coef_[0]:.4f}")
y_pred = lr.predict(x.reshape(-1, 1))
```

### Multiple Linear Regression

Model: Y = beta_0 + beta_1*X1 + beta_2*X2 + ... + beta_p*Xp + epsilon

```python
import pandas as pd
import statsmodels.api as sm
from statsmodels.formula.api import ols

# Example dataset
df = pd.DataFrame({
    'price': [300, 350, 400, 320, 380, 420, 450, 500, 310, 360],
    'sqft':  [1500, 1800, 2000, 1600, 1900, 2200, 2400, 2800, 1550, 1700],
    'beds':  [3, 3, 4, 3, 4, 4, 5, 5, 3, 3],
    'age':   [10, 15, 5, 20, 8, 3, 2, 1, 18, 12],
})

# Formula interface (recommended)
model = ols('price ~ sqft + beds + age', data=df).fit()
print(model.summary())
```

#### Assumptions and Diagnostics

```python
import matplotlib.pyplot as plt
from statsmodels.stats.stattools import durbin_watson
from statsmodels.stats.diagnostic import het_breuschpagan
from scipy import stats

# 1. Linearity: residuals vs. fitted values
fig, axes = plt.subplots(2, 2, figsize=(12, 10))

# Residuals vs Fitted
axes[0, 0].scatter(model.fittedvalues, model.resid, alpha=0.7)
axes[0, 0].axhline(y=0, color='red', linestyle='--')
axes[0, 0].set_xlabel('Fitted Values')
axes[0, 0].set_ylabel('Residuals')
axes[0, 0].set_title('Residuals vs Fitted')

# QQ plot of residuals (normality)
sm.qqplot(model.resid, line='45', ax=axes[0, 1])
axes[0, 1].set_title('Normal QQ Plot of Residuals')

# Scale-Location plot (homoscedasticity)
standardized_resid = model.get_influence().resid_studentized_internal
axes[1, 0].scatter(model.fittedvalues, np.sqrt(np.abs(standardized_resid)), alpha=0.7)
axes[1, 0].set_xlabel('Fitted Values')
axes[1, 0].set_ylabel('Sqrt(|Standardized Residuals|)')
axes[1, 0].set_title('Scale-Location Plot')

# Residuals vs Leverage (influence)
sm.graphics.influence_plot(model, ax=axes[1, 1])
axes[1, 1].set_title('Influence Plot')

plt.tight_layout()
plt.savefig("regression_diagnostics.png", dpi=150)

# 2. Independence of residuals (Durbin-Watson)
dw = durbin_watson(model.resid)
print(f"Durbin-Watson: {dw:.4f}")
# Values near 2 indicate no autocorrelation
# < 1 or > 3 indicates serial correlation

# 3. Homoscedasticity (Breusch-Pagan test)
bp_stat, bp_p, _, _ = het_breuschpagan(model.resid, model.model.exog)
print(f"Breusch-Pagan: stat = {bp_stat:.4f}, p = {bp_p:.4f}")
# p < 0.05 suggests heteroscedasticity

# 4. Normality of residuals
shapiro_stat, shapiro_p = stats.shapiro(model.resid)
print(f"Shapiro-Wilk: stat = {shapiro_stat:.4f}, p = {shapiro_p:.4f}")
```

#### Multicollinearity (VIF)

```python
from statsmodels.stats.outliers_influence import variance_inflation_factor

X = df[['sqft', 'beds', 'age']]
X_const = sm.add_constant(X)

vif_data = pd.DataFrame()
vif_data['Variable'] = X_const.columns
vif_data['VIF'] = [
    variance_inflation_factor(X_const.values, i)
    for i in range(X_const.shape[1])
]
print(vif_data)

# VIF interpretation:
# VIF = 1:     No multicollinearity
# VIF 1-5:     Moderate (usually acceptable)
# VIF 5-10:    High (investigate)
# VIF > 10:    Severe (take action: remove or combine variables)
```

#### Cook's Distance and Leverage

```python
influence = model.get_influence()
cooks_d = influence.cooks_distance[0]
leverage = influence.hat_matrix_diag

# Threshold for Cook's distance
threshold_cooks = 4 / len(df)
influential = np.where(cooks_d > threshold_cooks)[0]
print(f"Influential observations (Cook's d > {threshold_cooks:.4f}): {influential}")

# High leverage points
threshold_lev = 2 * (model.df_model + 1) / len(df)
high_leverage = np.where(leverage > threshold_lev)[0]
print(f"High leverage points: {high_leverage}")
```

#### Feature Selection

```python
# Forward selection using AIC
import statsmodels.api as sm

def forward_selection(df, target, candidates, criterion='aic'):
    """Simple forward selection based on AIC or BIC."""
    selected = []
    remaining = list(candidates)
    current_score = np.inf

    while remaining:
        scores = {}
        for var in remaining:
            formula = f"{target} ~ {' + '.join(selected + [var])}"
            model = ols(formula, data=df).fit()
            scores[var] = model.aic if criterion == 'aic' else model.bic

        best_var = min(scores, key=scores.get)
        best_score = scores[best_var]

        if best_score < current_score:
            selected.append(best_var)
            remaining.remove(best_var)
            current_score = best_score
        else:
            break

    return selected

# Usage
features = forward_selection(df, 'price', ['sqft', 'beds', 'age'])
print(f"Selected features: {features}")
```

#### Regularization

```python
from sklearn.linear_model import Ridge, Lasso, ElasticNet
from sklearn.preprocessing import StandardScaler
from sklearn.model_selection import cross_val_score

X = df[['sqft', 'beds', 'age']].values
y = df['price'].values

# Always scale features before regularization
scaler = StandardScaler()
X_scaled = scaler.fit_transform(X)

# Ridge Regression (L2 penalty)
ridge = Ridge(alpha=1.0)
ridge.fit(X_scaled, y)
print(f"Ridge coefficients: {ridge.coef_}")
# Shrinks coefficients toward zero but never exactly zero

# Lasso Regression (L1 penalty)
lasso = Lasso(alpha=0.5)
lasso.fit(X_scaled, y)
print(f"Lasso coefficients: {lasso.coef_}")
# Can set coefficients exactly to zero (automatic feature selection)

# ElasticNet (L1 + L2)
enet = ElasticNet(alpha=0.5, l1_ratio=0.5)
enet.fit(X_scaled, y)
print(f"ElasticNet coefficients: {enet.coef_}")

# Cross-validated selection of regularization strength
from sklearn.linear_model import RidgeCV, LassoCV
ridge_cv = RidgeCV(alphas=np.logspace(-3, 3, 50), cv=5)
ridge_cv.fit(X_scaled, y)
print(f"Best Ridge alpha: {ridge_cv.alpha_:.4f}")

lasso_cv = LassoCV(alphas=np.logspace(-3, 1, 50), cv=5, random_state=42)
lasso_cv.fit(X_scaled, y)
print(f"Best Lasso alpha: {lasso_cv.alpha_:.4f}")
```

---

### Logistic Regression

#### Binary Logistic Regression

Model: log(p / (1-p)) = beta_0 + beta_1*X1 + ... + beta_p*Xp

```python
import statsmodels.api as sm
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import (
    confusion_matrix, classification_report,
    roc_auc_score, roc_curve, precision_recall_curve
)

# Example data
df = pd.DataFrame({
    'age':      [25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 28, 33, 38, 42, 48],
    'income':   [30, 40, 45, 50, 55, 60, 65, 70, 75, 80, 35, 42, 48, 52, 58],
    'purchase': [0,  0,  0,  0,  1,  1,  1,  1,  1,  1,  0,  0,  0,  1,  1],
})

# --- statsmodels (for inference) ---
X = sm.add_constant(df[['age', 'income']])
y = df['purchase']
logit_model = sm.Logit(y, X).fit()
print(logit_model.summary())

# Odds ratios (exponentiate coefficients)
odds_ratios = np.exp(logit_model.params)
ci = np.exp(logit_model.conf_int())
ci.columns = ['OR_lower', 'OR_upper']
odds_table = pd.DataFrame({
    'Odds Ratio': odds_ratios,
    'OR_lower': ci['OR_lower'],
    'OR_upper': ci['OR_upper'],
    'p-value': logit_model.pvalues,
})
print(odds_table)

# Interpretation: OR = 1.05 for age means each additional year of age
# increases the odds of purchase by 5%.

# --- sklearn (for prediction) ---
X_sk = df[['age', 'income']].values
y_sk = df['purchase'].values

clf = LogisticRegression(random_state=42)
clf.fit(X_sk, y_sk)
y_pred = clf.predict(X_sk)
y_prob = clf.predict_proba(X_sk)[:, 1]

# Confusion matrix
cm = confusion_matrix(y_sk, y_pred)
print(f"Confusion Matrix:\n{cm}")
print(classification_report(y_sk, y_pred))

# ROC-AUC
auc = roc_auc_score(y_sk, y_prob)
print(f"ROC-AUC = {auc:.4f}")
```

---

### Other Regression Models

#### Polynomial Regression

```python
from sklearn.preprocessing import PolynomialFeatures

x = np.array([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]).reshape(-1, 1)
y = np.array([1, 4, 10, 15, 25, 35, 50, 62, 80, 100])

# Create polynomial features (degree 2)
poly = PolynomialFeatures(degree=2, include_bias=False)
X_poly = poly.fit_transform(x)

lr = LinearRegression()
lr.fit(X_poly, y)
print(f"Coefficients: {lr.coef_}, Intercept: {lr.intercept_:.4f}")

# Use cross-validation to select degree
from sklearn.model_selection import cross_val_score
for degree in range(1, 6):
    poly = PolynomialFeatures(degree=degree, include_bias=False)
    X_p = poly.fit_transform(x)
    scores = cross_val_score(LinearRegression(), X_p, y, cv=5, scoring='r2')
    print(f"Degree {degree}: R2 = {scores.mean():.4f} (+/- {scores.std():.4f})")
```

#### Poisson Regression (Count Data)

```python
import statsmodels.api as sm

# Count data example: number of insurance claims
df = pd.DataFrame({
    'claims':  [0, 1, 0, 2, 3, 1, 0, 4, 2, 1, 5, 0, 1, 2, 3],
    'age':     [25, 30, 35, 40, 50, 28, 22, 55, 45, 33, 60, 27, 32, 38, 48],
    'income':  [30, 40, 50, 45, 55, 35, 28, 60, 50, 42, 65, 32, 38, 46, 52],
})

poisson_model = sm.GLM(
    df['claims'],
    sm.add_constant(df[['age', 'income']]),
    family=sm.families.Poisson()
).fit()
print(poisson_model.summary())

# Incidence rate ratios (exponentiate coefficients)
irr = np.exp(poisson_model.params)
print(f"Incidence Rate Ratios:\n{irr}")
```

#### Quantile Regression

Estimates conditional quantiles (e.g., median regression) instead of
the conditional mean. Robust to outliers and useful when the effect
of predictors varies across the distribution.

```python
import statsmodels.api as sm

# Median regression (quantile = 0.5)
X = sm.add_constant(df[['age', 'income']])
quantile_model = sm.QuantReg(df['claims'], X).fit(q=0.5)
print(quantile_model.summary())

# Compare multiple quantiles
for q in [0.1, 0.25, 0.5, 0.75, 0.9]:
    qr = sm.QuantReg(df['claims'], X).fit(q=q)
    print(f"Quantile {q}: age coeff = {qr.params['age']:.4f}")
```

---

## Time Series Analysis

### Components of a Time Series

- **Trend:** Long-term increase or decrease
- **Seasonality:** Regular periodic pattern (e.g., monthly, quarterly)
- **Cyclical:** Irregular fluctuations (business cycles, not fixed period)
- **Residual (Irregular):** Random noise after removing other components

### Decomposition

```python
import pandas as pd
import numpy as np
from statsmodels.tsa.seasonal import seasonal_decompose, STL
import matplotlib.pyplot as plt

# Create example time series
np.random.seed(42)
dates = pd.date_range('2020-01-01', periods=120, freq='MS')
trend = np.linspace(100, 200, 120)
seasonal = 15 * np.sin(2 * np.pi * np.arange(120) / 12)
noise = np.random.normal(0, 5, 120)
ts = pd.Series(trend + seasonal + noise, index=dates)

# Classical decomposition
# Additive: Y = Trend + Seasonal + Residual (seasonal amplitude is constant)
decomp_add = seasonal_decompose(ts, model='additive', period=12)
decomp_add.plot()
plt.savefig("decomposition_additive.png", dpi=150)

# Multiplicative: Y = Trend * Seasonal * Residual (seasonal amplitude grows with trend)
decomp_mult = seasonal_decompose(ts, model='multiplicative', period=12)

# STL decomposition (more robust, handles outliers better)
stl = STL(ts, period=12, robust=True)
result = stl.fit()
result.plot()
plt.savefig("stl_decomposition.png", dpi=150)

# Access components
trend_component = result.trend
seasonal_component = result.seasonal
residual_component = result.resid
```

### Stationarity Testing

A time series is stationary if its statistical properties (mean, variance,
autocorrelation) do not change over time. Most time series models require
stationarity.

```python
from statsmodels.tsa.stattools import adfuller, kpss

# Augmented Dickey-Fuller test
# H0: series has a unit root (non-stationary)
# H1: series is stationary
adf_result = adfuller(ts, autolag='AIC')
print(f"ADF Statistic: {adf_result[0]:.4f}")
print(f"p-value: {adf_result[1]:.4f}")
print(f"Lags used: {adf_result[2]}")
for key, value in adf_result[4].items():
    print(f"  Critical Value ({key}): {value:.4f}")
# Reject H0 if p < 0.05 (series is stationary)

# KPSS test
# H0: series is stationary
# H1: series is non-stationary (has unit root)
kpss_result = kpss(ts, regression='ct')  # 'ct' for trend stationarity
print(f"KPSS Statistic: {kpss_result[0]:.4f}")
print(f"p-value: {kpss_result[1]:.4f}")
# Reject H0 if p < 0.05 (series is non-stationary)

# Interpretation:
# ADF rejects + KPSS fails to reject -> stationary
# ADF fails to reject + KPSS rejects -> non-stationary
# Both reject -> trend-stationary (difference once)
# Neither rejects -> inconclusive (get more data)
```

#### Differencing

```python
# First differencing (remove trend)
ts_diff1 = ts.diff().dropna()

# Second differencing (if first differencing is insufficient)
ts_diff2 = ts_diff1.diff().dropna()

# Seasonal differencing (remove seasonality)
ts_seasonal_diff = ts.diff(12).dropna()

# Check stationarity after differencing
adf_after = adfuller(ts_diff1, autolag='AIC')
print(f"ADF after first differencing: p = {adf_after[1]:.4f}")
```

### Autocorrelation

```python
from statsmodels.tsa.stattools import acf, pacf
from statsmodels.graphics.tsaplots import plot_acf, plot_pacf

fig, axes = plt.subplots(1, 2, figsize=(14, 5))

# ACF — helps identify MA order (q)
plot_acf(ts_diff1, lags=30, ax=axes[0])
axes[0].set_title("Autocorrelation Function (ACF)")

# PACF — helps identify AR order (p)
plot_pacf(ts_diff1, lags=30, method='ywm', ax=axes[1])
axes[1].set_title("Partial Autocorrelation Function (PACF)")

plt.tight_layout()
plt.savefig("acf_pacf.png", dpi=150)

# Interpreting ACF/PACF for ARIMA order selection:
# ACF cuts off after lag q, PACF decays -> MA(q)
# ACF decays, PACF cuts off after lag p -> AR(p)
# Both decay -> ARMA(p,q), use AIC to select
```

### ARIMA Models

ARIMA(p, d, q): AutoRegressive Integrated Moving Average

- p: number of autoregressive terms
- d: number of differences needed for stationarity
- q: number of moving average terms

```python
from statsmodels.tsa.arima.model import ARIMA
import warnings
warnings.filterwarnings('ignore')

# Fit ARIMA model
model = ARIMA(ts, order=(1, 1, 1))
result = model.fit()
print(result.summary())

# Automatic order selection using AIC
from itertools import product

best_aic = np.inf
best_order = None

for p, d, q in product(range(4), range(2), range(4)):
    try:
        model = ARIMA(ts, order=(p, d, q))
        fit = model.fit()
        if fit.aic < best_aic:
            best_aic = fit.aic
            best_order = (p, d, q)
    except:
        continue

print(f"Best ARIMA order: {best_order} (AIC = {best_aic:.2f})")

# Alternative: use pmdarima for auto_arima
# pip install pmdarima
import pmdarima as pm
auto_model = pm.auto_arima(
    ts, seasonal=False, stepwise=True,
    suppress_warnings=True, trace=True
)
print(auto_model.summary())

# Forecasting
best_model = ARIMA(ts, order=best_order).fit()
forecast = best_model.get_forecast(steps=12)
forecast_mean = forecast.predicted_mean
forecast_ci = forecast.conf_int()

print(f"Forecast:\n{forecast_mean}")
print(f"95% CI:\n{forecast_ci}")
```

### Seasonal ARIMA (SARIMA)

SARIMA(p, d, q)(P, D, Q, s): adds seasonal autoregressive and moving average terms.

```python
from statsmodels.tsa.statespace.sarimax import SARIMAX

model = SARIMAX(
    ts,
    order=(1, 1, 1),
    seasonal_order=(1, 1, 1, 12),  # P, D, Q, period
    enforce_stationarity=False,
    enforce_invertibility=False
)
result = model.fit(disp=False)
print(result.summary())

# Diagnostics
result.plot_diagnostics(figsize=(12, 8))
plt.savefig("sarima_diagnostics.png", dpi=150)

# Forecast
forecast = result.get_forecast(steps=24)
```

### Exponential Smoothing

```python
from statsmodels.tsa.holtwinters import (
    SimpleExpSmoothing, Holt, ExponentialSmoothing
)

# Simple Exponential Smoothing (level only, no trend/seasonality)
ses = SimpleExpSmoothing(ts).fit(smoothing_level=0.2, optimized=False)
ses_forecast = ses.forecast(12)

# Holt's method (level + trend, no seasonality)
holt = Holt(ts).fit(smoothing_level=0.8, smoothing_trend=0.2, optimized=False)
holt_forecast = holt.forecast(12)

# Holt-Winters (level + trend + seasonality)
hw = ExponentialSmoothing(
    ts,
    trend='add',        # 'add' or 'mul'
    seasonal='add',     # 'add' or 'mul'
    seasonal_periods=12
).fit()
hw_forecast = hw.forecast(12)

print(f"SES AIC: {ses.aic:.2f}")
print(f"Holt AIC: {holt.aic:.2f}")
print(f"Holt-Winters AIC: {hw.aic:.2f}")
```

### Change Point Detection

```python
# Using ruptures library: pip install ruptures
import ruptures as rpt

# Generate signal with change points
np.random.seed(42)
n_samples = 200
signal = np.concatenate([
    np.random.normal(0, 1, 50),
    np.random.normal(5, 1, 50),
    np.random.normal(2, 1, 50),
    np.random.normal(8, 1, 50),
])

# PELT algorithm (Pruned Exact Linear Time)
algo = rpt.Pelt(model="rbf").fit(signal)
change_points = algo.predict(pen=10)
print(f"Detected change points: {change_points}")

# Binary segmentation
algo_binseg = rpt.Binseg(model="l2").fit(signal)
change_points_binseg = algo_binseg.predict(n_bkps=3)
print(f"Binary segmentation: {change_points_binseg}")

# Visualization
rpt.display(signal, change_points, figsize=(12, 4))
plt.savefig("change_points.png", dpi=150)
```

---

## Sampling Methods

### Simple Random Sampling

Every member of the population has an equal probability of selection.

```python
import numpy as np
import pandas as pd

population = pd.DataFrame({
    'id': range(1000),
    'value': np.random.normal(50, 10, 1000),
})

# Simple random sample (without replacement)
sample = population.sample(n=100, random_state=42)

# With replacement (bootstrap-style)
sample_wr = population.sample(n=100, replace=True, random_state=42)
```

### Stratified Sampling

Divide the population into strata and sample proportionally from each.

```python
population['group'] = np.random.choice(['A', 'B', 'C'], size=1000, p=[0.5, 0.3, 0.2])

# Proportional stratified sampling
stratified_sample = population.groupby('group', group_keys=False).apply(
    lambda x: x.sample(frac=0.1, random_state=42)
)
print(stratified_sample['group'].value_counts())

# Fixed-size per stratum
stratified_fixed = population.groupby('group', group_keys=False).apply(
    lambda x: x.sample(n=min(30, len(x)), random_state=42)
)
```

### Cluster Sampling

Divide the population into clusters, randomly select clusters, then sample
all (or some) members within selected clusters.

```python
population['cluster'] = np.random.randint(0, 50, size=1000)

# Select 10 random clusters
selected_clusters = np.random.choice(population['cluster'].unique(), size=10, replace=False)
cluster_sample = population[population['cluster'].isin(selected_clusters)]
print(f"Cluster sample size: {len(cluster_sample)}")
```

### Systematic Sampling

Select every k-th element from the population.

```python
k = 10  # sampling interval
start = np.random.randint(0, k)
systematic_sample = population.iloc[start::k]
print(f"Systematic sample size: {len(systematic_sample)}")
```

### Sample Size Calculation

```python
from scipy import stats
import numpy as np

# For estimating a population mean with desired margin of error
def sample_size_mean(std_dev, margin_of_error, confidence=0.95):
    """Sample size for estimating a mean."""
    z = stats.norm.ppf(1 - (1 - confidence) / 2)
    n = (z * std_dev / margin_of_error) ** 2
    return int(np.ceil(n))

# Example: std_dev=10, margin of error=2, 95% confidence
n = sample_size_mean(std_dev=10, margin_of_error=2)
print(f"Required sample size: {n}")

# For estimating a proportion
def sample_size_proportion(p_hat, margin_of_error, confidence=0.95):
    """Sample size for estimating a proportion."""
    z = stats.norm.ppf(1 - (1 - confidence) / 2)
    n = (z ** 2 * p_hat * (1 - p_hat)) / margin_of_error ** 2
    return int(np.ceil(n))

# Example: expected proportion 0.5, margin of error 3%
n_prop = sample_size_proportion(p_hat=0.5, margin_of_error=0.03)
print(f"Required sample size for proportion: {n_prop}")
```

### Bootstrap Methods

Resampling with replacement to estimate the sampling distribution of a statistic.

```python
from scipy import stats

data = np.array([23, 25, 28, 22, 27, 26, 24, 29, 21, 25, 30, 19, 27, 24, 26])

def bootstrap_ci(data, stat_func=np.mean, n_boot=10000, ci=0.95, seed=42):
    """Bootstrap confidence interval for any statistic."""
    rng = np.random.default_rng(seed)
    boot_stats = np.array([
        stat_func(rng.choice(data, size=len(data), replace=True))
        for _ in range(n_boot)
    ])
    alpha = 1 - ci
    lower = np.percentile(boot_stats, 100 * alpha / 2)
    upper = np.percentile(boot_stats, 100 * (1 - alpha / 2))
    return lower, upper, boot_stats

# Bootstrap CI for the mean
lower, upper, boot_means = bootstrap_ci(data)
print(f"Bootstrap 95% CI for mean: ({lower:.2f}, {upper:.2f})")

# Bootstrap CI for the median
lower_med, upper_med, _ = bootstrap_ci(data, stat_func=np.median)
print(f"Bootstrap 95% CI for median: ({lower_med:.2f}, {upper_med:.2f})")

# BCa (bias-corrected and accelerated) bootstrap — more accurate
from scipy.stats import bootstrap
result = bootstrap(
    (data,),
    statistic=np.mean,
    n_resamples=10000,
    confidence_level=0.95,
    method='BCa',
    random_state=42
)
print(f"BCa 95% CI: ({result.confidence_interval.low:.2f}, "
      f"{result.confidence_interval.high:.2f})")

# Bootstrap hypothesis test
def bootstrap_hypothesis_test(group1, group2, n_boot=10000, seed=42):
    """Two-sample bootstrap permutation test."""
    rng = np.random.default_rng(seed)
    observed_diff = np.mean(group1) - np.mean(group2)
    combined = np.concatenate([group1, group2])
    n1 = len(group1)
    boot_diffs = []
    for _ in range(n_boot):
        perm = rng.permutation(combined)
        boot_diffs.append(np.mean(perm[:n1]) - np.mean(perm[n1:]))
    boot_diffs = np.array(boot_diffs)
    p_value = np.mean(np.abs(boot_diffs) >= np.abs(observed_diff))
    return observed_diff, p_value

group_a = np.array([23, 25, 28, 22, 27])
group_b = np.array([30, 33, 29, 35, 31])
diff, p = bootstrap_hypothesis_test(group_a, group_b)
print(f"Observed difference: {diff:.2f}, p-value: {p:.4f}")
```

### Jackknife

Leave-one-out resampling method. Useful for bias estimation and variance
estimation.

```python
def jackknife_estimate(data, stat_func=np.mean):
    """Jackknife estimate of bias and standard error."""
    n = len(data)
    stat_full = stat_func(data)

    # Leave-one-out statistics
    jack_stats = np.array([
        stat_func(np.delete(data, i)) for i in range(n)
    ])

    jack_mean = np.mean(jack_stats)
    bias = (n - 1) * (jack_mean - stat_full)
    variance = (n - 1) / n * np.sum((jack_stats - jack_mean) ** 2)
    se = np.sqrt(variance)

    return {
        'estimate': stat_full,
        'bias': bias,
        'bias_corrected': stat_full - bias,
        'se': se,
    }

result = jackknife_estimate(data)
print(f"Estimate: {result['estimate']:.4f}")
print(f"Bias: {result['bias']:.4f}")
print(f"Bias-corrected: {result['bias_corrected']:.4f}")
print(f"SE: {result['se']:.4f}")
```

---

## Bayesian Methods

### Bayes' Theorem

P(H | D) = P(D | H) * P(H) / P(D)

- **Prior P(H):** Belief about H before seeing data
- **Likelihood P(D | H):** Probability of the data given H
- **Posterior P(H | D):** Updated belief after seeing data
- **Evidence P(D):** Normalizing constant

### Conjugate Priors

When the posterior has the same distributional form as the prior, the prior
is called a conjugate prior. This makes Bayesian updating tractable.

| Likelihood   | Conjugate Prior | Posterior         | Use case                     |
|--------------|-----------------|-------------------|------------------------------|
| Binomial     | Beta            | Beta              | Proportions, conversion rates|
| Poisson      | Gamma           | Gamma             | Count rates                  |
| Normal (known var) | Normal    | Normal            | Means                        |
| Normal (known mean)| Inv-Gamma | Inv-Gamma        | Variances                    |
| Exponential  | Gamma           | Gamma             | Wait times                   |

### Bayesian Estimation of a Proportion

```python
from scipy import stats
import numpy as np
import matplotlib.pyplot as plt

# Example: A/B test. Observed 45 conversions out of 200 visitors.
successes = 45
trials = 200

# Prior: Beta(1, 1) = Uniform (uninformative)
alpha_prior = 1
beta_prior = 1

# Posterior: Beta(alpha_prior + successes, beta_prior + failures)
alpha_post = alpha_prior + successes
beta_post = beta_prior + (trials - successes)

posterior = stats.beta(alpha_post, beta_post)

# Posterior statistics
post_mean = posterior.mean()
post_median = posterior.median()
post_mode = (alpha_post - 1) / (alpha_post + beta_post - 2)

# 95% Credible Interval (Bayesian analog of confidence interval)
ci_lower, ci_upper = posterior.ppf(0.025), posterior.ppf(0.975)
print(f"Posterior mean: {post_mean:.4f}")
print(f"Posterior median: {post_median:.4f}")
print(f"Posterior mode: {post_mode:.4f}")
print(f"95% Credible Interval: ({ci_lower:.4f}, {ci_upper:.4f})")

# Probability that the true rate exceeds 0.25
p_exceeds = 1 - posterior.cdf(0.25)
print(f"P(rate > 0.25) = {p_exceeds:.4f}")
```

### Credible Intervals vs. Confidence Intervals

| Feature               | Credible Interval (Bayesian)          | Confidence Interval (Frequentist)     |
|-----------------------|---------------------------------------|---------------------------------------|
| Interpretation        | P(parameter in interval \| data) = 95% | 95% of such intervals contain the parameter |
| Requires prior?       | Yes                                   | No                                    |
| Fixed                 | Parameter is random, data is fixed    | Parameter is fixed, interval is random|
| Direct probability    | Yes, direct probability statement     | No, it's about the procedure          |

### Simple Bayesian A/B Testing

```python
def bayesian_ab_test(successes_a, trials_a, successes_b, trials_b,
                     n_simulations=100000, seed=42):
    """
    Bayesian A/B test using Beta-Binomial model.
    Returns the probability that B is better than A.
    """
    rng = np.random.default_rng(seed)

    # Posterior distributions (uninformative prior: Beta(1,1))
    samples_a = rng.beta(successes_a + 1, trials_a - successes_a + 1, n_simulations)
    samples_b = rng.beta(successes_b + 1, trials_b - successes_b + 1, n_simulations)

    # Probability that B is better than A
    prob_b_better = np.mean(samples_b > samples_a)

    # Expected lift
    lift_samples = (samples_b - samples_a) / samples_a
    expected_lift = np.mean(lift_samples)

    # Credible interval for the difference
    diff_samples = samples_b - samples_a
    ci_lower = np.percentile(diff_samples, 2.5)
    ci_upper = np.percentile(diff_samples, 97.5)

    return {
        'prob_b_better': prob_b_better,
        'expected_lift': expected_lift,
        'diff_ci': (ci_lower, ci_upper),
        'mean_a': np.mean(samples_a),
        'mean_b': np.mean(samples_b),
    }

# Example: control had 120/1000 conversions, variant had 145/1000
result = bayesian_ab_test(120, 1000, 145, 1000)
print(f"P(B > A) = {result['prob_b_better']:.4f}")
print(f"Expected lift = {result['expected_lift']:.2%}")
print(f"95% CI for difference: ({result['diff_ci'][0]:.4f}, {result['diff_ci'][1]:.4f})")
```

### Bayesian Estimation of a Normal Mean

```python
from scipy import stats
import numpy as np

# Data
data = np.array([52, 48, 55, 51, 49, 53, 47, 50, 54, 46])
n = len(data)
x_bar = np.mean(data)
s = np.std(data, ddof=1)

# Uninformative prior: mu ~ Normal(mu_0, sigma_0^2) with large sigma_0
mu_0 = 0        # prior mean (centered at 0 for uninformative)
sigma_0 = 1000  # prior std (very large = uninformative)
sigma_known = s  # using sample std as plug-in

# Posterior (Normal-Normal conjugate with known variance)
precision_prior = 1 / sigma_0**2
precision_data = n / sigma_known**2
precision_post = precision_prior + precision_data

mu_post = (precision_prior * mu_0 + precision_data * x_bar) / precision_post
sigma_post = np.sqrt(1 / precision_post)

posterior = stats.norm(mu_post, sigma_post)
print(f"Posterior mean: {mu_post:.4f}")
print(f"Posterior std: {sigma_post:.4f}")
print(f"95% Credible Interval: ({posterior.ppf(0.025):.4f}, {posterior.ppf(0.975):.4f})")

# With uninformative prior, posterior ≈ frequentist result
freq_ci = stats.t.interval(0.95, df=n-1, loc=x_bar, scale=s/np.sqrt(n))
print(f"Frequentist 95% CI: ({freq_ci[0]:.4f}, {freq_ci[1]:.4f})")
```

---

## Effect Size & Power

### Cohen's d

Standardized difference between two means.

```python
import numpy as np

def cohens_d(group1, group2):
    """Cohen's d for independent samples with pooled std."""
    n1, n2 = len(group1), len(group2)
    var1 = np.var(group1, ddof=1)
    var2 = np.var(group2, ddof=1)
    pooled_std = np.sqrt(((n1 - 1) * var1 + (n2 - 1) * var2) / (n1 + n2 - 2))
    return (np.mean(group1) - np.mean(group2)) / pooled_std

def cohens_d_paired(before, after):
    """Cohen's d for paired samples."""
    diffs = np.array(before) - np.array(after)
    return np.mean(diffs) / np.std(diffs, ddof=1)

# Hedges' g — bias-corrected version of Cohen's d
def hedges_g(group1, group2):
    """Hedges' g: bias-corrected Cohen's d. Preferred for small samples."""
    d = cohens_d(group1, group2)
    n1, n2 = len(group1), len(group2)
    correction = 1 - 3 / (4 * (n1 + n2) - 9)
    return d * correction
```

**Conventions (Cohen, 1988):**

| d     | Interpretation | Practical meaning                                  |
|-------|----------------|----------------------------------------------------|
| 0.2   | Small          | Hard to see with the naked eye                     |
| 0.5   | Medium         | Noticeable to a careful observer                   |
| 0.8   | Large          | Obvious to casual observation                      |
| 1.2   | Very large     | Clearly different groups                           |
| 2.0   | Huge           | Virtually no overlap between groups                |

### Eta-Squared and Partial Eta-Squared

```python
def eta_squared(ss_between, ss_total):
    """Proportion of total variance explained by the factor."""
    return ss_between / ss_total

def partial_eta_squared(ss_effect, ss_error):
    """Proportion of variance explained after removing other effects."""
    return ss_effect / (ss_effect + ss_error)

# From ANOVA table (statsmodels)
# anova_table = sm.stats.anova_lm(model, typ=2)
# eta_sq = anova_table['sum_sq']['factor'] / anova_table['sum_sq'].sum()
```

**Conventions:**

| Eta-squared | Interpretation |
|-------------|----------------|
| 0.01        | Small          |
| 0.06        | Medium         |
| 0.14        | Large          |

### Odds Ratio and Relative Risk

```python
def odds_ratio(a, b, c, d):
    """
    Odds ratio from a 2x2 table:
         Outcome+  Outcome-
    Exp+    a         b
    Exp-    c         d
    """
    return (a * d) / (b * c)

def relative_risk(a, b, c, d):
    """Relative risk from a 2x2 table."""
    risk_exposed = a / (a + b)
    risk_unexposed = c / (c + d)
    return risk_exposed / risk_unexposed

# Example: treatment vs control
# Treatment: 30 improved, 70 did not
# Control: 15 improved, 85 did not
or_val = odds_ratio(30, 70, 15, 85)
rr_val = relative_risk(30, 70, 15, 85)
print(f"Odds Ratio: {or_val:.2f}")
print(f"Relative Risk: {rr_val:.2f}")

# Confidence interval for odds ratio
import numpy as np
log_or = np.log(or_val)
se_log_or = np.sqrt(1/30 + 1/70 + 1/15 + 1/85)
ci_lower = np.exp(log_or - 1.96 * se_log_or)
ci_upper = np.exp(log_or + 1.96 * se_log_or)
print(f"95% CI for OR: ({ci_lower:.2f}, {ci_upper:.2f})")
```

### Power Analysis for Common Tests

```python
from statsmodels.stats.power import (
    TTestIndPower, TTestPower, NormalIndPower,
    GofChisquarePower, FTestAnovaPower
)
import numpy as np

# --- Independent t-test ---
tt_ind = TTestIndPower()

# Required sample size
n_required = tt_ind.solve_power(effect_size=0.5, alpha=0.05, power=0.8,
                                 ratio=1.0, alternative='two-sided')
print(f"Independent t-test: n per group = {int(np.ceil(n_required))}")

# Power with given sample size
power = tt_ind.solve_power(effect_size=0.5, alpha=0.05, nobs1=50,
                           ratio=1.0, alternative='two-sided')
print(f"Power with n=50: {power:.3f}")

# --- Paired t-test ---
tt_paired = TTestPower()
n_paired = tt_paired.solve_power(effect_size=0.5, alpha=0.05, power=0.8,
                                  alternative='two-sided')
print(f"Paired t-test: n = {int(np.ceil(n_paired))}")

# --- One-way ANOVA ---
f_anova = FTestAnovaPower()
n_anova = f_anova.solve_power(
    effect_size=0.25,   # Cohen's f (0.1=small, 0.25=medium, 0.4=large)
    alpha=0.05,
    power=0.8,
    k_groups=3
)
print(f"One-way ANOVA (3 groups): n per group = {int(np.ceil(n_anova))}")

# --- Chi-squared test ---
chi2_power = GofChisquarePower()
n_chi2 = chi2_power.solve_power(
    effect_size=0.3,   # Cohen's w (0.1=small, 0.3=medium, 0.5=large)
    alpha=0.05,
    power=0.8,
    n_bins=4           # degrees of freedom + 1
)
print(f"Chi-squared test: n = {int(np.ceil(n_chi2))}")

# --- Correlation ---
from statsmodels.stats.power import NormalIndPower

def power_correlation(r, alpha=0.05, power=0.8):
    """Sample size for testing correlation significance."""
    from scipy.stats import norm
    z_alpha = norm.ppf(1 - alpha / 2)
    z_beta = norm.ppf(power)
    # Fisher z transform
    z_r = np.arctanh(r)
    n = ((z_alpha + z_beta) / z_r) ** 2 + 3
    return int(np.ceil(n))

n_corr = power_correlation(r=0.3)
print(f"Correlation (r=0.3): n = {n_corr}")
```

### Sample Size Planning: Complete Workflow

```python
import numpy as np
from statsmodels.stats.power import TTestIndPower
import matplotlib.pyplot as plt

analysis = TTestIndPower()

# 1. Power curve: how power changes with sample size
sample_sizes = np.arange(10, 200)
effect_sizes = [0.2, 0.5, 0.8]

fig, ax = plt.subplots(figsize=(10, 6))
for d in effect_sizes:
    powers = [analysis.solve_power(effect_size=d, nobs1=n, alpha=0.05,
              ratio=1.0, alternative='two-sided') for n in sample_sizes]
    ax.plot(sample_sizes, powers, label=f"d = {d}")

ax.axhline(y=0.8, color='red', linestyle='--', label='Power = 0.8')
ax.set_xlabel('Sample Size per Group')
ax.set_ylabel('Power')
ax.set_title('Power Curves for Independent t-test')
ax.legend()
ax.grid(True, alpha=0.3)
plt.savefig("power_curves.png", dpi=150)

# 2. Sample size table
print("\nSample Size Requirements (alpha=0.05, power=0.80, two-sided):")
print(f"{'Effect Size':>12} {'Cohen d':>10} {'n per group':>12}")
print("-" * 36)
for label, d in [("Small", 0.2), ("Medium", 0.5), ("Large", 0.8)]:
    n = analysis.solve_power(effect_size=d, alpha=0.05, power=0.80,
                             ratio=1.0, alternative='two-sided')
    print(f"{label:>12} {d:>10.1f} {int(np.ceil(n)):>12}")
```

---

## Practical Guidelines

### When to Use Parametric vs. Non-Parametric

| Condition                       | Use              | Reasoning                                   |
|---------------------------------|------------------|---------------------------------------------|
| Large n (>30), roughly normal   | Parametric       | More powerful, CLT applies                  |
| Small n, clearly normal         | Parametric       | Assumptions met                             |
| Small n, non-normal             | Non-parametric   | Assumptions violated                        |
| Ordinal data                    | Non-parametric   | Parametric assumes interval/ratio scale     |
| Severe outliers                 | Non-parametric   | Robust to extreme values                    |
| Highly skewed, small n          | Non-parametric   | Mean not representative                     |
| Large n, moderate skew          | Parametric       | CLT ensures approximate normality of mean   |

**Rule of thumb:** When in doubt, run both. If they agree, report the
parametric test (more powerful). If they disagree, investigate why and consider
the non-parametric result more trustworthy.

### Common Mistakes in Statistical Analysis

1. **P-hacking:** Running many tests and reporting only significant ones.
   Fix: pre-register hypotheses, apply multiple testing corrections.

2. **Confusing statistical and practical significance:** A p-value of 0.001
   with Cohen's d = 0.05 means the effect is real but trivially small.
   Always report effect sizes.

3. **Assuming normality without checking:** Use Shapiro-Wilk test and QQ
   plots before choosing parametric tests for small samples.

4. **Ignoring assumptions:** Performing t-tests on ordinal Likert data,
   running ANOVA without checking equal variances, using Pearson r on
   non-linear data.

5. **Correlation implies causation:** It does not. Use causal inference
   methods or randomized experiments.

6. **Misinterpreting confidence intervals:** A 95% CI does not mean there is
   a 95% probability the parameter is in the interval. It means 95% of
   such intervals (across repeated sampling) would contain the parameter.

7. **Dropping outliers without justification:** Only remove outliers if you
   have a data-quality reason (measurement error, data entry mistake).
   Consider robust methods instead.

8. **Failing to account for multiple testing:** If you test 20 hypotheses at
   alpha=0.05, you expect 1 false positive by chance. Apply corrections.

9. **Dichotomizing continuous variables:** Converting a continuous variable
   to high/low categories loses information and reduces power. Use
   regression instead.

10. **Ignoring effect direction:** Reporting a significant ANOVA but not
    identifying which groups differ. Always follow up with post-hoc tests.

### Reporting Standards (APA Style)

```
# How to report common statistical tests in APA format:

# t-test
# "An independent-samples t-test indicated that scores were significantly
#  higher in the treatment group (M = 85.3, SD = 7.2) than in the control
#  group (M = 78.1, SD = 8.4), t(48) = 3.25, p = .002, d = 0.92."

# ANOVA
# "A one-way ANOVA revealed a significant effect of condition on performance,
#  F(2, 57) = 8.34, p < .001, eta_sq = .23. Post-hoc Tukey HSD tests showed
#  that Group A (M = 90.1) scored significantly higher than Group B (M = 82.4,
#  p = .003) and Group C (M = 79.8, p < .001)."

# Correlation
# "There was a significant positive correlation between study hours and
#  exam score, r(98) = .54, p < .001."

# Chi-squared
# "A chi-squared test of independence showed a significant association
#  between gender and preference, chi2(2, N = 200) = 12.45, p = .002,
#  Cramer's V = .25."

# Regression
# "A multiple regression analysis indicated that the model significantly
#  predicted sales, F(3, 96) = 24.5, p < .001, R_adj^2 = .42. Both
#  advertising spend (beta = .45, p < .001) and store size (beta = .32,
#  p = .004) were significant predictors."
```

**Formatting rules:**
- Italicize statistical symbols: *t*, *F*, *p*, *r*, *M*, *SD*, *N*, *n*
- Report exact p-values (p = .032), not p < .05, unless p < .001
- Two decimal places for most statistics, three for p-values
- Always include degrees of freedom
- Always include effect sizes alongside p-values
- Round to two decimal places; use leading zero for values that can exceed 1
  (M = 0.52) but not for values that cannot (p = .032, r = .45)

### P-Value Controversy and Alternatives

The American Statistical Association (ASA, 2016) stated: "Scientific conclusions
and business or policy decisions should not be based only on whether a p-value
passes a specific threshold."

**Problems with p-values:**
- Conflates effect size with sample size
- Does not tell you the probability that H0 is true
- Encourages binary (significant/not-significant) thinking
- Highly sensitive to sample size: trivial effects become "significant" with
  large n

**Alternatives and complements:**

1. **Effect sizes with confidence intervals:** Report Cohen's d, r, or
   eta-squared with their CIs. These communicate practical significance.

2. **Bayesian methods:** Compute Bayes factors or posterior probabilities.
   A Bayes factor quantifies the evidence ratio for H1 vs H0.

3. **Equivalence testing (TOST):** Instead of testing for difference, test
   whether the effect is practically zero (within a smallest effect size of
   interest, SESOI).

```python
# TOST (Two One-Sided Tests) for equivalence
from scipy import stats

def tost_equivalence(group1, group2, delta, alpha=0.05):
    """
    Test whether the difference between means is within +/- delta.
    delta: smallest effect size of interest (in raw units).
    """
    n1, n2 = len(group1), len(group2)
    diff = np.mean(group1) - np.mean(group2)
    se = np.sqrt(np.var(group1, ddof=1)/n1 + np.var(group2, ddof=1)/n2)
    df = n1 + n2 - 2  # approximate

    # Test 1: H0: diff <= -delta (lower bound)
    t1 = (diff + delta) / se
    p1 = 1 - stats.t.cdf(t1, df)

    # Test 2: H0: diff >= delta (upper bound)
    t2 = (diff - delta) / se
    p2 = stats.t.cdf(t2, df)

    p_tost = max(p1, p2)
    equivalent = p_tost < alpha

    return {
        'difference': diff,
        'p_tost': p_tost,
        'equivalent': equivalent,
        'delta': delta,
    }

# Example: are these groups equivalent within +/- 3 points?
g1 = np.array([50.1, 49.3, 51.2, 48.8, 50.5])
g2 = np.array([49.8, 50.2, 49.5, 50.8, 49.1])
result = tost_equivalence(g1, g2, delta=3)
print(f"Difference: {result['difference']:.2f}")
print(f"TOST p-value: {result['p_tost']:.4f}")
print(f"Equivalent within +/- {result['delta']}? {result['equivalent']}")
```

4. **Estimation approach:** Focus on parameter estimation with uncertainty
   quantification rather than yes/no hypothesis testing.

### Confidence Interval Interpretation

**Correct interpretation:** If we repeated the experiment many times and
computed a 95% CI each time, approximately 95% of those intervals would
contain the true population parameter.

**Incorrect interpretations:**
- "There is a 95% probability that the true value is in this interval."
  (The parameter is fixed; probability is about the procedure, not the parameter.)
- "95% of the data falls in this interval."
  (CIs are about the parameter, not individual data points.)

```python
# Confidence intervals for common statistics

from scipy import stats
import numpy as np

data = np.array([23, 25, 28, 22, 27, 26, 24, 29, 21, 25])
n = len(data)

# CI for the mean (t-based)
mean = np.mean(data)
se = stats.sem(data)
ci_mean = stats.t.interval(0.95, df=n-1, loc=mean, scale=se)
print(f"95% CI for mean: ({ci_mean[0]:.2f}, {ci_mean[1]:.2f})")

# CI for the variance (chi-squared-based)
s2 = np.var(data, ddof=1)
ci_var_lower = (n - 1) * s2 / stats.chi2.ppf(0.975, df=n-1)
ci_var_upper = (n - 1) * s2 / stats.chi2.ppf(0.025, df=n-1)
print(f"95% CI for variance: ({ci_var_lower:.2f}, {ci_var_upper:.2f})")

# CI for a proportion
successes = 45
trials = 200
p_hat = successes / trials

# Wilson score interval (preferred over Wald for small n or extreme p)
from statsmodels.stats.proportion import proportion_confint
ci_wilson = proportion_confint(successes, trials, alpha=0.05, method='wilson')
print(f"95% Wilson CI for proportion: ({ci_wilson[0]:.4f}, {ci_wilson[1]:.4f})")

# CI for the difference in means
g1 = np.array([23, 25, 28, 22, 27])
g2 = np.array([30, 33, 29, 35, 31])
from scipy.stats import ttest_ind
result = ttest_ind(g1, g2)
diff = np.mean(g1) - np.mean(g2)
se_diff = np.sqrt(np.var(g1, ddof=1)/len(g1) + np.var(g2, ddof=1)/len(g2))
df_welch = (np.var(g1, ddof=1)/len(g1) + np.var(g2, ddof=1)/len(g2))**2 / (
    (np.var(g1, ddof=1)/len(g1))**2/(len(g1)-1) +
    (np.var(g2, ddof=1)/len(g2))**2/(len(g2)-1)
)
ci_diff = stats.t.interval(0.95, df=df_welch, loc=diff, scale=se_diff)
print(f"95% CI for difference in means: ({ci_diff[0]:.2f}, {ci_diff[1]:.2f})")
```

---

## Quick Reference: Function Lookup Table

| Task                              | Function                                        |
|-----------------------------------|-------------------------------------------------|
| Mean                              | `np.mean(data)`                                 |
| Median                            | `np.median(data)`                               |
| Mode                              | `stats.mode(data, keepdims=True)`               |
| Std dev (sample)                  | `np.std(data, ddof=1)`                          |
| MAD                               | `stats.median_abs_deviation(data, scale='normal')` |
| IQR                               | `stats.iqr(data)`                               |
| Skewness                          | `stats.skew(data)`                              |
| Kurtosis                          | `stats.kurtosis(data)`                          |
| Shapiro-Wilk                      | `stats.shapiro(data)`                           |
| KS test                           | `stats.kstest(data, 'norm', args=(mu, sig))`    |
| One-sample t                      | `stats.ttest_1samp(data, popmean)`              |
| Two-sample t                      | `stats.ttest_ind(a, b, equal_var=False)`        |
| Paired t                          | `stats.ttest_rel(before, after)`                |
| Mann-Whitney U                    | `stats.mannwhitneyu(a, b)`                      |
| Wilcoxon signed-rank              | `stats.wilcoxon(before, after)`                 |
| One-way ANOVA                     | `stats.f_oneway(g1, g2, g3)`                    |
| Kruskal-Wallis                    | `stats.kruskal(g1, g2, g3)`                     |
| Chi-squared independence          | `stats.chi2_contingency(table)`                 |
| Fisher's exact                    | `stats.fisher_exact(table_2x2)`                 |
| Pearson correlation               | `stats.pearsonr(x, y)`                          |
| Spearman correlation              | `stats.spearmanr(x, y)`                         |
| Kendall's tau                     | `stats.kendalltau(x, y)`                        |
| Linear regression (stats)         | `sm.OLS(y, sm.add_constant(X)).fit()`           |
| Logistic regression (stats)       | `sm.Logit(y, sm.add_constant(X)).fit()`         |
| ADF stationarity test             | `adfuller(ts)`                                  |
| ARIMA                             | `ARIMA(ts, order=(p,d,q)).fit()`                |
| SARIMA                            | `SARIMAX(ts, order=..., seasonal_order=...).fit()` |
| Bootstrap CI                      | `bootstrap((data,), np.mean, method='BCa')`     |
| Multiple testing correction       | `multipletests(pvals, method='fdr_bh')`         |
| Power analysis (t-test)           | `TTestIndPower().solve_power(...)`              |

---

## Version Information

This reference is written for:
- Python 3.9+
- NumPy 1.21+
- SciPy 1.7+
- pandas 1.3+
- statsmodels 0.13+
- scikit-learn 1.0+

Optional packages referenced:
- pmdarima (auto_arima)
- ruptures (change point detection)
- scikit-posthocs (post-hoc tests)
- dcor (distance correlation)

---

*Statistical Methods Reference -- Data Analysis Suite Plugin*
