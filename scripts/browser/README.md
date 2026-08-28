# AI-assisted browser testing

DevArch configures Microsoft's official [Playwright MCP](https://github.com/microsoft/playwright-mcp) server in [`.pi/mcp.json`](../../.pi/mcp.json). It gives Pi an accessibility-tree-first browser that can navigate, click, type, inspect console/network activity, and capture screenshots.

The server is lazy-started and its browser is **headed by default**. You and the AI can therefore watch and interact with the same browser window while testing a local app.

## Activate it

Requirements:

- Node.js 18 or newer
- A graphical desktop session
- `pi-mcp-adapter` installed in Pi

From the DevArch repository, run `/reload` in Pi after first adding or changing the MCP configuration. Then check `/mcp` or ask the AI to search for Playwright browser tools.

The first browser launch may report that no browser executable is installed. Ask the AI to run Playwright's `browser_install` tool, then retry. The browser download is stored in Playwright's user cache, not this repository.

## Collaborative workflow

1. Start the target app, for example with `scripts/node/bootstrap.sh <app-name>` or the app's normal development command.
2. Tell the AI the URL and the behavior to verify.
3. Keep the headed browser visible. You can take over to enter credentials, solve a challenge, demonstrate a bug, or point out a visual issue.
4. Tell the AI what you changed or observed; it can continue from the current page state, inspect the page, and retest after code changes.

Example prompts:

```text
Open https://my-app.test and manually test the sign-up flow. Keep the browser visible, inspect console errors, and take a screenshot at each failure.
```

```text
Use Playwright to test the page I am building. Pause after navigation so I can demonstrate the bug in the browser, then inspect the resulting DOM and network activity.
```

```text
Retest the mobile navigation at 390x844, click every menu item, and report only reproducible failures with screenshots.
```

## Safety and scope

- The AI controls a real browser. Do not expose credentials you would not enter into that browser session.
- The default Playwright MCP profile persists browser state outside the repository. Use a test account for authenticated flows.
- Playwright MCP is for interactive inspection and manual verification. Add normal Playwright test files to an individual app when a regression must run repeatedly in CI.
- Local TLS warnings can be worked around per session when appropriate, but do not weaken production TLS settings.

## Verify the integration

Validate the tracked configuration without downloading anything:

```bash
bash scripts/browser/smoke-test.sh
```

Also resolve the configured package and check its CLI (requires network access the first time):

```bash
bash scripts/browser/smoke-test.sh --runtime
```
