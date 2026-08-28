PROFILE_DESCRIPTION='Route-handler-only API project with TypeScript'
SCAFFOLD=(npx --yes create-next-app@latest __APP__ --ts --eslint --api --src-dir --import-alias '@/*' --use-npm --disable-git)
DEVARCH_SCRIPT='next dev --hostname 0.0.0.0 --port 3000'
POST_INSTALL=0

configure_app() {
  local app_dir="$1" app_name="$2"
  DEVARCH_NEXT_ORIGIN="$app_name.test" node - "$app_dir/next.config.ts" <<'NODE'
const fs = require('node:fs');
const path = process.argv[2];
const origin = process.env.DEVARCH_NEXT_ORIGIN;
let source = fs.existsSync(path) ? fs.readFileSync(path, 'utf8') : 'import type { NextConfig } from "next";\n\nconst nextConfig: NextConfig = {};\n\nexport default nextConfig;\n';
if (!/allowedDevOrigins\s*:/.test(source)) source = source.replace(/const nextConfig:\s*NextConfig\s*=\s*\{\s*\n?/, (match) => `${match}\n  allowedDevOrigins: [${JSON.stringify(origin)}],`);
fs.writeFileSync(path, source);
NODE
}
