PROFILE_DESCRIPTION='TypeScript application with linting, formatting, unit, component, and browser tests'
SCAFFOLD=(npx --yes sv@latest create __APP__ --template minimal --types ts --add prettier eslint 'vitest=usages:unit,component' playwright --install npm)
DEVARCH_SCRIPT='vite dev --host 0.0.0.0 --port 3000'
POST_INSTALL=0
