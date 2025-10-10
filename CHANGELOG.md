# Changelog

All notable changes to this project will be documented in this file.

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-10-10)

### ⚠ BREAKING CHANGES

* **terraform:** Replace containerized data stores with AWS managed
services (RDS PostgreSQL, ElastiCache Redis, S3) for improved
reliability, scalability, and operational efficiency.

New Terraform Modules:
- RDS PostgreSQL with Multi-AZ, automated backups, encryption
- ElastiCache Redis with replication and automatic failover
- S3 buckets with versioning, encryption, lifecycle policies

Module Refactoring:
- ECS Service: Simplified to API-only container, removed Postgres/Redis/MinIO
- Security Groups: Added RDS and ElastiCache SGs, removed EFS
- Secrets Manager: Flexible payload structure per environment

Environment Migration:
Production:
- RDS db.r6g.large Multi-AZ, 100GB gp3, 7-day backups
- ElastiCache cache.r6g.large with 2 replicas
- S3 whatsmeow-production-media
- Cost: ~$575/month

Staging:
- RDS db.t4g.medium single-AZ, 20GB, 3-day backups
- ElastiCache cache.t4g.small with 1 replica
- FARGATE_SPOT enabled
- Cost: ~$126/month

Homolog:
- RDS db.t4g.small single-AZ, 10GB, 1-day backups
- ElastiCache cache.t4g.small no replicas
- FARGATE_SPOT, NAT Gateway disabled
- Cost: ~$84/month

Documentation:
- Updated architecture diagram with managed services
- New cost breakdown per environment
- Updated troubleshooting for RDS/ElastiCache/S3
- Removed EFS and container-based service documentation

Benefits:
- Automated backups and point-in-time recovery
- Managed patching and maintenance
- Better scalability with auto-scaling support
- Enhanced security with encryption at rest/transit
- Reduced operational complexity

### ✨ Features

* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-10-10)

### ⚠ BREAKING CHANGES

* **terraform:** Replace containerized data stores with AWS managed
services (RDS PostgreSQL, ElastiCache Redis, S3) for improved
reliability, scalability, and operational efficiency.

New Terraform Modules:
- RDS PostgreSQL with Multi-AZ, automated backups, encryption
- ElastiCache Redis with replication and automatic failover
- S3 buckets with versioning, encryption, lifecycle policies

Module Refactoring:
- ECS Service: Simplified to API-only container, removed Postgres/Redis/MinIO
- Security Groups: Added RDS and ElastiCache SGs, removed EFS
- Secrets Manager: Flexible payload structure per environment

Environment Migration:
Production:
- RDS db.r6g.large Multi-AZ, 100GB gp3, 7-day backups
- ElastiCache cache.r6g.large with 2 replicas
- S3 whatsmeow-production-media
- Cost: ~$575/month

Staging:
- RDS db.t4g.medium single-AZ, 20GB, 3-day backups
- ElastiCache cache.t4g.small with 1 replica
- FARGATE_SPOT enabled
- Cost: ~$126/month

Homolog:
- RDS db.t4g.small single-AZ, 10GB, 1-day backups
- ElastiCache cache.t4g.small no replicas
- FARGATE_SPOT, NAT Gateway disabled
- Cost: ~$84/month

Documentation:
- Updated architecture diagram with managed services
- New cost breakdown per environment
- Updated troubleshooting for RDS/ElastiCache/S3
- Removed EFS and container-based service documentation

Benefits:
- Automated backups and point-in-time recovery
- Managed patching and maintenance
- Better scalability with auto-scaling support
- Enhanced security with encryption at rest/transit
- Reduced operational complexity

### ✨ Features

* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-10-10)

### ⚠ BREAKING CHANGES

* **terraform:** Replace containerized data stores with AWS managed
services (RDS PostgreSQL, ElastiCache Redis, S3) for improved
reliability, scalability, and operational efficiency.

New Terraform Modules:
- RDS PostgreSQL with Multi-AZ, automated backups, encryption
- ElastiCache Redis with replication and automatic failover
- S3 buckets with versioning, encryption, lifecycle policies

Module Refactoring:
- ECS Service: Simplified to API-only container, removed Postgres/Redis/MinIO
- Security Groups: Added RDS and ElastiCache SGs, removed EFS
- Secrets Manager: Flexible payload structure per environment

Environment Migration:
Production:
- RDS db.r6g.large Multi-AZ, 100GB gp3, 7-day backups
- ElastiCache cache.r6g.large with 2 replicas
- S3 whatsmeow-production-media
- Cost: ~$575/month

Staging:
- RDS db.t4g.medium single-AZ, 20GB, 3-day backups
- ElastiCache cache.t4g.small with 1 replica
- FARGATE_SPOT enabled
- Cost: ~$126/month

Homolog:
- RDS db.t4g.small single-AZ, 10GB, 1-day backups
- ElastiCache cache.t4g.small no replicas
- FARGATE_SPOT, NAT Gateway disabled
- Cost: ~$84/month

Documentation:
- Updated architecture diagram with managed services
- New cost breakdown per environment
- Updated troubleshooting for RDS/ElastiCache/S3
- Removed EFS and container-based service documentation

Benefits:
- Automated backups and point-in-time recovery
- Managed patching and maintenance
- Better scalability with auto-scaling support
- Enhanced security with encryption at rest/transit
- Reduced operational complexity

### ✨ Features

* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-10-10)

### ⚠ BREAKING CHANGES

* **terraform:** Replace containerized data stores with AWS managed
services (RDS PostgreSQL, ElastiCache Redis, S3) for improved
reliability, scalability, and operational efficiency.

New Terraform Modules:
- RDS PostgreSQL with Multi-AZ, automated backups, encryption
- ElastiCache Redis with replication and automatic failover
- S3 buckets with versioning, encryption, lifecycle policies

Module Refactoring:
- ECS Service: Simplified to API-only container, removed Postgres/Redis/MinIO
- Security Groups: Added RDS and ElastiCache SGs, removed EFS
- Secrets Manager: Flexible payload structure per environment

Environment Migration:
Production:
- RDS db.r6g.large Multi-AZ, 100GB gp3, 7-day backups
- ElastiCache cache.r6g.large with 2 replicas
- S3 whatsmeow-production-media
- Cost: ~$575/month

Staging:
- RDS db.t4g.medium single-AZ, 20GB, 3-day backups
- ElastiCache cache.t4g.small with 1 replica
- FARGATE_SPOT enabled
- Cost: ~$126/month

Homolog:
- RDS db.t4g.small single-AZ, 10GB, 1-day backups
- ElastiCache cache.t4g.small no replicas
- FARGATE_SPOT, NAT Gateway disabled
- Cost: ~$84/month

Documentation:
- Updated architecture diagram with managed services
- New cost breakdown per environment
- Updated troubleshooting for RDS/ElastiCache/S3
- Removed EFS and container-based service documentation

Benefits:
- Automated backups and point-in-time recovery
- Managed patching and maintenance
- Better scalability with auto-scaling support
- Enhanced security with encryption at rest/transit
- Reduced operational complexity

### ✨ Features

* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))

## [2.0.0-homolog.3](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v2.0.0-homolog.2...v2.0.0-homolog.3) (2025-10-10)

### ✨ Features

* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))

## [2.0.0-homolog.2](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v2.0.0-homolog.1...v2.0.0-homolog.2) (2025-10-10)

### ✨ Features

* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))

## [2.0.0-homolog.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.1.0...v2.0.0-homolog.1) (2025-10-09)

### ⚠ BREAKING CHANGES

* **terraform:** Replace containerized data stores with AWS managed
services (RDS PostgreSQL, ElastiCache Redis, S3) for improved
reliability, scalability, and operational efficiency.

New Terraform Modules:
- RDS PostgreSQL with Multi-AZ, automated backups, encryption
- ElastiCache Redis with replication and automatic failover
- S3 buckets with versioning, encryption, lifecycle policies

Module Refactoring:
- ECS Service: Simplified to API-only container, removed Postgres/Redis/MinIO
- Security Groups: Added RDS and ElastiCache SGs, removed EFS
- Secrets Manager: Flexible payload structure per environment

Environment Migration:
Production:
- RDS db.r6g.large Multi-AZ, 100GB gp3, 7-day backups
- ElastiCache cache.r6g.large with 2 replicas
- S3 whatsmeow-production-media
- Cost: ~$575/month

Staging:
- RDS db.t4g.medium single-AZ, 20GB, 3-day backups
- ElastiCache cache.t4g.small with 1 replica
- FARGATE_SPOT enabled
- Cost: ~$126/month

Homolog:
- RDS db.t4g.small single-AZ, 10GB, 1-day backups
- ElastiCache cache.t4g.small no replicas
- FARGATE_SPOT, NAT Gateway disabled
- Cost: ~$84/month

Documentation:
- Updated architecture diagram with managed services
- New cost breakdown per environment
- Updated troubleshooting for RDS/ElastiCache/S3
- Removed EFS and container-based service documentation

Benefits:
- Automated backups and point-in-time recovery
- Managed patching and maintenance
- Better scalability with auto-scaling support
- Enhanced security with encryption at rest/transit
- Reduced operational complexity

### ✨ Features

* **observability:** add comprehensive metrics and async context helpers ([680c8a6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/680c8a602a528f7f14adf069121cab354245eb1d))
* **db:** add event outbox, DLQ, and media metadata schema ([321fa6b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/321fa6baa15eabfc4bc57dbeabc95f82da7b285e))
* **config:** add event system and media configuration ([d52469f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d52469fd2d2989847fbfd30dbf8e512fd04f81b3))
* **config:** add media cleanup and S3 URL configuration options ([858e789](https://github.com/Funnelchat20/whatsapp-api-golang/commit/858e78955e1484f94d195f66368966767e6e3cad))
* **handlers:** add media HTTP handler for local file serving ([42300f3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42300f3ebb30dcf382e9968e6a745b3e1558cdff))
* **media:** add presigned URL toggle and public URL support for S3 ([e9900ee](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e9900eeecffa399cf26bd8d84527e805051ef902))
* **events:** add undecryptable message support and contact metadata enrichment ([dce506d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dce506d2a18832f175a77f702f201188f9a2888a))
* **whatsmeow:** enhance contact metadata with photo details and presence system ([602f920](https://github.com/Funnelchat20/whatsapp-api-golang/commit/602f920507cd0783de6e95f91619e41366472148))
* **events:** enhance ZAPI transformer with complete webhook payload support ([c495b3e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c495b3e89b7f81fdd31f4337e0472cb7dbb8b2e7))
* **events:** enrich webhook payloads with contact metadata ([7d043d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7d043d68450669924ccc823af8c19829d458ef61))
* **media:** implement automated media cleanup system with distributed locking ([8a88244](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8a88244431a792ccfa2a182a2be65a9fc0500a73))
* **events:** implement comprehensive event processing system ([c84d840](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c84d8408400bcb71a510687b2d8ccd8101012470))
* **whatsmeow:** implement LID to phone number resolution system ([ca50b99](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ca50b99679091e541fcbc9f745ca9667ed8cca02))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **integration:** wire event system into WhatsApp client lifecycle ([98f787c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/98f787c038da4058bb73a42460f40ce32746b031))

### ♻️ Code Refactoring

* **handlers:** clean up import aliases for consistency ([e443649](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e4436490fef498028846471a91c75ea581423db4))
* **locks:** enhance circuit breaker metrics and tracking ([76a7ba6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/76a7ba614799dd117f36ab2412610dd3c88572f1))
* **integration:** finalize event system wiring and interfaces ([5831173](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5831173ce3f5a705ace8bb2b0907a8e53ed35a7c))
* **events:** persist internal events using JSON ([f9f9e6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f9f9e6cad3e1a46b6928356feb4fe1dae3de7fea))

### 📝 Documentation

* update code standards and add comprehensive development plan ([e59b86b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e59b86b8c2e0120e49c31ab129e96696b233c2e7))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-10-09)

### ⚠ BREAKING CHANGES

* **terraform:** Replace containerized data stores with AWS managed
services (RDS PostgreSQL, ElastiCache Redis, S3) for improved
reliability, scalability, and operational efficiency.

New Terraform Modules:
- RDS PostgreSQL with Multi-AZ, automated backups, encryption
- ElastiCache Redis with replication and automatic failover
- S3 buckets with versioning, encryption, lifecycle policies

Module Refactoring:
- ECS Service: Simplified to API-only container, removed Postgres/Redis/MinIO
- Security Groups: Added RDS and ElastiCache SGs, removed EFS
- Secrets Manager: Flexible payload structure per environment

Environment Migration:
Production:
- RDS db.r6g.large Multi-AZ, 100GB gp3, 7-day backups
- ElastiCache cache.r6g.large with 2 replicas
- S3 whatsmeow-production-media
- Cost: ~$575/month

Staging:
- RDS db.t4g.medium single-AZ, 20GB, 3-day backups
- ElastiCache cache.t4g.small with 1 replica
- FARGATE_SPOT enabled
- Cost: ~$126/month

Homolog:
- RDS db.t4g.small single-AZ, 10GB, 1-day backups
- ElastiCache cache.t4g.small no replicas
- FARGATE_SPOT, NAT Gateway disabled
- Cost: ~$84/month

Documentation:
- Updated architecture diagram with managed services
- New cost breakdown per environment
- Updated troubleshooting for RDS/ElastiCache/S3
- Removed EFS and container-based service documentation

Benefits:
- Automated backups and point-in-time recovery
- Managed patching and maintenance
- Better scalability with auto-scaling support
- Enhanced security with encryption at rest/transit
- Reduced operational complexity

### ✨ Features

* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-10-09)

### ⚠ BREAKING CHANGES

* **terraform:** Replace containerized data stores with AWS managed
services (RDS PostgreSQL, ElastiCache Redis, S3) for improved
reliability, scalability, and operational efficiency.

New Terraform Modules:
- RDS PostgreSQL with Multi-AZ, automated backups, encryption
- ElastiCache Redis with replication and automatic failover
- S3 buckets with versioning, encryption, lifecycle policies

Module Refactoring:
- ECS Service: Simplified to API-only container, removed Postgres/Redis/MinIO
- Security Groups: Added RDS and ElastiCache SGs, removed EFS
- Secrets Manager: Flexible payload structure per environment

Environment Migration:
Production:
- RDS db.r6g.large Multi-AZ, 100GB gp3, 7-day backups
- ElastiCache cache.r6g.large with 2 replicas
- S3 whatsmeow-production-media
- Cost: ~$575/month

Staging:
- RDS db.t4g.medium single-AZ, 20GB, 3-day backups
- ElastiCache cache.t4g.small with 1 replica
- FARGATE_SPOT enabled
- Cost: ~$126/month

Homolog:
- RDS db.t4g.small single-AZ, 10GB, 1-day backups
- ElastiCache cache.t4g.small no replicas
- FARGATE_SPOT, NAT Gateway disabled
- Cost: ~$84/month

Documentation:
- Updated architecture diagram with managed services
- New cost breakdown per environment
- Updated troubleshooting for RDS/ElastiCache/S3
- Removed EFS and container-based service documentation

Benefits:
- Automated backups and point-in-time recovery
- Managed patching and maintenance
- Better scalability with auto-scaling support
- Enhanced security with encryption at rest/transit
- Reduced operational complexity

### ✨ Features

* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))

## [1.2.0-develop.3](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.2...v1.2.0-develop.3) (2025-10-06)

### ✨ Features

* **config:** add media cleanup and S3 URL configuration options ([858e789](https://github.com/Funnelchat20/whatsapp-api-golang/commit/858e78955e1484f94d195f66368966767e6e3cad))
* **media:** add presigned URL toggle and public URL support for S3 ([e9900ee](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e9900eeecffa399cf26bd8d84527e805051ef902))
* **events:** add undecryptable message support and contact metadata enrichment ([dce506d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dce506d2a18832f175a77f702f201188f9a2888a))
* **whatsmeow:** enhance contact metadata with photo details and presence system ([602f920](https://github.com/Funnelchat20/whatsapp-api-golang/commit/602f920507cd0783de6e95f91619e41366472148))
* **events:** enhance ZAPI transformer with complete webhook payload support ([c495b3e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c495b3e89b7f81fdd31f4337e0472cb7dbb8b2e7))
* **media:** implement automated media cleanup system with distributed locking ([8a88244](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8a88244431a792ccfa2a182a2be65a9fc0500a73))
* **whatsmeow:** implement LID to phone number resolution system ([ca50b99](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ca50b99679091e541fcbc9f745ca9667ed8cca02))

## [1.2.0-develop.2](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.1...v1.2.0-develop.2) (2025-10-04)

### ✨ Features

* **events:** enrich webhook payloads with contact metadata ([7d043d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7d043d68450669924ccc823af8c19829d458ef61))

### ♻️ Code Refactoring

* **events:** persist internal events using JSON ([f9f9e6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f9f9e6cad3e1a46b6928356feb4fe1dae3de7fea))

## [1.2.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.1.0...v1.2.0-develop.1) (2025-10-03)

### ✨ Features

* **observability:** add comprehensive metrics and async context helpers ([680c8a6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/680c8a602a528f7f14adf069121cab354245eb1d))
* **db:** add event outbox, DLQ, and media metadata schema ([321fa6b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/321fa6baa15eabfc4bc57dbeabc95f82da7b285e))
* **config:** add event system and media configuration ([d52469f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d52469fd2d2989847fbfd30dbf8e512fd04f81b3))
* **handlers:** add media HTTP handler for local file serving ([42300f3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42300f3ebb30dcf382e9968e6a745b3e1558cdff))
* **events:** implement comprehensive event processing system ([c84d840](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c84d8408400bcb71a510687b2d8ccd8101012470))
* **integration:** wire event system into WhatsApp client lifecycle ([98f787c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/98f787c038da4058bb73a42460f40ce32746b031))

### ♻️ Code Refactoring

* **handlers:** clean up import aliases for consistency ([e443649](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e4436490fef498028846471a91c75ea581423db4))
* **locks:** enhance circuit breaker metrics and tracking ([76a7ba6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/76a7ba614799dd117f36ab2412610dd3c88572f1))
* **integration:** finalize event system wiring and interfaces ([5831173](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5831173ce3f5a705ace8bb2b0907a8e53ed35a7c))

### 📝 Documentation

* update code standards and add comprehensive development plan ([e59b86b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e59b86b8c2e0120e49c31ab129e96696b233c2e7))

## [1.1.0](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.0.5...v1.1.0) (2025-10-02)

### ✨ Features

* add docker dev workflow ([8c6249c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8c6249c31640d983bde0c7e3aee185a2dd785b0e))
* harden registry lifecycle and health checks ([008b1c8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/008b1c8dd2f07a4738c919f3c794616e3dd55266))

## [1.0.5](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.0.4...v1.0.5) (2025-10-01)

### 🐛 Bug Fixes

* test only API code in Docker build, not entire whatsmeow library ([b844208](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b8442081c1345feb24e8c1497453ba60f2fda60d))

## [1.0.4](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.0.3...v1.0.4) (2025-10-01)

### 🐛 Bug Fixes

* correct Docker Hub login parameter ([4f6348d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f6348d6258a774917d8dc2de928edb34861f8a5))

## [1.0.3](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.0.2...v1.0.3) (2025-10-01)

### 🐛 Bug Fixes

* resolve CI workflow failures ([ec21016](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ec21016c0530cdec54d0caa207e73f8604a23dc3))

## [1.0.2](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.0.1...v1.0.2) (2025-10-01)

### 🐛 Bug Fixes

* remove SARIF upload from security scan workflow ([4d62fe6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4d62fe648d52f41caf224b9159c339521a1b8d3e))

## [1.0.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.0.0...v1.0.1) (2025-10-01)

### 🐛 Bug Fixes

* format Go code and normalize JSON file line endings ([1b245e4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1b245e4499082733a76b679a809289732bea82d8))

## 1.0.0 (2025-10-01)

### ✨ Features

* add auto duplicate removal ([e06479e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e06479e01f0e747de2e5720cfa8391c82b4054e5))
* add REST API layer with Z-API compatibility ([6a1fb66](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6a1fb661ff97c92b07be702c585eac5942593d33))
* add robust support for hosted WhatsApp accounts ([34e9981](https://github.com/Funnelchat20/whatsapp-api-golang/commit/34e998186db9219ae53250fd238f0c9700215053))
* add support for additional nodes in SendRequestExtra ([1a14727](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1a147277a078bd92d483b1882b3e3faf559206a3))
* write proto to file ([18e75da](https://github.com/Funnelchat20/whatsapp-api-golang/commit/18e75da1156d5003858475a6be1c59a437642315))

### 🐛 Bug Fixes

* check if dialect is pgx for postgresql ([efd0d77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/efd0d7795c7953f64355d08b8241dc6753ac08a2))
* out of range when marking more than one message as read ([ed20d21](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ed20d21ffa816df410c33d7a609edb8e92b6edfc))
* panic when link/unlink community ([#437](https://github.com/Funnelchat20/whatsapp-api-golang/issues/437)) ([c313a80](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c313a80ab292f8ca701d7e68e0a21bfe047676b3))
* reactionMessage ([d51dc6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d51dc6efb0e9eaa01bbf4a3dc8d576de6c8d175f))

### 📝 Documentation

* add comprehensive CI/CD and deployment documentation ([e6af6a1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6af6a1623fe97550026fc327ef5ca01e7daa9f7))
