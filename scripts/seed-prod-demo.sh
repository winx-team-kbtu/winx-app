#!/bin/sh
# Fills the Winx database with demo data for Team Mode, or removes it again.
#
# Usage on the server (from /opt/winx, next to compose.yaml):
#   sh seed-prod-demo.sh            # (re)create demo data
#   sh seed-prod-demo.sh clean      # remove demo data only
#
# Optional environment:
#   DEMO_PASSWORD   password of all demo accounts (default: DemoPass123)
#   COMPOSE_FILE    compose file (default: compose.yaml)
#   POSTGRES_SERVICE, DB_USER, DB_NAME   defaults: postgres-winx, winx, winx-db
#
# Only accounts with an e-mail ending in @demo.winx.local are created or deleted.
set -eu

here="$(cd "$(dirname "$0")" && pwd)"
mode="${1:-seed}"
password="${DEMO_PASSWORD:-DemoPass123}"
compose_file="${COMPOSE_FILE:-compose.yaml}"
service="${POSTGRES_SERVICE:-postgres-winx}"
db_user="${DB_USER:-winx}"
db_name="${DB_NAME:-winx-db}"

# the SQL files live in scripts/ in the repository, or next to this script when copied
sql_dir="$here"
[ -f "$sql_dir/demo-clean.sql" ] || sql_dir="$here/scripts"

run_psql() {
  docker compose -p winx -f "$compose_file" exec -T "$service" \
    psql -U "$db_user" -d "$db_name" -v ON_ERROR_STOP=1 -v "demo_password=$password" -f -
}

case "$mode" in
  seed)
    cat "$sql_dir/demo-clean.sql" "$sql_dir/seed-prod-demo.sql" | run_psql
    echo
    echo "Demo data is ready. Sign in with demoNN@demo.winx.local (NN = 01..40), password: $password"
    echo "  demo16@demo.winx.local  - no team, has chats and saved candidates in Algorithms"
    echo "  demo01@demo.winx.local  - captain of the Forming team (code DEMO01)"
    echo "  demo06@demo.winx.local  - member of the Full team (code DEMO03)"
    echo "Join codes: DEMO01 Forming(2), DEMO02 Ready(3), DEMO03 Full(5), DEMO04 Ready(4), DEMO05 Forming(1)"
    ;;
  clean)
    run_psql < "$sql_dir/demo-clean.sql"
    echo "Demo data removed."
    ;;
  *)
    echo "Unknown mode: $mode (use 'seed' or 'clean')" >&2
    exit 2
    ;;
esac
