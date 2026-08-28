PROFILE_DESCRIPTION='App Router web application with TypeScript and Tailwind CSS'
SCAFFOLD=(npx --yes create-next-app@latest __APP__ --ts --eslint --tailwind --app --src-dir --import-alias '@/*' --use-npm --disable-git)
DEVARCH_SCRIPT='next dev --hostname 0.0.0.0 --port 3000'
POST_INSTALL=0

configure_app() {
  local app_dir="$1" app_name="$2"
  DEVARCH_NEXT_ORIGIN="$app_name.test" node - "$app_dir/next.config.ts" <<'NODE'
const fs = require('node:fs');
const path = process.argv[2];
const origin = process.env.DEVARCH_NEXT_ORIGIN;
let source;
if (fs.existsSync(path)) {
  source = fs.readFileSync(path, 'utf8');
  if (/allowedDevOrigins\s*:/.test(source)) process.exit(0);
  const objectStart = /const nextConfig:\s*NextConfig\s*=\s*\{\s*\n/;
  if (!objectStart.test(source)) throw new Error(`unsupported Next.js config shape: ${path}`);
  source = source.replace(objectStart, (match) => `${match}  allowedDevOrigins: [${JSON.stringify(origin)}],\n`);
} else {
  source = `import type { NextConfig } from "next";\n\nconst nextConfig: NextConfig = {\n  allowedDevOrigins: [${JSON.stringify(origin)}],\n};\n\nexport default nextConfig;\n`;
}
fs.writeFileSync(path, source);
NODE
}
