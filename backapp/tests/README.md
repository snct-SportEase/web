# Backend Test Layout

Backend tests are grouped by the layer they exercise.

- `handler/`: HTTP handler tests and endpoint-level behavior.
- `repository/`: repository tests, mostly SQL and persistence behavior.
- `middleware/`: middleware tests that can exercise exported middleware from outside the package.

Tests that need unexported functions, fields, or package-internal channels should stay beside the implementation under `internal/...`.

Examples:

- `internal/handler/*_test.go` covers private auth and SQL dump helpers.
- `internal/websocket/*_test.go` covers package-internal hub/client channels.

Prefer adding new black-box tests under `backapp/tests/<layer>/`. Add tests under `backapp/internal/...` only when package-private access is required.
