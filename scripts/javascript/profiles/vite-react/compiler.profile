PROFILE_DESCRIPTION='React SPA with TypeScript and React Compiler enabled'
SCAFFOLD=(npm create vite@latest -- __APP__ --template react-compiler-ts --no-interactive)
DEVARCH_SCRIPT='vite --host 0.0.0.0 --port 3000'
POST_INSTALL=1
