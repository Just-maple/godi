# Multi-Level Nested Container Hooks Example

This example demonstrates how the hooks mechanism works with multi-level nested containers.

## Container Hierarchy

```
app (top level)
├── services (middle layer)
│   └── infra (bottom layer)
├── Config
├── ServiceA (build)
└── ServiceB (build)
```

## Key Mechanisms

### 1. Hook Trigger Rules

- Each container maintains its own hooks independently
- When a type is injected, hooks trigger on **all** containers in the injection chain
- The `provided` counter is tracked per type ID independently

### 2. Hook vs HookOnce

- `Hook`: Triggers on every injection (can check `provided` parameter for count)
- `HookOnce`: Only triggers on first injection (returns nil when `provided > 0`)

### 3. Injection Chain Propagation

When injecting `Database` from the `app` container:
1. `app` looks for provider → delegates to `services`
2. `services` looks for provider → delegates to `infra`
3. `infra` provides `Database`
4. Hooks trigger on **each container** in the chain

## Running the Example

```bash
cd examples/08-nested-container-hooks
go run main.go
```

## Expected Behavior

- Each type triggers Init Hook only once per container level (when provided=0)
- HookOnce ensures cleanup logic is registered only once
- Re-injection does not increase HookOnce call count

## Verification

The example tracks hook calls for each container level and prints a summary:
- Infra container: Database, Cache hooks
- Services container: Database, Cache, Logger hooks
- App container: All types including Config, ServiceA, ServiceB

All counters should show exactly 1 call per type per container, demonstrating proper singleton behavior and hook propagation.

## Related: Runtime Container Add

For an example of dynamically adding containers during Build execution, see [examples/14-runtime-container-add](../14-runtime-container-add/).

This pattern allows:
- Conditional dependency selection based on runtime values
- Plugin architecture with dynamic registration
- Feature flag-based dependency injection
