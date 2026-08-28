PROFILE_DESCRIPTION='Minimal universal Vue application'
SCAFFOLD=(npm create nuxt@latest -- __APP__ --packageManager npm --template v4 --no-gitInit --no-modules)
DEVARCH_SCRIPT='nuxt dev --host 0.0.0.0 --port 3000'
POST_INSTALL=0
