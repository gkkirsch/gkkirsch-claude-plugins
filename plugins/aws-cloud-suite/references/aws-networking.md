# AWS Networking Reference

Comprehensive guide to VPC design, subnet architecture, routing, Transit Gateway, PrivateLink, DNS, and network security.

---

## VPC Architecture

### Standard 3-Tier VPC Design

```
VPC CIDR: 10.0.0.0/16 (65,536 IPs)

AZ-a (us-east-1a)          AZ-b (us-east-1b)          AZ-c (us-east-1c)
┌─────────────────┐        ┌─────────────────┐        ┌─────────────────┐
│ Public Subnet   │        │ Public Subnet   │        │ Public Subnet   │
│ 10.0.0.0/20     │        │ 10.0.16.0/20    │        │ 10.0.32.0/20    │
│ (4,091 IPs)     │        │ (4,091 IPs)     │        │ (4,091 IPs)     │
│ ALB, NAT GW     │        │ ALB, NAT GW     │        │ ALB, NAT GW     │
├─────────────────┤        ├─────────────────┤        ├─────────────────┤
│ Private Subnet  │        │ Private Subnet  │        │ Private Subnet  │
│ 10.0.48.0/20    │        │ 10.0.64.0/20    │        │ 10.0.80.0/20    │
│ (4,091 IPs)     │        │ (4,091 IPs)     │        │ (4,091 IPs)     │
│ App servers,    │        │ App servers,    │        │ App servers,    │
│ Lambda ENIs     │        │ Lambda ENIs     │        │ Lambda ENIs     │
├─────────────────┤        ├─────────────────┤        ├─────────────────┤
│ Isolated Subnet │        │ Isolated Subnet │        │ Isolated Subnet │
│ 10.0.96.0/20    │        │ 10.0.112.0/20   │        │ 10.0.128.0/20   │
│ (4,091 IPs)     │        │ (4,091 IPs)     │        │ (4,091 IPs)     │
│ RDS, ElastiCache│        │ RDS, ElastiCache│        │ RDS, ElastiCache│
└─────────────────┘        └─────────────────┘        └─────────────────┘

Reserved: 10.0.144.0/16 through 10.0.255.255 for future use
```

**Subnet tiers:**
- **Public**: Has a route to Internet Gateway. For ALBs, NAT Gateways, bastion hosts
- **Private**: Has a route to NAT Gateway for outbound internet. For application workloads
- **Isolated**: No internet access at all. For databases and caches — maximum security

### CIDR Planning for Multi-Account

```
Network          │ CIDR            │ Size      │ Purpose
─────────────────┼─────────────────┼───────────┼──────────────────────
Hub/Shared       │ 10.0.0.0/16     │ 65,536    │ Transit GW, shared services
Production       │ 10.1.0.0/16     │ 65,536    │ Prod workloads
Staging          │ 10.2.0.0/16     │ 65,536    │ Staging workloads
Development      │ 10.3.0.0/16     │ 65,536    │ Dev workloads
Sandbox          │ 10.4.0.0/14     │ 262,144   │ Developer sandboxes (10.4-10.7)
Reserved         │ 10.8.0.0/13     │ 524,288   │ Future growth (10.8-10.15)
On-Premises      │ 172.16.0.0/12   │ 1,048,576 │ Corporate network
```

**Rules:**
- Never overlap CIDRs across accounts/VPCs that might need to communicate
- Use /16 per environment at minimum — IP exhaustion is a real problem
- Reserve space for growth — you can't expand a VPC CIDR retroactively (only add secondary CIDRs)
- Document every CIDR allocation in a central registry

### VPC Creation with CloudFormation

```yaml
VPC:
  Type: AWS::EC2::VPC
  Properties:
    CidrBlock: !Ref VpcCidr
    EnableDnsHostnames: true
    EnableDnsSupport: true
    Tags:
      - Key: Name
        Value: !Sub "${Environment}-vpc"

InternetGateway:
  Type: AWS::EC2::InternetGateway
  Properties:
    Tags:
      - Key: Name
        Value: !Sub "${Environment}-igw"

IGWAttachment:
  Type: AWS::EC2::VPCGatewayAttachment
  Properties:
    VpcId: !Ref VPC
    InternetGatewayId: !Ref InternetGateway

# Public route table — shared across all public subnets
PublicRouteTable:
  Type: AWS::EC2::RouteTable
  Properties:
    VpcId: !Ref VPC
    Tags:
      - Key: Name
        Value: !Sub "${Environment}-public-rt"

PublicDefaultRoute:
  Type: AWS::EC2::Route
  DependsOn: IGWAttachment
  Properties:
    RouteTableId: !Ref PublicRouteTable
    DestinationCidrBlock: 0.0.0.0/0
    GatewayId: !Ref InternetGateway

# Each AZ gets its own private route table (for AZ-specific NAT GW)
PrivateRouteTableAZ1:
  Type: AWS::EC2::RouteTable
  Properties:
    VpcId: !Ref VPC
    Tags:
      - Key: Name
        Value: !Sub "${Environment}-private-rt-az1"

PrivateDefaultRouteAZ1:
  Type: AWS::EC2::Route
  Properties:
    RouteTableId: !Ref PrivateRouteTableAZ1
    DestinationCidrBlock: 0.0.0.0/0
    NatGatewayId: !Ref NATGatewayAZ1

# Isolated route table — NO default route
IsolatedRouteTable:
  Type: AWS::EC2::RouteTable
  Properties:
    VpcId: !Ref VPC
    Tags:
      - Key: Name
        Value: !Sub "${Environment}-isolated-rt"
# No routes added — completely isolated from internet
```

---

## NAT Gateway

### Cost Optimization

NAT Gateways are expensive: $0.045/hour + $0.045/GB processed.

**A single NAT Gateway costs ~$32/month + data transfer.**
**Three NAT Gateways (HA) cost ~$97/month + data transfer.**

```
Strategy                    │ Cost/Month │ Availability │ When to Use
────────────────────────────┼────────────┼──────────────┼──────────────────
Single NAT GW               │ ~$32+      │ Single-AZ    │ Dev, non-critical
NAT GW per AZ (HA)          │ ~$97+      │ Multi-AZ     │ Production
NAT Instance (t4g.nano)     │ ~$3        │ Single-AZ    │ Sandbox, hobby
VPC Endpoints (replace NAT) │ $7-10/each │ Managed      │ AWS-only traffic
No NAT (private + endpoints)│ $0         │ N/A          │ When no internet needed
```

**Reduce NAT costs with VPC Endpoints:**
```bash
# Gateway endpoints are FREE — always use for S3 and DynamoDB
aws ec2 create-vpc-endpoint \
  --vpc-id vpc-123 \
  --service-name com.amazonaws.us-east-1.s3 \
  --route-table-ids rtb-private1 rtb-private2

aws ec2 create-vpc-endpoint \
  --vpc-id vpc-123 \
  --service-name com.amazonaws.us-east-1.dynamodb \
  --route-table-ids rtb-private1 rtb-private2

# Interface endpoints cost $7.20/month/AZ but avoid NAT Gateway data processing
# Use for services your workloads call frequently
aws ec2 create-vpc-endpoint \
  --vpc-id vpc-123 \
  --service-name com.amazonaws.us-east-1.execute-api \
  --vpc-endpoint-type Interface \
  --subnet-ids subnet-priv1 subnet-priv2 \
  --security-group-ids sg-vpce \
  --private-dns-enabled
```

### Common VPC Endpoints to Create

```
Endpoint                          │ Type      │ Cost/AZ/Month │ Why
──────────────────────────────────┼───────────┼───────────────┼──────────────────
com.amazonaws.REGION.s3           │ Gateway   │ Free          │ Always create
com.amazonaws.REGION.dynamodb     │ Gateway   │ Free          │ Always create
com.amazonaws.REGION.ecr.api      │ Interface │ $7.20         │ ECS pulls images
com.amazonaws.REGION.ecr.dkr      │ Interface │ $7.20         │ ECS pulls images
com.amazonaws.REGION.logs         │ Interface │ $7.20         │ CloudWatch Logs
com.amazonaws.REGION.monitoring   │ Interface │ $7.20         │ CloudWatch Metrics
com.amazonaws.REGION.secretsmanager│Interface │ $7.20         │ Secrets access
com.amazonaws.REGION.ssm          │ Interface │ $7.20         │ SSM Parameter Store
com.amazonaws.REGION.ssmmessages  │ Interface │ $7.20         │ SSM Session Manager
com.amazonaws.REGION.ec2messages  │ Interface │ $7.20         │ SSM Run Command
com.amazonaws.REGION.sts          │ Interface │ $7.20         │ IAM role assumption
com.amazonaws.REGION.kms          │ Interface │ $7.20         │ Encryption ops
com.amazonaws.REGION.sqs          │ Interface │ $7.20         │ SQS messaging
com.amazonaws.REGION.sns          │ Interface │ $7.20         │ SNS messaging
```

---

## Transit Gateway

### Hub-and-Spoke Architecture

```bash
# Create Transit Gateway
aws ec2 create-transit-gateway \
  --description "Central hub" \
  --options '{
    "AmazonSideAsn": 64512,
    "AutoAcceptSharedAttachments": "enable",
    "DefaultRouteTableAssociation": "disable",
    "DefaultRouteTablePropagation": "disable",
    "DnsSupport": "enable",
    "VpnEcmpSupport": "enable",
    "MulticastSupport": "disable"
  }' \
  --tags Key=Name,Value=central-tgw

# Attach VPCs
aws ec2 create-transit-gateway-vpc-attachment \
  --transit-gateway-id tgw-0123456789abcdef0 \
  --vpc-id vpc-prod-123 \
  --subnet-ids subnet-tgw-az1 subnet-tgw-az2 subnet-tgw-az3 \
  --options '{
    "DnsSupport": "enable",
    "ApplianceModeSupport": "enable"
  }' \
  --tags Key=Name,Value=prod-attachment
```

### Route Table Segmentation

```
Route Table: production-rt
  Associations: prod-vpc-attachment
  Routes:
    10.0.0.0/16  → shared-services-attachment  (shared services)
    10.1.0.0/16  → local (production VPC)
    0.0.0.0/0    → inspection-vpc-attachment    (centralized egress)
  Propagations:
    shared-services-attachment
    # NOTE: NO propagation from non-prod — production is isolated

Route Table: non-production-rt
  Associations: staging-vpc-attachment, dev-vpc-attachment
  Routes:
    10.0.0.0/16  → shared-services-attachment  (shared services)
    10.2.0.0/16  → local (staging)
    10.3.0.0/16  → local (dev)
    0.0.0.0/0    → inspection-vpc-attachment    (centralized egress)
  Propagations:
    staging-vpc-attachment
    dev-vpc-attachment
    shared-services-attachment
    # NOTE: NO propagation from prod — non-prod can't reach prod

Route Table: shared-services-rt
  Associations: shared-services-attachment
  Routes:
    10.0.0.0/16  → local
    10.1.0.0/16  → prod-attachment
    10.2.0.0/16  → staging-attachment
    10.3.0.0/16  → dev-attachment
    172.16.0.0/12→ vpn-attachment  (on-premises)
  Propagations:
    All attachments (shared services can reach everything)
```

### Cross-Region Transit Gateway Peering

```bash
# Create peering attachment
aws ec2 create-transit-gateway-peering-attachment \
  --transit-gateway-id tgw-us-east-1 \
  --peer-transit-gateway-id tgw-eu-west-1 \
  --peer-region eu-west-1 \
  --peer-account-id 123456789012 \
  --tags Key=Name,Value=us-east-1-to-eu-west-1

# Accept in the peer region
aws ec2 accept-transit-gateway-peering-attachment \
  --transit-gateway-attachment-id tgw-attach-peer-123 \
  --region eu-west-1

# Add routes in both regions
aws ec2 create-transit-gateway-route \
  --transit-gateway-route-table-id tgw-rtb-us-east-1 \
  --destination-cidr-block 10.100.0.0/16 \
  --transit-gateway-attachment-id tgw-attach-peer-123

aws ec2 create-transit-gateway-route \
  --transit-gateway-route-table-id tgw-rtb-eu-west-1 \
  --destination-cidr-block 10.0.0.0/8 \
  --transit-gateway-attachment-id tgw-attach-peer-123 \
  --region eu-west-1
```

---

## AWS PrivateLink

### Consuming Third-Party Services

```bash
# Create interface endpoint for a partner service
aws ec2 create-vpc-endpoint \
  --vpc-id vpc-123 \
  --service-name com.amazonaws.vpce.us-east-1.vpce-svc-partner123 \
  --vpc-endpoint-type Interface \
  --subnet-ids subnet-priv1 subnet-priv2 \
  --security-group-ids sg-vpce \
  --private-dns-enabled false

# After acceptance, access via endpoint DNS
# vpce-abc123.vpce-svc-partner123.us-east-1.vpce.amazonaws.com
```

### Exposing Your Service via PrivateLink

```yaml
# Create a Network Load Balancer (required for PrivateLink)
NLB:
  Type: AWS::ElasticLoadBalancingV2::LoadBalancer
  Properties:
    Type: network
    Scheme: internal
    Subnets:
      - !Ref PrivateSubnet1
      - !Ref PrivateSubnet2

# Create VPC Endpoint Service
EndpointService:
  Type: AWS::EC2::VPCEndpointService
  Properties:
    AcceptanceRequired: true
    NetworkLoadBalancerArns:
      - !Ref NLB

# Allow specific accounts to connect
EndpointServicePermission:
  Type: AWS::EC2::VPCEndpointServicePermissions
  Properties:
    ServiceId: !Ref EndpointService
    AllowedPrincipals:
      - arn:aws:iam::111111111111:root
      - arn:aws:iam::222222222222:root
```

---

## Route 53 DNS

### Hosted Zone Configuration

```bash
# Create a public hosted zone
aws route53 create-hosted-zone \
  --name example.com \
  --caller-reference "$(date +%s)" \
  --hosted-zone-config Comment="Production domain"

# Create a private hosted zone for internal DNS
aws route53 create-hosted-zone \
  --name internal.example.com \
  --caller-reference "$(date +%s)" \
  --vpc VPCRegion=us-east-1,VPCId=vpc-123 \
  --hosted-zone-config Comment="Internal services",PrivateZone=true
```

### Routing Policies

```
Policy              │ Use Case                           │ Example
────────────────────┼────────────────────────────────────┼────────────────────
Simple              │ Single resource                    │ api.example.com → ALB
Weighted            │ A/B testing, canary                │ 90% v1, 10% v2
Latency             │ Multi-region, lowest latency       │ Route to nearest region
Failover            │ Active-passive DR                  │ Primary → Secondary
Geolocation         │ Compliance, localization           │ EU users → EU region
Geoproximity        │ Fine-grained geographic routing    │ Bias toward regions
Multivalue Answer   │ Simple load balancing with health  │ Return multiple IPs
IP-based            │ Route by client IP ranges          │ ISP-specific routing
```

### Health Checks with Failover

```bash
# Create health check for primary endpoint
aws route53 create-health-check \
  --caller-reference "primary-$(date +%s)" \
  --health-check-config '{
    "FullyQualifiedDomainName": "api-us-east-1.example.com",
    "Port": 443,
    "Type": "HTTPS",
    "ResourcePath": "/health",
    "RequestInterval": 10,
    "FailureThreshold": 2,
    "MeasureLatency": true,
    "EnableSNI": true,
    "Regions": ["us-east-1", "eu-west-1", "ap-southeast-1"]
  }'

# Primary record (failover)
aws route53 change-resource-record-sets \
  --hosted-zone-id Z1234567890 \
  --change-batch '{
    "Changes": [
      {
        "Action": "CREATE",
        "ResourceRecordSet": {
          "Name": "api.example.com",
          "Type": "A",
          "SetIdentifier": "primary",
          "Failover": "PRIMARY",
          "AliasTarget": {
            "HostedZoneId": "Z35SXDOTRQ7X7K",
            "DNSName": "us-east-1-alb.elb.amazonaws.com",
            "EvaluateTargetHealth": true
          },
          "HealthCheckId": "health-check-id-primary"
        }
      },
      {
        "Action": "CREATE",
        "ResourceRecordSet": {
          "Name": "api.example.com",
          "Type": "A",
          "SetIdentifier": "secondary",
          "Failover": "SECONDARY",
          "AliasTarget": {
            "HostedZoneId": "Z2FDTNDATAQYW2",
            "DNSName": "eu-west-1-alb.elb.amazonaws.com",
            "EvaluateTargetHealth": true
          }
        }
      }
    ]
  }'
```

---

## Network Access Control Lists (NACLs)

### NACL vs Security Groups

```
Feature              │ NACL                        │ Security Group
─────────────────────┼─────────────────────────────┼──────────────────────
Level                │ Subnet                      │ ENI (instance)
Rules                │ Allow AND Deny              │ Allow only
Evaluation           │ Rules processed in order    │ All rules evaluated
Statefulness         │ Stateless (must define      │ Stateful (return
                     │ both inbound and outbound)  │ traffic auto-allowed)
Default              │ Allow all                   │ Deny all inbound
Association          │ 1 NACL per subnet           │ Multiple SGs per ENI
When to use          │ Subnet-level deny rules     │ Primary access control
```

### Production NACL Configuration

```yaml
# Public subnet NACL
PublicSubnetNACL:
  Type: AWS::EC2::NetworkAcl
  Properties:
    VpcId: !Ref VPC
    Tags:
      - Key: Name
        Value: public-subnet-nacl

# Inbound: Allow HTTPS from internet, ephemeral ports for return traffic
PublicInbound100:
  Type: AWS::EC2::NetworkAclEntry
  Properties:
    NetworkAclId: !Ref PublicSubnetNACL
    RuleNumber: 100
    Protocol: 6  # TCP
    RuleAction: allow
    CidrBlock: 0.0.0.0/0
    PortRange: { From: 443, To: 443 }

PublicInbound110:
  Type: AWS::EC2::NetworkAclEntry
  Properties:
    NetworkAclId: !Ref PublicSubnetNACL
    RuleNumber: 110
    Protocol: 6
    RuleAction: allow
    CidrBlock: 0.0.0.0/0
    PortRange: { From: 80, To: 80 }

PublicInbound120:
  Type: AWS::EC2::NetworkAclEntry
  Properties:
    NetworkAclId: !Ref PublicSubnetNACL
    RuleNumber: 120
    Protocol: 6
    RuleAction: allow
    CidrBlock: 0.0.0.0/0
    PortRange: { From: 1024, To: 65535 }  # Ephemeral ports for return traffic

# Outbound: Allow all (security groups handle fine-grained control)
PublicOutbound100:
  Type: AWS::EC2::NetworkAclEntry
  Properties:
    NetworkAclId: !Ref PublicSubnetNACL
    RuleNumber: 100
    Protocol: -1  # All
    Egress: true
    RuleAction: allow
    CidrBlock: 0.0.0.0/0
```

---

## Direct Connect

### Connection Architecture

```
On-Premises                    AWS
┌──────────┐    ┌─────────┐   ┌──────────────────┐
│ Router   │────│ Partner │───│ Direct Connect   │
│          │    │ Cage    │   │ Location         │
└──────────┘    └─────────┘   └────────┬─────────┘
                                       │
                              ┌────────▼─────────┐
                              │ Virtual Interfaces│
                              ├──────────────────┤
                              │ Private VIF ──→ VPC (via VGW or TGW)
                              │ Transit VIF ──→ Transit Gateway
                              │ Public VIF  ──→ AWS public services
                              └──────────────────┘
```

### Resilient Direct Connect

```
Pattern                         │ Connections            │ SLA
────────────────────────────────┼────────────────────────┼──────────
Development (non-critical)      │ 1 DX + VPN backup      │ No SLA
Maximum Resiliency (critical)   │ 2 DX (different locs)  │ 99.99%
High Resiliency (important)     │ 2 DX (same location)   │ 99.9%
```

```bash
# Create a Transit Virtual Interface for Transit Gateway attachment
aws directconnect create-transit-virtual-interface \
  --connection-id dxcon-abc123 \
  --new-transit-virtual-interface '{
    "virtualInterfaceName": "prod-transit-vif",
    "vlan": 100,
    "asn": 65000,
    "authKey": "your-bgp-auth-key",
    "amazonAddress": "169.254.100.1/30",
    "customerAddress": "169.254.100.2/30",
    "directConnectGatewayId": "dx-gw-123"
  }'
```

---

## Load Balancing

### ALB vs NLB vs GWLB

```
Feature           │ ALB                  │ NLB                 │ GWLB
──────────────────┼──────────────────────┼─────────────────────┼──────────────
Layer             │ 7 (HTTP/HTTPS)       │ 4 (TCP/UDP/TLS)     │ 3 (IP)
Latency           │ ~50-100ms added      │ ~<1ms added          │ Transparent
Static IP         │ No (use Global Acc.) │ Yes (Elastic IP)    │ No
WebSocket         │ Yes                  │ Yes                 │ No
gRPC              │ Yes                  │ No (use TLS pass)   │ No
PrivateLink       │ No                   │ Yes (endpoint svc)  │ Yes
WAF               │ Yes                  │ No                  │ No
Auth (Cognito/OIDC)│ Yes                 │ No                  │ No
Use case          │ Web APIs, microsvcs  │ High perf, TCP      │ Security appliances
```

### ALB Advanced Configuration

```yaml
ALB:
  Type: AWS::ElasticLoadBalancingV2::LoadBalancer
  Properties:
    Type: application
    Scheme: internet-facing
    SecurityGroups:
      - !Ref ALBSecurityGroup
    Subnets:
      - !Ref PublicSubnet1
      - !Ref PublicSubnet2
      - !Ref PublicSubnet3
    LoadBalancerAttributes:
      - Key: routing.http2.enabled
        Value: "true"
      - Key: idle_timeout.timeout_seconds
        Value: "60"
      - Key: routing.http.drop_invalid_header_fields.enabled
        Value: "true"
      - Key: deletion_protection.enabled
        Value: "true"
      - Key: access_logs.s3.enabled
        Value: "true"
      - Key: access_logs.s3.bucket
        Value: !Ref ALBLogBucket
      - Key: connection_logs.s3.enabled
        Value: "true"
      - Key: connection_logs.s3.bucket
        Value: !Ref ALBLogBucket

HTTPSListener:
  Type: AWS::ElasticLoadBalancingV2::Listener
  Properties:
    LoadBalancerArn: !Ref ALB
    Port: 443
    Protocol: HTTPS
    SslPolicy: ELBSecurityPolicy-TLS13-1-2-2021-06
    Certificates:
      - CertificateArn: !Ref Certificate
    DefaultActions:
      - Type: fixed-response
        FixedResponseConfig:
          StatusCode: "404"
          ContentType: text/plain
          MessageBody: "Not Found"

# Path-based routing
APIRoute:
  Type: AWS::ElasticLoadBalancingV2::ListenerRule
  Properties:
    ListenerArn: !Ref HTTPSListener
    Priority: 10
    Conditions:
      - Field: path-pattern
        PathPatternConfig:
          Values: ["/api/*"]
      - Field: http-request-method
        HttpRequestMethodConfig:
          Values: [GET, POST, PUT, DELETE, PATCH]
    Actions:
      - Type: forward
        TargetGroupArn: !Ref APITargetGroup

# Host-based routing for multi-tenant
TenantRoute:
  Type: AWS::ElasticLoadBalancingV2::ListenerRule
  Properties:
    ListenerArn: !Ref HTTPSListener
    Priority: 20
    Conditions:
      - Field: host-header
        HostHeaderConfig:
          Values: ["*.app.example.com"]
    Actions:
      - Type: forward
        TargetGroupArn: !Ref AppTargetGroup

# HTTP → HTTPS redirect
HTTPRedirect:
  Type: AWS::ElasticLoadBalancingV2::Listener
  Properties:
    LoadBalancerArn: !Ref ALB
    Port: 80
    Protocol: HTTP
    DefaultActions:
      - Type: redirect
        RedirectConfig:
          Protocol: HTTPS
          Port: "443"
          StatusCode: HTTP_301
```

---

## AWS Network Firewall

### Centralized Inspection

```yaml
NetworkFirewall:
  Type: AWS::NetworkFirewall::Firewall
  Properties:
    FirewallName: central-inspection
    FirewallPolicyArn: !Ref FirewallPolicy
    VpcId: !Ref InspectionVPC
    SubnetMappings:
      - SubnetId: !Ref FirewallSubnet1
      - SubnetId: !Ref FirewallSubnet2

FirewallPolicy:
  Type: AWS::NetworkFirewall::FirewallPolicy
  Properties:
    FirewallPolicyName: central-policy
    FirewallPolicy:
      StatelessDefaultActions: [aws:forward_to_sfe]
      StatelessFragmentDefaultActions: [aws:forward_to_sfe]
      StatefulEngineOptions:
        RuleOrder: STRICT_ORDER
      StatefulDefaultActions: [aws:drop_strict]
      StatefulRuleGroupReferences:
        - ResourceArn: !Ref AllowedDomains
          Priority: 100
        - ResourceArn: !Ref AWSManagedThreatSignatures
          Priority: 200

AllowedDomains:
  Type: AWS::NetworkFirewall::RuleGroup
  Properties:
    RuleGroupName: allowed-outbound-domains
    Type: STATEFUL
    Capacity: 100
    RuleGroup:
      RulesSource:
        RulesSourceList:
          TargetTypes: [TLS_SNI, HTTP_HOST]
          Targets:
            - ".amazonaws.com"
            - ".aws.amazon.com"
            - "registry.npmjs.org"
            - "pypi.org"
            - "github.com"
            - ".githubusercontent.com"
            - ".docker.io"
            - ".docker.com"
          GeneratedRulesType: ALLOWLIST

AWSManagedThreatSignatures:
  Type: AWS::NetworkFirewall::RuleGroup
  Properties:
    RuleGroupName: aws-managed-threats
    Type: STATEFUL
    Capacity: 1000
    RuleGroup:
      RulesSource:
        RulesSourceList:
          TargetTypes: [TLS_SNI, HTTP_HOST]
          Targets: [".malware-domain.com", ".phishing-site.com"]
          GeneratedRulesType: DENYLIST
```

---

## Network Troubleshooting

### VPC Flow Logs Analysis

```bash
# Enable flow logs to S3 (cheapest option for long-term storage)
aws ec2 create-flow-log \
  --resource-type VPC \
  --resource-ids vpc-123 \
  --traffic-type ALL \
  --log-destination-type s3 \
  --log-destination arn:aws:s3:::vpc-flow-logs-bucket \
  --max-aggregation-interval 60 \
  --log-format '${version} ${account-id} ${interface-id} ${srcaddr} ${dstaddr} ${srcport} ${dstport} ${protocol} ${packets} ${bytes} ${start} ${end} ${action} ${log-status} ${vpc-id} ${subnet-id} ${az-id} ${pkt-srcaddr} ${pkt-dstaddr} ${region} ${flow-direction} ${traffic-path}'
```

### Reachability Analyzer

```bash
# Check if traffic can flow between two resources
aws ec2 create-network-insights-path \
  --source eni-source123 \
  --destination eni-dest456 \
  --protocol TCP \
  --destination-port 443 \
  --tags Key=Name,Value=test-api-to-db

aws ec2 start-network-insights-analysis \
  --network-insights-path-id nip-123

# Check results
aws ec2 describe-network-insights-analyses \
  --network-insights-analysis-ids nia-123 \
  --query 'NetworkInsightsAnalyses[0].{Status:Status,PathFound:NetworkPathFound,Explanations:Explanations}'
```

### Common Networking Issues

```
Symptom                              │ Likely Cause                  │ Fix
─────────────────────────────────────┼───────────────────────────────┼──────────
Lambda can't reach internet          │ No NAT GW or no route         │ Add NAT GW route
Lambda timeout to AWS service        │ In VPC without endpoint       │ Add VPC endpoint
EC2 can't reach RDS                  │ Security group misconfigured  │ Allow SG-to-SG
ALB health check failing             │ Security group doesn't allow  │ Allow ALB SG ingress
                                     │ ALB health check port         │
Cross-account access denied          │ Missing resource policy       │ Add resource policy
DNS resolution fails in VPC          │ EnableDnsSupport = false      │ Enable DNS support
```
