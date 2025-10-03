# Changelog

All notable changes to this project will be documented in this file.

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
