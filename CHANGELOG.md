# Changelog

All notable changes to this project will be documented in this file.

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2026-01-14)

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

* **echo:** add API echo emitter for webhook notifications ([7fb6265](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7fb62650d04860e5567111e9dcb8554923707d62))
* **events:** add API echo support in event pipeline ([54adef8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54adef80a620ab958cf3cac764a34aade95c21c7))
* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **config:** add APIEcho configuration section ([632c4a8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/632c4a8781eab6bb2398228daae656b1ad1ad90e))
* **whatsmeow:** add automatic call rejection with ZuckZapGo pattern ([539b1fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/539b1fc579f9fe6f2b51fcb0fdd92491ee452877))
* **carousel:** add body text at root InteractiveMessage level ([9e44475](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e444751ee434731f42b751b7e14472d8b62d5ff))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* **api:** add config endpoint to expose WHATSAPP_CLIENT_TOKEN ([859d34e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/859d34ebaed57d60570a2981c07abde2a183a015))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **queue:** add custom link preview override support ([d792207](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d792207ab3f8ab4a9f352820e1a1c7c6f8a9feb3))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* **handlers:** add HTTP endpoints for new message types ([e9a9c42](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e9a9c427df57ea95fd0e03fb1a14c25508ed450f))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **components:** add instance management UI components ([7f4d47a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f4d47a130e592b03e290c2bc0e666f927befcb7))
* **schemas:** add isBusiness field to ConnectedCallback ([580d3d5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/580d3d503d5a610274fbd0bffc99a88c4f102d3d))
* **dispatch:** add isBusiness flag to event transformer ([fcc8a2e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/fcc8a2ed72d5a9139c9834a44961cce56a1938d0))
* **messages:** add media support to send-button-actions endpoint ([00e1c96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/00e1c96f6bc0f12300b5ef8b2de2346007e1a38c))
* **api:** add message management endpoints ([aadf080](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aadf080a1c356d5294a4a011b56aab87d01648cc))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **api:** add privacy settings handler ([049ea20](https://github.com/Funnelchat20/whatsapp-api-golang/commit/049ea2031e825850ffb14729de1bca44cd6b795d))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **api:** add proxy routes for send-text, queue, and message status stats ([64d1baf](https://github.com/Funnelchat20/whatsapp-api-golang/commit/64d1bafa9bcc138c76f67a5a6f5ea4041c947348))
* **queue:** add PTV processor for circular video messages ([50dcdcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/50dcdcb20b1475c77c3dbd3831d1d51077113935))
* **statuscache:** add QueryAll to retrieve all cached entries without filters ([87fada0](https://github.com/Funnelchat20/whatsapp-api-golang/commit/87fada064f73c30ff7a8041ff3e12e06f05518c1))
* **api:** add queue, status-cache, and metrics API clients ([59037fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/59037fc6a45063b83d66e35fab74b423870cb0d1))
* **types:** add queue, status-cache, and metrics type definitions ([80ad331](https://github.com/Funnelchat20/whatsapp-api-golang/commit/80ad3318ea76dbb0ef7cf01175abf8c14ff15130))
* **send:** add review_and_pay button support for native flow messages ([3cbf4c1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3cbf4c105b1fb912f92fc4465b44b798cb3dc093))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **events:** add SourceLibAPI constant for API-originated events ([9f922b1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9f922b154480c03072f0e8c2abfee57d49c8b00c))
* **api:** add status cache core implementation ([6f0d776](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6f0d7761aa96d1975a5a0aa0f8ea2022177c5729))
* **api:** add status cache HTTP handlers and routes ([acf5910](https://github.com/Funnelchat20/whatsapp-api-golang/commit/acf5910023ab465b126c142bd547f8f088e31f60))
* **handlers:** add Status endpoints for WhatsApp Stories ([78d4ed2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/78d4ed29b7c6393a691440adf4df7f1c644c5b91))
* **queue:** add Status message type constants ([e81714a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e81714a6bdcb42ad8c44ad0bc2b96628ce69fd53))
* **api:** add StatusCache configuration ([2d7989c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d7989cd34986971b67929355ca48e9a3a67d54d))
* **manager:** add StatusCache metrics tab component ([d85936d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d85936d9e2e5c82563acd21740f0a46651f54942))
* **manager:** add StatusCache metrics transformer ([8f9be0f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8f9be0f2f3a30d6e44a6e6424834643b46f9c1a9))
* **manager:** add StatusCache metrics type definitions ([023d17f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/023d17f3673dfb5cdb9977e33d815e37a61dbac8))
* **api:** add StatusCache Prometheus metrics ([8b32dcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8b32dcb7fd8b72ea5534ccb19f09aa4d551ae610))
* **queue:** add StatusProcessor for WhatsApp Status messages ([bd08589](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bd08589ce1caa93763dccfb74808de4d51490045))
* **queue:** add sticker processor with WebP conversion ([7d4f94a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7d4f94ac1a3de1fe3cd3d49196cd1171f28fb2c7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **hooks:** add SWR hooks for queue, status-cache, metrics, and client token ([bb7bb5b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb7bb5b73a712e08d208b7aaf94f8921505e06be))
* **pages:** add Test tab and cURL copy to instance detail page ([ef80626](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef80626d1aefbcbc97cd960b59c0320842a48fb1))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* **version:** add version package with dynamic version injection ([bfe83b5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bfe83b54d0df260dd1b2e2992df1d604c0a3e42e))
* **audio:** add waveform visualization for voice notes ([70d8f3e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/70d8f3e46be6a8a70a2dedb6e5700b3e06efc2d3))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* **interactive:** add Z-API compatible message building package ([8a5ade5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8a5ade52d9e6da0baac5a13cfe72c839ae752576))
* **queue:** add Z-API interactive message processors ([fb4ba84](https://github.com/Funnelchat20/whatsapp-api-golang/commit/fb4ba8408355479605852fce42b252617640931d))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **manager:** display API version in sidebar ([5f25ab3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5f25ab36795ce1b55d23b10a943c5e940d4468b9))
* **terraform:** enable StatusCache in ECS service configuration ([475e7ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/475e7ba842a06586ba9720339914dbe88c360804))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export StatusCacheTab from metrics barrel ([d62ed6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d62ed6e94fbc7539a92fa1ec18d0bf2f9d0db14b))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* **queue:** extend message models for sticker, PTV, and Z-API types ([6584d50](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6584d508fd9dacdff916c93dce5dbb667ca513f1))
* **server:** implement IsBusiness lookup for instance adapter ([dcd1edd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dcd1eddaab0c6b3e989732a044b57a45e2b51b53))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **transformer:** include isBusiness in connected webhook payload ([0a2feda](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0a2fedae4c616ec2765eb7b465f74448e2c25bb8))
* **server:** initialize and wire API echo emitter ([f265e5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f265e5a49c5e1419a1684957d96117375ba88715))
* **api:** initialize and wire status cache system ([f3eeac4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3eeac4459271b74b6f12cf40d8a7fa4014c89e5))
* **queue:** integrate echo emitter into message processors ([89dcd5c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89dcd5c2a72f4fcd35f6ecc801531e92e05dc260))
* **queue:** integrate new processors and add GetClient helper ([2006337](https://github.com/Funnelchat20/whatsapp-api-golang/commit/200633715d16481132fa88c0e7a45cc515210bea))
* **api:** integrate privacy handler into router ([22ed83b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/22ed83bb0feea86d25a4db55c60ea7cd848d7dbb))
* **api:** integrate status cache interceptor with event dispatch ([40632e5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/40632e589538d021d8cbe3ac0e71feb66921b672))
* **manager:** integrate StatusCache tab into metrics dashboard ([ef8c429](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef8c429d071443565a21980e123bd31fceded647))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **manager:** mask sensitive tokens in UI display ([911485b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911485b013a99fbda868c667536599be80d0b4e6))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **queue:** register StatusProcessor in message router ([b3b7e1b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b3b7e1b655d5469253ad1f862a119caf225e2f4b))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **deploy:** add --load flag and correct ECS service name ([9a58192](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9a581928b2e5ea476af9afa447fb0049349721e6))
* **deploy:** add --load flag to docker build command ([6dc8494](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6dc84946c939c1d45b58caf137ab252ebcec0f7c))
* **deploy:** add --load flag to docker build command ([8731d9d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8731d9dcde4d18775e5a5845095d5c158c266fae))
* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **queue:** add missing MessageTypeCarousel case to processor ([a630251](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a630251365f41692d587c803dc8a8c97b804a5d2))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **instances:** allow empty webhook value to clear configuration ([db722f8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/db722f8da1eec870cea6efbd1cb68b61a9f7e775))
* **manager:** always send all webhook fields to enable clearing ([4b0f5e6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4b0f5e62b0927786bee657280f2ebd8e9d0afa4e))
* **webhooks:** change HTTP method from POST to PUT for webhook updates ([5b9392b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5b9392b4ce45e9c1e5dbff6e8908de77d1fc9286))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** correct webhook form field mappings and API integration ([57e9c7c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57e9c7ca039c419959370e840cf3bc4b7be3b36c))
* **groups:** fetch TopicID before updating group description ([e3adce5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e3adce58844120dfe2e23af05e45d06b418c047e))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **dispatch:** handle status cache only events without webhook delivery ([a2224a5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a2224a5599d6c5a9ccad58cca96e40f7b0b4c4d3))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve metrics transformer and tab components ([0d9d220](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0d9d2208218d4c200738fcbf4398d551ce5ee24b))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **groups:** improve ParseGroupID validation and add IsValidDescriptionID ([a0545fb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a0545fb3a7fb7d066512cbe65f85246519a09a19))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **capture:** persist receipt events for status cache when no webhook configured ([45d94df](https://github.com/Funnelchat20/whatsapp-api-golang/commit/45d94df7928fa4a3ad9f425db77b71a5a1ee28fa))
* **manager:** persist webhook clearing to server and improve Clear All button ([c00a024](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c00a02404296ad3771c53beab887da2bec3cc644))
* **manager:** preserve empty strings in webhook schema for clearing ([9758642](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9758642bd21bfc8a3481b5ac0483f3b3d6d3a7bf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **pages:** remove console.log handlers from instances list page ([7ea6439](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7ea6439d015ea8e2f712fd68aac910be3acc2091))
* **manager:** resolve all lint errors and warnings ([c0660bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c0660bbecbd448eb0718c183e5426375696d2235))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** resolve password reset flow in AWS ALB ([a66b332](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a66b332df39d39c85d1852ed6dae58babbc64ff3))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **webhooks:** route messages to delivery_url and receipts to message_status_url ([0a6eb09](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0a6eb096c9fdec3dca35e8c849ce581014d14655))
* **webhooks:** route messages to receivedAndDeliveryCallbackUrl when notifySentByMe enabled ([dfaac3f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dfaac3fdc80afa76c1825dab95bd49edeeb025d6))
* **whatsmeow:** strip device part from JID when sending rejection message ([cc52493](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cc524930f9c1a9e79381621872290f8404676391))
* **auth:** support multiple trusted origins for Better Auth ([f1b259a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f1b259a4e7291b3bb09d2002291f4f72afc36d55))
* **statuscache:** use case-insensitive comparison in service shouldCacheStatusType ([543433d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/543433dd2549d3998f41bf472e80f50aeffdca0c))
* **statuscache:** use case-insensitive comparison in shouldCacheStatusType ([8668f45](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8668f4561928caabe1b3a12b6c3fd8b97a84f389))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **manager:** extract error from useInstance hook ([27982b1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/27982b12065a2a6483cf2709cb5b0e8a583fed1c))
* **groups:** implement context-aware methods in client adapter ([5285153](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5285153b9d22da9033fb8e0d141a69e0bdd32c00))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **communities:** replace SetGroupDescription with SetGroupTopic ([8d34970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8d34970fb618aabeb49345174f041b94b6d1c371))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))
* **groups:** update Client interface for context-aware methods ([aac162c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aac162cca6964f7dd98b022b1331c613b612dae4))
* **communities:** use SetGroupTopic for description updates ([7f3ae8b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f3ae8b0d10ee1e28b74442d44242dc3412c241b))

### 📝 Documentation

* **examples:** add media examples for button-actions and carousel ([49df2fa](https://github.com/Funnelchat20/whatsapp-api-golang/commit/49df2fa8559b980af17fc7ce175365eab490f571))
* **openapi:** add message field to carousel examples ([8466133](https://github.com/Funnelchat20/whatsapp-api-golang/commit/84661333c6b054a8849b68c14df5dfc29ed2d43d))
* **openapi:** add new endpoint specifications ([f688e9a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f688e9aa834384d1b9ef288d3d1e0facdcd3cd8b))
* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* **api:** add status cache API documentation ([bb293f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb293f56976338dd18ab6bb8bc762eb768a88c15))
* **openapi:** add Status endpoints documentation ([2a5c592](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2a5c592e44c161bb854211512b8ab495452ef9a0))
* **api:** add StatusCache environment variables to .env.example ([c57da24](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c57da24c2a50aad7eaddb0d201dd5a114834ea23))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* **openapi:** clarify StatusCache webhook requirements and usage modes ([5e083f8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e083f83d397443b5e332badc8ed1866c8029396))
* **openapi:** update API documentation for new endpoints ([9e2642d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e2642dd9707e9e2a7402a6a4cac643e44a4a532))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2026-01-13)

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

* **echo:** add API echo emitter for webhook notifications ([7fb6265](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7fb62650d04860e5567111e9dcb8554923707d62))
* **events:** add API echo support in event pipeline ([54adef8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54adef80a620ab958cf3cac764a34aade95c21c7))
* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **config:** add APIEcho configuration section ([632c4a8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/632c4a8781eab6bb2398228daae656b1ad1ad90e))
* **whatsmeow:** add automatic call rejection with ZuckZapGo pattern ([539b1fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/539b1fc579f9fe6f2b51fcb0fdd92491ee452877))
* **carousel:** add body text at root InteractiveMessage level ([9e44475](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e444751ee434731f42b751b7e14472d8b62d5ff))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* **api:** add config endpoint to expose WHATSAPP_CLIENT_TOKEN ([859d34e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/859d34ebaed57d60570a2981c07abde2a183a015))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **queue:** add custom link preview override support ([d792207](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d792207ab3f8ab4a9f352820e1a1c7c6f8a9feb3))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* **handlers:** add HTTP endpoints for new message types ([e9a9c42](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e9a9c427df57ea95fd0e03fb1a14c25508ed450f))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **components:** add instance management UI components ([7f4d47a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f4d47a130e592b03e290c2bc0e666f927befcb7))
* **schemas:** add isBusiness field to ConnectedCallback ([580d3d5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/580d3d503d5a610274fbd0bffc99a88c4f102d3d))
* **dispatch:** add isBusiness flag to event transformer ([fcc8a2e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/fcc8a2ed72d5a9139c9834a44961cce56a1938d0))
* **messages:** add media support to send-button-actions endpoint ([00e1c96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/00e1c96f6bc0f12300b5ef8b2de2346007e1a38c))
* **api:** add message management endpoints ([aadf080](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aadf080a1c356d5294a4a011b56aab87d01648cc))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **api:** add privacy settings handler ([049ea20](https://github.com/Funnelchat20/whatsapp-api-golang/commit/049ea2031e825850ffb14729de1bca44cd6b795d))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **api:** add proxy routes for send-text, queue, and message status stats ([64d1baf](https://github.com/Funnelchat20/whatsapp-api-golang/commit/64d1bafa9bcc138c76f67a5a6f5ea4041c947348))
* **queue:** add PTV processor for circular video messages ([50dcdcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/50dcdcb20b1475c77c3dbd3831d1d51077113935))
* **statuscache:** add QueryAll to retrieve all cached entries without filters ([87fada0](https://github.com/Funnelchat20/whatsapp-api-golang/commit/87fada064f73c30ff7a8041ff3e12e06f05518c1))
* **api:** add queue, status-cache, and metrics API clients ([59037fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/59037fc6a45063b83d66e35fab74b423870cb0d1))
* **types:** add queue, status-cache, and metrics type definitions ([80ad331](https://github.com/Funnelchat20/whatsapp-api-golang/commit/80ad3318ea76dbb0ef7cf01175abf8c14ff15130))
* **send:** add review_and_pay button support for native flow messages ([3cbf4c1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3cbf4c105b1fb912f92fc4465b44b798cb3dc093))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **events:** add SourceLibAPI constant for API-originated events ([9f922b1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9f922b154480c03072f0e8c2abfee57d49c8b00c))
* **api:** add status cache core implementation ([6f0d776](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6f0d7761aa96d1975a5a0aa0f8ea2022177c5729))
* **api:** add status cache HTTP handlers and routes ([acf5910](https://github.com/Funnelchat20/whatsapp-api-golang/commit/acf5910023ab465b126c142bd547f8f088e31f60))
* **handlers:** add Status endpoints for WhatsApp Stories ([78d4ed2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/78d4ed29b7c6393a691440adf4df7f1c644c5b91))
* **queue:** add Status message type constants ([e81714a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e81714a6bdcb42ad8c44ad0bc2b96628ce69fd53))
* **api:** add StatusCache configuration ([2d7989c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d7989cd34986971b67929355ca48e9a3a67d54d))
* **manager:** add StatusCache metrics tab component ([d85936d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d85936d9e2e5c82563acd21740f0a46651f54942))
* **manager:** add StatusCache metrics transformer ([8f9be0f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8f9be0f2f3a30d6e44a6e6424834643b46f9c1a9))
* **manager:** add StatusCache metrics type definitions ([023d17f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/023d17f3673dfb5cdb9977e33d815e37a61dbac8))
* **api:** add StatusCache Prometheus metrics ([8b32dcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8b32dcb7fd8b72ea5534ccb19f09aa4d551ae610))
* **queue:** add StatusProcessor for WhatsApp Status messages ([bd08589](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bd08589ce1caa93763dccfb74808de4d51490045))
* **queue:** add sticker processor with WebP conversion ([7d4f94a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7d4f94ac1a3de1fe3cd3d49196cd1171f28fb2c7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **hooks:** add SWR hooks for queue, status-cache, metrics, and client token ([bb7bb5b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb7bb5b73a712e08d208b7aaf94f8921505e06be))
* **pages:** add Test tab and cURL copy to instance detail page ([ef80626](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef80626d1aefbcbc97cd960b59c0320842a48fb1))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* **audio:** add waveform visualization for voice notes ([70d8f3e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/70d8f3e46be6a8a70a2dedb6e5700b3e06efc2d3))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* **interactive:** add Z-API compatible message building package ([8a5ade5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8a5ade52d9e6da0baac5a13cfe72c839ae752576))
* **queue:** add Z-API interactive message processors ([fb4ba84](https://github.com/Funnelchat20/whatsapp-api-golang/commit/fb4ba8408355479605852fce42b252617640931d))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enable StatusCache in ECS service configuration ([475e7ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/475e7ba842a06586ba9720339914dbe88c360804))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export StatusCacheTab from metrics barrel ([d62ed6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d62ed6e94fbc7539a92fa1ec18d0bf2f9d0db14b))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* **queue:** extend message models for sticker, PTV, and Z-API types ([6584d50](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6584d508fd9dacdff916c93dce5dbb667ca513f1))
* **server:** implement IsBusiness lookup for instance adapter ([dcd1edd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dcd1eddaab0c6b3e989732a044b57a45e2b51b53))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **transformer:** include isBusiness in connected webhook payload ([0a2feda](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0a2fedae4c616ec2765eb7b465f74448e2c25bb8))
* **server:** initialize and wire API echo emitter ([f265e5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f265e5a49c5e1419a1684957d96117375ba88715))
* **api:** initialize and wire status cache system ([f3eeac4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3eeac4459271b74b6f12cf40d8a7fa4014c89e5))
* **queue:** integrate echo emitter into message processors ([89dcd5c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89dcd5c2a72f4fcd35f6ecc801531e92e05dc260))
* **queue:** integrate new processors and add GetClient helper ([2006337](https://github.com/Funnelchat20/whatsapp-api-golang/commit/200633715d16481132fa88c0e7a45cc515210bea))
* **api:** integrate privacy handler into router ([22ed83b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/22ed83bb0feea86d25a4db55c60ea7cd848d7dbb))
* **api:** integrate status cache interceptor with event dispatch ([40632e5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/40632e589538d021d8cbe3ac0e71feb66921b672))
* **manager:** integrate StatusCache tab into metrics dashboard ([ef8c429](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef8c429d071443565a21980e123bd31fceded647))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **manager:** mask sensitive tokens in UI display ([911485b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911485b013a99fbda868c667536599be80d0b4e6))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **queue:** register StatusProcessor in message router ([b3b7e1b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b3b7e1b655d5469253ad1f862a119caf225e2f4b))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **deploy:** add --load flag and correct ECS service name ([9a58192](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9a581928b2e5ea476af9afa447fb0049349721e6))
* **deploy:** add --load flag to docker build command ([6dc8494](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6dc84946c939c1d45b58caf137ab252ebcec0f7c))
* **deploy:** add --load flag to docker build command ([8731d9d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8731d9dcde4d18775e5a5845095d5c158c266fae))
* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **queue:** add missing MessageTypeCarousel case to processor ([a630251](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a630251365f41692d587c803dc8a8c97b804a5d2))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **instances:** allow empty webhook value to clear configuration ([db722f8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/db722f8da1eec870cea6efbd1cb68b61a9f7e775))
* **manager:** always send all webhook fields to enable clearing ([4b0f5e6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4b0f5e62b0927786bee657280f2ebd8e9d0afa4e))
* **webhooks:** change HTTP method from POST to PUT for webhook updates ([5b9392b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5b9392b4ce45e9c1e5dbff6e8908de77d1fc9286))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **groups:** fetch TopicID before updating group description ([e3adce5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e3adce58844120dfe2e23af05e45d06b418c047e))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **dispatch:** handle status cache only events without webhook delivery ([a2224a5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a2224a5599d6c5a9ccad58cca96e40f7b0b4c4d3))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve metrics transformer and tab components ([0d9d220](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0d9d2208218d4c200738fcbf4398d551ce5ee24b))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **groups:** improve ParseGroupID validation and add IsValidDescriptionID ([a0545fb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a0545fb3a7fb7d066512cbe65f85246519a09a19))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **capture:** persist receipt events for status cache when no webhook configured ([45d94df](https://github.com/Funnelchat20/whatsapp-api-golang/commit/45d94df7928fa4a3ad9f425db77b71a5a1ee28fa))
* **manager:** persist webhook clearing to server and improve Clear All button ([c00a024](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c00a02404296ad3771c53beab887da2bec3cc644))
* **manager:** preserve empty strings in webhook schema for clearing ([9758642](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9758642bd21bfc8a3481b5ac0483f3b3d6d3a7bf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **pages:** remove console.log handlers from instances list page ([7ea6439](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7ea6439d015ea8e2f712fd68aac910be3acc2091))
* **manager:** resolve all lint errors and warnings ([c0660bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c0660bbecbd448eb0718c183e5426375696d2235))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** resolve password reset flow in AWS ALB ([a66b332](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a66b332df39d39c85d1852ed6dae58babbc64ff3))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **webhooks:** route messages to receivedAndDeliveryCallbackUrl when notifySentByMe enabled ([dfaac3f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dfaac3fdc80afa76c1825dab95bd49edeeb025d6))
* **whatsmeow:** strip device part from JID when sending rejection message ([cc52493](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cc524930f9c1a9e79381621872290f8404676391))
* **auth:** support multiple trusted origins for Better Auth ([f1b259a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f1b259a4e7291b3bb09d2002291f4f72afc36d55))
* **statuscache:** use case-insensitive comparison in service shouldCacheStatusType ([543433d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/543433dd2549d3998f41bf472e80f50aeffdca0c))
* **statuscache:** use case-insensitive comparison in shouldCacheStatusType ([8668f45](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8668f4561928caabe1b3a12b6c3fd8b97a84f389))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **manager:** extract error from useInstance hook ([27982b1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/27982b12065a2a6483cf2709cb5b0e8a583fed1c))
* **groups:** implement context-aware methods in client adapter ([5285153](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5285153b9d22da9033fb8e0d141a69e0bdd32c00))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **communities:** replace SetGroupDescription with SetGroupTopic ([8d34970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8d34970fb618aabeb49345174f041b94b6d1c371))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))
* **groups:** update Client interface for context-aware methods ([aac162c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aac162cca6964f7dd98b022b1331c613b612dae4))
* **communities:** use SetGroupTopic for description updates ([7f3ae8b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f3ae8b0d10ee1e28b74442d44242dc3412c241b))

### 📝 Documentation

* **examples:** add media examples for button-actions and carousel ([49df2fa](https://github.com/Funnelchat20/whatsapp-api-golang/commit/49df2fa8559b980af17fc7ce175365eab490f571))
* **openapi:** add message field to carousel examples ([8466133](https://github.com/Funnelchat20/whatsapp-api-golang/commit/84661333c6b054a8849b68c14df5dfc29ed2d43d))
* **openapi:** add new endpoint specifications ([f688e9a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f688e9aa834384d1b9ef288d3d1e0facdcd3cd8b))
* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* **api:** add status cache API documentation ([bb293f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb293f56976338dd18ab6bb8bc762eb768a88c15))
* **openapi:** add Status endpoints documentation ([2a5c592](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2a5c592e44c161bb854211512b8ab495452ef9a0))
* **api:** add StatusCache environment variables to .env.example ([c57da24](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c57da24c2a50aad7eaddb0d201dd5a114834ea23))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* **openapi:** clarify StatusCache webhook requirements and usage modes ([5e083f8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e083f83d397443b5e332badc8ed1866c8029396))
* **openapi:** update API documentation for new endpoints ([9e2642d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e2642dd9707e9e2a7402a6a4cac643e44a4a532))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2026-01-10)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **whatsmeow:** add automatic call rejection with ZuckZapGo pattern ([539b1fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/539b1fc579f9fe6f2b51fcb0fdd92491ee452877))
* **carousel:** add body text at root InteractiveMessage level ([9e44475](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e444751ee434731f42b751b7e14472d8b62d5ff))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* **api:** add config endpoint to expose WHATSAPP_CLIENT_TOKEN ([859d34e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/859d34ebaed57d60570a2981c07abde2a183a015))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **queue:** add custom link preview override support ([d792207](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d792207ab3f8ab4a9f352820e1a1c7c6f8a9feb3))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* **handlers:** add HTTP endpoints for new message types ([e9a9c42](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e9a9c427df57ea95fd0e03fb1a14c25508ed450f))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **components:** add instance management UI components ([7f4d47a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f4d47a130e592b03e290c2bc0e666f927befcb7))
* **messages:** add media support to send-button-actions endpoint ([00e1c96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/00e1c96f6bc0f12300b5ef8b2de2346007e1a38c))
* **api:** add message management endpoints ([aadf080](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aadf080a1c356d5294a4a011b56aab87d01648cc))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **api:** add privacy settings handler ([049ea20](https://github.com/Funnelchat20/whatsapp-api-golang/commit/049ea2031e825850ffb14729de1bca44cd6b795d))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **api:** add proxy routes for send-text, queue, and message status stats ([64d1baf](https://github.com/Funnelchat20/whatsapp-api-golang/commit/64d1bafa9bcc138c76f67a5a6f5ea4041c947348))
* **queue:** add PTV processor for circular video messages ([50dcdcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/50dcdcb20b1475c77c3dbd3831d1d51077113935))
* **api:** add queue, status-cache, and metrics API clients ([59037fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/59037fc6a45063b83d66e35fab74b423870cb0d1))
* **types:** add queue, status-cache, and metrics type definitions ([80ad331](https://github.com/Funnelchat20/whatsapp-api-golang/commit/80ad3318ea76dbb0ef7cf01175abf8c14ff15130))
* **send:** add review_and_pay button support for native flow messages ([3cbf4c1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3cbf4c105b1fb912f92fc4465b44b798cb3dc093))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **api:** add status cache core implementation ([6f0d776](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6f0d7761aa96d1975a5a0aa0f8ea2022177c5729))
* **api:** add status cache HTTP handlers and routes ([acf5910](https://github.com/Funnelchat20/whatsapp-api-golang/commit/acf5910023ab465b126c142bd547f8f088e31f60))
* **handlers:** add Status endpoints for WhatsApp Stories ([78d4ed2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/78d4ed29b7c6393a691440adf4df7f1c644c5b91))
* **queue:** add Status message type constants ([e81714a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e81714a6bdcb42ad8c44ad0bc2b96628ce69fd53))
* **api:** add StatusCache configuration ([2d7989c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d7989cd34986971b67929355ca48e9a3a67d54d))
* **manager:** add StatusCache metrics tab component ([d85936d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d85936d9e2e5c82563acd21740f0a46651f54942))
* **manager:** add StatusCache metrics transformer ([8f9be0f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8f9be0f2f3a30d6e44a6e6424834643b46f9c1a9))
* **manager:** add StatusCache metrics type definitions ([023d17f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/023d17f3673dfb5cdb9977e33d815e37a61dbac8))
* **api:** add StatusCache Prometheus metrics ([8b32dcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8b32dcb7fd8b72ea5534ccb19f09aa4d551ae610))
* **queue:** add StatusProcessor for WhatsApp Status messages ([bd08589](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bd08589ce1caa93763dccfb74808de4d51490045))
* **queue:** add sticker processor with WebP conversion ([7d4f94a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7d4f94ac1a3de1fe3cd3d49196cd1171f28fb2c7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **hooks:** add SWR hooks for queue, status-cache, metrics, and client token ([bb7bb5b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb7bb5b73a712e08d208b7aaf94f8921505e06be))
* **pages:** add Test tab and cURL copy to instance detail page ([ef80626](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef80626d1aefbcbc97cd960b59c0320842a48fb1))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* **audio:** add waveform visualization for voice notes ([70d8f3e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/70d8f3e46be6a8a70a2dedb6e5700b3e06efc2d3))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* **interactive:** add Z-API compatible message building package ([8a5ade5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8a5ade52d9e6da0baac5a13cfe72c839ae752576))
* **queue:** add Z-API interactive message processors ([fb4ba84](https://github.com/Funnelchat20/whatsapp-api-golang/commit/fb4ba8408355479605852fce42b252617640931d))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enable StatusCache in ECS service configuration ([475e7ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/475e7ba842a06586ba9720339914dbe88c360804))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export StatusCacheTab from metrics barrel ([d62ed6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d62ed6e94fbc7539a92fa1ec18d0bf2f9d0db14b))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* **queue:** extend message models for sticker, PTV, and Z-API types ([6584d50](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6584d508fd9dacdff916c93dce5dbb667ca513f1))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **api:** initialize and wire status cache system ([f3eeac4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3eeac4459271b74b6f12cf40d8a7fa4014c89e5))
* **queue:** integrate new processors and add GetClient helper ([2006337](https://github.com/Funnelchat20/whatsapp-api-golang/commit/200633715d16481132fa88c0e7a45cc515210bea))
* **api:** integrate privacy handler into router ([22ed83b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/22ed83bb0feea86d25a4db55c60ea7cd848d7dbb))
* **api:** integrate status cache interceptor with event dispatch ([40632e5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/40632e589538d021d8cbe3ac0e71feb66921b672))
* **manager:** integrate StatusCache tab into metrics dashboard ([ef8c429](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef8c429d071443565a21980e123bd31fceded647))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **manager:** mask sensitive tokens in UI display ([911485b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911485b013a99fbda868c667536599be80d0b4e6))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **queue:** register StatusProcessor in message router ([b3b7e1b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b3b7e1b655d5469253ad1f862a119caf225e2f4b))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **deploy:** add --load flag and correct ECS service name ([9a58192](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9a581928b2e5ea476af9afa447fb0049349721e6))
* **deploy:** add --load flag to docker build command ([8731d9d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8731d9dcde4d18775e5a5845095d5c158c266fae))
* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **queue:** add missing MessageTypeCarousel case to processor ([a630251](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a630251365f41692d587c803dc8a8c97b804a5d2))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **webhooks:** change HTTP method from POST to PUT for webhook updates ([5b9392b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5b9392b4ce45e9c1e5dbff6e8908de77d1fc9286))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve metrics transformer and tab components ([0d9d220](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0d9d2208218d4c200738fcbf4398d551ce5ee24b))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **pages:** remove console.log handlers from instances list page ([7ea6439](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7ea6439d015ea8e2f712fd68aac910be3acc2091))
* **manager:** resolve all lint errors and warnings ([c0660bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c0660bbecbd448eb0718c183e5426375696d2235))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** resolve password reset flow in AWS ALB ([a66b332](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a66b332df39d39c85d1852ed6dae58babbc64ff3))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **whatsmeow:** strip device part from JID when sending rejection message ([cc52493](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cc524930f9c1a9e79381621872290f8404676391))
* **auth:** support multiple trusted origins for Better Auth ([f1b259a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f1b259a4e7291b3bb09d2002291f4f72afc36d55))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* **examples:** add media examples for button-actions and carousel ([49df2fa](https://github.com/Funnelchat20/whatsapp-api-golang/commit/49df2fa8559b980af17fc7ce175365eab490f571))
* **openapi:** add message field to carousel examples ([8466133](https://github.com/Funnelchat20/whatsapp-api-golang/commit/84661333c6b054a8849b68c14df5dfc29ed2d43d))
* **openapi:** add new endpoint specifications ([f688e9a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f688e9aa834384d1b9ef288d3d1e0facdcd3cd8b))
* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* **api:** add status cache API documentation ([bb293f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb293f56976338dd18ab6bb8bc762eb768a88c15))
* **openapi:** add Status endpoints documentation ([2a5c592](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2a5c592e44c161bb854211512b8ab495452ef9a0))
* **api:** add StatusCache environment variables to .env.example ([c57da24](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c57da24c2a50aad7eaddb0d201dd5a114834ea23))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* **openapi:** update API documentation for new endpoints ([9e2642d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e2642dd9707e9e2a7402a6a4cac643e44a4a532))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2026-01-09)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **whatsmeow:** add automatic call rejection with ZuckZapGo pattern ([539b1fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/539b1fc579f9fe6f2b51fcb0fdd92491ee452877))
* **carousel:** add body text at root InteractiveMessage level ([9e44475](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e444751ee434731f42b751b7e14472d8b62d5ff))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* **api:** add config endpoint to expose WHATSAPP_CLIENT_TOKEN ([859d34e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/859d34ebaed57d60570a2981c07abde2a183a015))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **queue:** add custom link preview override support ([d792207](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d792207ab3f8ab4a9f352820e1a1c7c6f8a9feb3))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* **handlers:** add HTTP endpoints for new message types ([e9a9c42](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e9a9c427df57ea95fd0e03fb1a14c25508ed450f))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **components:** add instance management UI components ([7f4d47a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f4d47a130e592b03e290c2bc0e666f927befcb7))
* **messages:** add media support to send-button-actions endpoint ([00e1c96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/00e1c96f6bc0f12300b5ef8b2de2346007e1a38c))
* **api:** add message management endpoints ([aadf080](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aadf080a1c356d5294a4a011b56aab87d01648cc))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **api:** add privacy settings handler ([049ea20](https://github.com/Funnelchat20/whatsapp-api-golang/commit/049ea2031e825850ffb14729de1bca44cd6b795d))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **api:** add proxy routes for send-text, queue, and message status stats ([64d1baf](https://github.com/Funnelchat20/whatsapp-api-golang/commit/64d1bafa9bcc138c76f67a5a6f5ea4041c947348))
* **queue:** add PTV processor for circular video messages ([50dcdcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/50dcdcb20b1475c77c3dbd3831d1d51077113935))
* **api:** add queue, status-cache, and metrics API clients ([59037fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/59037fc6a45063b83d66e35fab74b423870cb0d1))
* **types:** add queue, status-cache, and metrics type definitions ([80ad331](https://github.com/Funnelchat20/whatsapp-api-golang/commit/80ad3318ea76dbb0ef7cf01175abf8c14ff15130))
* **send:** add review_and_pay button support for native flow messages ([3cbf4c1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3cbf4c105b1fb912f92fc4465b44b798cb3dc093))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **api:** add status cache core implementation ([6f0d776](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6f0d7761aa96d1975a5a0aa0f8ea2022177c5729))
* **api:** add status cache HTTP handlers and routes ([acf5910](https://github.com/Funnelchat20/whatsapp-api-golang/commit/acf5910023ab465b126c142bd547f8f088e31f60))
* **handlers:** add Status endpoints for WhatsApp Stories ([78d4ed2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/78d4ed29b7c6393a691440adf4df7f1c644c5b91))
* **queue:** add Status message type constants ([e81714a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e81714a6bdcb42ad8c44ad0bc2b96628ce69fd53))
* **api:** add StatusCache configuration ([2d7989c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d7989cd34986971b67929355ca48e9a3a67d54d))
* **manager:** add StatusCache metrics tab component ([d85936d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d85936d9e2e5c82563acd21740f0a46651f54942))
* **manager:** add StatusCache metrics transformer ([8f9be0f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8f9be0f2f3a30d6e44a6e6424834643b46f9c1a9))
* **manager:** add StatusCache metrics type definitions ([023d17f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/023d17f3673dfb5cdb9977e33d815e37a61dbac8))
* **api:** add StatusCache Prometheus metrics ([8b32dcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8b32dcb7fd8b72ea5534ccb19f09aa4d551ae610))
* **queue:** add StatusProcessor for WhatsApp Status messages ([bd08589](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bd08589ce1caa93763dccfb74808de4d51490045))
* **queue:** add sticker processor with WebP conversion ([7d4f94a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7d4f94ac1a3de1fe3cd3d49196cd1171f28fb2c7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **hooks:** add SWR hooks for queue, status-cache, metrics, and client token ([bb7bb5b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb7bb5b73a712e08d208b7aaf94f8921505e06be))
* **pages:** add Test tab and cURL copy to instance detail page ([ef80626](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef80626d1aefbcbc97cd960b59c0320842a48fb1))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* **audio:** add waveform visualization for voice notes ([70d8f3e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/70d8f3e46be6a8a70a2dedb6e5700b3e06efc2d3))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* **interactive:** add Z-API compatible message building package ([8a5ade5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8a5ade52d9e6da0baac5a13cfe72c839ae752576))
* **queue:** add Z-API interactive message processors ([fb4ba84](https://github.com/Funnelchat20/whatsapp-api-golang/commit/fb4ba8408355479605852fce42b252617640931d))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enable StatusCache in ECS service configuration ([475e7ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/475e7ba842a06586ba9720339914dbe88c360804))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export StatusCacheTab from metrics barrel ([d62ed6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d62ed6e94fbc7539a92fa1ec18d0bf2f9d0db14b))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* **queue:** extend message models for sticker, PTV, and Z-API types ([6584d50](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6584d508fd9dacdff916c93dce5dbb667ca513f1))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **api:** initialize and wire status cache system ([f3eeac4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3eeac4459271b74b6f12cf40d8a7fa4014c89e5))
* **queue:** integrate new processors and add GetClient helper ([2006337](https://github.com/Funnelchat20/whatsapp-api-golang/commit/200633715d16481132fa88c0e7a45cc515210bea))
* **api:** integrate privacy handler into router ([22ed83b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/22ed83bb0feea86d25a4db55c60ea7cd848d7dbb))
* **api:** integrate status cache interceptor with event dispatch ([40632e5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/40632e589538d021d8cbe3ac0e71feb66921b672))
* **manager:** integrate StatusCache tab into metrics dashboard ([ef8c429](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef8c429d071443565a21980e123bd31fceded647))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **manager:** mask sensitive tokens in UI display ([911485b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911485b013a99fbda868c667536599be80d0b4e6))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **queue:** register StatusProcessor in message router ([b3b7e1b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b3b7e1b655d5469253ad1f862a119caf225e2f4b))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **deploy:** add --load flag and correct ECS service name ([9a58192](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9a581928b2e5ea476af9afa447fb0049349721e6))
* **deploy:** add --load flag to docker build command ([6dc8494](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6dc84946c939c1d45b58caf137ab252ebcec0f7c))
* **deploy:** add --load flag to docker build command ([8731d9d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8731d9dcde4d18775e5a5845095d5c158c266fae))
* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **queue:** add missing MessageTypeCarousel case to processor ([a630251](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a630251365f41692d587c803dc8a8c97b804a5d2))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **webhooks:** change HTTP method from POST to PUT for webhook updates ([5b9392b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5b9392b4ce45e9c1e5dbff6e8908de77d1fc9286))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve metrics transformer and tab components ([0d9d220](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0d9d2208218d4c200738fcbf4398d551ce5ee24b))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **pages:** remove console.log handlers from instances list page ([7ea6439](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7ea6439d015ea8e2f712fd68aac910be3acc2091))
* **manager:** resolve all lint errors and warnings ([c0660bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c0660bbecbd448eb0718c183e5426375696d2235))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** resolve password reset flow in AWS ALB ([a66b332](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a66b332df39d39c85d1852ed6dae58babbc64ff3))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **auth:** support multiple trusted origins for Better Auth ([f1b259a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f1b259a4e7291b3bb09d2002291f4f72afc36d55))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* **examples:** add media examples for button-actions and carousel ([49df2fa](https://github.com/Funnelchat20/whatsapp-api-golang/commit/49df2fa8559b980af17fc7ce175365eab490f571))
* **openapi:** add message field to carousel examples ([8466133](https://github.com/Funnelchat20/whatsapp-api-golang/commit/84661333c6b054a8849b68c14df5dfc29ed2d43d))
* **openapi:** add new endpoint specifications ([f688e9a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f688e9aa834384d1b9ef288d3d1e0facdcd3cd8b))
* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* **api:** add status cache API documentation ([bb293f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb293f56976338dd18ab6bb8bc762eb768a88c15))
* **openapi:** add Status endpoints documentation ([2a5c592](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2a5c592e44c161bb854211512b8ab495452ef9a0))
* **api:** add StatusCache environment variables to .env.example ([c57da24](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c57da24c2a50aad7eaddb0d201dd5a114834ea23))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* **openapi:** update API documentation for new endpoints ([9e2642d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e2642dd9707e9e2a7402a6a4cac643e44a4a532))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2026-01-09)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **carousel:** add body text at root InteractiveMessage level ([9e44475](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e444751ee434731f42b751b7e14472d8b62d5ff))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* **api:** add config endpoint to expose WHATSAPP_CLIENT_TOKEN ([859d34e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/859d34ebaed57d60570a2981c07abde2a183a015))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **queue:** add custom link preview override support ([d792207](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d792207ab3f8ab4a9f352820e1a1c7c6f8a9feb3))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* **handlers:** add HTTP endpoints for new message types ([e9a9c42](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e9a9c427df57ea95fd0e03fb1a14c25508ed450f))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **components:** add instance management UI components ([7f4d47a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f4d47a130e592b03e290c2bc0e666f927befcb7))
* **messages:** add media support to send-button-actions endpoint ([00e1c96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/00e1c96f6bc0f12300b5ef8b2de2346007e1a38c))
* **api:** add message management endpoints ([aadf080](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aadf080a1c356d5294a4a011b56aab87d01648cc))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **api:** add privacy settings handler ([049ea20](https://github.com/Funnelchat20/whatsapp-api-golang/commit/049ea2031e825850ffb14729de1bca44cd6b795d))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **api:** add proxy routes for send-text, queue, and message status stats ([64d1baf](https://github.com/Funnelchat20/whatsapp-api-golang/commit/64d1bafa9bcc138c76f67a5a6f5ea4041c947348))
* **queue:** add PTV processor for circular video messages ([50dcdcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/50dcdcb20b1475c77c3dbd3831d1d51077113935))
* **api:** add queue, status-cache, and metrics API clients ([59037fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/59037fc6a45063b83d66e35fab74b423870cb0d1))
* **types:** add queue, status-cache, and metrics type definitions ([80ad331](https://github.com/Funnelchat20/whatsapp-api-golang/commit/80ad3318ea76dbb0ef7cf01175abf8c14ff15130))
* **send:** add review_and_pay button support for native flow messages ([3cbf4c1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3cbf4c105b1fb912f92fc4465b44b798cb3dc093))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **api:** add status cache core implementation ([6f0d776](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6f0d7761aa96d1975a5a0aa0f8ea2022177c5729))
* **api:** add status cache HTTP handlers and routes ([acf5910](https://github.com/Funnelchat20/whatsapp-api-golang/commit/acf5910023ab465b126c142bd547f8f088e31f60))
* **handlers:** add Status endpoints for WhatsApp Stories ([78d4ed2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/78d4ed29b7c6393a691440adf4df7f1c644c5b91))
* **queue:** add Status message type constants ([e81714a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e81714a6bdcb42ad8c44ad0bc2b96628ce69fd53))
* **api:** add StatusCache configuration ([2d7989c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d7989cd34986971b67929355ca48e9a3a67d54d))
* **manager:** add StatusCache metrics tab component ([d85936d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d85936d9e2e5c82563acd21740f0a46651f54942))
* **manager:** add StatusCache metrics transformer ([8f9be0f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8f9be0f2f3a30d6e44a6e6424834643b46f9c1a9))
* **manager:** add StatusCache metrics type definitions ([023d17f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/023d17f3673dfb5cdb9977e33d815e37a61dbac8))
* **api:** add StatusCache Prometheus metrics ([8b32dcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8b32dcb7fd8b72ea5534ccb19f09aa4d551ae610))
* **queue:** add StatusProcessor for WhatsApp Status messages ([bd08589](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bd08589ce1caa93763dccfb74808de4d51490045))
* **queue:** add sticker processor with WebP conversion ([7d4f94a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7d4f94ac1a3de1fe3cd3d49196cd1171f28fb2c7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **hooks:** add SWR hooks for queue, status-cache, metrics, and client token ([bb7bb5b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb7bb5b73a712e08d208b7aaf94f8921505e06be))
* **pages:** add Test tab and cURL copy to instance detail page ([ef80626](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef80626d1aefbcbc97cd960b59c0320842a48fb1))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* **audio:** add waveform visualization for voice notes ([70d8f3e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/70d8f3e46be6a8a70a2dedb6e5700b3e06efc2d3))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* **interactive:** add Z-API compatible message building package ([8a5ade5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8a5ade52d9e6da0baac5a13cfe72c839ae752576))
* **queue:** add Z-API interactive message processors ([fb4ba84](https://github.com/Funnelchat20/whatsapp-api-golang/commit/fb4ba8408355479605852fce42b252617640931d))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enable StatusCache in ECS service configuration ([475e7ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/475e7ba842a06586ba9720339914dbe88c360804))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export StatusCacheTab from metrics barrel ([d62ed6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d62ed6e94fbc7539a92fa1ec18d0bf2f9d0db14b))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* **queue:** extend message models for sticker, PTV, and Z-API types ([6584d50](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6584d508fd9dacdff916c93dce5dbb667ca513f1))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **api:** initialize and wire status cache system ([f3eeac4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3eeac4459271b74b6f12cf40d8a7fa4014c89e5))
* **queue:** integrate new processors and add GetClient helper ([2006337](https://github.com/Funnelchat20/whatsapp-api-golang/commit/200633715d16481132fa88c0e7a45cc515210bea))
* **api:** integrate privacy handler into router ([22ed83b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/22ed83bb0feea86d25a4db55c60ea7cd848d7dbb))
* **api:** integrate status cache interceptor with event dispatch ([40632e5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/40632e589538d021d8cbe3ac0e71feb66921b672))
* **manager:** integrate StatusCache tab into metrics dashboard ([ef8c429](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef8c429d071443565a21980e123bd31fceded647))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **manager:** mask sensitive tokens in UI display ([911485b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911485b013a99fbda868c667536599be80d0b4e6))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **queue:** register StatusProcessor in message router ([b3b7e1b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b3b7e1b655d5469253ad1f862a119caf225e2f4b))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **deploy:** add --load flag and correct ECS service name ([9a58192](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9a581928b2e5ea476af9afa447fb0049349721e6))
* **deploy:** add --load flag to docker build command ([6dc8494](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6dc84946c939c1d45b58caf137ab252ebcec0f7c))
* **deploy:** add --load flag to docker build command ([8731d9d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8731d9dcde4d18775e5a5845095d5c158c266fae))
* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **queue:** add missing MessageTypeCarousel case to processor ([a630251](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a630251365f41692d587c803dc8a8c97b804a5d2))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **webhooks:** change HTTP method from POST to PUT for webhook updates ([5b9392b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5b9392b4ce45e9c1e5dbff6e8908de77d1fc9286))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve metrics transformer and tab components ([0d9d220](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0d9d2208218d4c200738fcbf4398d551ce5ee24b))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **pages:** remove console.log handlers from instances list page ([7ea6439](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7ea6439d015ea8e2f712fd68aac910be3acc2091))
* **manager:** resolve all lint errors and warnings ([c0660bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c0660bbecbd448eb0718c183e5426375696d2235))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** resolve password reset flow in AWS ALB ([a66b332](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a66b332df39d39c85d1852ed6dae58babbc64ff3))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **auth:** support multiple trusted origins for Better Auth ([f1b259a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f1b259a4e7291b3bb09d2002291f4f72afc36d55))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* **examples:** add media examples for button-actions and carousel ([49df2fa](https://github.com/Funnelchat20/whatsapp-api-golang/commit/49df2fa8559b980af17fc7ce175365eab490f571))
* **openapi:** add message field to carousel examples ([8466133](https://github.com/Funnelchat20/whatsapp-api-golang/commit/84661333c6b054a8849b68c14df5dfc29ed2d43d))
* **openapi:** add new endpoint specifications ([f688e9a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f688e9aa834384d1b9ef288d3d1e0facdcd3cd8b))
* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* **api:** add status cache API documentation ([bb293f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb293f56976338dd18ab6bb8bc762eb768a88c15))
* **openapi:** add Status endpoints documentation ([2a5c592](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2a5c592e44c161bb854211512b8ab495452ef9a0))
* **api:** add StatusCache environment variables to .env.example ([c57da24](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c57da24c2a50aad7eaddb0d201dd5a114834ea23))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* **openapi:** update API documentation for new endpoints ([9e2642d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e2642dd9707e9e2a7402a6a4cac643e44a4a532))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2026-01-09)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **carousel:** add body text at root InteractiveMessage level ([9e44475](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e444751ee434731f42b751b7e14472d8b62d5ff))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* **api:** add config endpoint to expose WHATSAPP_CLIENT_TOKEN ([859d34e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/859d34ebaed57d60570a2981c07abde2a183a015))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **queue:** add custom link preview override support ([d792207](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d792207ab3f8ab4a9f352820e1a1c7c6f8a9feb3))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* **handlers:** add HTTP endpoints for new message types ([e9a9c42](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e9a9c427df57ea95fd0e03fb1a14c25508ed450f))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **components:** add instance management UI components ([7f4d47a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f4d47a130e592b03e290c2bc0e666f927befcb7))
* **messages:** add media support to send-button-actions endpoint ([00e1c96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/00e1c96f6bc0f12300b5ef8b2de2346007e1a38c))
* **api:** add message management endpoints ([aadf080](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aadf080a1c356d5294a4a011b56aab87d01648cc))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **api:** add privacy settings handler ([049ea20](https://github.com/Funnelchat20/whatsapp-api-golang/commit/049ea2031e825850ffb14729de1bca44cd6b795d))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **api:** add proxy routes for send-text, queue, and message status stats ([64d1baf](https://github.com/Funnelchat20/whatsapp-api-golang/commit/64d1bafa9bcc138c76f67a5a6f5ea4041c947348))
* **queue:** add PTV processor for circular video messages ([50dcdcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/50dcdcb20b1475c77c3dbd3831d1d51077113935))
* **api:** add queue, status-cache, and metrics API clients ([59037fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/59037fc6a45063b83d66e35fab74b423870cb0d1))
* **types:** add queue, status-cache, and metrics type definitions ([80ad331](https://github.com/Funnelchat20/whatsapp-api-golang/commit/80ad3318ea76dbb0ef7cf01175abf8c14ff15130))
* **send:** add review_and_pay button support for native flow messages ([3cbf4c1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3cbf4c105b1fb912f92fc4465b44b798cb3dc093))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **api:** add status cache core implementation ([6f0d776](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6f0d7761aa96d1975a5a0aa0f8ea2022177c5729))
* **api:** add status cache HTTP handlers and routes ([acf5910](https://github.com/Funnelchat20/whatsapp-api-golang/commit/acf5910023ab465b126c142bd547f8f088e31f60))
* **api:** add StatusCache configuration ([2d7989c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d7989cd34986971b67929355ca48e9a3a67d54d))
* **manager:** add StatusCache metrics tab component ([d85936d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d85936d9e2e5c82563acd21740f0a46651f54942))
* **manager:** add StatusCache metrics transformer ([8f9be0f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8f9be0f2f3a30d6e44a6e6424834643b46f9c1a9))
* **manager:** add StatusCache metrics type definitions ([023d17f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/023d17f3673dfb5cdb9977e33d815e37a61dbac8))
* **api:** add StatusCache Prometheus metrics ([8b32dcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8b32dcb7fd8b72ea5534ccb19f09aa4d551ae610))
* **queue:** add sticker processor with WebP conversion ([7d4f94a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7d4f94ac1a3de1fe3cd3d49196cd1171f28fb2c7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **hooks:** add SWR hooks for queue, status-cache, metrics, and client token ([bb7bb5b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb7bb5b73a712e08d208b7aaf94f8921505e06be))
* **pages:** add Test tab and cURL copy to instance detail page ([ef80626](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef80626d1aefbcbc97cd960b59c0320842a48fb1))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* **audio:** add waveform visualization for voice notes ([70d8f3e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/70d8f3e46be6a8a70a2dedb6e5700b3e06efc2d3))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* **interactive:** add Z-API compatible message building package ([8a5ade5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8a5ade52d9e6da0baac5a13cfe72c839ae752576))
* **queue:** add Z-API interactive message processors ([fb4ba84](https://github.com/Funnelchat20/whatsapp-api-golang/commit/fb4ba8408355479605852fce42b252617640931d))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enable StatusCache in ECS service configuration ([475e7ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/475e7ba842a06586ba9720339914dbe88c360804))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export StatusCacheTab from metrics barrel ([d62ed6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d62ed6e94fbc7539a92fa1ec18d0bf2f9d0db14b))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* **queue:** extend message models for sticker, PTV, and Z-API types ([6584d50](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6584d508fd9dacdff916c93dce5dbb667ca513f1))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **api:** initialize and wire status cache system ([f3eeac4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3eeac4459271b74b6f12cf40d8a7fa4014c89e5))
* **queue:** integrate new processors and add GetClient helper ([2006337](https://github.com/Funnelchat20/whatsapp-api-golang/commit/200633715d16481132fa88c0e7a45cc515210bea))
* **api:** integrate privacy handler into router ([22ed83b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/22ed83bb0feea86d25a4db55c60ea7cd848d7dbb))
* **api:** integrate status cache interceptor with event dispatch ([40632e5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/40632e589538d021d8cbe3ac0e71feb66921b672))
* **manager:** integrate StatusCache tab into metrics dashboard ([ef8c429](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef8c429d071443565a21980e123bd31fceded647))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **manager:** mask sensitive tokens in UI display ([911485b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911485b013a99fbda868c667536599be80d0b4e6))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **deploy:** add --load flag and correct ECS service name ([9a58192](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9a581928b2e5ea476af9afa447fb0049349721e6))
* **deploy:** add --load flag to docker build command ([6dc8494](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6dc84946c939c1d45b58caf137ab252ebcec0f7c))
* **deploy:** add --load flag to docker build command ([8731d9d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8731d9dcde4d18775e5a5845095d5c158c266fae))
* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **queue:** add missing MessageTypeCarousel case to processor ([a630251](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a630251365f41692d587c803dc8a8c97b804a5d2))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **webhooks:** change HTTP method from POST to PUT for webhook updates ([5b9392b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5b9392b4ce45e9c1e5dbff6e8908de77d1fc9286))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve metrics transformer and tab components ([0d9d220](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0d9d2208218d4c200738fcbf4398d551ce5ee24b))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **pages:** remove console.log handlers from instances list page ([7ea6439](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7ea6439d015ea8e2f712fd68aac910be3acc2091))
* **manager:** resolve all lint errors and warnings ([c0660bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c0660bbecbd448eb0718c183e5426375696d2235))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** resolve password reset flow in AWS ALB ([a66b332](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a66b332df39d39c85d1852ed6dae58babbc64ff3))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **auth:** support multiple trusted origins for Better Auth ([f1b259a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f1b259a4e7291b3bb09d2002291f4f72afc36d55))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* **examples:** add media examples for button-actions and carousel ([49df2fa](https://github.com/Funnelchat20/whatsapp-api-golang/commit/49df2fa8559b980af17fc7ce175365eab490f571))
* **openapi:** add message field to carousel examples ([8466133](https://github.com/Funnelchat20/whatsapp-api-golang/commit/84661333c6b054a8849b68c14df5dfc29ed2d43d))
* **openapi:** add new endpoint specifications ([f688e9a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f688e9aa834384d1b9ef288d3d1e0facdcd3cd8b))
* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* **api:** add status cache API documentation ([bb293f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb293f56976338dd18ab6bb8bc762eb768a88c15))
* **api:** add StatusCache environment variables to .env.example ([c57da24](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c57da24c2a50aad7eaddb0d201dd5a114834ea23))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* **openapi:** update API documentation for new endpoints ([9e2642d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e2642dd9707e9e2a7402a6a4cac643e44a4a532))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2026-01-09)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **carousel:** add body text at root InteractiveMessage level ([9e44475](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e444751ee434731f42b751b7e14472d8b62d5ff))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* **api:** add config endpoint to expose WHATSAPP_CLIENT_TOKEN ([859d34e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/859d34ebaed57d60570a2981c07abde2a183a015))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **queue:** add custom link preview override support ([d792207](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d792207ab3f8ab4a9f352820e1a1c7c6f8a9feb3))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* **handlers:** add HTTP endpoints for new message types ([e9a9c42](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e9a9c427df57ea95fd0e03fb1a14c25508ed450f))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **components:** add instance management UI components ([7f4d47a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f4d47a130e592b03e290c2bc0e666f927befcb7))
* **messages:** add media support to send-button-actions endpoint ([00e1c96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/00e1c96f6bc0f12300b5ef8b2de2346007e1a38c))
* **api:** add message management endpoints ([aadf080](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aadf080a1c356d5294a4a011b56aab87d01648cc))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **api:** add privacy settings handler ([049ea20](https://github.com/Funnelchat20/whatsapp-api-golang/commit/049ea2031e825850ffb14729de1bca44cd6b795d))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **api:** add proxy routes for send-text, queue, and message status stats ([64d1baf](https://github.com/Funnelchat20/whatsapp-api-golang/commit/64d1bafa9bcc138c76f67a5a6f5ea4041c947348))
* **queue:** add PTV processor for circular video messages ([50dcdcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/50dcdcb20b1475c77c3dbd3831d1d51077113935))
* **api:** add queue, status-cache, and metrics API clients ([59037fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/59037fc6a45063b83d66e35fab74b423870cb0d1))
* **types:** add queue, status-cache, and metrics type definitions ([80ad331](https://github.com/Funnelchat20/whatsapp-api-golang/commit/80ad3318ea76dbb0ef7cf01175abf8c14ff15130))
* **send:** add review_and_pay button support for native flow messages ([3cbf4c1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3cbf4c105b1fb912f92fc4465b44b798cb3dc093))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **api:** add status cache core implementation ([6f0d776](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6f0d7761aa96d1975a5a0aa0f8ea2022177c5729))
* **api:** add status cache HTTP handlers and routes ([acf5910](https://github.com/Funnelchat20/whatsapp-api-golang/commit/acf5910023ab465b126c142bd547f8f088e31f60))
* **api:** add StatusCache configuration ([2d7989c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d7989cd34986971b67929355ca48e9a3a67d54d))
* **manager:** add StatusCache metrics tab component ([d85936d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d85936d9e2e5c82563acd21740f0a46651f54942))
* **manager:** add StatusCache metrics transformer ([8f9be0f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8f9be0f2f3a30d6e44a6e6424834643b46f9c1a9))
* **manager:** add StatusCache metrics type definitions ([023d17f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/023d17f3673dfb5cdb9977e33d815e37a61dbac8))
* **api:** add StatusCache Prometheus metrics ([8b32dcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8b32dcb7fd8b72ea5534ccb19f09aa4d551ae610))
* **queue:** add sticker processor with WebP conversion ([7d4f94a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7d4f94ac1a3de1fe3cd3d49196cd1171f28fb2c7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **hooks:** add SWR hooks for queue, status-cache, metrics, and client token ([bb7bb5b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb7bb5b73a712e08d208b7aaf94f8921505e06be))
* **pages:** add Test tab and cURL copy to instance detail page ([ef80626](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef80626d1aefbcbc97cd960b59c0320842a48fb1))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* **audio:** add waveform visualization for voice notes ([70d8f3e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/70d8f3e46be6a8a70a2dedb6e5700b3e06efc2d3))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* **interactive:** add Z-API compatible message building package ([8a5ade5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8a5ade52d9e6da0baac5a13cfe72c839ae752576))
* **queue:** add Z-API interactive message processors ([fb4ba84](https://github.com/Funnelchat20/whatsapp-api-golang/commit/fb4ba8408355479605852fce42b252617640931d))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enable StatusCache in ECS service configuration ([475e7ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/475e7ba842a06586ba9720339914dbe88c360804))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export StatusCacheTab from metrics barrel ([d62ed6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d62ed6e94fbc7539a92fa1ec18d0bf2f9d0db14b))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* **queue:** extend message models for sticker, PTV, and Z-API types ([6584d50](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6584d508fd9dacdff916c93dce5dbb667ca513f1))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **api:** initialize and wire status cache system ([f3eeac4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3eeac4459271b74b6f12cf40d8a7fa4014c89e5))
* **queue:** integrate new processors and add GetClient helper ([2006337](https://github.com/Funnelchat20/whatsapp-api-golang/commit/200633715d16481132fa88c0e7a45cc515210bea))
* **api:** integrate privacy handler into router ([22ed83b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/22ed83bb0feea86d25a4db55c60ea7cd848d7dbb))
* **api:** integrate status cache interceptor with event dispatch ([40632e5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/40632e589538d021d8cbe3ac0e71feb66921b672))
* **manager:** integrate StatusCache tab into metrics dashboard ([ef8c429](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef8c429d071443565a21980e123bd31fceded647))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **manager:** mask sensitive tokens in UI display ([911485b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911485b013a99fbda868c667536599be80d0b4e6))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **deploy:** add --load flag and correct ECS service name ([9a58192](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9a581928b2e5ea476af9afa447fb0049349721e6))
* **deploy:** add --load flag to docker build command ([6dc8494](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6dc84946c939c1d45b58caf137ab252ebcec0f7c))
* **deploy:** add --load flag to docker build command ([8731d9d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8731d9dcde4d18775e5a5845095d5c158c266fae))
* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **queue:** add missing MessageTypeCarousel case to processor ([a630251](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a630251365f41692d587c803dc8a8c97b804a5d2))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **webhooks:** change HTTP method from POST to PUT for webhook updates ([5b9392b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5b9392b4ce45e9c1e5dbff6e8908de77d1fc9286))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve metrics transformer and tab components ([0d9d220](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0d9d2208218d4c200738fcbf4398d551ce5ee24b))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **pages:** remove console.log handlers from instances list page ([7ea6439](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7ea6439d015ea8e2f712fd68aac910be3acc2091))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** resolve password reset flow in AWS ALB ([a66b332](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a66b332df39d39c85d1852ed6dae58babbc64ff3))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **auth:** support multiple trusted origins for Better Auth ([f1b259a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f1b259a4e7291b3bb09d2002291f4f72afc36d55))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* **examples:** add media examples for button-actions and carousel ([49df2fa](https://github.com/Funnelchat20/whatsapp-api-golang/commit/49df2fa8559b980af17fc7ce175365eab490f571))
* **openapi:** add message field to carousel examples ([8466133](https://github.com/Funnelchat20/whatsapp-api-golang/commit/84661333c6b054a8849b68c14df5dfc29ed2d43d))
* **openapi:** add new endpoint specifications ([f688e9a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f688e9aa834384d1b9ef288d3d1e0facdcd3cd8b))
* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* **api:** add status cache API documentation ([bb293f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb293f56976338dd18ab6bb8bc762eb768a88c15))
* **api:** add StatusCache environment variables to .env.example ([c57da24](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c57da24c2a50aad7eaddb0d201dd5a114834ea23))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* **openapi:** update API documentation for new endpoints ([9e2642d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e2642dd9707e9e2a7402a6a4cac643e44a4a532))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2026-01-09)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **carousel:** add body text at root InteractiveMessage level ([9e44475](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e444751ee434731f42b751b7e14472d8b62d5ff))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* **api:** add config endpoint to expose WHATSAPP_CLIENT_TOKEN ([859d34e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/859d34ebaed57d60570a2981c07abde2a183a015))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **queue:** add custom link preview override support ([d792207](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d792207ab3f8ab4a9f352820e1a1c7c6f8a9feb3))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* **handlers:** add HTTP endpoints for new message types ([e9a9c42](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e9a9c427df57ea95fd0e03fb1a14c25508ed450f))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **components:** add instance management UI components ([7f4d47a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f4d47a130e592b03e290c2bc0e666f927befcb7))
* **messages:** add media support to send-button-actions endpoint ([00e1c96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/00e1c96f6bc0f12300b5ef8b2de2346007e1a38c))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **api:** add proxy routes for send-text, queue, and message status stats ([64d1baf](https://github.com/Funnelchat20/whatsapp-api-golang/commit/64d1bafa9bcc138c76f67a5a6f5ea4041c947348))
* **queue:** add PTV processor for circular video messages ([50dcdcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/50dcdcb20b1475c77c3dbd3831d1d51077113935))
* **api:** add queue, status-cache, and metrics API clients ([59037fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/59037fc6a45063b83d66e35fab74b423870cb0d1))
* **types:** add queue, status-cache, and metrics type definitions ([80ad331](https://github.com/Funnelchat20/whatsapp-api-golang/commit/80ad3318ea76dbb0ef7cf01175abf8c14ff15130))
* **send:** add review_and_pay button support for native flow messages ([3cbf4c1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3cbf4c105b1fb912f92fc4465b44b798cb3dc093))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **api:** add status cache core implementation ([6f0d776](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6f0d7761aa96d1975a5a0aa0f8ea2022177c5729))
* **api:** add status cache HTTP handlers and routes ([acf5910](https://github.com/Funnelchat20/whatsapp-api-golang/commit/acf5910023ab465b126c142bd547f8f088e31f60))
* **api:** add StatusCache configuration ([2d7989c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d7989cd34986971b67929355ca48e9a3a67d54d))
* **manager:** add StatusCache metrics tab component ([d85936d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d85936d9e2e5c82563acd21740f0a46651f54942))
* **manager:** add StatusCache metrics transformer ([8f9be0f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8f9be0f2f3a30d6e44a6e6424834643b46f9c1a9))
* **manager:** add StatusCache metrics type definitions ([023d17f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/023d17f3673dfb5cdb9977e33d815e37a61dbac8))
* **api:** add StatusCache Prometheus metrics ([8b32dcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8b32dcb7fd8b72ea5534ccb19f09aa4d551ae610))
* **queue:** add sticker processor with WebP conversion ([7d4f94a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7d4f94ac1a3de1fe3cd3d49196cd1171f28fb2c7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **hooks:** add SWR hooks for queue, status-cache, metrics, and client token ([bb7bb5b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb7bb5b73a712e08d208b7aaf94f8921505e06be))
* **pages:** add Test tab and cURL copy to instance detail page ([ef80626](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef80626d1aefbcbc97cd960b59c0320842a48fb1))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* **audio:** add waveform visualization for voice notes ([70d8f3e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/70d8f3e46be6a8a70a2dedb6e5700b3e06efc2d3))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* **interactive:** add Z-API compatible message building package ([8a5ade5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8a5ade52d9e6da0baac5a13cfe72c839ae752576))
* **queue:** add Z-API interactive message processors ([fb4ba84](https://github.com/Funnelchat20/whatsapp-api-golang/commit/fb4ba8408355479605852fce42b252617640931d))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enable StatusCache in ECS service configuration ([475e7ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/475e7ba842a06586ba9720339914dbe88c360804))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export StatusCacheTab from metrics barrel ([d62ed6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d62ed6e94fbc7539a92fa1ec18d0bf2f9d0db14b))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* **queue:** extend message models for sticker, PTV, and Z-API types ([6584d50](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6584d508fd9dacdff916c93dce5dbb667ca513f1))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **api:** initialize and wire status cache system ([f3eeac4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3eeac4459271b74b6f12cf40d8a7fa4014c89e5))
* **queue:** integrate new processors and add GetClient helper ([2006337](https://github.com/Funnelchat20/whatsapp-api-golang/commit/200633715d16481132fa88c0e7a45cc515210bea))
* **api:** integrate status cache interceptor with event dispatch ([40632e5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/40632e589538d021d8cbe3ac0e71feb66921b672))
* **manager:** integrate StatusCache tab into metrics dashboard ([ef8c429](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef8c429d071443565a21980e123bd31fceded647))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **deploy:** add --load flag and correct ECS service name ([9a58192](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9a581928b2e5ea476af9afa447fb0049349721e6))
* **deploy:** add --load flag to docker build command ([6dc8494](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6dc84946c939c1d45b58caf137ab252ebcec0f7c))
* **deploy:** add --load flag to docker build command ([8731d9d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8731d9dcde4d18775e5a5845095d5c158c266fae))
* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **queue:** add missing MessageTypeCarousel case to processor ([a630251](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a630251365f41692d587c803dc8a8c97b804a5d2))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **webhooks:** change HTTP method from POST to PUT for webhook updates ([5b9392b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5b9392b4ce45e9c1e5dbff6e8908de77d1fc9286))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve metrics transformer and tab components ([0d9d220](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0d9d2208218d4c200738fcbf4398d551ce5ee24b))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **pages:** remove console.log handlers from instances list page ([7ea6439](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7ea6439d015ea8e2f712fd68aac910be3acc2091))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **auth:** support multiple trusted origins for Better Auth ([f1b259a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f1b259a4e7291b3bb09d2002291f4f72afc36d55))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* **examples:** add media examples for button-actions and carousel ([49df2fa](https://github.com/Funnelchat20/whatsapp-api-golang/commit/49df2fa8559b980af17fc7ce175365eab490f571))
* **openapi:** add message field to carousel examples ([8466133](https://github.com/Funnelchat20/whatsapp-api-golang/commit/84661333c6b054a8849b68c14df5dfc29ed2d43d))
* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* **api:** add status cache API documentation ([bb293f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb293f56976338dd18ab6bb8bc762eb768a88c15))
* **api:** add StatusCache environment variables to .env.example ([c57da24](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c57da24c2a50aad7eaddb0d201dd5a114834ea23))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* **openapi:** update API documentation for new endpoints ([9e2642d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e2642dd9707e9e2a7402a6a4cac643e44a4a532))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2026-01-09)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **carousel:** add body text at root InteractiveMessage level ([9e44475](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e444751ee434731f42b751b7e14472d8b62d5ff))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* **api:** add config endpoint to expose WHATSAPP_CLIENT_TOKEN ([859d34e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/859d34ebaed57d60570a2981c07abde2a183a015))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **queue:** add custom link preview override support ([d792207](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d792207ab3f8ab4a9f352820e1a1c7c6f8a9feb3))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* **handlers:** add HTTP endpoints for new message types ([e9a9c42](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e9a9c427df57ea95fd0e03fb1a14c25508ed450f))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **components:** add instance management UI components ([7f4d47a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f4d47a130e592b03e290c2bc0e666f927befcb7))
* **messages:** add media support to send-button-actions endpoint ([00e1c96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/00e1c96f6bc0f12300b5ef8b2de2346007e1a38c))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **api:** add proxy routes for send-text, queue, and message status stats ([64d1baf](https://github.com/Funnelchat20/whatsapp-api-golang/commit/64d1bafa9bcc138c76f67a5a6f5ea4041c947348))
* **queue:** add PTV processor for circular video messages ([50dcdcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/50dcdcb20b1475c77c3dbd3831d1d51077113935))
* **api:** add queue, status-cache, and metrics API clients ([59037fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/59037fc6a45063b83d66e35fab74b423870cb0d1))
* **types:** add queue, status-cache, and metrics type definitions ([80ad331](https://github.com/Funnelchat20/whatsapp-api-golang/commit/80ad3318ea76dbb0ef7cf01175abf8c14ff15130))
* **send:** add review_and_pay button support for native flow messages ([3cbf4c1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3cbf4c105b1fb912f92fc4465b44b798cb3dc093))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **api:** add status cache core implementation ([6f0d776](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6f0d7761aa96d1975a5a0aa0f8ea2022177c5729))
* **api:** add status cache HTTP handlers and routes ([acf5910](https://github.com/Funnelchat20/whatsapp-api-golang/commit/acf5910023ab465b126c142bd547f8f088e31f60))
* **api:** add StatusCache configuration ([2d7989c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d7989cd34986971b67929355ca48e9a3a67d54d))
* **manager:** add StatusCache metrics tab component ([d85936d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d85936d9e2e5c82563acd21740f0a46651f54942))
* **manager:** add StatusCache metrics transformer ([8f9be0f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8f9be0f2f3a30d6e44a6e6424834643b46f9c1a9))
* **manager:** add StatusCache metrics type definitions ([023d17f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/023d17f3673dfb5cdb9977e33d815e37a61dbac8))
* **api:** add StatusCache Prometheus metrics ([8b32dcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8b32dcb7fd8b72ea5534ccb19f09aa4d551ae610))
* **queue:** add sticker processor with WebP conversion ([7d4f94a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7d4f94ac1a3de1fe3cd3d49196cd1171f28fb2c7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **hooks:** add SWR hooks for queue, status-cache, metrics, and client token ([bb7bb5b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb7bb5b73a712e08d208b7aaf94f8921505e06be))
* **pages:** add Test tab and cURL copy to instance detail page ([ef80626](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef80626d1aefbcbc97cd960b59c0320842a48fb1))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* **audio:** add waveform visualization for voice notes ([70d8f3e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/70d8f3e46be6a8a70a2dedb6e5700b3e06efc2d3))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* **interactive:** add Z-API compatible message building package ([8a5ade5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8a5ade52d9e6da0baac5a13cfe72c839ae752576))
* **queue:** add Z-API interactive message processors ([fb4ba84](https://github.com/Funnelchat20/whatsapp-api-golang/commit/fb4ba8408355479605852fce42b252617640931d))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enable StatusCache in ECS service configuration ([475e7ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/475e7ba842a06586ba9720339914dbe88c360804))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export StatusCacheTab from metrics barrel ([d62ed6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d62ed6e94fbc7539a92fa1ec18d0bf2f9d0db14b))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* **queue:** extend message models for sticker, PTV, and Z-API types ([6584d50](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6584d508fd9dacdff916c93dce5dbb667ca513f1))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **api:** initialize and wire status cache system ([f3eeac4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3eeac4459271b74b6f12cf40d8a7fa4014c89e5))
* **queue:** integrate new processors and add GetClient helper ([2006337](https://github.com/Funnelchat20/whatsapp-api-golang/commit/200633715d16481132fa88c0e7a45cc515210bea))
* **api:** integrate status cache interceptor with event dispatch ([40632e5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/40632e589538d021d8cbe3ac0e71feb66921b672))
* **manager:** integrate StatusCache tab into metrics dashboard ([ef8c429](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef8c429d071443565a21980e123bd31fceded647))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **deploy:** add --load flag and correct ECS service name ([9a58192](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9a581928b2e5ea476af9afa447fb0049349721e6))
* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **queue:** add missing MessageTypeCarousel case to processor ([a630251](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a630251365f41692d587c803dc8a8c97b804a5d2))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **webhooks:** change HTTP method from POST to PUT for webhook updates ([5b9392b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5b9392b4ce45e9c1e5dbff6e8908de77d1fc9286))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve metrics transformer and tab components ([0d9d220](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0d9d2208218d4c200738fcbf4398d551ce5ee24b))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **pages:** remove console.log handlers from instances list page ([7ea6439](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7ea6439d015ea8e2f712fd68aac910be3acc2091))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **auth:** support multiple trusted origins for Better Auth ([f1b259a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f1b259a4e7291b3bb09d2002291f4f72afc36d55))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* **examples:** add media examples for button-actions and carousel ([49df2fa](https://github.com/Funnelchat20/whatsapp-api-golang/commit/49df2fa8559b980af17fc7ce175365eab490f571))
* **openapi:** add message field to carousel examples ([8466133](https://github.com/Funnelchat20/whatsapp-api-golang/commit/84661333c6b054a8849b68c14df5dfc29ed2d43d))
* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* **api:** add status cache API documentation ([bb293f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb293f56976338dd18ab6bb8bc762eb768a88c15))
* **api:** add StatusCache environment variables to .env.example ([c57da24](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c57da24c2a50aad7eaddb0d201dd5a114834ea23))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* **openapi:** update API documentation for new endpoints ([9e2642d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e2642dd9707e9e2a7402a6a4cac643e44a4a532))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2026-01-09)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **carousel:** add body text at root InteractiveMessage level ([9e44475](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e444751ee434731f42b751b7e14472d8b62d5ff))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* **api:** add config endpoint to expose WHATSAPP_CLIENT_TOKEN ([859d34e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/859d34ebaed57d60570a2981c07abde2a183a015))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **queue:** add custom link preview override support ([d792207](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d792207ab3f8ab4a9f352820e1a1c7c6f8a9feb3))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* **handlers:** add HTTP endpoints for new message types ([e9a9c42](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e9a9c427df57ea95fd0e03fb1a14c25508ed450f))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **components:** add instance management UI components ([7f4d47a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f4d47a130e592b03e290c2bc0e666f927befcb7))
* **messages:** add media support to send-button-actions endpoint ([00e1c96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/00e1c96f6bc0f12300b5ef8b2de2346007e1a38c))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **api:** add proxy routes for send-text, queue, and message status stats ([64d1baf](https://github.com/Funnelchat20/whatsapp-api-golang/commit/64d1bafa9bcc138c76f67a5a6f5ea4041c947348))
* **queue:** add PTV processor for circular video messages ([50dcdcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/50dcdcb20b1475c77c3dbd3831d1d51077113935))
* **api:** add queue, status-cache, and metrics API clients ([59037fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/59037fc6a45063b83d66e35fab74b423870cb0d1))
* **types:** add queue, status-cache, and metrics type definitions ([80ad331](https://github.com/Funnelchat20/whatsapp-api-golang/commit/80ad3318ea76dbb0ef7cf01175abf8c14ff15130))
* **send:** add review_and_pay button support for native flow messages ([3cbf4c1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3cbf4c105b1fb912f92fc4465b44b798cb3dc093))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **api:** add status cache core implementation ([6f0d776](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6f0d7761aa96d1975a5a0aa0f8ea2022177c5729))
* **api:** add status cache HTTP handlers and routes ([acf5910](https://github.com/Funnelchat20/whatsapp-api-golang/commit/acf5910023ab465b126c142bd547f8f088e31f60))
* **api:** add StatusCache configuration ([2d7989c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d7989cd34986971b67929355ca48e9a3a67d54d))
* **manager:** add StatusCache metrics tab component ([d85936d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d85936d9e2e5c82563acd21740f0a46651f54942))
* **manager:** add StatusCache metrics transformer ([8f9be0f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8f9be0f2f3a30d6e44a6e6424834643b46f9c1a9))
* **manager:** add StatusCache metrics type definitions ([023d17f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/023d17f3673dfb5cdb9977e33d815e37a61dbac8))
* **api:** add StatusCache Prometheus metrics ([8b32dcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8b32dcb7fd8b72ea5534ccb19f09aa4d551ae610))
* **queue:** add sticker processor with WebP conversion ([7d4f94a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7d4f94ac1a3de1fe3cd3d49196cd1171f28fb2c7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **hooks:** add SWR hooks for queue, status-cache, metrics, and client token ([bb7bb5b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb7bb5b73a712e08d208b7aaf94f8921505e06be))
* **pages:** add Test tab and cURL copy to instance detail page ([ef80626](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef80626d1aefbcbc97cd960b59c0320842a48fb1))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* **interactive:** add Z-API compatible message building package ([8a5ade5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8a5ade52d9e6da0baac5a13cfe72c839ae752576))
* **queue:** add Z-API interactive message processors ([fb4ba84](https://github.com/Funnelchat20/whatsapp-api-golang/commit/fb4ba8408355479605852fce42b252617640931d))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enable StatusCache in ECS service configuration ([475e7ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/475e7ba842a06586ba9720339914dbe88c360804))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export StatusCacheTab from metrics barrel ([d62ed6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d62ed6e94fbc7539a92fa1ec18d0bf2f9d0db14b))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* **queue:** extend message models for sticker, PTV, and Z-API types ([6584d50](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6584d508fd9dacdff916c93dce5dbb667ca513f1))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **api:** initialize and wire status cache system ([f3eeac4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3eeac4459271b74b6f12cf40d8a7fa4014c89e5))
* **queue:** integrate new processors and add GetClient helper ([2006337](https://github.com/Funnelchat20/whatsapp-api-golang/commit/200633715d16481132fa88c0e7a45cc515210bea))
* **api:** integrate status cache interceptor with event dispatch ([40632e5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/40632e589538d021d8cbe3ac0e71feb66921b672))
* **manager:** integrate StatusCache tab into metrics dashboard ([ef8c429](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef8c429d071443565a21980e123bd31fceded647))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **deploy:** add --load flag and correct ECS service name ([9a58192](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9a581928b2e5ea476af9afa447fb0049349721e6))
* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **queue:** add missing MessageTypeCarousel case to processor ([a630251](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a630251365f41692d587c803dc8a8c97b804a5d2))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **webhooks:** change HTTP method from POST to PUT for webhook updates ([5b9392b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5b9392b4ce45e9c1e5dbff6e8908de77d1fc9286))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve metrics transformer and tab components ([0d9d220](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0d9d2208218d4c200738fcbf4398d551ce5ee24b))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **pages:** remove console.log handlers from instances list page ([7ea6439](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7ea6439d015ea8e2f712fd68aac910be3acc2091))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **auth:** support multiple trusted origins for Better Auth ([f1b259a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f1b259a4e7291b3bb09d2002291f4f72afc36d55))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* **examples:** add media examples for button-actions and carousel ([49df2fa](https://github.com/Funnelchat20/whatsapp-api-golang/commit/49df2fa8559b980af17fc7ce175365eab490f571))
* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* **api:** add status cache API documentation ([bb293f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb293f56976338dd18ab6bb8bc762eb768a88c15))
* **api:** add StatusCache environment variables to .env.example ([c57da24](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c57da24c2a50aad7eaddb0d201dd5a114834ea23))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* **openapi:** update API documentation for new endpoints ([9e2642d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9e2642dd9707e9e2a7402a6a4cac643e44a4a532))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2026-01-09)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* **api:** add config endpoint to expose WHATSAPP_CLIENT_TOKEN ([859d34e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/859d34ebaed57d60570a2981c07abde2a183a015))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **components:** add instance management UI components ([7f4d47a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f4d47a130e592b03e290c2bc0e666f927befcb7))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **api:** add proxy routes for send-text, queue, and message status stats ([64d1baf](https://github.com/Funnelchat20/whatsapp-api-golang/commit/64d1bafa9bcc138c76f67a5a6f5ea4041c947348))
* **api:** add queue, status-cache, and metrics API clients ([59037fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/59037fc6a45063b83d66e35fab74b423870cb0d1))
* **types:** add queue, status-cache, and metrics type definitions ([80ad331](https://github.com/Funnelchat20/whatsapp-api-golang/commit/80ad3318ea76dbb0ef7cf01175abf8c14ff15130))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **api:** add status cache core implementation ([6f0d776](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6f0d7761aa96d1975a5a0aa0f8ea2022177c5729))
* **api:** add status cache HTTP handlers and routes ([acf5910](https://github.com/Funnelchat20/whatsapp-api-golang/commit/acf5910023ab465b126c142bd547f8f088e31f60))
* **api:** add StatusCache configuration ([2d7989c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d7989cd34986971b67929355ca48e9a3a67d54d))
* **manager:** add StatusCache metrics tab component ([d85936d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d85936d9e2e5c82563acd21740f0a46651f54942))
* **manager:** add StatusCache metrics transformer ([8f9be0f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8f9be0f2f3a30d6e44a6e6424834643b46f9c1a9))
* **manager:** add StatusCache metrics type definitions ([023d17f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/023d17f3673dfb5cdb9977e33d815e37a61dbac8))
* **api:** add StatusCache Prometheus metrics ([8b32dcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8b32dcb7fd8b72ea5534ccb19f09aa4d551ae610))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **hooks:** add SWR hooks for queue, status-cache, metrics, and client token ([bb7bb5b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb7bb5b73a712e08d208b7aaf94f8921505e06be))
* **pages:** add Test tab and cURL copy to instance detail page ([ef80626](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef80626d1aefbcbc97cd960b59c0320842a48fb1))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enable StatusCache in ECS service configuration ([475e7ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/475e7ba842a06586ba9720339914dbe88c360804))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export StatusCacheTab from metrics barrel ([d62ed6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d62ed6e94fbc7539a92fa1ec18d0bf2f9d0db14b))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **api:** initialize and wire status cache system ([f3eeac4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3eeac4459271b74b6f12cf40d8a7fa4014c89e5))
* **api:** integrate status cache interceptor with event dispatch ([40632e5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/40632e589538d021d8cbe3ac0e71feb66921b672))
* **manager:** integrate StatusCache tab into metrics dashboard ([ef8c429](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef8c429d071443565a21980e123bd31fceded647))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **webhooks:** change HTTP method from POST to PUT for webhook updates ([5b9392b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5b9392b4ce45e9c1e5dbff6e8908de77d1fc9286))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve metrics transformer and tab components ([0d9d220](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0d9d2208218d4c200738fcbf4398d551ce5ee24b))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **pages:** remove console.log handlers from instances list page ([7ea6439](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7ea6439d015ea8e2f712fd68aac910be3acc2091))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* **api:** add status cache API documentation ([bb293f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb293f56976338dd18ab6bb8bc762eb768a88c15))
* **api:** add StatusCache environment variables to .env.example ([c57da24](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c57da24c2a50aad7eaddb0d201dd5a114834ea23))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-30)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* **api:** add config endpoint to expose WHATSAPP_CLIENT_TOKEN ([859d34e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/859d34ebaed57d60570a2981c07abde2a183a015))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **components:** add instance management UI components ([7f4d47a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f4d47a130e592b03e290c2bc0e666f927befcb7))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **api:** add proxy routes for send-text, queue, and message status stats ([64d1baf](https://github.com/Funnelchat20/whatsapp-api-golang/commit/64d1bafa9bcc138c76f67a5a6f5ea4041c947348))
* **api:** add queue, status-cache, and metrics API clients ([59037fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/59037fc6a45063b83d66e35fab74b423870cb0d1))
* **types:** add queue, status-cache, and metrics type definitions ([80ad331](https://github.com/Funnelchat20/whatsapp-api-golang/commit/80ad3318ea76dbb0ef7cf01175abf8c14ff15130))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **api:** add status cache core implementation ([6f0d776](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6f0d7761aa96d1975a5a0aa0f8ea2022177c5729))
* **api:** add status cache HTTP handlers and routes ([acf5910](https://github.com/Funnelchat20/whatsapp-api-golang/commit/acf5910023ab465b126c142bd547f8f088e31f60))
* **api:** add StatusCache configuration ([2d7989c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d7989cd34986971b67929355ca48e9a3a67d54d))
* **manager:** add StatusCache metrics tab component ([d85936d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d85936d9e2e5c82563acd21740f0a46651f54942))
* **manager:** add StatusCache metrics transformer ([8f9be0f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8f9be0f2f3a30d6e44a6e6424834643b46f9c1a9))
* **manager:** add StatusCache metrics type definitions ([023d17f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/023d17f3673dfb5cdb9977e33d815e37a61dbac8))
* **api:** add StatusCache Prometheus metrics ([8b32dcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8b32dcb7fd8b72ea5534ccb19f09aa4d551ae610))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **hooks:** add SWR hooks for queue, status-cache, metrics, and client token ([bb7bb5b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb7bb5b73a712e08d208b7aaf94f8921505e06be))
* **pages:** add Test tab and cURL copy to instance detail page ([ef80626](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef80626d1aefbcbc97cd960b59c0320842a48fb1))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enable StatusCache in ECS service configuration ([475e7ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/475e7ba842a06586ba9720339914dbe88c360804))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export StatusCacheTab from metrics barrel ([d62ed6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d62ed6e94fbc7539a92fa1ec18d0bf2f9d0db14b))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **api:** initialize and wire status cache system ([f3eeac4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3eeac4459271b74b6f12cf40d8a7fa4014c89e5))
* **api:** integrate status cache interceptor with event dispatch ([40632e5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/40632e589538d021d8cbe3ac0e71feb66921b672))
* **manager:** integrate StatusCache tab into metrics dashboard ([ef8c429](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef8c429d071443565a21980e123bd31fceded647))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **webhooks:** change HTTP method from POST to PUT for webhook updates ([5b9392b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5b9392b4ce45e9c1e5dbff6e8908de77d1fc9286))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve metrics transformer and tab components ([0d9d220](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0d9d2208218d4c200738fcbf4398d551ce5ee24b))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **pages:** remove console.log handlers from instances list page ([7ea6439](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7ea6439d015ea8e2f712fd68aac910be3acc2091))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* **api:** add status cache API documentation ([bb293f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb293f56976338dd18ab6bb8bc762eb768a88c15))
* **api:** add StatusCache environment variables to .env.example ([c57da24](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c57da24c2a50aad7eaddb0d201dd5a114834ea23))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-27)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* **api:** add config endpoint to expose WHATSAPP_CLIENT_TOKEN ([859d34e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/859d34ebaed57d60570a2981c07abde2a183a015))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **components:** add instance management UI components ([7f4d47a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7f4d47a130e592b03e290c2bc0e666f927befcb7))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **api:** add proxy routes for send-text, queue, and message status stats ([64d1baf](https://github.com/Funnelchat20/whatsapp-api-golang/commit/64d1bafa9bcc138c76f67a5a6f5ea4041c947348))
* **api:** add queue, status-cache, and metrics API clients ([59037fc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/59037fc6a45063b83d66e35fab74b423870cb0d1))
* **types:** add queue, status-cache, and metrics type definitions ([80ad331](https://github.com/Funnelchat20/whatsapp-api-golang/commit/80ad3318ea76dbb0ef7cf01175abf8c14ff15130))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **api:** add status cache core implementation ([6f0d776](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6f0d7761aa96d1975a5a0aa0f8ea2022177c5729))
* **api:** add status cache HTTP handlers and routes ([acf5910](https://github.com/Funnelchat20/whatsapp-api-golang/commit/acf5910023ab465b126c142bd547f8f088e31f60))
* **api:** add StatusCache configuration ([2d7989c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d7989cd34986971b67929355ca48e9a3a67d54d))
* **manager:** add StatusCache metrics tab component ([d85936d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d85936d9e2e5c82563acd21740f0a46651f54942))
* **manager:** add StatusCache metrics transformer ([8f9be0f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8f9be0f2f3a30d6e44a6e6424834643b46f9c1a9))
* **manager:** add StatusCache metrics type definitions ([023d17f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/023d17f3673dfb5cdb9977e33d815e37a61dbac8))
* **api:** add StatusCache Prometheus metrics ([8b32dcb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8b32dcb7fd8b72ea5534ccb19f09aa4d551ae610))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **hooks:** add SWR hooks for queue, status-cache, metrics, and client token ([bb7bb5b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb7bb5b73a712e08d208b7aaf94f8921505e06be))
* **pages:** add Test tab and cURL copy to instance detail page ([ef80626](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef80626d1aefbcbc97cd960b59c0320842a48fb1))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enable StatusCache in ECS service configuration ([475e7ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/475e7ba842a06586ba9720339914dbe88c360804))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export StatusCacheTab from metrics barrel ([d62ed6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d62ed6e94fbc7539a92fa1ec18d0bf2f9d0db14b))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **api:** initialize and wire status cache system ([f3eeac4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3eeac4459271b74b6f12cf40d8a7fa4014c89e5))
* **api:** integrate status cache interceptor with event dispatch ([40632e5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/40632e589538d021d8cbe3ac0e71feb66921b672))
* **manager:** integrate StatusCache tab into metrics dashboard ([ef8c429](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ef8c429d071443565a21980e123bd31fceded647))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve metrics transformer and tab components ([0d9d220](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0d9d2208218d4c200738fcbf4398d551ce5ee24b))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **pages:** remove console.log handlers from instances list page ([7ea6439](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7ea6439d015ea8e2f712fd68aac910be3acc2091))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* **api:** add status cache API documentation ([bb293f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bb293f56976338dd18ab6bb8bc762eb768a88c15))
* **api:** add StatusCache environment variables to .env.example ([c57da24](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c57da24c2a50aad7eaddb0d201dd5a114834ea23))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-21)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **manager:** improve Media by Instance section with avatar and phone display ([ffb1616](https://github.com/Funnelchat20/whatsapp-api-golang/commit/ffb161633fe6f65303dd7962d6c1a61390faec68))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* **manager:** improve responsive layout for instance cards in metrics tabs ([01324c3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/01324c32fad646c0d20d64afda6e4075559d8323))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-21)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **manager:** improve Queue by Instance section with avatar and phone display ([373153f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/373153f874a77fab773d86acfc6aa5e020b73315))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-21)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **metrics:** add queue metrics instrumentation for enqueue operations ([843c06b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/843c06b067de0418f3bd7b058f2c3ca50c2b6b28))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **metrics:** pass metrics to transport registry for HTTP transport instrumentation ([9328a9c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/9328a9c5c1810a2f4d8a712e29bde3f638f750cf))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-21)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **manager:** add instance details with avatar, phone, and friendly names ([708a298](https://github.com/Funnelchat20/whatsapp-api-golang/commit/708a298d1ec11f2f2173f39549728f6c9d50e282))
* **metrics:** add message queue size and worker metrics to coordinator ([f291549](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f291549ceb26ff2696cb5fa9032926a7ce1c4769))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* **metrics:** add periodic EventOutboxBacklog gauge updates ([3ed9f96](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3ed9f968904f2afc58e18ab02d67fcdf42f3bd99))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **metrics:** add processing metrics to message queue worker ([1d7a53f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1d7a53f515d5f1b6ef862c2e0cf116c92ab063ee))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **manager:** add transport metrics parser to transformer ([533cf85](https://github.com/Funnelchat20/whatsapp-api-golang/commit/533cf85652d147b7c4ed4fab9cb21644ede5cc88))
* **manager:** add transport metrics tab component ([0f8765b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0f8765b2ddaca1e8ca7a2cf31984d9448f68f0ff))
* **manager:** add transport metrics type definitions ([4dcaefd](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4dcaefd484d2170d770393a0e87e69306f2abfa9))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **manager:** export TransportMetricsTab from metrics barrel ([24b7c43](https://github.com/Funnelchat20/whatsapp-api-golang/commit/24b7c43c45ebb27c8a9fded23bfa43fca8826c72))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **metrics:** implement transport delivery metrics in HTTP transport ([c8ec2d9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c8ec2d9113efb1867583fb83a118885f0a0fdd9a))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **manager:** improve HTTP status code display with friendly labels ([8267bd4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8267bd4b8068d4bc19b2ca60bc7bcdc0ff1fe2f5))
* **manager:** integrate transport tab into metrics dashboard ([12f27f4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/12f27f4041b3b65969a53c15fa0441c86b2ab238))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **metrics:** add all 24 event types from backend ([2d61fb8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d61fb89dee833b509ec390d96cba9f3e617dee2))
* **manager:** add informative message when queue metrics unavailable ([af50e60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/af50e60a9d12c8a534c8469b6976f52d406d46b4))
* **metrics:** add missing attempt label to EventRetries metric ([f684e6e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f684e6e67697e9bd5f91392de88424ba2f5d1575))
* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **metrics:** correct metric field names in message service ([f8eb273](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f8eb273eb9b9d6c3b9cb11ade86be8353da23653))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **metrics:** handle whatsmeow_api_ prefix and _total suffix variations ([5e92ddb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5e92ddb433da78b4adfeaa2a6848d2d8c8dd247a))
* **metrics:** improve chart colors for better UX ([14eaf63](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14eaf63391e522fd3c23511ebbe1f895fd626831))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **metrics:** improve Events tab UX with friendly names and colors ([89d7fa1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/89d7fa1a16690f9bac4439f75a23e79b0ad5f0c9))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* **metrics:** improve page responsiveness and add to mobile nav ([8175279](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8175279a03fd4e3234d9e17eff2969d9c2e34dcc))
* **metrics:** improve Queue tab empty state with debug info ([52ec572](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ec57253fd6ac7b32abf55cb6b13e1ec75ff897))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))
* **metrics:** use worker_type label instead of instance_id in dispatch worker ([a3912f6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a3912f6536003f0a8f1a8f853e3a56e9d40e11f3))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **phone:** centralize phone formatting with international support ([0869588](https://github.com/Funnelchat20/whatsapp-api-golang/commit/08695884bbc6ccc34b943cf436ae255ce0b21d70))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-21)

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

* **metrics:** add API route to fetch and parse prometheus metrics ([52ffe10](https://github.com/Funnelchat20/whatsapp-api-golang/commit/52ffe100c68aa89c88dfaad92b704b6d8a4b0d5d))
* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* **metrics:** add comprehensive metrics dashboard components ([2d8e0e7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/2d8e0e7bc3dbdec664febc8e9d0cd6ee8631eb43))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* **metrics:** add metrics dashboard page with tabbed navigation ([087cfab](https://github.com/Funnelchat20/whatsapp-api-golang/commit/087cfab983246a18d97403c46446a5f5c69ddc2b))
* **metrics:** add metrics navigation link to sidebar ([c3f1a5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c3f1a5af1edbab4cdce8c2bd5bf3322b7dfa0059))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **metrics:** add prometheus parser and transformer utilities ([79f76d6](https://github.com/Funnelchat20/whatsapp-api-golang/commit/79f76d67cbf51360380aad173b1847e7b61fb8c8))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* **metrics:** add SWR hook for metrics data fetching with polling ([5d00145](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5d001451411aae973ca665cccd89d3f441233362))
* **metrics:** add typescript interfaces for prometheus metrics ([e6eb88c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e6eb88cbcdcf1d2d9306b31f0ff2c8117360e667))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-20)

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

* **manager:** add clickable rows to instance table for navigation ([dae855c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dae855cae1b5289147e6f191b346570fe32e5d5c))
* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-20)

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

* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **manager:** improve dashboard UI with modern minimal design ([3af5de3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3af5de32c40c12961157464f0d629ce9e7643d68))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-20)

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

* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **manager:** add Settings to instance actions and fix tab navigation ([1bdb965](https://github.com/Funnelchat20/whatsapp-api-golang/commit/1bdb965e5c96e6ca7a8c1cb13e275e9f15f0860f))
* **manager:** fix email URLs and translate subjects to English ([6cdc6a3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6cdc6a35a898dac3cfec32900ba8441a8e214fab))
* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* **manager:** improve instance detail navigation with tab query params ([63454d2](https://github.com/Funnelchat20/whatsapp-api-golang/commit/63454d202f73226337b8121c709738408e8cc059))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **manager:** prioritize APP_URL for email links and format email modules ([72d9da1](https://github.com/Funnelchat20/whatsapp-api-golang/commit/72d9da19be50700179c0f12d11ecf063f17c954e))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **manager:** resolve Biome linter warnings ([0db4f0d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0db4f0dad01775ea607f3acdb329c07225397af6))
* **manager:** resolve eslint warnings and improve accessibility ([cb0f74f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cb0f74fa33cbc45dc8a8553c0b1f708eae2f97df))
* **manager:** resolve lint warnings and improve accessibility ([90b3372](https://github.com/Funnelchat20/whatsapp-api-golang/commit/90b3372ecc4c4cc9d93a5f2556dffdfa7fd6895a))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** remove redundant webhooks and settings routes ([3e4b18c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3e4b18cceec196e9b11e1de9294cb96f93085d19))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-20)

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

* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **terraform:** add dedicated ALB module for Manager frontend ([3108f0b](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3108f0bb7975b705d1fbc90fac0789fd1b0872a6))
* **manager:** add deployment and setup scripts ([f97bf46](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f97bf465f15028cab47c9838aad4b7501b6d8f76))
* **api:** add deployment script for API backend ([73969f5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/73969f56befa4e6532ff3c537141219533f5d0e6))
* **manager:** add Dockerfile for containerized deployment ([3c8e993](https://github.com/Funnelchat20/whatsapp-api-golang/commit/3c8e99339cf37ef631da3c65fa20225966154f3d))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* **terraform:** add ECS service module for Manager frontend ([afd46ba](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afd46ba414171499337dd5c4eb7649a2a6646dce))
* **ci:** add GitHub Actions workflow for Manager deployment ([aa10df9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/aa10df98eb0cc3c388d823b6cfc8113126007abb))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* **terraform:** add security group rule for Manager port 3000 ([b237970](https://github.com/Funnelchat20/whatsapp-api-golang/commit/b237970dfee1e68722e175ffbe04f8781cf141f7))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** configure homolog environment with dedicated Manager ALB ([42dec6f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/42dec6fe91826514c519773c40f3a0f2063f8bf7))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* **manager:** fix secure cookies and login alert email ([7460a29](https://github.com/Funnelchat20/whatsapp-api-golang/commit/7460a29721e7041944898080957f59bf9b8cc6cd))
* **manager:** fix seed.ts user creation order ([d8438c4](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d8438c447d5aa7d85fe9f7e25128212b1e7f906b))
* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))
* **manager:** use window.location.origin for auth client baseURL ([54d3141](https://github.com/Funnelchat20/whatsapp-api-golang/commit/54d3141cef92f3a1ee2e9ee6d6552cd5a0f19f72))

### ♻️ Code Refactoring

* **terraform:** allow environment suffix in secrets module ([e572a83](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e572a830b941612a7bed87c4b53d8eefaf467c50))
* **terraform:** remove Manager routing from API ALB ([641add8](https://github.com/Funnelchat20/whatsapp-api-golang/commit/641add83213a19e6b05cfa1eb5284323bd6b215c))
* **manager:** simplify auto-refresh indicator component ([c35f37f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c35f37f396c184830916e6a18c548204d2e1b987))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-19)

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

* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))
* **manager:** remove backend check from health endpoint ([15f3de7](https://github.com/Funnelchat20/whatsapp-api-golang/commit/15f3de7ff4f800f9b1b7b33f7fd5bd86554ebeb2))
* **manager:** restore backend check in health endpoint for frontend monitoring ([57ab6bb](https://github.com/Funnelchat20/whatsapp-api-golang/commit/57ab6bbbaa617f5f04d57ac2a22f77c9d0883c6f))
* **manager:** return correct health response format for frontend ([e80a558](https://github.com/Funnelchat20/whatsapp-api-golang/commit/e80a5587b0e892626bf47aae8265e86951fbede2))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-19)

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

* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **manager:** improve health check endpoint for container orchestration ([dc55f78](https://github.com/Funnelchat20/whatsapp-api-golang/commit/dc55f78d56744f576ba6d01c59bc6cf03afc1971))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-19)

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

* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* add phone validation endpoints ([c96a8f9](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c96a8f99cce50d4362542e9c3a0fa69540486463))
* add WhatsApp Manager web application ([baacc6c](https://github.com/Funnelchat20/whatsapp-api-golang/commit/baacc6cd0f4f4f6bc78990af663dab9099ec1f9b))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-19)

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

* add contacts service with phone validation ([5bacdcc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5bacdcce46a5ac071bbb18b654939fd5ff3ad182))
* **docs:** add dynamic OpenAPI specification generation ([d6f6b86](https://github.com/Funnelchat20/whatsapp-api-golang/commit/d6f6b866b34807402544a0df01c6f83392351a53))
* add instance configuration settings for calls and messages ([5ef7f60](https://github.com/Funnelchat20/whatsapp-api-golang/commit/5ef7f60d793cf85d1912df99ae4e98a2830b9cd5))
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* implement pairing code cache with TTL ([14ab61d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/14ab61d0b53404fafac21a98222fc393df026dec))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 🐛 Bug Fixes

* improve S3 credential handling for IAM roles ([911ae5a](https://github.com/Funnelchat20/whatsapp-api-golang/commit/911ae5a67485960c7888530617c69aac64ef051d))

### 📝 Documentation

* add OpenAPI schemas for contacts and instance settings ([afdfe05](https://github.com/Funnelchat20/whatsapp-api-golang/commit/afdfe05da377e81778d36f3cd5faa61a7f8434c0))
* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))
* update endpoint implementation status tracking ([8eae8c5](https://github.com/Funnelchat20/whatsapp-api-golang/commit/8eae8c55210169cac61f6d7503ca9f2b5058969c))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-18)

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
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 📝 Documentation

* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-12-18)

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
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **core:** sync whatsmeow core library with upstream changes ([0512728](https://github.com/Funnelchat20/whatsapp-api-golang/commit/0512728ae7a423543b187a1225cc45430503c72e))
* **api:** update API internals with improved configuration and error handling ([bc11341](https://github.com/Funnelchat20/whatsapp-api-golang/commit/bc1134114be8351ed8b42ab0a2697bd159667f2f))
* **proto:** update protobuf definitions from upstream whatsmeow ([a56671f](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a56671f39a46f1eb6a1996fc742500973ac3a4e0))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))
* **socket:** update socket layer and store with upstream improvements ([591bd77](https://github.com/Funnelchat20/whatsapp-api-golang/commit/591bd7701278818625438c541f4721a22f94124f))

### 📝 Documentation

* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))

## [2.0.0-develop.1](https://github.com/Funnelchat20/whatsapp-api-golang/compare/v1.2.0-develop.3...v2.0.0-develop.1) (2025-11-13)

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
* add PDF processing and image manipulation dependencies ([869074e](https://github.com/Funnelchat20/whatsapp-api-golang/commit/869074e31349f7fb5e3d700db39d7a5b139f7649))
* add z-api services, queues, and poll events ([cd21306](https://github.com/Funnelchat20/whatsapp-api-golang/commit/cd213062fab5ba6f64e55f86e7370d8441fd2cd8))
* **terraform:** enhance configuration for S3 and media handling ([a9de795](https://github.com/Funnelchat20/whatsapp-api-golang/commit/a9de795bb7fda9bc4e250eb4c27f573091c8da16))
* enrich group membership and interactive payloads ([f3487bc](https://github.com/Funnelchat20/whatsapp-api-golang/commit/f3487bc34331e049d0e9fc448e9a6380f768e5c6))
* **terraform:** migrate to AWS managed services architecture ([768916d](https://github.com/Funnelchat20/whatsapp-api-golang/commit/768916d2b924e40644676bda2f5e4e686f42d2db))
* **terraform:** update S3 URL expiration and enhance media storage configuration ([4f7b2d3](https://github.com/Funnelchat20/whatsapp-api-golang/commit/4f7b2d33368a80c57c1812711efec302eb254f3d))

### 📝 Documentation

* add z-api playbooks and handler references ([c330d54](https://github.com/Funnelchat20/whatsapp-api-golang/commit/c330d54886a6af1c8938ca883ef6f8d253685d33))

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
* add REST API layer ([6a1fb66](https://github.com/Funnelchat20/whatsapp-api-golang/commit/6a1fb661ff97c92b07be702c585eac5942593d33))
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
