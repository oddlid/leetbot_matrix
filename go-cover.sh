go_cover() {
  local coverfile
  coverfile=$(mktemp) || return 1
  go test -tags goolm -count=1 -timeout=10s -coverprofile="${coverfile}" "$@"
  go tool cover -func=${coverfile}
  rm -f ${coverfile}
}

go_cover "$@"
