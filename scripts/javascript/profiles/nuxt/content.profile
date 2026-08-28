PROFILE_DESCRIPTION='Content-driven site with the official Nuxt Content module'
SCAFFOLD=(npm create nuxt@latest -- __APP__ --packageManager npm --template content --no-gitInit --no-modules)
DEVARCH_SCRIPT='nuxt dev --host 0.0.0.0 --port 3000'
POST_INSTALL=0

configure_app() {
  local app_dir="$1"
  (cd "$app_dir" && npm install better-sqlite3)
}
