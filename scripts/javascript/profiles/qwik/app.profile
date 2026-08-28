PROFILE_DESCRIPTION='Lean resumable TypeScript application'
SCAFFOLD=(npm create qwik@latest -- empty __APP__ --installDeps)
DEVARCH_SCRIPT='vite --mode ssr --host 0.0.0.0 --port 3000'
POST_INSTALL=0
