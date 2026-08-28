PROFILE_DESCRIPTION='Publishable Svelte component library'
SCAFFOLD=(npx --yes sv@latest create __APP__ --template library --types ts --no-add-ons --install npm)
DEVARCH_SCRIPT='vite dev --host 0.0.0.0 --port 3000'
POST_INSTALL=0
