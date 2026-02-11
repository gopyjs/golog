# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.0.0] - 2026-02-10

### forked: https://github.com/gopyjs/golog/v3

### Added

- **New Configuration: `splitByLevel`**
  - Controls whether logs are split into separate files by level
  - Default: `false` (all logs go to a single file)
  - When `true`: Each level is written to its own file only (no duplication)
  - Files generated: `app_debug.log`, `app_info.log`, `app_warn.log`, `app_error.log`, etc.

- **New Configuration: `isOutputFile`**
  - Controls whether logs are written to files
  - Default: `true` (logs are saved to files)
  - When `false`: Logs are not written to files (console only if `isOutputStdout=true`)

- **Environment Variable Support**
  - Added `InitFromEnv()` method to configure logger from environment variables
  - Supported variables:
    - `LOG_LEVEL` - Set log level (debug, info, warn, error, fatal, panic)
    - `LOG_SHORT` - Use short caller path (true/false)
    - `LOG_JSON` - Output JSON format (true/false)
    - `LOG_SPLIT_BY_LEVEL` - Split logs by level (true/false)
    - `LOG_OUTPUT_FILE` - Enable file output (true/false)
    - `LOG_OUTPUT_STDOUT` - Enable stdout output (true/false)
    - `LOG_PATH` - Log directory path
    - `LOG_FILE_NAME` - Log file name prefix
    - `LOG_MAX_SIZE_MB` - Max log file size before rotation
    - `LOG_MAX_BACKUPS` - Max number of backup files
    - `LOG_MAX_AGE_DAY` - Max age of log files in days

### Changed

- **Refactored `InitLogger()`**
  - Reduced function length from ~130 lines to ~30 lines
  - Extracted helper methods:
    - `buildEncoder()` - Create zap encoder based on config
    - `createFileCore()` - Create file output core
    - `createExactLevelCore()` - Create level-specific core with exact level matching
    - `createStdoutCore()` - Create stdout output core
    - `buildCores()` - Build all output cores
    - `buildLogFileName()` - Generate log file names

- **Performance Improvement**
  - Eliminated IO duplication when using split-level logging
  - Each log message is now written exactly once, regardless of configuration

### Fixed

- Fixed potential IO overhead when logs were written to multiple files simultaneously

## [2.0.0] - Previous Release

### origin: https://github.com/hunterhug/golog/v2

### Features

- Based on Uber ZapLog for high performance
- Support console and file output
- File rotation with configurable size, backup count, and max age
- Log level control (Debug, Info, Warn, Error, Fatal, Panic)
- JSON or console format output
- Short or full caller path
- Context support with custom field injection
- Structured logging with fields
