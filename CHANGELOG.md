# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-09-09
### Added
- Connect to D-Bus session bus
- Discover active MPRIS players (e.g., Spotify)
- Retrieve metadata and playback status
- Subscribe to `PropertiesChanged` signals for real-time updates
- Handle player lifecycle with `NameOwnerChanged` signals
- Goroutine-based signal listening with cleanup and error recovery
