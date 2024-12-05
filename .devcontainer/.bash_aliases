
function dse-simer() {
    ( if test -d "$1"; then cd "$1" && shift; fi && docker run -it --rm -v $(pwd):/sim -p 2159:2159 -p 6379:6379 $DSE_SIMER_IMAGE "$@"; )
}
export -f dse-simer

function dse-simer-host() {
    ( if test -d "$1"; then cd "$1" && shift; fi && docker run -it --rm --network=host -v $(pwd):/sim $DSE_SIMER_IMAGE "$@"; )
}
export -f dse-simer-host

alias dse-env='env | grep ^DSE | sort'
alias simer=dse-simer
alias h='history'
alias hi='history | grep $1'
