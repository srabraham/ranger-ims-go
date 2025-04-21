#!/usr/bin/env bash

set -e

reporoot=$(git rev-parse --show-toplevel)

cd "${reporoot}"

extdir="web/static/ext"
mkdir -p "${extdir}"

curl -o "${extdir}/bootstrap.min.css" "https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css"
curl -o "${extdir}/bootstrap.bundle.min.js" "https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js"
curl -o "${extdir}/jquery-3.1.0.min.js" "https://code.jquery.com/jquery-3.1.0.min.js"
curl -o "${extdir}/dataTables.min.js" "https://cdn.datatables.net/2.2.2/js/dataTables.min.js"
curl -o "${extdir}/dataTables.bootstrap5.min.js" "https://cdn.datatables.net/2.2.2/js/dataTables.bootstrap5.min.js"
curl -o "${extdir}/dataTables.bootstrap5.min.css" "https://cdn.datatables.net/2.2.2/css/dataTables.bootstrap5.min.css"