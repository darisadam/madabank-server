# 🏗️ System Architecture

## Overview

MadaBank Server is designed as a **modular monolith** with a clear separation of concerns, ensuring high performance, maintainability, and scalability. It leverages AWS Cloud-Native services for robustness.

## 🧩 High-Level Design

```mermaid
graph TD
    Client[Client (Mobile/Web)] -->|HTTPS| ALB[Application LoadBalancer]
    ALB -->|Route| ECS[AWS ECS Fargate Cluster]
    
    subgraph "VPC (Private Subnet)"
        ECS -->|Gorm| RDS[(AWS RDS PostgreSQL)]
        ECS -->|Cache/Session| Redis[(AWS ElastiCache Redis)]
    end
    
    subgraph "External Services"
        ECS -->|Logging| CW[CloudWatch Logs]
        ECS -->|Metrics| Prom[Prometheus/Grafana]
    end
```

## 🛠️ Technology Choices

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **Language** | **Go 1.24+** | High concurency, strong typing, fast startup (ideal for Fargate). |
| **Framework** | **Gin** | Minimalist, high-performance HTTP web framework. |
| **Database** | **PostgreSQL 16** | ACID compliance is non-negotiable for financial transactions. |
| **Caching** | **Redis 7** | Sub-millisecond latency for session management and rate limiting. |
| **Infrastructure** | **Terraform** | Infrastructure as Code (IaC) ensures reproducible environments. |
| **Platform** | **AWS ECS Fargate** | Serverless containers reduce operational overhead compared to EC2/EKS. |

## 🔐 Security Architecture

1.  **Network Isolation**:
    *   Database and Redis reside in **Private Subnets**, inaccessible from the public internet.
    *   Only the Load Balancer receives public traffic (port 80/443).
    *   NAT Gateways provide controlled outbound access for updates/APIs.

2.  **Data Security**:
    *   **At Rest**: AWS KMS encryption for RDS and ElastiCache.
    *   **In Transit**: TLS 1.2+ enforced everywhere.
    *   **Application**:
        *   **JWT (RS256)**: Asymmetric signing for tamper-proof authentication.
        *   **E2EE**: Application-layer encryption for sensitive card data (AES-256 + RSA-2048).

3.  **Secrets Management**:
    *   No hardcoded secrets. All credentials injected via **AWS Secrets Manager** or Environment Variables at runtime.

## 🔄 Data Flow (Transaction Example)

1.  User initiates transfer.
2.  Request hits **ALB** -> Forwarded to **API Container**.
3.  **Middleware** validates JWT & Rate Limits (Redis).
4.  **Service Layer** starts a DB Transaction.
5.  **Repository Layer** updates Receiver/Sender balances (Atomic lock).
6.  **Audit Logger** records the event.
7.  Transaction committed & Response sent.
