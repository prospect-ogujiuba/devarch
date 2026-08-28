PROFILE_DESCRIPTION='Vue SPA with TypeScript'
SCAFFOLD=(npm create vite@latest -- __APP__ --template vue-ts --no-interactive)
DEVARCH_SCRIPT='vite --host 0.0.0.0 --port 3000'
POST_INSTALL=1
