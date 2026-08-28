PROFILE_DESCRIPTION='Lit web-component application with TypeScript'
SCAFFOLD=(npm create vite@latest -- __APP__ --template lit-ts --no-interactive)
DEVARCH_SCRIPT='vite --host 0.0.0.0 --port 3000'
POST_INSTALL=1
