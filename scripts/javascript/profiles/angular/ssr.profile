PROFILE_DESCRIPTION='Server-rendered application with hydration, routing, and SCSS'
SCAFFOLD=(npx --yes @angular/cli@latest new __APP__ --directory __APP__ --routing --style scss --strict --ssr --skip-git --package-manager npm --defaults)
DEVARCH_SCRIPT='ng serve --host 0.0.0.0 --port 3000 --allowed-hosts'
POST_INSTALL=0
