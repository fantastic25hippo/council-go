# Packages

- `internal/worker`: Defines the core worker capabilities, outcome types, and generic interfaces.
- `local`: Provides an adapter for local execution, such as running commands via shell.
- `jules`: Provides an adapter to interface with Jules's API and dispatch messages.
- `plan`: Manages code changes and override behaviors during execution.
- `route`: Validates payload formats and resolves structural requirements.
- `dispatch`: Handles task delivery to specific handlers based on action types.
- `state`: Represents meeting phases and execution state transitions.
- `evidence`: Collects and scopes data like diffs, test results, or git history.
- `verify`: Checks evidence items against constraints to ensure safety and completeness.
- `config`: Handles environment settings and startup prerequisites for the system.
- `gate`: Provides access control and concurrency boundaries for resources.
- `prompt`: Centralizes structural rules and text building.
- `proc`: Spawns and manages OS-level subprocesses securely.
- `relay`: Wires components together and coordinates proposals until agreement.
- `check`: Provides utilities for standardizing assertions and constraints.
- `cmd/council`: Contains the entry point for the static binary, managing signals and worker lifecycle.
