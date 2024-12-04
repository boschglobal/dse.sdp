
function dse-simer() {
    ( if test -d "$1"; then cd "$1" && shift; fi && docker run -it --rm -v $(pwd):/sim -p 2159:2159 -p 6379:6379 $DSE_SIMER_IMAGE "$@"; )
}
export -f dse-simer

alias dse-env='env | grep ^DSE'
alias simer=dse-simer
