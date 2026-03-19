# /cloud-design

Design cloud architectures, review infrastructure, and optimize cloud deployments.

## Usage

```
/cloud-design [subcommand] [options]
```

## Subcommands

### `architecture` — Design a Cloud Architecture

Design a complete cloud architecture for a new system or feature.

```
/cloud-design architecture
```

**Process:**
1. Gather requirements (functional, non-functional, constraints)
2. Select cloud provider and services
3. Design network topology
4. Define compute, storage, and database architecture
5. Plan security controls (IAM, encryption, network segmentation)
6. Estimate costs
7. Produce architecture documentation

**Options:**
- `--provider aws|gcp|azure|multi` — Target cloud provider (default: aws)
- `--scale small|medium|large` — Expected scale
- `--compliance hipaa|pci|soc2|gdpr` — Compliance requirements

### `review` — Review Existing Architecture

Review an existing cloud architecture against the Well-Architected Framework.

```
/cloud-design review
```

**Process:**
1. Analyze existing Terraform/CloudFormation/infrastructure code
2. Evaluate against 6 pillars (Security, Reliability, Performance, Cost, Operations, Sustainability)
3. Identify risks, anti-patterns, and improvement opportunities
4. Prioritize findings by impact and effort
5. Provide specific remediation recommendations with code

### `terraform` — Generate or Review Terraform

Create new Terraform configurations or review existing ones.

```
/cloud-design terraform
```

**Capabilities:**
- Generate Terraform for a new resource or module
- Review existing Terraform for best practices
- Design module interfaces (variables, outputs)
- Configure remote state backends
- Set up CI/CD pipelines for Terraform
- Import existing resources into Terraform

### `cost` — Analyze and Optimize Costs

Analyze cloud spend and recommend optimizations.

```
/cloud-design cost
```

**Process:**
1. Review current architecture and resource sizing
2. Identify unused or underutilized resources
3. Recommend right-sizing changes
4. Suggest pricing model optimizations (RI, Savings Plans, Spot)
5. Estimate potential monthly savings
6. Generate implementation plan

### `containers` — Container and Kubernetes Design

Design containerized deployments and Kubernetes configurations.

```
/cloud-design containers
```

**Capabilities:**
- Optimize Dockerfiles (multi-stage builds, size reduction)
- Design Kubernetes manifests (deployments, services, ingress)
- Author Helm charts
- Configure auto-scaling (HPA, VPA, KEDA, Karpenter)
- Set up deployment strategies (rolling, blue-green, canary)
- Design GitOps workflows (ArgoCD, Flux)

### `migrate` — Plan Cloud Migration

Plan migration of workloads to the cloud or between clouds.

```
/cloud-design migrate
```

**Process:**
1. Assess current workload (architecture, dependencies, data)
2. Recommend migration strategy (7 Rs)
3. Design target architecture
4. Plan migration phases
5. Identify risks and mitigation strategies
6. Estimate timeline and costs

### `security` — Security Architecture Review

Review and improve cloud security posture.

```
/cloud-design security
```

**Capabilities:**
- IAM policy review (least privilege analysis)
- Network security assessment (security groups, NACLs, WAF)
- Encryption audit (at rest, in transit, key management)
- Compliance gap analysis
- Security group rule optimization
- Generate security-hardened configurations

### `network` — Network Architecture Design

Design or review cloud networking.

```
/cloud-design network
```

**Capabilities:**
- VPC design with CIDR planning
- Multi-account networking (Transit Gateway, VPC Peering)
- Hybrid connectivity (VPN, Direct Connect)
- Service mesh configuration (Istio, App Mesh)
- DNS architecture (Route 53, Cloud DNS)
- CDN and edge network design

---

## Examples

```
# Design a new architecture
/cloud-design architecture --provider aws --scale medium

# Review existing Terraform
/cloud-design review

# Optimize costs
/cloud-design cost

# Generate Kubernetes manifests
/cloud-design containers

# Plan a migration
/cloud-design migrate --provider aws

# Security review
/cloud-design security
```

---

## Agents

This command uses the following specialist agents:

| Agent | Expertise |
|-------|-----------|
| `cloud-architect` | Architecture design, Well-Architected Framework, multi-cloud |
| `terraform-engineer` | Terraform/OpenTofu, IaC, state management, CI/CD |
| `cost-optimizer` | Cost optimization, FinOps, right-sizing, pricing models |
| `container-orchestrator` | Docker, Kubernetes, ECS/EKS, Helm, deployment strategies |

---

## References

The following reference materials are available:

- `aws-services-guide` — Deep reference for core AWS services
- `kubernetes-patterns` — K8s deployment patterns, operators, CRDs
- `iac-best-practices` — Infrastructure as Code patterns and anti-patterns

---

## Tips

1. **Start with requirements.** Tell me what you're building, the expected scale, and any constraints.
2. **Share existing code.** I can review and improve your Terraform, Kubernetes manifests, or Dockerfiles.
3. **Be specific about compliance.** HIPAA, PCI-DSS, SOC2, and GDPR each have different infrastructure requirements.
4. **Include budget constraints.** Cost-aware architecture decisions are better than cost-optimizing after the fact.
5. **Mention your team.** Team size and cloud experience affect architecture recommendations.
