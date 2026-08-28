PROFILE_DESCRIPTION='Minimal TypeScript content site'
SCAFFOLD=(npm create astro@latest -- __APP__ --template minimal --install --no-git --yes)
DEVARCH_SCRIPT='astro dev --host 0.0.0.0 --port 3000'
POST_INSTALL=0
