---
name: cloud-networking
description: >
  Cloud networking patterns — VPC design, subnet strategies, security groups,
  load balancers, DNS with Route53, CDN with CloudFront, VPN/peering,
  service mesh, and network troubleshooting.
  Triggers: "vpc", "subnet", "security group", "load balancer", "route53",
  "cloudfront", "cdn", "network", "dns", "vpn", "peering".
  NOT for: Cost optimization or billing (use cloud-cost-optimization).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Cloud Networking

## VPC Design (AWS)

```hcl
# Multi-AZ VPC with public, private, and isolated subnets
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = { Name = "production-vpc" }
}

# Subnet layout:
# 10.0.0.0/20  = Public  (AZ-a) — 4,094 IPs
# 10.0.16.0/20 = Public  (AZ-b)
# 10.0.32.0/20 = Public  (AZ-c)
# 10.0.48.0/20 = Private (AZ-a) — app servers
# 10.0.64.0/20 = Private (AZ-b)
# 10.0.80.0/20 = Private (AZ-c)
# 10.0.96.0/20 = Isolated (AZ-a) — databases (no internet)
# 10.0.112.0/20 = Isolated (AZ-b)
# 10.0.128.0/20 = Isolated (AZ-c)

locals {
  azs = ["us-east-1a", "us-east-1b", "us-east-1c"]

  public_subnets   = ["10.0.0.0/20", "10.0.16.0/20", "10.0.32.0/20"]
  private_subnets  = ["10.0.48.0/20", "10.0.64.0/20", "10.0.80.0/20"]
  isolated_subnets = ["10.0.96.0/20", "10.0.112.0/20", "10.0.128.0/20"]
}

resource "aws_subnet" "public" {
  count                   = length(local.azs)
  vpc_id                  = aws_vpc.main.id
  cidr_block              = local.public_subnets[count.index]
  availability_zone       = local.azs[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = "public-${local.azs[count.index]}"
    Tier = "public"
    "kubernetes.io/role/elb" = "1"  # For K8s service type LoadBalancer
  }
}

resource "aws_subnet" "private" {
  count             = length(local.azs)
  vpc_id            = aws_vpc.main.id
  cidr_block        = local.private_subnets[count.index]
  availability_zone = local.azs[count.index]

  tags = {
    Name = "private-${local.azs[count.index]}"
    Tier = "private"
    "kubernetes.io/role/internal-elb" = "1"
  }
}

resource "aws_subnet" "isolated" {
  count             = length(local.azs)
  vpc_id            = aws_vpc.main.id
  cidr_block        = local.isolated_subnets[count.index]
  availability_zone = local.azs[count.index]

  tags = {
    Name = "isolated-${local.azs[count.index]}"
    Tier = "isolated"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
}

# NAT Gateways (one per AZ for HA)
resource "aws_eip" "nat" {
  count  = length(local.azs)
  domain = "vpc"
}

resource "aws_nat_gateway" "main" {
  count         = length(local.azs)
  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id
}

# Route tables
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
}

resource "aws_route_table" "private" {
  count  = length(local.azs)
  vpc_id = aws_vpc.main.id
  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main[count.index].id
  }
}

# Isolated subnets: NO route to internet (no NAT, no IGW)
resource "aws_route_table" "isolated" {
  vpc_id = aws_vpc.main.id
  # No routes — only VPC-internal traffic
}
```

## Security Groups

```hcl
# Application Load Balancer SG
resource "aws_security_group" "alb" {
  name_prefix = "alb-"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS from internet"
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP (redirect to HTTPS)"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle { create_before_destroy = true }
}

# App server SG — only accepts traffic from ALB
resource "aws_security_group" "app" {
  name_prefix = "app-"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
    description     = "App port from ALB only"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Database SG — only accepts from app servers
resource "aws_security_group" "db" {
  name_prefix = "db-"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.app.id]
    description     = "PostgreSQL from app servers only"
  }

  # No egress to internet — isolated subnet handles this
}
```

## Application Load Balancer

```hcl
resource "aws_lb" "main" {
  name               = "api-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.public[*].id

  enable_deletion_protection = true

  access_logs {
    bucket  = aws_s3_bucket.alb_logs.id
    prefix  = "api-alb"
    enabled = true
  }
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = aws_acm_certificate.api.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api.arn
  }
}

resource "aws_lb_listener" "http_redirect" {
  load_balancer_arn = aws_lb.main.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb_target_group" "api" {
  name        = "api-tg"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"  # For ECS Fargate / K8s pods

  health_check {
    path                = "/health"
    port                = 8080
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 15
    matcher             = "200"
  }

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 86400
    enabled         = false  # Enable only if needed
  }

  deregistration_delay = 30  # Seconds to drain connections
}

# Path-based routing
resource "aws_lb_listener_rule" "api_v2" {
  listener_arn = aws_lb_listener.https.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api_v2.arn
  }

  condition {
    path_pattern {
      values = ["/api/v2/*"]
    }
  }
}
```

## DNS with Route53

```hcl
resource "aws_route53_zone" "main" {
  name = "example.com"
}

# A record with alias to ALB
resource "aws_route53_record" "api" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "api.example.com"
  type    = "A"

  alias {
    name                   = aws_lb.main.dns_name
    zone_id                = aws_lb.main.zone_id
    evaluate_target_health = true
  }
}

# Weighted routing (for canary/blue-green)
resource "aws_route53_record" "api_blue" {
  zone_id        = aws_route53_zone.main.zone_id
  name           = "api.example.com"
  type           = "A"
  set_identifier = "blue"

  weighted_routing_policy {
    weight = 90
  }

  alias {
    name    = aws_lb.blue.dns_name
    zone_id = aws_lb.blue.zone_id
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "api_green" {
  zone_id        = aws_route53_zone.main.zone_id
  name           = "api.example.com"
  type           = "A"
  set_identifier = "green"

  weighted_routing_policy {
    weight = 10
  }

  alias {
    name    = aws_lb.green.dns_name
    zone_id = aws_lb.green.zone_id
    evaluate_target_health = true
  }
}

# Health check for failover
resource "aws_route53_health_check" "api" {
  fqdn              = "api.example.com"
  port               = 443
  type               = "HTTPS"
  resource_path      = "/health"
  failure_threshold  = 3
  request_interval   = 10

  tags = { Name = "api-health-check" }
}
```

## CDN with CloudFront

```hcl
resource "aws_cloudfront_distribution" "main" {
  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = "index.html"
  price_class         = "PriceClass_100"  # US/EU only (cheapest)
  aliases             = ["app.example.com"]

  origin {
    domain_name = aws_lb.main.dns_name
    origin_id   = "api"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  origin {
    domain_name              = aws_s3_bucket.static.bucket_regional_domain_name
    origin_id                = "s3-static"
    origin_access_control_id = aws_cloudfront_origin_access_control.s3.id
  }

  # API requests — no caching, pass through
  ordered_cache_behavior {
    path_pattern     = "/api/*"
    target_origin_id = "api"
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]

    forwarded_values {
      query_string = true
      headers      = ["Authorization", "Accept", "Origin"]
      cookies { forward = "none" }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 0
    max_ttl                = 0
  }

  # Static assets — aggressive caching
  default_cache_behavior {
    target_origin_id       = "s3-static"
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    viewer_protocol_policy = "redirect-to-https"

    forwarded_values {
      query_string = false
      cookies { forward = "none" }
    }

    min_ttl     = 0
    default_ttl = 86400    # 1 day
    max_ttl     = 31536000 # 1 year
    compress    = true
  }

  # SPA fallback
  custom_error_response {
    error_code         = 404
    response_code      = 200
    response_page_path = "/index.html"
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn      = aws_acm_certificate.cdn.arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }
}
```

## Network Troubleshooting

```bash
# VPC Flow Logs analysis
aws ec2 describe-flow-logs --filter "Name=resource-id,Values=vpc-abc123"

# Check security group rules for a specific port
aws ec2 describe-security-groups --group-ids sg-abc123 \
  --query 'SecurityGroups[].IpPermissions[?FromPort==`5432`]'

# Check if an instance can reach the internet
aws ec2 describe-route-tables \
  --filters "Name=association.subnet-id,Values=subnet-abc123" \
  --query 'RouteTables[].Routes[]'

# DNS resolution test
dig api.example.com
nslookup api.example.com
host api.example.com

# Test connectivity from within a pod (K8s)
kubectl run debug --rm -it --image=nicolaka/netshoot -- bash
# Inside the pod:
curl -v https://api.example.com/health
nslookup api-server.default.svc.cluster.local
traceroute api-server

# Check ALB target health
aws elbv2 describe-target-health --target-group-arn arn:aws:...

# VPC Reachability Analyzer
aws ec2 create-network-insights-path \
  --source i-abc123 \
  --destination i-xyz789 \
  --protocol tcp \
  --destination-port 5432
```

## Gotchas

1. **NAT Gateway costs add up fast** — Each NAT Gateway costs ~$32/month plus $0.045/GB processed. Three NAT Gateways (one per AZ for HA) = ~$96/month before any data transfer. For non-production environments, use a single NAT Gateway or NAT instances. Consider VPC endpoints for AWS service traffic (S3, DynamoDB, ECR) to avoid NAT charges entirely.

2. **Security group rules are stateful but NACLs are not** — Security groups automatically allow return traffic. NACLs don't — you must explicitly allow both inbound AND outbound on the correct ports. Forgetting the ephemeral port range (1024-65535) on the outbound NACL rule blocks all responses.

3. **Cross-AZ data transfer charges** — Traffic between subnets in different AZs costs $0.01/GB each way ($0.02 round trip). Database replication, service-to-service calls, and log shipping across AZs add up. Place frequently communicating services in the same AZ when possible, or use AZ-aware routing.

4. **Route53 alias records don't support TTL** — Alias records to AWS resources (ALB, CloudFront, S3) inherit the target's TTL, which is typically 60 seconds. You can't set a longer TTL to reduce DNS queries. For non-AWS targets, use CNAME records with custom TTL instead.

5. **CloudFront caches error responses** — A 502 or 503 from your origin gets cached by CloudFront for the `min_ttl` duration. Set `custom_error_response` with a short `error_caching_min_ttl` (e.g., 10 seconds) to prevent stale errors being served for minutes.

6. **VPC peering doesn't support transitive routing** — If VPC-A peers with VPC-B and VPC-B peers with VPC-C, VPC-A cannot reach VPC-C through VPC-B. You need a direct peering connection or use Transit Gateway for hub-and-spoke networking with many VPCs.
