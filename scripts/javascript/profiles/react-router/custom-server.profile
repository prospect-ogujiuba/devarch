PROFILE_DESCRIPTION='Full-stack application with a custom Node server'
SCAFFOLD=(npx --yes create-react-router@latest __APP__ --template remix-run/react-router-templates/node-custom-server --yes --package-manager npm --no-git-init --no-agent-skills)
DEVARCH_SCRIPT='cross-env NODE_ENV=development node --conditions development server.js'
POST_INSTALL=0

configure_app() {
  local app_dir="$1"
  [[ -f "$app_dir/server.js" ]] || return 0
  node - "$app_dir/server.js" <<'NODE'
const fs = require('node:fs');
const path = process.argv[2];
let source = fs.readFileSync(path, 'utf8');
const replacements = [
  [
    'import compression from "compression";\n',
    'import { createServer as createHttpServer } from "node:http";\nimport compression from "compression";\n',
  ],
  [
    'const app = express();\n',
    'const app = express();\nconst httpServer = createHttpServer(app);\n',
  ],
  [
    'server: { middlewareMode: true },',
    'server: {\n        middlewareMode: true,\n        hmr: { server: httpServer, clientPort: 443, protocol: "wss" },\n      },',
  ],
  ['app.listen(PORT, () => {', 'httpServer.listen(PORT, () => {'],
];
for (const [before, after] of replacements) {
  if (!source.includes(before)) throw new Error(`unsupported custom-server template shape: missing ${before}`);
  source = source.replace(before, after);
}
fs.writeFileSync(path, source);
NODE
}
