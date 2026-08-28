PROFILE_DESCRIPTION='Minimal TypeScript application without add-ons'
SCAFFOLD=(npx --yes sv@latest create __APP__ --template minimal --types ts --no-add-ons --install npm)
DEVARCH_SCRIPT='vite dev --host 0.0.0.0 --port 3000'
POST_INSTALL=0
