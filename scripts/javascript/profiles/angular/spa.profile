PROFILE_DESCRIPTION='Client-rendered application with routing, SCSS, and strict mode'
SCAFFOLD=(npx --yes @angular/cli@latest new __APP__ --directory __APP__ --routing --style scss --strict --skip-git --package-manager npm --defaults)
DEVARCH_SCRIPT='ng serve --host 0.0.0.0 --port 3000 --allowed-hosts'
POST_INSTALL=0
